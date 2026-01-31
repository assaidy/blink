include .env

sqlc:
	@go tool sqlc generate

tailwind:
	@tailwindcss --minify -i ./app/web/tailwind_input.css -o ./app/web/public/css/style.css

build: sqlc tailwind
	@go build -o ./bin/api ./cmd/api/

run: build
	@./bin/api

WATCH_CMD = watchexec --ignore-nothing
watch:
	# tailwind watch mode didn't work when i put '&' at the end
	@$(WATCH_CMD) -w ./app/web/tailwind_input.css \
								-w ./app/web/public/js/   -e js \
								-w ./app/web/components/  -e go \
								-- tailwindcss --minify -i ./app/web/tailwind_input.css -o ./app/web/public/css/style.css &
	@$(WATCH_CMD) -w ./app/db/ -e sql -- go tool sqlc generate &
	@$(WATCH_CMD) --restart \
								-w ./app/            -e go    \
								-w ./cmd/api/        -e go    \
								-w ./app/web/public/ -e js,js \
								-- "go build -o ./bin/api ./cmd/api/ && ./bin/api"

clean:
	@rm -rf ./bin/

docker-up-dev:
	@docker compose up postgres_db valkey_cache papercut_smtp

docker-up-prod:
	@docker compose up --build --scale api=4

docker-down:
	@docker compose down

GOOSE_ENV = GOOSE_DRIVER="postgres" GOOSE_DBSTRING="$(DB_URL_LOCAL)" GOOSE_MIGRATION_DIR="./app/db/migrations/"
goose-up:
	@$(GOOSE_ENV) goose up

goose-down:
	@$(GOOSE_ENV) goose down

goose-reset:
	@$(GOOSE_ENV) goose reset

goose-new:
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
