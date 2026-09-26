package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"signet-backend/internal/storage"
)

func TestKycServeFile(t *testing.T) {
	store := &storage.Local{Root: t.TempDir()}
	png := "\x89PNG\r\n\x1a\n" + strings.Repeat("x", 20)
	store.Put(context.Background(), "kyc/nic_front/a.png", strings.NewReader(png), int64(len(png)), "")

	r := chi.NewRouter()
	r.Get("/storage/kyc/*", kycServeFile(store))

	cases := map[string]int{
		"/storage/kyc/nic_front/a.png":        http.StatusOK,
		"/storage/kyc/nic_front/missing.png":  http.StatusNotFound,
		"/storage/kyc/../../go.mod":           http.StatusNotFound,
		"/storage/kyc/nic_front/..%2f..%2fx":  http.StatusNotFound,
		"/storage/kyc/nic_front%5c..%5cx.png": http.StatusNotFound,
	}
	for url, want := range cases {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
		if rec.Code != want {
			t.Errorf("%s: status %d, want %d", url, rec.Code, want)
		}
		if want == http.StatusOK {
			if rec.Body.String() != png || rec.Header().Get("Content-Type") != "image/png" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Errorf("%s: body/headers wrong: %q %v", url, rec.Body.String(), rec.Header())
			}
		}
	}
}

func TestKycSafeExt(t *testing.T) {
	for in, want := range map[string]string{
		"photo.JPG": ".jpg", "scan.jpeg": ".jpeg", "a.png": ".png", "noext": "",
		"x.p hp": "", "x.<svg>": "", "x.toolongext": "", "x.": "",
	} {
		if got := kycSafeExt(in); got != want {
			t.Errorf("kycSafeExt(%q) = %q, want %q", in, got, want)
		}
	}
}
