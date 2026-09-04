// kyc_handler.go covers KycController (api_spec.md "## KycController").
//
// Auth decision: the original left verify/unverify/create/edit/store/
// update/destroy completely unguarded (anyone, including guests, could
// verify a KYC record) and index/show relied on `auth()->check()`
// defensively with a guest-facing branch. Per BACKEND_CONVENTIONS.md this
// port requires a session on every kyc.* route, including index/show —
// the "simplest/safest, be consistent" option explicitly offered in the
// task brief. That means the original's guest branch of index/show never
// triggers here; the role branch (company vs non-company) is preserved.
package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/httpx"
	"signet-backend/internal/models"
)

const kycMaxUploadBytes = 20 << 20 // 20MB total multipart form

func RegisterKycRoutes(r chi.Router, d *app.Deps) {
	// Serves files saved by kycStoreFile below (mirrors the original's
	// `public` disk under kyc/nic_front, kyc/nic_back, kyc/passport).
	r.Get("/storage/kyc/*", http.StripPrefix("/storage/kyc/", http.FileServer(http.Dir("./storage/kyc"))).ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAuth) // deliberately required on every route below — see file header

		r.Get("/api/v1/kyc", kycIndexHandler(d))
		r.Get("/api/v1/kyc/show", kycIndexHandler(d)) // kyc.show mirrors kyc.index's logic exactly, per api_spec.md
		r.Get("/api/v1/kyc/create", kycCreateHandler(d))
		r.Get("/api/v1/kyc/{id}/edit", kycEditHandler(d))
		r.Post("/api/v1/kyc", kycStoreHandler(d))
		r.Put("/api/v1/kyc/{id}", kycUpdateHandler(d))
		r.Delete("/api/v1/kyc/{id}", kycDestroyHandler(d))
		r.Post("/api/v1/kyc/{id}/verify", kycVerifyHandler(d))
		r.Post("/api/v1/kyc/{id}/unverify", kycUnverifyHandler(d))

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireRole("company"))
			r.Get("/api/v1/kyc/verified", kycVerifiedHandler(d))
		})
	})
}

// GET /kyc (kyc.index) — KycController@index (also mounted at /kyc/show,
// see RegisterKycRoutes).
//
// If the logged-in user's role != 'company': fetch their own single Kyc
// row (0 or 1 of them). Else (role == 'company'): paginated list of
// unverified KYCs belonging to non-company users.
func kycIndexHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())

		if user.Role != "company" {
			var kycs []models.Kyc
			_ = d.DB.Select(&kycs, "SELECT * FROM kycs WHERE user_id = ? LIMIT 1", user.ID)

			alert, message := "info", "No KYC submitted yet."
			if len(kycs) == 1 {
				if kycs[0].IsVerified {
					alert, message = "success", "Your KYC is verified."
				} else {
					alert, message = "warning", "Your KYC is pending verification."
				}
			}
			httpx.OK(w, map[string]interface{}{"status": "success", "alert": alert, "message": message, "kycs": kycs})
			return
		}

		page, perPage, offset := httpx.PageParams(r, 10)
		var total int
		_ = d.DB.Get(&total, `
			SELECT COUNT(*) FROM kycs k JOIN users u ON u.id = k.user_id
			WHERE k.is_verified = 0 AND u.role != 'company'`)
		var kycs []models.Kyc
		_ = d.DB.Select(&kycs, `
			SELECT k.* FROM kycs k JOIN users u ON u.id = k.user_id
			WHERE k.is_verified = 0 AND u.role != 'company'
			ORDER BY k.id DESC LIMIT ? OFFSET ?`, perPage, offset)

		httpx.OK(w, map[string]interface{}{"status": "success", "kycs": httpx.Paginate(kycs, total, page, perPage)})
	}
}

// GET /kyc/verified (kyc.verified) — KycController@verified
func kycVerifiedHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, perPage, offset := httpx.PageParams(r, 10)
		var total int
		_ = d.DB.Get(&total, `
			SELECT COUNT(*) FROM kycs k JOIN users u ON u.id = k.user_id
			WHERE k.is_verified = 1 AND u.role != 'company'`)
		var kycs []models.Kyc
		_ = d.DB.Select(&kycs, `
			SELECT k.* FROM kycs k JOIN users u ON u.id = k.user_id
			WHERE k.is_verified = 1 AND u.role != 'company'
			ORDER BY k.id DESC LIMIT ? OFFSET ?`, perPage, offset)

		httpx.OK(w, map[string]interface{}{"status": "success", "kycs": httpx.Paginate(kycs, total, page, perPage)})
	}
}

