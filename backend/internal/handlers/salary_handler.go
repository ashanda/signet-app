// salary_handler.go covers SalaryController (api_spec.md "## SalaryController").
package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/httpx"
	"signet-backend/internal/models"
	"signet-backend/internal/wallet"
)

func RegisterSalaryRoutes(r chi.Router, d *app.Deps) {
	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAuth)
		r.Use(auth.RequireRole("company"))
		r.Get("/api/v1/salaries", salariesIndexHandler(d))
		r.Get("/api/v1/salaries/search-users", salariesSearchUsersHandler(d))
		r.Post("/api/v1/salaries", salariesStoreHandler(d))
		r.Delete("/api/v1/salaries/{id}", salariesDestroyHandler(d))
	})
}

// salariesIndexHandler ports SalaryController@index.
func salariesIndexHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dateFrom := r.URL.Query().Get("date_from")
		dateTo := r.URL.Query().Get("date_to")

		where := "1=1"
		var args []interface{}
		if dateFrom != "" {
			where += " AND s.salary_date >= ?"
			args = append(args, dateFrom)
		}
		if dateTo != "" {
			where += " AND s.salary_date <= ?"
			args = append(args, dateTo)
		}

		var totalAmount models.NullFloat64
		_ = d.DB.Get(&totalAmount, "SELECT COALESCE(SUM(s.amount),0) FROM salaries s WHERE "+where, args...)

		var total int
		_ = d.DB.Get(&total, "SELECT COUNT(*) FROM salaries s WHERE "+where, args...)

		page, perPage, offset := httpx.PageParams(r, 15)
		listArgs := append(append([]interface{}{}, args...), perPage, offset)

		type salaryRow struct {
			models.Salary
			UserName     models.NullString `db:"user_name" json:"user_name"`
			UserWhatsapp models.NullString `db:"user_whatsapp" json:"user_whatsapp"`
		}
		var salaries []salaryRow
		listQuery := `
			SELECT s.*, u.name AS user_name, u.whatsapp_number AS user_whatsapp
			FROM salaries s LEFT JOIN users u ON u.id = s.user_id
			WHERE ` + where + ` ORDER BY s.salary_date DESC LIMIT ? OFFSET ?`
		_ = d.DB.Select(&salaries, listQuery, listArgs...)

		httpx.OK(w, map[string]interface{}{
			"salaries":     httpx.Paginate(salaries, total, page, perPage),
			"total_amount": totalAmount.Float64,
		})
	}
}

// salariesSearchUsersHandler ports SalaryController@searchUsers — Select2-
// style AJAX search matching name/whatsapp_number LIKE, plus a numeric
// SIG-00N id match when `term` ends in digits.
func salariesSearchUsersHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		term := strings.TrimSpace(r.URL.Query().Get("term"))

		type row struct {
			ID       uint64            `db:"id"`
			Name     string            `db:"name"`
			Whatsapp models.NullString `db:"whatsapp_number"`
		}
		var rows []row
		if term == "" {
			_ = d.DB.Select(&rows, "SELECT id, name, whatsapp_number FROM users LIMIT 15")
		} else {
			like := "%" + term + "%"
			digits := salaryTrailingDigits(term)
			if digits != "" {
				numericID := strings.TrimLeft(digits, "0")
				if numericID == "" {
					numericID = "0"
				}
				_ = d.DB.Select(&rows, `
					SELECT id, name, whatsapp_number FROM users
					WHERE name LIKE ? OR whatsapp_number LIKE ? OR id = ?
					LIMIT 15`, like, like, numericID)
			} else {
				_ = d.DB.Select(&rows, `
					SELECT id, name, whatsapp_number FROM users
					WHERE name LIKE ? OR whatsapp_number LIKE ?
					LIMIT 15`, like, like)
			}
		}

		results := make([]map[string]interface{}, 0, len(rows))
		for _, u := range rows {
			text := u.Name + " (" + models.SignetID(u.ID) + ")"
			if u.Whatsapp.Valid && u.Whatsapp.String != "" {
				text += " - " + u.Whatsapp.String
			}
			results = append(results, map[string]interface{}{"id": u.ID, "text": text})
		}
		httpx.OK(w, results)
	}
}

func salaryTrailingDigits(s string) string {
	i := len(s)
	for i > 0 && s[i-1] >= '0' && s[i-1] <= '9' {
		i--
	}
	return s[i:]
}

// salariesStoreHandler ports SalaryController@store: create the Salary row,
// then credit the wallet via wallet.UpdateWallet (WalletService::updateWallet
// — the CAP-GATED path, see financial_engine.md §1). The original wraps
// both in DB::transaction(), but WalletService::updateWallet opens its own
// internal transaction, so — as documented in BACKEND_CONVENTIONS.md — this
// is two sequential steps rather than one nested *sqlx.Tx: not truly atomic
// with each other, matching what the original actually does under the
// hood despite the outer DB::transaction() wrapper.
func salariesStoreHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			UserID     uint64  `json:"user_id"`
			Amount     float64 `json:"amount"`
			SalaryDate string  `json:"salary_date"`
			Remarks    string  `json:"remarks"`
		}
		if err := decodeJSON(r, &body); err != nil {
			httpx.JSON(w, http.StatusInternalServerError, map[string]interface{}{"success": false, "message": "Could not add salary: invalid request body"})
			return
		}

		errs := map[string][]string{}
		var userExists int
		if body.UserID != 0 {
			_ = d.DB.Get(&userExists, "SELECT COUNT(*) FROM users WHERE id = ?", body.UserID)
		}
		if body.UserID == 0 || userExists == 0 {
			errs["user_id"] = []string{"The selected user id is invalid."}
		}
		if body.Amount < 0 {
			errs["amount"] = []string{"The amount field must be at least 0."}
		}
		salaryDate, dateErr := time.Parse("2006-01-02", body.SalaryDate)
		if body.SalaryDate == "" || dateErr != nil {
			errs["salary_date"] = []string{"The salary date field must be a valid date."}
		}
		if len(body.Remarks) > 255 {
			errs["remarks"] = []string{"The remarks field must not be greater than 255 characters."}
		}
		if len(errs) > 0 {
			httpx.ValidationError(w, errs)
			return
		}

		if _, err := d.DB.Exec(`INSERT INTO salaries (user_id, amount, salary_date, remarks, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
			body.UserID, body.Amount, salaryDate, salaryNullableString(body.Remarks)); err != nil {
			httpx.JSON(w, http.StatusInternalServerError, map[string]interface{}{"success": false, "message": "Could not add salary: " + err.Error()})
			return
		}

		if err := wallet.UpdateWallet(d.DB, body.UserID, body.Amount, "Salary credited"); err != nil {
			httpx.JSON(w, http.StatusInternalServerError, map[string]interface{}{"success": false, "message": "Could not add salary: " + err.Error()})
			return
		}

		httpx.OK(w, map[string]interface{}{"success": true, "message": "Salary added successfully"})
	}
}

func salaryNullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// salariesDestroyHandler ports SalaryController@destroy.
func salariesDestroyHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.JSON(w, http.StatusNotFound, map[string]interface{}{"success": false, "message": "Record not found"})
			return
		}
		res, err := d.DB.Exec("DELETE FROM salaries WHERE id = ?", id)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Database error")
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			httpx.JSON(w, http.StatusNotFound, map[string]interface{}{"success": false, "message": "Record not found"})
			return
		}
		httpx.OK(w, map[string]interface{}{"success": true})
	}
}
