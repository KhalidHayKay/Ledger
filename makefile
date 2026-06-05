APP_NAME := smply
CLI_DIR := ./cmd/cli
SERVER_DIR := ./cmd/app

.PHONY: test lint cli server build

cli:
	docker compose exec app go run $(CLI_DIR) $(filter-out $@,$(MAKECMDGOALS))

logs:
	$(eval ARGS := $(filter-out $@,$(MAKECMDGOALS)))
	$(eval SERVICE := $(word 1,$(ARGS)))
	$(eval TAIL := $(or $(word 2,$(ARGS)),100))
	docker compose logs -f --tail=$(TAIL) $(SERVICE)

test:
	docker compose exec app go test -v ./...

lint:
	docker compose exec app golangci-lint run ./...

%:
	@: