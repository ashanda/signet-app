// Package handlers implements the HTTP surface of the app, one file per
// original controller domain (see docs/analysis/api_spec.md, which this
// package follows section-by-section). auth_handler.go covers
// AuthController + Auth/PasswordResetController.
package handlers

import (
	"database/sql"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/httpx"
	"signet-backend/internal/models"
	"signet-backend/internal/tree"
)

var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func RegisterAuthRoutes(r chi.Router, d *app.Deps) {
	r.Post("/api/v1/login", loginHandler(d))
	r.Post("/api/v1/logout", d.Auth.RequireAuth(http.HandlerFunc(logoutHandler(d))).ServeHTTP)
	r.Get("/api/v1/me", d.Auth.OptionalAuth(http.HandlerFunc(meHandler(d))).ServeHTTP)

	r.Get("/api/v1/register/referral", registerReferralHandler(d))
	r.Post("/api/v1/register/step1", registerStep1Handler(d))
	r.Get("/api/v1/register/step2/{id}", registerStep2FormHandler(d))
	r.Post("/api/v1/register/step2", registerStep2SubmitHandler(d))
	r.Get("/api/v1/register/status", registerStatusHandler(d))

	r.Post("/api/v1/password/email", passwordEmailHandler(d))
	r.Get("/api/v1/password/reset/{token}", passwordResetTokenHandler(d))
	r.Post("/api/v1/password/reset", passwordResetHandler(d))
}

// --- login/logout/me ---

func loginHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := decodeJSON(r, &body); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		errs := map[string][]string{}
		if strings.TrimSpace(body.Email) == "" || !emailRe.MatchString(body.Email) {
			errs["email"] = []string{"The email field must be a valid email address."}
		}
		if body.Password == "" {
			errs["password"] = []string{"The password field is required."}
		}
		if len(errs) > 0 {
			httpx.ValidationError(w, errs)
			return
		}

		var user models.User
		err := d.DB.Get(&user, "SELECT * FROM users WHERE email = ?", body.Email)
		if err != nil || !auth.CheckPassword(user.Password, body.Password) {
			httpx.JSON(w, http.StatusOK, map[string]interface{}{"status": "error", "message": "Invalid credentials"})
			return
		}

		if err := d.Auth.IssueSession(w, &user); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not start session")
			return
		}

		var packageCount int
		_ = d.DB.Get(&packageCount, "SELECT COUNT(*) FROM user_packages WHERE user_id = ?", user.ID)

		if packageCount > 1 {
			httpx.OK(w, map[string]interface{}{"status": "success", "role": user.Role, "dashboard": user.Role + ".dashboard"})
			return
		}
		if packageCount == 1 {
			var pendingID models.NullInt64
			_ = d.DB.Get(&pendingID, "SELECT id FROM user_packages WHERE user_id = ? AND status = 'pending' LIMIT 1", user.ID)
			if pendingID.Valid {
				httpx.OK(w, map[string]interface{}{"status": "error", "message": "You're package is pending"})
				return
			}
			httpx.OK(w, map[string]interface{}{"status": "success", "role": user.Role, "dashboard": user.Role + ".dashboard"})
			return
		}
		// packageCount == 0 — user IS authenticated (session set above) but
		// has no package at all; matches the original exactly (not a logout).
		httpx.OK(w, map[string]interface{}{"status": "error", "message": "You do not have any package"})
	}
}

func logoutHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d.Auth.ClearSession(w)
		httpx.OK(w, map[string]interface{}{"status": "success"})
	}
}

func meHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user == nil {
			httpx.JSON(w, http.StatusOK, map[string]interface{}{"authenticated": false})
			return
		}
		httpx.OK(w, map[string]interface{}{
			"authenticated": true,
			"user": map[string]interface{}{
				"id":          user.ID,
				"signet_id":   models.SignetID(user.ID),
				"name":        user.Name,
				"email":       user.Email,
				"role":        user.Role,
				"status":      user.Status,
				"on_vacation": user.OnVacation,
				"roc_status":  user.RocStatus.String,
			},
		})
	}
}

// --- registration ---

func registerReferralHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("ref")
		var referral models.ReferralCode
		if err := d.DB.Get(&referral, "SELECT * FROM referral_codes WHERE code = ?", code); err != nil {
			httpx.OK(w, map[string]interface{}{"status": "error", "message": "Invalid referral code"})
			return
		}

		var lastRefUser models.User
		err := d.DB.Get(&lastRefUser, "SELECT * FROM users WHERE referred_by = ? ORDER BY id DESC LIMIT 1", referral.UserID)
		if err == nil {
			var activeCount int
			_ = d.DB.Get(&activeCount, "SELECT COUNT(*) FROM user_packages WHERE user_id = ? AND status = 'active'", lastRefUser.ID)
			if activeCount == 0 {
				httpx.OK(w, map[string]interface{}{"status": "error", "message": "The last user in this referral does not have an active package. Please contact your referral person."})
				return
			}
		}

		var countries []models.Country
		_ = d.DB.Select(&countries, "SELECT * FROM countries WHERE deleted_at IS NULL ORDER BY id")
		var leaders []models.User
		_ = d.DB.Select(&leaders, "SELECT * FROM users WHERE leader_status = 'active'")

		httpx.OK(w, map[string]interface{}{
			"status":        "success",
			"referral_code": code,
			"countries":     countries,
			"leaders":       leaderOptions(leaders),
		})
	}
}

func leaderOptions(leaders []models.User) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(leaders))
	for _, l := range leaders {
		out = append(out, map[string]interface{}{"id": l.ID, "signet_id": models.SignetID(l.ID), "name": l.Name})
	}
	return out
}

func registerStep1Handler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name            string `json:"name"`
			Email           string `json:"email"`
			Password        string `json:"password"`
			PasswordConfirm string `json:"password_confirmation"`
			WhatsappNumber  string `json:"whatsapp_number"`
			CountryCode     string `json:"country_code"`
			BinancePayID    string `json:"binance_pay_id"`
			ReferralCode    string `json:"referral_code"`
			Country         string `json:"country"`
			LeaderCode      string `json:"leader_code"`
			ExecutiveCode   string `json:"executive_code"`
		}
		if err := decodeJSON(r, &body); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		errs := map[string][]string{}
		if !emailRe.MatchString(body.Email) {
			errs["email"] = []string{"The email field must be a valid email address."}
		}
		if len(body.Password) < 6 {
			errs["password"] = []string{"The password field must be at least 6 characters."}
		} else if body.Password != body.PasswordConfirm {
			errs["password"] = []string{"The password field confirmation does not match."}
		}
		if body.WhatsappNumber == "" {
			errs["whatsapp_number"] = []string{"The whatsapp number field is required."}
		}
		if body.CountryCode == "" {
			errs["country_code"] = []string{"The country code field is required."}
		}
		if body.BinancePayID == "" {
			errs["binance_pay_id"] = []string{"The binance pay id field is required."}
		}
		if body.Country == "" {
			errs["country"] = []string{"The country field is required."}
		}
		var referral models.ReferralCode
		if err := d.DB.Get(&referral, "SELECT * FROM referral_codes WHERE code = ?", body.ReferralCode); err != nil {
			errs["referral_code"] = []string{"The selected referral code is invalid."}
		}
		if body.LeaderCode != "" && body.LeaderCode == body.ExecutiveCode {
			errs["executive_code"] = []string{"The executive code and leader code must be different."}
		}
		if len(errs) > 0 {
			httpx.ValidationError(w, errs)
			return
		}

		whatsapp := body.CountryCode + body.WhatsappNumber

		var dupe int
		_ = d.DB.Get(&dupe, "SELECT COUNT(*) FROM users WHERE whatsapp_number = ?", whatsapp)
		if dupe > 0 {
			httpx.OK(w, map[string]interface{}{"status": "error", "message": "The WhatsApp number is already registered."})
			return
		}
		_ = d.DB.Get(&dupe, "SELECT COUNT(*) FROM users WHERE binance_pay_id = ?", body.BinancePayID)
		if dupe > 0 {
			httpx.OK(w, map[string]interface{}{"status": "error", "message": "The Binance Pay ID is already registered."})
			return
		}
		_ = d.DB.Get(&dupe, "SELECT COUNT(*) FROM users WHERE email = ?", body.Email)
		if dupe > 0 {
			httpx.OK(w, map[string]interface{}{"status": "error", "message": "Email already exists"})
			return
		}

		hashed, err := auth.HashPassword(body.Password)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not process password")
			return
		}

		countryID, _ := strconv.ParseUint(body.Country, 10, 64)

		tx, err := d.DB.Beginx()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Database error")
			return
		}
		defer tx.Rollback()

		res, err := tx.Exec(`INSERT INTO users
			(name, email, password, whatsapp_number, binance_pay_id, status, leader_code, executive_code, referred_by, country_id, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?, NOW(), NOW())`,
			body.Name, body.Email, hashed, whatsapp, body.BinancePayID,
			ltrimZero(body.LeaderCode), ltrimZero(body.ExecutiveCode), referral.UserID, countryID)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not create account")
			return
		}
		newID, _ := res.LastInsertId()

		if _, err := tx.Exec("INSERT INTO wallets (user_id, balance, created_at, updated_at) VALUES (?, 0, NOW(), NOW())", newID); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not create wallet")
			return
		}

		newReferralCode := randomUpperString(6)
		if _, err := tx.Exec("INSERT INTO referral_codes (user_id, code, created_at, updated_at) VALUES (?, ?, NOW(), NOW())", newID, newReferralCode); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not create referral code")
			return
		}

		plainKey := "USER-" + strconv.FormatInt(newID, 10) + "-" + randomString(40)
		hashedKey, err := auth.HashPassword(plainKey)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not create secret key")
			return
		}
		if _, err := tx.Exec("INSERT INTO user_secret_keys (user_id, secret_key, created_at, updated_at) VALUES (?, ?, NOW(), NOW())", newID, hashedKey); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not create secret key")
			return
		}

		if err := tx.Commit(); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Database error")
			return
		}

		httpx.OK(w, map[string]interface{}{"status": "success", "user_id": newID})
	}
}

