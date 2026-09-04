// google_auth_handler.go covers GoogleAuthenticatorController.
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/httpx"
	"signet-backend/internal/totp"
)

func RegisterGoogleAuthRoutes(r chi.Router, d *app.Deps) {
	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAuth)
		r.Use(auth.RequireRole("company"))
		r.Get("/api/v1/admin/{userId}/setup-google-auth", setupGoogleAuthHandler(d))
	})
}

// setupGoogleAuthHandler ports GoogleAuthenticatorController::setupGoogleAuthenticator.
// Regenerates (overwrites) the target user's secret on every call — preserved
// verbatim, see api_spec.md's note on this invalidating any previously
// configured authenticator app. QR rendering itself is left to the Vue
// frontend (a client-side QR library renders the otpauth:// URL) rather than
// server-side SVG, since no bundled Go QR renderer is needed for that.
func setupGoogleAuthHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := parseUintParam(chi.URLParam(r, "userId"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid user id")
			return
		}
		var email string
		if err := d.DB.Get(&email, "SELECT email FROM users WHERE id = ?", userID); err != nil {
			httpx.OK(w, map[string]interface{}{"status": "error", "message": "User not found"})
			return
		}

		secret := totp.GenerateSecret()
		if _, err := d.DB.Exec("UPDATE users SET google_authenticator_secret = ?, updated_at = NOW() WHERE id = ?", secret, userID); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not save secret")
			return
		}

		otpauthURL := totp.OtpauthURL("signetint", email, secret)
		httpx.OK(w, map[string]interface{}{
			"status":      "success",
			"secret":      secret,
			"otpauth_url": otpauthURL,
		})
	}
}
