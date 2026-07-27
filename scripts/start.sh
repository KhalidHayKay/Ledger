#!/bin/sh
if [ "$APP_ENV" = "local" ]; then
    go mod download
    exec air
else
    exec ./cmd/app
fi