func registerStep2FormHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var packages []models.Package
		_ = d.DB.Select(&packages, "SELECT * FROM packages WHERE status = 'active'")
		httpx.OK(w, map[string]interface{}{"status": "success", "id": id, "packages": packages})
	}
}

func registerStep2SubmitHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Package   string `json:"package"`
			NewUserID string `json:"newUserID"`
		}
		if err := decodeJSON(r, &body); err != nil || body.Package == "" || body.NewUserID == "" {
			httpx.ValidationError(w, map[string][]string{"package": {"The package field is required."}})
			return
		}
		newUserID, ok := parseUintParam(body.NewUserID)
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid user id")
			return
		}
		packageID, err := strconv.ParseUint(body.Package, 10, 64)
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid package id")
			return
		}

		var userData models.User
		if err := d.DB.Get(&userData, "SELECT * FROM users WHERE id = ?", newUserID); err != nil {
			httpx.Error(w, http.StatusNotFound, "User not found")
			return
		}

		var createdIDs []uint64
		referredBy := uint64(0)
		if userData.ReferredBy.Valid {
			referredBy = uint64(userData.ReferredBy.Int64)
		}
		parentID, err := tree.ParentFind(d.DB, referredBy, packageID, newUserID, 1, &createdIDs)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not resolve upline placement")
			return
		}

		var userparent uint64
		if parentID == 1 {
			userparent = referredBy
		} else {
			userparent = parentID
		}

		tx, err := d.DB.Beginx()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Database error")
			return
		}
		defer tx.Rollback()

		res, err := tx.Exec(`INSERT INTO user_parents (user_id, virtual_id, parent_id, node, created_at, updated_at)
			VALUES (?, ?, ?, 'deactive', NOW(), NOW())`, newUserID, referredBy, userparent)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not place user in tree")
			return
		}
		rootRowID, _ := res.LastInsertId()
		createdIDs = append(createdIDs, uint64(rootRowID))

		// Binary-spillover correction: if the referrer now has 2+ children
		// created at this exact created_at instant, keep the oldest as
		// "gratitude" (copying the newest's placement) and mark a fresh
		// "correct" row — see financial_engine.md / api_spec.md step 5.
		if referredBy != 0 {
			var siblings []struct {
				ID        uint64    `db:"id"`
				UserID    uint64    `db:"user_id"`
				VirtualID uint64    `db:"virtual_id"`
				ParentID  uint64    `db:"parent_id"`
				CreatedAt time.Time `db:"created_at"`
			}
			_ = tx.Select(&siblings, `SELECT id, user_id, virtual_id, parent_id, created_at FROM user_parents
				WHERE virtual_id = ? ORDER BY created_at ASC FOR UPDATE`, referredBy)
			if len(siblings) >= 2 {
				gratitudeRow := siblings[0]
				activeRow := siblings[len(siblings)-1]
				_, _ = tx.Exec(`UPDATE user_parents SET user_id=?, virtual_id=?, parent_id=?, updated_at=NOW() WHERE id=?`,
					activeRow.UserID, activeRow.VirtualID, activeRow.ParentID, gratitudeRow.ID)
				res, err := tx.Exec(`INSERT INTO user_parents (user_id, virtual_id, parent_id, node, created_at, updated_at)
					VALUES (?, ?, ?, 'correct', NOW(), NOW())`, activeRow.UserID, activeRow.UserID, activeRow.ParentID)
				if err == nil {
					correctID, _ := res.LastInsertId()
					createdIDs = append(createdIDs, uint64(correctID))
				}
				_, _ = tx.Exec("DELETE FROM user_parents WHERE id = ?", activeRow.ID)
			}
		}

		var findParent struct {
			ParentID uint64 `db:"parent_id"`
		}
		_ = tx.Get(&findParent, "SELECT parent_id FROM user_parents WHERE user_id = ? LIMIT 1", newUserID)

		var parentData models.User
		_ = tx.Get(&parentData, "SELECT * FROM users WHERE id = ?", parentID)

		var activeExists int
		_ = tx.Get(&activeExists, "SELECT COUNT(*) FROM user_packages WHERE user_id = ? AND status = 'active'", newUserID)
		if activeExists > 0 {
			httpx.OK(w, map[string]interface{}{"status": "error", "message": "You already have an active package", "redirect": "login"})
			return
		}

		var buyPackage models.Package
		if err := tx.Get(&buyPackage, "SELECT * FROM packages WHERE id = ?", packageID); err != nil {
			httpx.Error(w, http.StatusNotFound, "Package not found")
			return
		}

		var existingUserPackageID models.NullInt64
		_ = tx.Get(&existingUserPackageID, "SELECT id FROM user_packages WHERE user_id = ? LIMIT 1", newUserID)
		var newUserPackageID int64
		if existingUserPackageID.Valid {
			newUserPackageID = existingUserPackageID.Int64
			_, err = tx.Exec(`UPDATE user_packages SET package=?, status='pending', ref_id=?, sale='first', updated_at=NOW() WHERE id=?`,
				strconv.FormatUint(buyPackage.ID, 10), strconv.FormatUint(findParent.ParentID, 10), newUserPackageID)
		} else {
			var r2 sql.Result
			r2, err = tx.Exec(`INSERT INTO user_packages (user_id, package, status, ref_id, sale, created_at, updated_at)
				VALUES (?, ?, 'pending', ?, 'first', NOW(), NOW())`, newUserID, strconv.FormatUint(buyPackage.ID, 10), strconv.FormatUint(findParent.ParentID, 10))
			if err == nil {
				newUserPackageID, _ = r2.LastInsertId()
			}
		}
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not create package purchase")
			return
		}

		_, _ = tx.Exec(`UPDATE super_parent_logs SET user_package = ? WHERE gratitude_user = ? AND created_at >= ?`,
			newUserPackageID, newUserID, time.Now().Add(-1*time.Minute))

		createdIDsJSON := "["
		for i, id := range createdIDs {
			if i > 0 {
				createdIDsJSON += ","
			}
			createdIDsJSON += strconv.FormatUint(id, 10)
		}
		createdIDsJSON += "]"
		_, _ = tx.Exec(`INSERT INTO user_parent_map_logs (user_id, parent_id, created_row_ids, note, created_at, updated_at)
			VALUES (?, ?, ?, 'UserParent rows created in processStep2', NOW(), NOW())`, newUserID, findParent.ParentID, createdIDsJSON)

		if err := tx.Commit(); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Database error")
			return
		}

		// Global Share Wallet bootstrap for the new user (best-effort,
		// mirrors the original's inline duplicate of the same logic).
		_ = bootstrapNewUserGlobalShareWallet(d, newUserID)

		// NewUserPackageMail — best effort, does not block the response.
		go sendNewUserPackageMail(d, &userData, &parentData, &buyPackage)

		httpx.OK(w, map[string]interface{}{
			"status": "success",
			"parent": map[string]interface{}{
				"binance_pay_id":  parentData.BinancePayID.String,
				"whatsapp_number": parentData.WhatsappNumber.String,
				"on_vacation":     parentData.OnVacation,
			},
			"user_id": newUserID,
		})
	}
}

