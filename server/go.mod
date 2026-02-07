module server

go 1.25.5

replace shared => ../shared

require (
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.11.1
	shared v0.0.0-00010101000000-000000000000
)

require golang.org/x/crypto v0.47.0 // indirect
