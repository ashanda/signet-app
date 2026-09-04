// apiv1_handler.go covers Api\AuthController (api_spec.md "## Api\AuthController
// (Sanctum API — `/api/*`)") — the Sanctum-guard B2B integration API,
// distinct from the end-user session (see api_spec.md Cross-cutting note
// #1). Route paths here are namespaced under /api/v1/external/... for the
// two that would otherwise collide conceptually with the end-user session's
// /api/v1/users routes registered elsewhere; method/behavior matches
// api_spec.md exactly.
package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/httpx"
	"signet-backend/internal/models"
)

func RegisterAPIV1Routes(r chi.Router, d *app.Deps) {
	// Public — issues the token (api_spec.md "### POST /api/token").
	r.Post("/api/v1/token", apiTokenHandler(d))

	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAPIToken)
		r.Post("/api/v1/check-user", apiCheckUserHandler(d))
		r.Get("/api/v1/external/users", apiExternalUsersHandler(d))
		r.Post("/api/v1/external/user-details", apiExternalUserDetailsHandler(d))
	})
}

// --- POST /api/v1/token — Api\AuthController@token ---

func apiTokenHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := decodeJSON(r, &body); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		errs := map[string][]string{}
		if body.Username == "" {
			errs["username"] = []string{"The username field is required."}
		}
		if body.Password == "" {
			errs["password"] = []string{"The password field is required."}
		}
		if len(errs) > 0 {
			httpx.ValidationError(w, errs)
			return
		}

		var apiUser models.ApiUser
		err := d.DB.Get(&apiUser, "SELECT * FROM api_users WHERE username = ?", body.Username)
		if err != nil || !auth.CheckPassword(apiUser.Password, body.Password) {
			httpx.JSON(w, http.StatusUnauthorized, map[string]interface{}{"message": "Invalid credentials"})
			return
		}

		token, err := d.Auth.NewAPIToken(apiUser.ID, "api_token")
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not issue token")
			return
		}
		httpx.OK(w, map[string]interface{}{"token": token})
	}
}

// --- POST /api/v1/check-user — Api\AuthController@checkUser ---

func apiCheckUserHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email string `json:"email"`
		}
		if err := decodeJSON(r, &body); err != nil || body.Email == "" {
			httpx.ValidationError(w, map[string][]string{"email": {"The email field is required."}})
			return
		}

		// Looks up the User table (end customers), NOT api_users — see
		// api_spec.md's note on this being intentional.
		var user models.User
		if err := d.DB.Get(&user, "SELECT * FROM users WHERE email = ?", body.Email); err != nil {
			httpx.OK(w, map[string]interface{}{"status": "fail", "message": "User not found"})
			return
		}

		var packageCount int
		_ = d.DB.Get(&packageCount, "SELECT COUNT(*) FROM user_packages WHERE user_id = ?", user.ID)

		httpx.OK(w, map[string]interface{}{"status": "success", "email": user.Email, "has_package": packageCount > 0})
	}
}

// --- GET /api/v1/external/users — Api\AuthController@getAllUsers ---

func apiExternalUsersHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, perPage, offset := httpx.PageParams(r, 20)

		var total int
		if err := d.DB.Get(&total, "SELECT COUNT(*) FROM users"); err != nil {
			httpx.JSON(w, http.StatusInternalServerError, map[string]interface{}{"status": "error", "message": "Something went wrong.", "error": err.Error()})
			return
		}

		// user_id here is the OPAQUE bcrypt secret_key string from
		// user_secret_keys — read through as-is, never re-hashed; see
		// api_spec.md Cross-cutting note #1 / #5.
		rows, err := d.DB.Queryx(`
			SELECT u.id, u.name, u.email,
			       sk.secret_key AS secret_key,
			       (SELECT COUNT(*) FROM user_packages up WHERE up.user_id = u.id) AS package_count,
			       m.total_token AS total_token
			FROM users u
			LEFT JOIN user_secret_keys sk ON sk.user_id = u.id
			LEFT JOIN user_minings m ON m.user_id = u.id
			ORDER BY u.id DESC
			LIMIT ? OFFSET ?`, perPage, offset)
		if err != nil {
			httpx.JSON(w, http.StatusInternalServerError, map[string]interface{}{"status": "error", "message": "Something went wrong.", "error": err.Error()})
			return
		}
		defer rows.Close()

		var users []map[string]interface{}
		for rows.Next() {
			var id uint64
			var name, email string
			var secretKey models.NullString
			var packageCount int
			var totalToken models.NullInt64
			if err := rows.Scan(&id, &name, &email, &secretKey, &packageCount, &totalToken); err != nil {
				continue
			}
			var userIDOpaque interface{}
			if secretKey.Valid {
				userIDOpaque = secretKey.String
			}
			var bookedToken interface{}
			if totalToken.Valid {
				bookedToken = totalToken.Int64
			}
			users = append(users, map[string]interface{}{
				"id": id, "user_id": userIDOpaque, "name": name, "email": email,
				"has_package": packageCount > 0, "booked_token": bookedToken,
			})
		}

		httpx.OK(w, map[string]interface{}{"status": "success", "users": httpx.Paginate(users, total, page, perPage)})
	}
}

// --- POST /api/v1/external/user-details — Api\AuthController@getSpecificUser ---

func apiExternalUserDetailsHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			UserID string `json:"user_id"`
		}
		if err := decodeJSON(r, &body); err != nil || body.UserID == "" {
			httpx.ValidationError(w, map[string][]string{"user_id": {"The user id field is required."}})
			return
		}

		// user_id is NOT the numeric users.id — it's the exact opaque
		// secret_key string returned by getAllUsers/registration. Matched
		// as a stored-string lookup, never re-hashed or verified — see
		// api_spec.md's note on this being an opaque token, not a real
		// hash comparison.
		var row struct {
			ID    uint64 `db:"id"`
			Name  string `db:"name"`
			Email string `db:"email"`
		}
		err := d.DB.Get(&row, `
			SELECT u.id, u.name, u.email
			FROM users u
			JOIN user_secret_keys sk ON sk.user_id = u.id
			WHERE sk.secret_key = ?
			LIMIT 1`, body.UserID)
		if err == sql.ErrNoRows {
			httpx.OK(w, map[string]interface{}{"status": "fail", "message": "User not found"})
			return
		}
		if err != nil {
			httpx.JSON(w, http.StatusInternalServerError, map[string]interface{}{"status": "error", "message": "Something went wrong.", "error": err.Error()})
			return
		}

		var packageCount int
		_ = d.DB.Get(&packageCount, "SELECT COUNT(*) FROM user_packages WHERE user_id = ?", row.ID)

		var totalToken models.NullInt64
		_ = d.DB.Get(&totalToken, "SELECT total_token FROM user_minings WHERE user_id = ? LIMIT 1", row.ID)
		var bookedToken interface{}
		if totalToken.Valid {
			bookedToken = totalToken.Int64
		}

		httpx.OK(w, map[string]interface{}{"status": "success", "user": map[string]interface{}{
			"id": row.ID, "user_id": body.UserID, "name": row.Name, "email": row.Email,
			"has_package": packageCount > 0, "booked_token": bookedToken,
		}})
	}
}
