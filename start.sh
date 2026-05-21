#!/bin/sh
if [ "$APP_ENV" = "local" ]; then
    exec air
else
    exec ./cmd/app
fi