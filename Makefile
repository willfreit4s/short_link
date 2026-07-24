# Database

DB_URL=postgres://postgres:postgres@localhost:5432/short_link?sslmode=disable

migUP:
	migrate -path=sql/migrations -database "$(DB_URL)" -verbose up

migDOWN:
	migrate -path=sql/migrations -database "$(DB_URL)" -verbose down

sqlc:
	sqlc generate

# Run the application
run:
	go run cmd/server/main.go


.PHONY: migUP migDOWN sqlc