// GET /kyc/create (kyc.create) — KycController@create — no data, just the form.
func kycCreateHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.OK(w, map[string]interface{}{"status": "success"})
	}
}

// GET /kyc/{id}/edit (kyc.edit) — Kyc::where('user_id', auth()->id())->findOrFail($id)
func kycEditHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid kyc id")
			return
		}
		var kyc models.Kyc
		if err := d.DB.Get(&kyc, "SELECT * FROM kycs WHERE id = ? AND user_id = ?", id, user.ID); err != nil {
			httpx.Error(w, http.StatusNotFound, "Kyc not found")
			return
		}
		httpx.OK(w, map[string]interface{}{"status": "success", "kyc": kyc})
	}
}

// POST /kyc (kyc.store) — KycController@store
func kycStoreHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if err := r.ParseMultipartForm(kycMaxUploadBytes); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid form data")
			return
		}

		fullName := strings.TrimSpace(r.FormValue("full_name"))
		email := strings.TrimSpace(r.FormValue("email"))
		contact1 := strings.TrimSpace(r.FormValue("contact_number1"))
		contact2 := strings.TrimSpace(r.FormValue("contact_number2"))
		address := r.FormValue("address")
		telegram := r.FormValue("telegram_username")
		docType := r.FormValue("document_type")
		docNumber := strings.TrimSpace(r.FormValue("document_number"))

		errs := map[string][]string{}
		if fullName == "" {
			errs["full_name"] = []string{"The full name field is required."}
		}
		if !emailRe.MatchString(email) {
			errs["email"] = []string{"The email field must be a valid email address."}
		}
		if contact1 == "" {
			errs["contact_number1"] = []string{"The contact number1 field is required."}
		}
		if docType != "nic" && docType != "passport" {
			errs["document_type"] = []string{"The selected document type is invalid."}
		}
		if docNumber == "" {
			errs["document_number"] = []string{"The document number field is required."}
		} else {
			var dupe int
			_ = d.DB.Get(&dupe, "SELECT COUNT(*) FROM kycs WHERE document_number = ?", docNumber)
			if dupe > 0 {
				errs["document_number"] = []string{"The document number has already been taken."}
			}
		}
		if docType == "nic" {
			if _, _, err := r.FormFile("nic_front"); err != nil {
				errs["nic_front"] = []string{"The nic front field is required."}
			}
			if _, _, err := r.FormFile("nic_back"); err != nil {
				errs["nic_back"] = []string{"The nic back field is required."}
			}
		}
		if docType == "passport" {
			if _, _, err := r.FormFile("passport_image"); err != nil {
				errs["passport_image"] = []string{"The passport image field is required."}
			}
		}
		if len(errs) > 0 {
			httpx.ValidationError(w, errs)
			return
		}

		nicFrontPath, err := kycStoreFile(r, "nic_front", "nic_front")
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not store nic_front")
			return
		}
		nicBackPath, err := kycStoreFile(r, "nic_back", "nic_back")
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not store nic_back")
			return
		}
		passportPath, err := kycStoreFile(r, "passport_image", "passport")
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not store passport_image")
			return
		}

		_, err = d.DB.Exec(`
			INSERT INTO kycs (user_id, full_name, email, contact_number1, contact_number2, address,
				telegram_username, document_type, document_number, nic_front, nic_back, passport_image,
				is_verified, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,0,NOW(),NOW())`,
			user.ID, fullName, email, contact1, kycNullStr(contact2), address, telegram, docType, docNumber,
			kycNullStr(nicFrontPath), kycNullStr(nicBackPath), kycNullStr(passportPath))
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not submit KYC")
			return
		}

		httpx.OK(w, map[string]interface{}{"status": "success", "message": "KYC submitted. Pending verification."})
	}
}

