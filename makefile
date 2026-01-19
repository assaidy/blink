include .env

GOOSE_ENV = GOOSE_DRIVER="postgres" GOOSE_DBSTRING="$(DB_URL_LOCAL)" GOOSE_MIGRATION_DIR="./app/db/migrations/"

all: build

sqlc:
	@go tool sqlc generate

build: sqlc
	@go build -o ./bin/app ./cmd/app/

run: sqlc
	@go run ./cmd/app/

watch:
	@watchexec -w ./app/db/ -e sql -- make sqlc &
	@watchexec --restart -w ./app/ -e go -w ./cmd/app/ -e go -- go run ./cmd/app/

clean:
	@rm -rf ./bin/

comp-up-prod:
	 @docker compose up --build

comp-up-dev:
	 @docker compose up postgres_db valkey_cache papercut_smtp

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
