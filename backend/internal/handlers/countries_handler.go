// countries_handler.go covers CountriesController (api_spec.md
// "## CountriesController"). Only index/store/update/destroy are
// implemented on the original controller — create/edit/show are NOT
// defined there and are intentionally omitted here too.
package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/httpx"
	"signet-backend/internal/models"
)

func RegisterCountriesRoutes(r chi.Router, d *app.Deps) {
	r.Group(func(r chi.Router) {
		r.Use(d.Auth.RequireAuth)
		r.Use(auth.RequireRole("company"))

		r.Get("/api/v1/countries", countriesIndexHandler(d))
		r.Post("/api/v1/countries", countriesStoreHandler(d))
		r.Put("/api/v1/countries/{id}", countriesUpdateHandler(d))
		r.Patch("/api/v1/countries/{id}", countriesUpdateHandler(d))
		r.Delete("/api/v1/countries/{id}", countriesDestroyHandler(d))
	})
}

func countriesIndexHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, perPage, offset := httpx.PageParams(r, 10)

		var total int
		_ = d.DB.Get(&total, "SELECT COUNT(*) FROM countries WHERE deleted_at IS NULL")

		var countries []models.Country
		_ = d.DB.Select(&countries, `
			SELECT * FROM countries WHERE deleted_at IS NULL
			ORDER BY created_at DESC LIMIT ? OFFSET ?`, perPage, offset)

		httpx.OK(w, map[string]interface{}{
			"status":    "success",
			"countries": httpx.Paginate(countries, total, page, perPage),
		})
	}
}

func countriesStoreHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Code string `json:"code"`
			Name string `json:"name"`
		}
		if err := decodeJSON(r, &body); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		errs := countriesValidate(d, body.Code, body.Name, 0)
		if len(errs) > 0 {
			httpx.ValidationError(w, errs)
			return
		}

		if _, err := d.DB.Exec(
			"INSERT INTO countries (code, name, created_at, updated_at) VALUES (?, ?, NOW(), NOW())",
			body.Code, body.Name,
		); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not create country")
			return
		}

		httpx.OK(w, map[string]interface{}{"status": "success", "message": "Country Create Success"})
	}
}

func countriesUpdateHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid country id")
			return
		}

		var exists int
		_ = d.DB.Get(&exists, "SELECT COUNT(*) FROM countries WHERE id = ? AND deleted_at IS NULL", id)
		if exists == 0 {
			httpx.Error(w, http.StatusNotFound, "Country not found")
			return
		}

		var body struct {
			Code string `json:"code"`
			Name string `json:"name"`
		}
		if err := decodeJSON(r, &body); err != nil {
			httpx.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		errs := countriesValidate(d, body.Code, body.Name, id)
		if len(errs) > 0 {
			httpx.ValidationError(w, errs)
			return
		}

		if _, err := d.DB.Exec(
			"UPDATE countries SET code = ?, name = ?, updated_at = NOW() WHERE id = ?",
			body.Code, body.Name, id,
		); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not update country")
			return
		}

		httpx.OK(w, map[string]interface{}{"status": "success", "message": "Country updated successfully."})
	}
}

func countriesDestroyHandler(d *app.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUintParam(chi.URLParam(r, "id"))
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "Invalid country id")
			return
		}

		res, err := d.DB.Exec("UPDATE countries SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", id)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "Could not delete country")
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			httpx.Error(w, http.StatusNotFound, "Country not found")
			return
		}

		httpx.OK(w, map[string]interface{}{"status": "success", "message": "Country deleted successfully."})
	}
}

// countriesValidate mirrors CountriesController@store/@update's validation:
// code required|string|max:10|unique:countries,code[,{id}]; name
// required|string|max:255. Laravel's default `unique` rule checks the raw
// table regardless of soft-delete state, so the uniqueness check here does
// NOT filter on deleted_at either.
func countriesValidate(d *app.Deps, code, name string, excludeID uint64) map[string][]string {
	errs := map[string][]string{}
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)

	if code == "" {
		errs["code"] = []string{"The code field is required."}
	} else if len(code) > 10 {
		errs["code"] = []string{"The code field must not be greater than 10 characters."}
	}
	if name == "" {
		errs["name"] = []string{"The name field is required."}
	} else if len(name) > 255 {
		errs["name"] = []string{"The name field must not be greater than 255 characters."}
	}

	if _, ok := errs["code"]; !ok {
		var dupe int
		if excludeID == 0 {
			_ = d.DB.Get(&dupe, "SELECT COUNT(*) FROM countries WHERE code = ?", code)
		} else {
			_ = d.DB.Get(&dupe, "SELECT COUNT(*) FROM countries WHERE code = ? AND id != ?", code, excludeID)
		}
		if dupe > 0 {
			errs["code"] = []string{"The code has already been taken."}
		}
	}

	return errs
}
