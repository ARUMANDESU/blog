IMAGE := blog
GARAGE_CONTAINER ?= blog-garage-s3-1
COMPOSE_ENV_FILE ?= .env.compose

ifeq (garage,$(firstword $(MAKECMDGOALS)))
  GARAGE_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(GARAGE_ARGS):;@:)
endif

.PHONY: dev run sqlc templ clean-db clean-run docker-build docker-run docker-build-run docker-clean-db

clean-db:
	rm -f data/*

sqlc:
	sqlc generate

templ:
	templ generate

run: sqlc templ
	@set -a && . ./.env && set +a && go run ./cmd/blog

clean-run: clean-db run

dev:
	templ generate --watch --proxy="http://localhost:8080" --cmd="go run ./cmd/blog"

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm -p 8080:8080 -v $(IMAGE)-data:/app/data $(IMAGE)

docker-build-run: docker-build docker-run

docker-clean-db:
	docker volume rm $(IMAGE)-data

compose-up:
	docker compose --env-file=${COMPOSE_ENV_FILE} up -d

compose-down:
	docker compose --env-file=${COMPOSE_ENV_FILE} down

garage-up:
	docker compose --env-file=${COMPOSE_ENV_FILE} up -d garage-s3

garage-down:
	docker compose --env-file=${COMPOSE_ENV_FILE} down garage-s3

garage:
	docker exec -ti $(GARAGE_CONTAINER) /garage $(GARAGE_ARGS)

debug-up:
	docker compose --env-file=${COMPOSE_ENV_FILE} --profile debug up -d

debug-down:
	docker compose --env-file=${COMPOSE_ENV_FILE} --profile debug down