// PUT /kyc/{id} (kyc.update) — KycController@update. Old files on disk are
// NOT deleted when a new one is uploaded — preserved verbatim per
// api_spec.md's documented storage-leak note, not fixed here.
func kycUpdateHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid kyc id")
			return
		}
		var kyc models.Kyc
		if err := d.DB.Get(&kyc, "SELECT * FROM kycs WHERE id = ? AND user_id = ?", id, user.ID); err != nil {
			httpx.Error(w, http.StatusNotFound, "Kyc not found")
			return
		}

		if err := r.ParseMultipartForm(kycMaxUploadBytes); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid form data")
			return
		}

		fullName := strings.TrimSpace(r.FormValue("full_name"))
		email := strings.TrimSpace(r.FormValue("email"))
		contact1 := strings.TrimSpace(r.FormValue("contact_number1"))
		contact2 := strings.TrimSpace(r.FormValue("contact_number2"))
		address := r.FormValue("address")
		telegram := r.FormValue("telegram_username")
		docType := r.FormValue("document_type")
		docNumber := strings.TrimSpace(r.FormValue("document_number"))

		errs := map[string][]string{}
		if fullName == "" {
			errs["full_name"] = []string{"The full name field is required."}
		}
		if !emailRe.MatchString(email) {
			errs["email"] = []string{"The email field must be a valid email address."}
		}
		if contact1 == "" {
			errs["contact_number1"] = []string{"The contact number1 field is required."}
		}
		if docType != "nic" && docType != "passport" {
			errs["document_type"] = []string{"The selected document type is invalid."}
		}
		if docNumber == "" {
			errs["document_number"] = []string{"The document number field is required."}
		} else {
			var dupe int
			_ = d.DB.Get(&dupe, "SELECT COUNT(*) FROM kycs WHERE document_number = ? AND id != ?", docNumber, id)
			if dupe > 0 {
				errs["document_number"] = []string{"The document number has already been taken."}
			}
		}
		if len(errs) > 0 {
			httpx.ValidationError(w, errs)
			return
		}

		nicFront := kyc.NicFront
		if p, err := kycStoreFile(r, "nic_front", "nic_front"); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not store nic_front")
			return
		} else if p != "" {
			nicFront = models.NullString{String: p, Valid: true}
		}
		nicBack := kyc.NicBack
		if p, err := kycStoreFile(r, "nic_back", "nic_back"); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not store nic_back")
			return
		} else if p != "" {
			nicBack = models.NullString{String: p, Valid: true}
		}
		passport := kyc.PassportImage
		if p, err := kycStoreFile(r, "passport_image", "passport"); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not store passport_image")
			return
		} else if p != "" {
			passport = models.NullString{String: p, Valid: true}
		}

		_, err := d.DB.Exec(`
			UPDATE kycs SET full_name=?, email=?, contact_number1=?, contact_number2=?, address=?,
				telegram_username=?, document_type=?, document_number=?, nic_front=?, nic_back=?,
				passport_image=?, updated_at=NOW()
			WHERE id = ?`,
			fullName, email, contact1, kycNullStr(contact2), address, telegram, docType, docNumber,
			nicFront, nicBack, passport, id)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not update KYC")
			return
		}

		httpx.OK(w, map[string]interface{}{"status": "success", "message": "KYC updated"})
	}
}

// DELETE /kyc/{id} (kyc.destroy) — Kyc::where('user_id',auth()->id())->findOrFail($id)->delete()
func kycDestroyHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid kyc id")
			return
		}
		res, err := d.DB.Exec("DELETE FROM kycs WHERE id = ? AND user_id = ?", id, user.ID)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not delete KYC")
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			httpx.Error(w, http.StatusNotFound, "Kyc not found")
			return
		}
		httpx.OK(w, map[string]interface{}{"status": "success", "message": "KYC deleted."})
	}
}

// POST /kyc/{id}/verify (kyc.verify) — no ownership/role check in the
// original beyond findOrFail; here it's gated to any authenticated user
// (the deliberate fix), not narrowed to company/admin, since the original
// had no role gate at all — only the missing-auth gap is closed.
func kycVerifyHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid kyc id")
			return
		}
		res, err := d.DB.Exec("UPDATE kycs SET is_verified = 1, updated_at = NOW() WHERE id = ?", id)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not verify KYC")
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			httpx.Error(w, http.StatusNotFound, "Kyc not found")
			return
		}
		httpx.OK(w, map[string]interface{}{"status": "success", "message": "KYC verified successfully."})
	}
}

// POST /kyc/{id}/unverify (kyc.unverify)
func kycUnverifyHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid kyc id")
			return
		}
		res, err := d.DB.Exec("UPDATE kycs SET is_verified = 0, updated_at = NOW() WHERE id = ?", id)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not unverify KYC")
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			httpx.Error(w, http.StatusNotFound, "Kyc not found")
			return
		}
		httpx.OK(w, map[string]interface{}{"status": "success", "message": "KYC unverified successfully."})
	}
}

// kycStoreFile reads an optional uploaded file from the given multipart
// field and saves it under ./storage/kyc/{subdir}/, returning the relative
// path (e.g. "kyc/nic_front/<hex>.jpg") to store in the DB. Returns ("",
// nil) if the field was not present in the request.
func kycStoreFile(r *http.Request, field, subdir string) (string, error) {
	file, header, err := r.FormFile(field)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	dir := filepath.Join("storage", "kyc", subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	ext := filepath.Ext(header.Filename)
	name := randomHex(16) + ext
	dst, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return "kyc/" + subdir + "/" + name, nil
}

func kycNullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
