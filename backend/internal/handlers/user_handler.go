// user_handler.go covers UserController (api_spec.md "## UserController"),
// EXCLUDING GET /user/dashboard (UserController@index) — that route is
// owned by the dashboard-aggregation handler file elsewhere in this
// codebase, not here.
package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/httpx"
	"signet-backend/internal/models"
)

func RegisterUserRoutes(r chi.Router, d *app.Deps) {
	// POST /toggle-vacation — deliberately required auth (unguarded in the
	// original, see ARCHITECTURE.md); no role restriction, matching the
	// original's plain `auth()->user()` check.
	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAuth)
		r.Post("/api/v1/toggle-vacation", toggleVacationHandler(d))
	})

	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAuth)
		r.Use(auth.RequireRole("company", "admin"))
		r.Get("/api/v1/users", allUsersHandler(d))
		r.Get("/api/v1/users/search/{id}", userSearchHandler(d))
		r.Post("/api/v1/users/update-leader-code/{id}", updateLeaderCodeHandler(d))
		r.Post("/api/v1/users/update-executive-code/{id}", updateExecutiveCodeHandler(d))
		r.Get("/api/v1/leader-code-logs", codeLogsHandler(d, "leader_code_logs", "old_leader_code", "new_leader_code"))
		r.Get("/api/v1/executive-code-logs", codeLogsHandler(d, "executive_code_logs", "old_executive_code", "new_executive_code"))
	})

	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAuth)
		r.Use(auth.RequireRole("company"))
		r.Post("/api/v1/users/update/{id}", userSimpleStatusUpdateHandler(d, "status"))
		r.Post("/api/v1/users/update-roc/{id}", userSimpleStatusUpdateHandler(d, "roc_status"))
		r.Post("/api/v1/users/update-leader-status/{id}", userSimpleStatusUpdateHandler(d, "leader_status"))
		r.Post("/api/v1/users/update-global-director-share/{id}", updateGlobalDirectorShareHandler(d))
	})
}

// POST /toggle-vacation (user.toggleVacation) — UserController@toggleVacation
func toggleVacationHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		newVal := !user.OnVacation
		if _, err := d.DB.Exec("UPDATE users SET on_vacation = ?, updated_at = NOW() WHERE id = ?", newVal, user.ID); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not update vacation status")
			return
		}
		httpx.OK(w, map[string]interface{}{"status": true, "on_vacation": newVal})
	}
}

// GET /users (company.users) — UserController@allUsers
type userListRow struct {
	models.User
	CountryCode   models.NullString `db:"country_code"`
	CountryName   models.NullString `db:"country_name"`
	LeaderName    models.NullString `db:"leader_name"`
	ExecutiveName models.NullString `db:"executive_name"`
}

func allUsersHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := strings.TrimSpace(r.URL.Query().Get("search"))
		page, perPage, offset := httpx.PageParams(r, 10)

		// Final whereIn(status,[...]) applies unconditionally; the search
		// OR-clause, when present, is an additional AND'd filter — matches
		// the original's query-builder chain exactly.
		where := "WHERE u.status IN ('active','inactive','pending')"
		args := []interface{}{}
		if search != "" {
			like := "%" + search + "%"
			where += " AND (u.name LIKE ? OR u.whatsapp_number LIKE ? OR u.status LIKE ? OR c.code LIKE ? OR CONCAT('SIG-00', u.id) LIKE ?)"
			args = append(args, like, like, like, like, like)
		}

		baseFrom := `FROM users u
			LEFT JOIN countries c ON c.id = u.country_id
			LEFT JOIN users lu ON lu.id = CAST(NULLIF(u.leader_code,'') AS UNSIGNED)
			LEFT JOIN users eu ON eu.id = CAST(NULLIF(u.executive_code,'') AS UNSIGNED)`

		var total int
		_ = d.DB.Get(&total, "SELECT COUNT(*) "+baseFrom+" "+where, args...)

		listQuery := "SELECT u.*, c.code AS country_code, c.name AS country_name, lu.name AS leader_name, eu.name AS executive_name " +
			baseFrom + " " + where + " ORDER BY u.id DESC LIMIT ? OFFSET ?"
		listArgs := append(append([]interface{}{}, args...), perPage, offset)
		var rows []userListRow
		_ = d.DB.Select(&rows, listQuery, listArgs...)

		items := make([]map[string]interface{}, 0, len(rows))
		for _, row := range rows {
			items = append(items, map[string]interface{}{
				"id":              row.ID,
				"signet_id":       models.SignetID(row.ID),
				"name":            row.Name,
				"email":           row.Email,
				"whatsapp_number": row.WhatsappNumber.String,
				"status":          row.Status,
				"role":            row.Role,
				"leader_status":   row.LeaderStatus,
				"leader_code":     row.LeaderCode.String,
				"leader_name":     row.LeaderName.String,
				"executive_code":  row.ExecutiveCode.String,
				"executive_name":  row.ExecutiveName.String,
				"country": map[string]interface{}{
					"code": row.CountryCode.String,
					"name": row.CountryName.String,
				},
				"created_at": row.CreatedAt,
			})
		}

		var leaders []models.User
		_ = d.DB.Select(&leaders, "SELECT * FROM users WHERE leader_status = 'active' ORDER BY name")

		httpx.OK(w, map[string]interface{}{
			"status":  "success",
			"users":   httpx.Paginate(items, total, page, perPage),
			"leaders": leaderOptions(leaders),
		})
	}
}

