// earnlog_handler.go covers EarnLogController (api_spec.md "##
// EarnLogController"). The route had NO auth middleware in the original
// (relied on auth()->id() and would 500 for a guest) — this is one of the
// deliberate, disclosed fixes: we require a session here.
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/httpx"
	"signet-backend/internal/models"
)

func RegisterEarnLogRoutes(r chi.Router, d *app.Deps) {
	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAuth) // deliberately required — unguarded in the original, see ARCHITECTURE.md
		r.Get("/api/v1/earn/history", earnHistoryHandler(d))
	})
}

func earnHistoryHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		dateFrom := r.URL.Query().Get("date_from")
		dateTo := r.URL.Query().Get("date_to")

		// earn_logs.user_id is a VARCHAR column (see schema.md) — compare
		// against the string form of the id, not the numeric one.
		where := "WHERE user_id = ?"
		args := []interface{}{itoaU(user.ID)}
		if dateFrom != "" {
			where += " AND created_at >= ?"
			args = append(args, dateFrom)
		}
		if dateTo != "" {
			where += " AND created_at <= ?"
			args = append(args, dateTo)
		}

		var total int
		_ = d.DB.Get(&total, "SELECT COUNT(*) FROM earn_logs "+where, args...)

		// earn_logs.amount is also VARCHAR — MySQL will not type-juggle a
		// raw SUM() the way PHP's collection ->sum('amount') would, so the
		// values are explicitly cast before summing.
		var totalAmount models.NullFloat64
		_ = d.DB.Get(&totalAmount, "SELECT COALESCE(SUM(CAST(amount AS DECIMAL(15,2))),0) FROM earn_logs "+where, args...)

		page, perPage, offset := httpx.PageParams(r, 10)
		listArgs := append(append([]interface{}{}, args...), perPage, offset)
		var earns []models.EarnLog
		_ = d.DB.Select(&earns, "SELECT * FROM earn_logs "+where+" ORDER BY created_at DESC LIMIT ? OFFSET ?", listArgs...)

		httpx.OK(w, map[string]interface{}{
			"status":       "success",
			"earns":        httpx.Paginate(earns, total, page, perPage),
			"total_amount": totalAmount.Float64,
		})
	}
}
