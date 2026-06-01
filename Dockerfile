FROM golang:1.26-alpine

WORKDIR /var/www

ARG APP_ENV

RUN if [ "$APP_ENV" = "local" ]; then \
    go install github.com/air-verse/air@latest; \
    fi

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN if [ "$APP_ENV" != "local" ]; then \
    go build -o cmd/app/main ./cmd/app; \
    fi

ENTRYPOINT ["/bin/sh", "start.sh"]