// GET /users/search/{id} — UserController@search
func userSearchHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid user id")
			return
		}

		var user models.User
		if err := d.DB.Get(&user, "SELECT * FROM users WHERE id = ? AND status IN ('active','inactive')", id); err != nil {
			httpx.OK(w, map[string]interface{}{"success": false, "message": "User not found"})
			return
		}

		var userparent int
		_ = d.DB.Get(&userparent, "SELECT COUNT(*) FROM user_parents WHERE virtual_id = ?", id)

		var directSale int
		_ = d.DB.Get(&directSale, `
			SELECT COUNT(*) FROM user_parents
			WHERE parent_id = ? AND node IN ('active','gratitude') AND user_id != ?`, id, id)

		var totalSales models.NullFloat64
		_ = d.DB.Get(&totalSales, `
			SELECT COALESCE(SUM(p.price),0) FROM user_packages up
			JOIN packages p ON p.id = CAST(up.package AS UNSIGNED)
			WHERE up.user_id = ? AND up.status IN ('active','deactive')`, id)

		var walletBalance models.NullFloat64
		werr := d.DB.Get(&walletBalance, "SELECT balance FROM wallets WHERE user_id = ?", id)

		// Preserve the original's PHP operator-precedence quirk exactly:
		// `$totalSales*4 - $user->wallet?->balance ?? 0` parses as
		// `($totalSales*4 - $user->wallet?->balance) ?? 0`, NOT
		// `$totalSales*4 - ($user->wallet?->balance ?? 0)`. When there is
		// no wallet row, `?->` yields null, the whole subtraction becomes
		// null, and `?? 0` replaces the ENTIRE expression — discarding
		// totalSales*4 entirely, not just the missing balance. So: no
		// wallet row -> total_wallet is exactly 0.
		var totalWallet int64
		if werr == nil {
			totalWallet = int64(totalSales.Float64*4 - walletBalance.Float64)
		} else {
			totalWallet = 0
		}

		var firstPkgName, lastPkgName models.NullString
		_ = d.DB.Get(&firstPkgName, `
			SELECT p.name FROM user_packages up JOIN packages p ON p.id = CAST(up.package AS UNSIGNED)
			WHERE up.user_id = ? ORDER BY up.created_at ASC LIMIT 1`, id)
		_ = d.DB.Get(&lastPkgName, `
			SELECT p.name FROM user_packages up JOIN packages p ON p.id = CAST(up.package AS UNSIGNED)
			WHERE up.user_id = ? ORDER BY up.created_at DESC LIMIT 1`, id)

		var mining models.UserMining
		miningOut := map[string]interface{}{"total_token": 0, "mining_token": 0.0, "daily_mining": 0}
		if err := d.DB.Get(&mining, "SELECT * FROM user_minings WHERE user_id = ?", id); err == nil {
			miningOut = map[string]interface{}{
				"total_token":  mining.TotalToken,
				"mining_token": mining.MiningToken,
				"daily_mining": mining.DailyMining,
			}
		}

		httpx.OK(w, map[string]interface{}{
			"success": true,
			"user": map[string]interface{}{
				"id":                           user.ID,
				"name":                         user.Name,
				"email":                        user.Email,
				"status":                       user.Status,
				"roc_status":                   user.RocStatus.String,
				"global_director_share":        user.GlobalDirectorShare,
				"global_director_share_status": user.GlobalDirectorShareStatus,
			},
			"packages": map[string]interface{}{"first": firstPkgName.String, "last": lastPkgName.String},
			"mining":   miningOut,
			"sales":    map[string]interface{}{"total_sales": totalSales.Float64, "direct_sales": directSale},
			"wallet":   map[string]interface{}{"total_wallet": totalWallet},
		})
	}
}

