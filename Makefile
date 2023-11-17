export DB_NAME=test
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=test
export DB_PASSWORD=test
export TOPIC=crud

ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=test password=test dbname=test host=localhost port=5432 sslmode=disable
endif

INTERNAL_PKG_PATH=$(CURDIR)/internal/
MIGRATION_FOLDER=$(INTERNAL_PKG_PATH)/db/migrations

.PHONY: migration-create
migration-create:
	goose -dir "$(MIGRATION_FOLDER)" create "$(name)" sql

.PHONY: test-migration-up
test-migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" up

.PHONY: test-migration-down
test-migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" down

.PHONY: test-environment
test-environment:
	docker-compose up -d

.PHONY: test-unit-test
test-unit-test:
	go test ./... -cover

.PHONY: test-integration-test
test-integration-test:
	go test -tags integration ./... -cover
