IMAGE := blog

.PHONY: dev run sqlc templ clean-db docker-build docker-run docker-build-run docker-clean-db

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

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm -p 8080:8080 -v $(IMAGE)-data:/app/data $(IMAGE)

docker-build-run: docker-build docker-run

docker-clean-db:
	docker volume rm $(IMAGE)-data
