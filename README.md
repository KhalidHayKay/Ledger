# Ledger

A small Payment Gateway service that mediates between a fictional e-commerce platform (FicMart) and a bank integration layer.

## Overview

Ledger accepts and manages payment intents, communicates with an external bank API, persists payment events, and runs background workers for async processing and retries. It is intentionally minimal and organized to make it easy to understand and extend.

## Architecture

- HTTP API server: exposes payment endpoints and health checks ([cmd/app](cmd/app/main.go#L1-L200)).
- CLI: lightweight migration helper using bundled SQL migrations ([cmd/cli](cmd/cli/main.go#L1-L120)).
- Worker: Asynq-based background worker that processes payment tasks ([cmd/worker](cmd/worker/main.go#L1-L200)).
- Postgres: persistent storage for payment intents and events.
- Redis: used for idempotency keys, notifier channels and Asynq.

Core domain packages live under `internal/domain/` and background queue/workers under `internal/jobs/`.

## Services (Compose)

This repository ships a Compose file at [compose.yaml](compose.yaml#L1-L40). The compose services are:

- `app` — the HTTP server (built from this repository).
- `worker` — runs background processors using the same image.
- `db` — Postgres (postgres:16-alpine).
- `redis` — Redis (redis:8-alpine).

Ports are driven by environment variables. The app maps `${APP_PORT}:${APP_PORT}` in Compose; set `APP_PORT` in your `.env` (commonly `8000`).

## Quickstart (Docker)

1. Copy or create a `.env` in the project root and fill required variables (see the Environment section).

2. Start services with Docker Compose:

```bash
docker compose up --build
```

3. Run the DB migrations using the `Makefile` (preferred):

```bash
make cli db:migrate
```

The HTTP API will be available at `http://localhost:${APP_PORT}`.

### Makefile (primary entrypoint)

This repository exposes a `Makefile` intended as the primary, discoverable entrypoint for running common commands against the running containers. Prefer using `make` targets instead of typing `docker compose exec ...` manually.

Examples (run after `docker compose up -d --build`):

```bash
# run the CLI migrations
make cli db:migrate

# tail container logs
make logs

# run tests inside the app service
make test

# run linter inside the app service
make lint
```

## Development (local)

For local development the Docker image installs `air` when `APP_ENV=local`. Use a compose override that mounts your source and sets `APP_ENV=local`, or run the server directly.

To run the app locally (without Docker):

```bash
# set up .env with local config
export APP_ENV=local
export APP_PORT=8000
# run the HTTP server
go run ./cmd/app
# run the worker in another terminal
go run ./cmd/worker
```

The repository includes start scripts used by the container: [scripts/start.app.sh](scripts/start.app.sh#L1-L10) and [scripts/start.worker.sh](scripts/start.worker.sh#L1-L10).

### Compose override (local development)

Create a `compose.override.yaml` in the project root to mount your source and enable hot reload with `air`:

```yaml
services:
	app:
		volumes:
			- ./:/var/www
			- go-module-cache:/root/go/pkg/mod
	worker:
		volumes:
			- ./:/var/www
			- go-module-cache:/root/go/pkg/mod

volumes:
	go-module-cache:
```

This file is intentionally not committed in many projects because it mounts local files into the container (development-only).

## HTTP Endpoints

The service exposes these endpoints:

- `GET /health` — health check
- `GET /payment/intent` — fetch payment intents (simple listing)
- `POST /payment/intent` — create payment intent
- `POST /payment/intent/capture` — capture a payment intent
- `POST /payment/intent/refund` — refund a payment
- `POST /payment/intent/cancel` — cancel a payment intent

The POST endpoints expect JSON payloads; see the `internal/domain/paymentintent` package for DTO shapes.

## Environment

The application relies on environment variables (loadable via a `.env` file). Required vars include:

- `APP_NAME`, `APP_VERSION`, `APP_ENV`, `APP_KEY`, `APP_PORT`, `APP_URL`
- `DB_URL` or `DB_CONNECTION`, `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD`
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- `BANK_API_BASE_URL` — the external bank API base URL used by the integration.

See [internal/platform/config/env.go](internal/platform/config/env.go#L1-L200) for how values are loaded and validated.

## Database

The repository includes SQL migrations embedded in the CLI at [cmd/cli](cmd/cli/main.go#L1-L200). Run them with:

```bash
go run ./cmd/cli db:migrate
```

Migrations create two core tables: `payment_intents` and `payment_events`.

## Background Jobs

Background processing uses Asynq with Redis. The worker is implemented in [cmd/worker](cmd/worker/main.go#L1-L200) and maps task types to handlers under `internal/jobs/`.
