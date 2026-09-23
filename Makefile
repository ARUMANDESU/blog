.PHONY: dev run
dev:
	templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/blog"
run:
	templ generate
	go run ./cmd/blog
