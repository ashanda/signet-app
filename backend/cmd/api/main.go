// Command api is the Signet backend entrypoint.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"signet-backend/internal/app"
	"signet-backend/internal/auth"
	"signet-backend/internal/config"
	"signet-backend/internal/db"
	"signet-backend/internal/handlers"
	"signet-backend/internal/jobs"
)

func main() {
	cfg := config.Load()

	conn, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer conn.Close()

	authSvc := auth.New(conn, cfg.SessionSecret, cfg.SessionLifetime, cfg.AppEnv == "production")
	deps := &app.Deps{DB: conn, Auth: authSvc, Cfg: cfg}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware(cfg.FrontendOrigin))

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	handlers.RegisterAuthRoutes(r, deps)
	handlers.RegisterGoogleAuthRoutes(r, deps)
	handlers.RegisterTokenRoutes(r, deps)
	handlers.RegisterDashboardRoutes(r, deps)
	handlers.RegisterPackageRoutes(r, deps)
	handlers.RegisterCountriesRoutes(r, deps)
	handlers.RegisterKycRoutes(r, deps)
	handlers.RegisterGeneologyRoutes(r, deps)
	handlers.RegisterMiningRoutes(r, deps)
	handlers.RegisterRocRoutes(r, deps)
	handlers.RegisterSalaryRoutes(r, deps)
	handlers.RegisterDirectShareRoutes(r, deps)
	handlers.RegisterLeaderExecutiveRoutes(r, deps)
	handlers.RegisterUserRoutes(r, deps)
	handlers.RegisterUserParentMapsLogRoutes(r, deps)
	handlers.RegisterEarnLogRoutes(r, deps)
	handlers.RegisterAPIV1Routes(r, deps)

	if cfg.EnableScheduler {
		jobs.StartScheduler(conn, cfg)
	}

	log.Printf("Signet API listening on :%s (env=%s, db=%s@%s)", cfg.Port, cfg.AppEnv, cfg.DBDatabase, cfg.DBHost)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
