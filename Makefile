.PHONY: dev run sqlc templ clean-db
clean-db:
	rm data/*
sqlc:
	sqlc generate
templ:
	templ generate
run: sqlc templ
	go run ./cmd/blog
dev:
	templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/blog"