func bootstrapNewUserGlobalShareWallet(d *app.Deps, userID uint64) error {
	var exists int
	if err := d.DB.Get(&exists, "SELECT COUNT(*) FROM global_share_wallets WHERE user_id = ?", userID); err != nil || exists > 0 {
		return err
	}
	var highest models.NullInt64
	_ = d.DB.Get(&highest, `
		SELECT MAX(p.price) FROM user_packages up JOIN packages p ON p.id = CAST(up.package AS UNSIGNED)
		WHERE up.user_id = ? AND p.price IN (5000,10000,25000,50000,100000,500000,1000000)`, userID)
	if !highest.Valid {
		return nil
	}
	maxOut := float64(highest.Int64) * 1.5
	_, err := d.DB.Exec("INSERT INTO global_share_wallets (user_id, balance, max_out, created_at, updated_at) VALUES (?, 0, ?, NOW(), NOW())", userID, maxOut)
	return err
}

func registerStatusHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var user models.User
		if err := d.DB.Get(&user, "SELECT * FROM users WHERE id = ?", id); err != nil {
			httpx.Error(w, http.StatusNotFound, "User not found")
			return
		}
		httpx.OK(w, map[string]interface{}{"status": "success", "user": map[string]interface{}{"id": user.ID, "status": user.Status}})
	}
}

