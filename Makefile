.PHONY: dev run sqlc templ
dev:
	templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/blog"
run: sqlc templ
	go run ./cmd/blog
sqlc:
	sqlc generate
templ:
	templ generate
