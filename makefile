include .env

GOOSE_ENV = GOOSE_DRIVER="postgres" GOOSE_DBSTRING="$(DB_URL_LOCAL)" GOOSE_MIGRATION_DIR="./app/db/migrations/"

all: build

sqlc:
	@go tool sqlc generate

templ:
	@go tool templ generate

tailwind:
	@npx tailwindcss -i ./app/web/input.css -o ./app/web/public/css/style.css --minify

build: sqlc templ tailwind
	@go build -o ./bin/api ./cmd/api/

run: build
	@./bin/api

clean:
	@rm -rf ./bin/

comp-up-dev:
	 @docker compose up postgres_db valkey_cache papercut_smtp

comp-up-prod:
	 @docker compose up --build --scale api=4

comp-down:
	 @docker compose down

goose-up:
	@$(GOOSE_ENV) goose up

goose-down:
	@$(GOOSE_ENV) goose down

goose-reset:
	@$(GOOSE_ENV) goose reset

goose-migration:
	@if [ -z "$(name)" ]; then \
		echo "error: 'name' variable is required."; \
		echo "usage: make goose-migration name=<migration_name_without_spaces>"; \
		exit 1; \
	fi
	@$(GOOSE_ENV) goose create -s $(name) sql

pg:
	@pgcli $(DB_URL_LOCAL)

vk:
	@valkey-cli -p $(VALKEY_PORT)