// --- password reset ---

func passwordEmailHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email string `json:"email"`
		}
		if err := decodeJSON(r, &body); err != nil || !emailRe.MatchString(body.Email) {
			httpx.ValidationError(w, map[string][]string{"email": {"The email field must be a valid email address."}})
			return
		}
		var exists int
		_ = d.DB.Get(&exists, "SELECT COUNT(*) FROM users WHERE email = ?", body.Email)
		if exists == 0 {
			httpx.ValidationError(w, map[string][]string{"email": {"The selected email is invalid."}})
			return
		}

		token := randomString(60)
		_, err := d.DB.Exec(`INSERT INTO password_reset (email, token, created_at) VALUES (?, ?, NOW())
			ON DUPLICATE KEY UPDATE token = VALUES(token), created_at = NOW()`, body.Email, token)
		if err != nil {
			// password_reset has no PK/unique key (preserved verbatim per
			// schema.md), so ON DUPLICATE KEY UPDATE is a no-op guard, not
			// a real upsert — fall back to plain insert on any error shape
			// other than a real DB failure.
			_, err = d.DB.Exec("INSERT INTO password_reset (email, token, created_at) VALUES (?, ?, NOW())", body.Email, token)
			if err != nil {
				httpx.Error(w, http.StatusInternalServerError, "Could not create reset token")
				return
			}
		}

		go sendPasswordResetMail(d, body.Email, token)

		httpx.OK(w, map[string]interface{}{"status": "success", "message": "We have emailed your password reset link!"})
	}
}

func passwordResetTokenHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		var email string
		if err := d.DB.Get(&email, "SELECT email FROM password_reset WHERE token = ? LIMIT 1", token); err != nil {
			httpx.OK(w, map[string]interface{}{"status": "error", "message": "This password reset token is invalid."})
			return
		}
		httpx.OK(w, map[string]interface{}{"status": "success", "token": token, "email": email})
	}
}

func passwordResetHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email           string `json:"email"`
			Password        string `json:"password"`
			PasswordConfirm string `json:"password_confirmation"`
			Token           string `json:"token"`
		}
		if err := decodeJSON(r, &body); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		errs := map[string][]string{}
		if !emailRe.MatchString(body.Email) {
			errs["email"] = []string{"The email field must be a valid email address."}
		}
		if len(body.Password) < 8 {
			errs["password"] = []string{"The password field must be at least 8 characters."}
		} else if body.Password != body.PasswordConfirm {
			errs["password"] = []string{"The password field confirmation does not match."}
		}
		if body.Token == "" {
			errs["token"] = []string{"The token field is required."}
		}
		if len(errs) > 0 {
			httpx.ValidationError(w, errs)
			return
		}

		var resetEmail string
		if err := d.DB.Get(&resetEmail, "SELECT email FROM password_reset WHERE token = ? LIMIT 1", body.Token); err != nil {
			httpx.ValidationError(w, map[string][]string{"token": {"This password reset token is invalid."}})
			return
		}
		var user models.User
		if err := d.DB.Get(&user, "SELECT * FROM users WHERE email = ?", resetEmail); err != nil {
			httpx.ValidationError(w, map[string][]string{"email": {"No user found with that email address."}})
			return
		}
		hashed, err := auth.HashPassword(body.Password)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not process password")
			return
		}
		if _, err := d.DB.Exec("UPDATE users SET password = ?, updated_at = NOW() WHERE id = ?", hashed, user.ID); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not update password")
			return
		}
		_, _ = d.DB.Exec("DELETE FROM password_reset WHERE email = ?", resetEmail)

		httpx.OK(w, map[string]interface{}{"status": "success", "message": "Your password has been reset!"})
	}
}
