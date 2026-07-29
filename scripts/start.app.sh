#!/bin/sh
if [ "$APP_ENV" = "local" ]; then
    go mod download
     exec air -c .air.app.toml
else
    exec ../cmd/app
fi