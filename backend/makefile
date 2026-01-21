include .env

GOOSE_ENV = GOOSE_DRIVER="postgres" GOOSE_DBSTRING="$(DB_URL_LOCAL)" GOOSE_MIGRATION_DIR="./app/db/migrations/"

all: build

sqlc:
	@go tool sqlc generate

build: sqlc
	@go build -o ./bin/app ./cmd/api/

run: sqlc
	@go run ./cmd/api/

watch:
	@watchexec -w ./app/db/ -e sql -- make sqlc &
	@watchexec --restart -w ./app/ -e go -w ./cmd/api/ -e go -- go run ./cmd/api/

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
