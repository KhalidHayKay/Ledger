APP_NAME := smply
CLI_DIR := ./cmd/cli
SERVER_DIR := ./cmd/app

.PHONY: test lint cli server build

cli:
	docker compose exec app go run $(CLI_DIR) $(filter-out $@,$(MAKECMDGOALS))

logs:
	docker compose logs -f --tail=10

test:
	docker compose exec app go test -v ./...

lint:
	docker compose exec app golangci-lint run ./...

%:
	@: