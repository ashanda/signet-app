module signet-backend

go 1.24.7

require (
	github.com/go-chi/chi/v5 v5.3.2
	github.com/go-sql-driver/mysql v1.10.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/jmoiron/sqlx v1.4.0
	golang.org/x/crypto v0.31.0
)

require filippo.io/edwards25519 v1.2.0 // indirect

replace golang.org/x/crypto => github.com/golang/crypto v0.31.0

replace filippo.io/edwards25519 => github.com/FiloSottile/edwards25519 v1.1.0
