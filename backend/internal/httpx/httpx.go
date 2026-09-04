// Package httpx holds small shared HTTP helpers: JSON responses and a
// paginator shape matching Laravel's default paginate() JSON envelope, so
// existing frontend pagination-consuming code (per ui_spec.md's recurring
// "Tables" pattern) needs no reshaping.
package httpx

import (
	"encoding/json"
	"math"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func OK(w http.ResponseWriter, payload interface{}) {
	JSON(w, http.StatusOK, payload)
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

func ValidationError(w http.ResponseWriter, errors map[string][]string) {
	JSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
		"message": "The given data was invalid.",
		"errors":  errors,
	})
}

// Paginated mirrors the subset of Laravel's LengthAwarePaginator JSON shape
// that the original Blade views actually rely on (per api_spec.md §8).
type Paginated struct {
	CurrentPage int         `json:"current_page"`
	Data        interface{} `json:"data"`
	From        int         `json:"from"`
	To          int         `json:"to"`
	LastPage    int         `json:"last_page"`
	PerPage     int         `json:"per_page"`
	Total       int         `json:"total"`
}

func Paginate(data interface{}, total, page, perPage int) Paginated {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}
	from := (page-1)*perPage + 1
	to := from + perPage - 1
	if total == 0 {
		from, to = 0, 0
	} else if to > total {
		to = total
	}
	return Paginated{
		CurrentPage: page,
		Data:        data,
		From:        from,
		To:          to,
		LastPage:    lastPage,
		PerPage:     perPage,
		Total:       total,
	}
}

// PageParams reads standard `page`/`per_page` query params with sane
// defaults, and returns the SQL OFFSET/LIMIT pair alongside the page number.
func PageParams(r *http.Request, defaultPerPage int) (page, perPage, offset int) {
	page = intQuery(r, "page", 1)
	perPage = intQuery(r, "per_page", defaultPerPage)
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	offset = (page - 1) * perPage
	return
}

func intQuery(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n := 0
	neg := false
	for i, c := range v {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c < '0' || c > '9' {
			return fallback
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	return n
}
