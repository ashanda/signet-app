// leaderexecutive_handler.go covers LeaderController and
// ExecutiveController (api_spec.md "## LeaderController" / "##
// ExecutiveController") — near-identical report queries keyed on
// leader_code vs executive_code respectively, all role:company. Shared
// date-range helpers (reportsDateFilter, reportsPreviousMonthBounds,
// reportsCurrentMonthBounds) live in directshare_handler.go.
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/httpx"
	"signet-backend/internal/models"
)

func RegisterLeaderExecutiveRoutes(r chi.Router, d *app.Deps) {
	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAuth)
		r.Use(auth.RequireRole("company"))
		r.Get("/api/v1/leaders/gain", leadersGainHandler(d))
		r.Get("/api/v1/executives/gain", executivesGainHandler(d))
		r.Get("/api/v1/leadership-bonus-log", leadershipBonusLogHandler(d))
	})
}

// leaderExecCodeGainQuery runs the shared "who gets credit for their
// downline's first-sale package volume, this period" report used by both
// leaders.gain and executives.gain (api_spec.md documents them as
// near-identical, keyed on `codeColumn` = "leader_code" or
// "executive_code"). Per api_spec.md: user_packages.status IN
// ('active','deactivate') — that literal spelling (the enum value really is
// 'deactivate', see schema.md, not 'deactive' as used elsewhere for tokens)
// — sale='first', created_at within [from 00:00:00, to 23:59:59].
func leaderExecCodeGainQuery(d *app.Deps, r *http.Request, codeColumn string) (interface{}, string, string) {
	defaultFrom, defaultTo := reportsCurrentMonthBounds()
	from := r.URL.Query().Get("from")
	if from == "" {
		from = defaultFrom.Format(reportsDateLayout)
	}
	to := r.URL.Query().Get("to")
	if to == "" {
		to = defaultTo.Format(reportsDateLayout)
	}
	filterIDParam := "leader_id"
	if codeColumn == "executive_code" {
		filterIDParam = "executive_id"
	}

	membershipFilter := "u.id IN (SELECT DISTINCT CAST(" + codeColumn + " AS UNSIGNED) FROM users WHERE " + codeColumn + " IS NOT NULL AND " + codeColumn + " <> '' AND CAST(" + codeColumn + " AS UNSIGNED) > 0)"
	filterSQL := membershipFilter
	var filterArgs []interface{}
	if idParam := r.URL.Query().Get(filterIDParam); idParam != "" {
		if id, ok := parseUintParam(idParam); ok {
			filterSQL += " AND u.id = ?"
			filterArgs = append(filterArgs, id)
		}
	}

	page, perPage, offset := httpx.PageParams(r, 10)

	var total int
	_ = d.DB.Get(&total, "SELECT COUNT(*) FROM users u WHERE "+filterSQL, filterArgs...)

	listArgs := append([]interface{}{from + " 00:00:00", to + " 23:59:59"}, filterArgs...)
	listArgs = append(listArgs, perPage, offset)

	rows, err := d.DB.Queryx(`
		SELECT u.id, u.name, u.email,
			(SELECT COALESCE(SUM(p.price),0)
			 FROM user_packages up
			 JOIN users m ON m.id = up.user_id
			 JOIN packages p ON p.id = CAST(up.package AS UNSIGNED)
			 WHERE CAST(m.`+codeColumn+` AS UNSIGNED) = u.id
			   AND up.status IN ('active','deactivate')
			   AND up.sale = 'first'
			   AND up.created_at BETWEEN ? AND ?
			) AS total_package,
			(SELECT COUNT(*) FROM users m2 WHERE CAST(m2.`+codeColumn+` AS UNSIGNED) = u.id) AS total_members
		FROM users u
		WHERE `+filterSQL+`
		ORDER BY u.name
		LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return nil, from, to
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var id uint64
		var name, email string
		var totalPackage float64
		var totalMembers int
		if err := rows.Scan(&id, &name, &email, &totalPackage, &totalMembers); err != nil {
			continue
		}
		out = append(out, map[string]interface{}{
			"id": id, "name": name, "email": email,
			"total_package": totalPackage, "total_members": totalMembers,
		})
	}
	return httpx.Paginate(out, total, page, perPage), from, to
}

func leadersGainHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paginated, from, to := leaderExecCodeGainQuery(d, r, "leader_code")
		if paginated == nil {
			httpx.Error(w, http.StatusInternalServerError, "Database error")
			return
		}
		httpx.OK(w, map[string]interface{}{"status": "success", "leaders": paginated, "from": from, "to": to})
	}
}

func executivesGainHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paginated, from, to := leaderExecCodeGainQuery(d, r, "executive_code")
		if paginated == nil {
			httpx.Error(w, http.StatusInternalServerError, "Database error")
			return
		}
		httpx.OK(w, map[string]interface{}{"status": "success", "executives": paginated, "from": from, "to": to})
	}
}

// --- GET /leadership-bonus-log ---

func leadershipBonusLogHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defaultStart, defaultEnd := reportsPreviousMonthBounds()
		startDate := r.URL.Query().Get("start_date")
		if startDate == "" {
			startDate = defaultStart.Format(reportsDateLayout)
		}
		endDate := r.URL.Query().Get("end_date")
		if endDate == "" {
			endDate = defaultEnd.Format(reportsDateLayout)
		}

		page, perPage, offset := httpx.PageParams(r, 20)

		var total int
		_ = d.DB.Get(&total, `SELECT COUNT(*) FROM earn_logs WHERE description = 'Leadership Bonus' AND DATE(created_at) >= ? AND DATE(created_at) <= ?`, startDate, endDate)

		rows, err := d.DB.Queryx(`
			SELECT el.id, el.user_id, el.amount, el.description, el.created_at, u.name AS user_name
			FROM earn_logs el
			LEFT JOIN users u ON u.id = CAST(el.user_id AS UNSIGNED)
			WHERE el.description = 'Leadership Bonus' AND DATE(el.created_at) >= ? AND DATE(el.created_at) <= ?
			ORDER BY el.created_at DESC LIMIT ? OFFSET ?`, startDate, endDate, perPage, offset)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Database error")
			return
		}
		defer rows.Close()

		var pools []map[string]interface{}
		for rows.Next() {
			var id uint64
			var userID, amount string
			var description models.NullString
			var createdAt models.NullTime
			var userName models.NullString
			if err := rows.Scan(&id, &userID, &amount, &description, &createdAt, &userName); err != nil {
				continue
			}
			pools = append(pools, map[string]interface{}{
				"id": id, "user_id": userID, "amount": amount, "description": description.String,
				"created_at": createdAt.Time, "user_name": userName.String,
			})
		}

		httpx.OK(w, map[string]interface{}{
			"status":     "success",
			"pools":      httpx.Paginate(pools, total, page, perPage),
			"start_date": startDate,
			"end_date":   endDate,
		})
	}
}
