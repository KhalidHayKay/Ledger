FROM golang:1.26-alpine

RUN apk add --no-cache gcc musl-dev git make bash

WORKDIR /var/www

ARG APP_ENV

RUN if [ "$APP_ENV" = "local" ]; then \
    go install github.com/air-verse/air@v1.65.3 && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2; \
    fi

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN if [ "$APP_ENV" != "local" ]; then \
    go build -o cmd/app/main ./cmd/app; \
    fi

ENTRYPOINT ["/bin/sh", "start.sh"]