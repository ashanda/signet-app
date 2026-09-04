// geneology_handler.go covers GeneologyController (api_spec.md "##
// GeneologyController"). Both routes had NO auth middleware in the
// original (relied on Auth()->user()/Auth::id() and would 500 for a
// guest) — this is one of the deliberate, disclosed fixes: we require a
// session on both.
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/httpx"
	"signet-backend/internal/models"
)

func RegisterGeneologyRoutes(r chi.Router, d *app.Deps) {
	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAuth) // deliberately required — unguarded in the original, see ARCHITECTURE.md
		r.Get("/api/v1/my-geneology", myGeneologyHandler(d))
		r.Get("/api/v1/geneology/{userId}", viewGeneologyHandler(d))
	})
}

// GET /my-geneology (my.geneology) — GeneologyController@index
//
// UserParent::where('parent_id', auth()->id())->with('user')
//
//	->whereIn('node', ['active','gratitude'])->where('user_id', '!=', auth()->id())->get()
func myGeneologyHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())

		type child struct {
			ID            uint64            `db:"id"`
			UserID        uint64            `db:"user_id"`
			VirtualID     uint64            `db:"virtual_id"`
			ParentID      uint64            `db:"parent_id"`
			Node          string            `db:"node"`
			ChildName     string            `db:"child_name"`
			ChildEmail    string            `db:"child_email"`
			ChildStatus   string            `db:"child_status"`
			ChildWhatsapp models.NullString `db:"child_whatsapp"`
		}

		var rows []child
		_ = d.DB.Select(&rows, `
			SELECT up.id, up.user_id, up.virtual_id, up.parent_id, up.node,
			       u.name AS child_name, u.email AS child_email, u.status AS child_status,
			       u.whatsapp_number AS child_whatsapp
			FROM user_parents up
			JOIN users u ON u.id = up.user_id
			WHERE up.parent_id = ? AND up.node IN ('active','gratitude') AND up.user_id != ?
			ORDER BY up.id`, user.ID, user.ID)

		childerns := make([]map[string]interface{}, 0, len(rows))
		for _, c := range rows {
			childerns = append(childerns, map[string]interface{}{
				"id":         c.ID,
				"user_id":    c.UserID,
				"virtual_id": c.VirtualID,
				"parent_id":  c.ParentID,
				"node":       c.Node,
				"user": map[string]interface{}{
					"id":              c.UserID,
					"signet_id":       models.SignetID(c.UserID),
					"name":            c.ChildName,
					"email":           c.ChildEmail,
					"status":          c.ChildStatus,
					"whatsapp_number": c.ChildWhatsapp.String,
				},
			})
		}

		httpx.OK(w, map[string]interface{}{"status": "success", "childerns": childerns})
	}
}

// GET /geneology/{userId} (geneology.show) — GeneologyController@viewGeneology
//
// $userdata = User::find($userId)
// $userPackage = UserPackage::where('user_id',$userId)->with('userpackage')->get()
// $parentData = User::where('id', $userdata->referred_by)->first()
func viewGeneologyHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := parseUintParam(chi.URLParam(r, "userId"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid user id")
			return
		}

		var userdata models.User
		if err := d.DB.Get(&userdata, "SELECT * FROM users WHERE id = ?", userID); err != nil {
			httpx.Error(w, http.StatusNotFound, "User not found")
			return
		}

		type pkgRow struct {
			models.UserPackage
			PackageName  models.NullString `db:"package_name"`
			PackagePrice models.NullInt64  `db:"package_price"`
		}
		var pkgRows []pkgRow
		_ = d.DB.Select(&pkgRows, `
			SELECT up.*, p.name AS package_name, p.price AS package_price
			FROM user_packages up
			LEFT JOIN packages p ON p.id = CAST(up.package AS UNSIGNED)
			WHERE up.user_id = ?
			ORDER BY up.id`, userID)

		userPackage := make([]map[string]interface{}, 0, len(pkgRows))
		for _, p := range pkgRows {
			userPackage = append(userPackage, map[string]interface{}{
				"id":            p.ID,
				"status":        p.Status,
				"sale":          p.Sale,
				"earn":          p.Earn,
				"activated_at":  p.ActivatedAt,
				"created_at":    p.CreatedAt,
				"package_name":  p.PackageName.String,
				"package_price": p.PackagePrice.Int64,
			})
		}

		var parentData *models.User
		if userdata.ReferredBy.Valid {
			var pd models.User
			if err := d.DB.Get(&pd, "SELECT * FROM users WHERE id = ?", userdata.ReferredBy.Int64); err == nil {
				parentData = &pd
			}
		}

		var parentOut interface{}
		if parentData != nil {
			parentOut = map[string]interface{}{
				"id":        parentData.ID,
				"signet_id": models.SignetID(parentData.ID),
				"name":      parentData.Name,
				"email":     parentData.Email,
			}
		}

		httpx.OK(w, map[string]interface{}{
			"status": "success",
			"userdata": map[string]interface{}{
				"id":              userdata.ID,
				"signet_id":       models.SignetID(userdata.ID),
				"name":            userdata.Name,
				"email":           userdata.Email,
				"status":          userdata.Status,
				"whatsapp_number": userdata.WhatsappNumber.String,
				"created_at":      userdata.CreatedAt,
			},
			"user_package": userPackage,
			"parent_data":  parentOut,
		})
	}
}