// POST /users/update/{id}, /users/update-roc/{id}, /users/update-leader-status/{id}
// — UserController@update / @updateRocStatus / @updateLeaderStatus, which
// are all the same "validate status in:active,inactive, set column, save"
// shape against a different column.
func userSimpleStatusUpdateHandler(d *app.Deps, column string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid user id")
			return
		}
		var body struct {
			Status string `json:"status"`
		}
		if err := decodeJSON(r, &body); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if body.Status != "active" && body.Status != "inactive" {
			httpx.ValidationError(w, map[string][]string{"status": {"The selected status is invalid."}})
			return
		}

		query := fmt.Sprintf("UPDATE users SET %s = ?, updated_at = NOW() WHERE id = ?", column)
		res, err := d.DB.Exec(query, body.Status, id)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Database error")
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			httpx.OK(w, map[string]interface{}{"success": false, "message": "User not found"})
			return
		}
		httpx.OK(w, map[string]interface{}{"success": true})
	}
}

// POST /users/update-leader-code/{id} — UserController@updateLeaderCode
func updateLeaderCodeHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			LeaderCode string `json:"leader_code"`
		}
		_ = decodeJSON(r, &body)
		updateUserCode(w, r, d, body.LeaderCode, "leader_code_logs", "old_leader_code", "new_leader_code", "leader_code", "leader")
	}
}

// POST /users/update-executive-code/{id} — UserController@updateExecutiveCode
func updateExecutiveCodeHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ExecutiveCode string `json:"executive_code"`
		}
		_ = decodeJSON(r, &body)
		updateUserCode(w, r, d, body.ExecutiveCode, "executive_code_logs", "old_executive_code", "new_executive_code", "executive_code", "executive")
	}
}

// updateUserCode implements the shared logic behind updateLeaderCode and
// updateExecutiveCode: reject self-assignment, no-op (and no log row) if
// unchanged, otherwise log the change and write the new code.
// logTable/oldCol/newCol/userCol are all fixed literals supplied by this
// file's two callers above — never request-derived — so building SQL with
// fmt.Sprintf here carries no injection risk.
func updateUserCode(w http.ResponseWriter, r *http.Request, d *app.Deps, rawCode, logTable, oldCol, newCol, userCol, respKey string) {
	actingUser := auth.UserFromContext(r.Context())
	id, ok := parseUintParam(chi.URLParam(r, "id"))
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "Invalid user id")
		return
	}
	newCode := signetIDToNumeric(strings.TrimSpace(rawCode))

	var user models.User
	if err := d.DB.Get(&user, "SELECT * FROM users WHERE id = ?", id); err != nil {
		httpx.OK(w, map[string]interface{}{"success": false, "message": "User not found"})
		return
	}

	if newCode != "" {
		if newIDVal, okID := parseUintParam(newCode); okID && newIDVal == id {
			msg := "A user cannot be their own leader"
			if respKey == "executive" {
				msg = "A user cannot be their own executive"
			}
			httpx.OK(w, map[string]interface{}{"success": false, "message": msg})
			return
		}
		var exists int
		_ = d.DB.Get(&exists, "SELECT COUNT(*) FROM users WHERE id = ?", newCode)
		if exists == 0 {
			httpx.ValidationError(w, map[string][]string{userCol: {"The selected " + userCol + " is invalid."}})
			return
		}
	}

	var oldCode string
	if userCol == "leader_code" {
		oldCode = user.LeaderCode.String
	} else {
		oldCode = user.ExecutiveCode.String
	}

	if newCode == oldCode {
		httpx.OK(w, map[string]interface{}{"success": true})
		return
	}

	tx, err := d.DB.Beginx()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer tx.Rollback()

	logQuery := fmt.Sprintf("INSERT INTO %s (user_id, %s, %s, changed_by, created_at, updated_at) VALUES (?,?,?,?,NOW(),NOW())", logTable, oldCol, newCol)
	if _, err := tx.Exec(logQuery, id, userNullStr(oldCode), userNullStr(newCode), actingUser.ID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Could not log code change")
		return
	}
	updateQuery := fmt.Sprintf("UPDATE users SET %s = ?, updated_at = NOW() WHERE id = ?", userCol)
	if _, err := tx.Exec(updateQuery, userNullStr(newCode), id); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Could not update user")
		return
	}
	if err := tx.Commit(); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	var newName models.NullString
	if newCode != "" {
		_ = d.DB.Get(&newName, "SELECT name FROM users WHERE id = ?", newCode)
	}
	var nameOut interface{}
	if newName.Valid {
		nameOut = newName.String
	}
	httpx.OK(w, map[string]interface{}{"success": true, respKey: nameOut})
}

// GET /leader-code-logs, /executive-code-logs — UserController@leaderCodeLogs / @executiveCodeLogs
func codeLogsHandler(d *app.Deps, table, oldCol, newCol string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := strings.TrimSpace(r.URL.Query().Get("search"))
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		page, perPage, offset := httpx.PageParams(r, 15)

		where := "WHERE 1=1"
		args := []interface{}{}
		if search != "" {
			like := "%" + search + "%"
			where += " AND (u.name LIKE ? OR CONCAT('SIG-00', u.id) LIKE ?)"
			args = append(args, like, like)
		}
		if from != "" {
			where += " AND l.created_at >= ?"
			args = append(args, from)
		}
		if to != "" {
			where += " AND l.created_at <= ?"
			args = append(args, to)
		}

		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s l JOIN users u ON u.id = l.user_id %s", table, where)
		var total int
		_ = d.DB.Get(&total, countQuery, args...)

		listQuery := fmt.Sprintf(`
			SELECT l.id, l.user_id, l.%s AS old_code, l.%s AS new_code, l.changed_by, l.created_at,
			       u.name AS user_name,
			       ou.name AS old_name,
			       nu.name AS new_name,
			       cb.name AS changed_by_name
			FROM %s l
			JOIN users u ON u.id = l.user_id
			LEFT JOIN users ou ON ou.id = CAST(NULLIF(l.%s,'') AS UNSIGNED)
			LEFT JOIN users nu ON nu.id = CAST(NULLIF(l.%s,'') AS UNSIGNED)
			LEFT JOIN users cb ON cb.id = l.changed_by
			%s
			ORDER BY l.created_at DESC LIMIT ? OFFSET ?`, oldCol, newCol, table, oldCol, newCol, where)

		type row struct {
			ID            uint64            `db:"id"`
			UserID        uint64            `db:"user_id"`
			OldCode       models.NullString `db:"old_code"`
			NewCode       models.NullString `db:"new_code"`
			ChangedBy     models.NullInt64  `db:"changed_by"`
			CreatedAt     models.NullTime   `db:"created_at"`
			UserName      string            `db:"user_name"`
			OldName       models.NullString `db:"old_name"`
			NewName       models.NullString `db:"new_name"`
			ChangedByName models.NullString `db:"changed_by_name"`
		}
		listArgs := append(append([]interface{}{}, args...), perPage, offset)
		var rows []row
		_ = d.DB.Select(&rows, listQuery, listArgs...)

		logs := make([]map[string]interface{}, 0, len(rows))
		for _, rr := range rows {
			logs = append(logs, map[string]interface{}{
				"id": rr.ID,
				"user": map[string]interface{}{
					"id":        rr.UserID,
					"signet_id": models.SignetID(rr.UserID),
					"name":      rr.UserName,
				},
				"old_code":   rr.OldCode.String,
				"old_name":   rr.OldName.String,
				"new_code":   rr.NewCode.String,
				"new_name":   rr.NewName.String,
				"changed_by": rr.ChangedByName.String,
				"created_at": rr.CreatedAt,
			})
		}

		httpx.OK(w, map[string]interface{}{"status": "success", "logs": httpx.Paginate(logs, total, page, perPage)})
	}
}

// POST /users/update-global-director-share/{id} — UserController@updateGlobalDirectorShare
func updateGlobalDirectorShareHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid user id")
			return
		}
		var body struct {
			Value  float64 `json:"value"`
			Status string  `json:"status"` // "1" or "0", matching the original's in:1,0 rule
		}
		if err := decodeJSON(r, &body); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		errs := map[string][]string{}
		if body.Value < 0 {
			errs["value"] = []string{"The value field must be at least 0."}
		}
		if body.Status != "1" && body.Status != "0" {
			errs["status"] = []string{"The selected status is invalid."}
		}
		if len(errs) > 0 {
			httpx.ValidationError(w, errs)
			return
		}

		res, err := d.DB.Exec(
			"UPDATE users SET global_director_share = ?, global_director_share_status = ?, updated_at = NOW() WHERE id = ?",
			body.Value, body.Status == "1", id)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Database error")
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			httpx.OK(w, map[string]interface{}{"success": false, "message": "User not found"})
			return
		}
		httpx.OK(w, map[string]interface{}{"success": true})
	}
}

func userNullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
