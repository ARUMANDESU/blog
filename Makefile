.PHONY: dev
dev:
	templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/blog"
