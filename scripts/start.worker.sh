#!/bin/sh
if [ "$APP_ENV" = "local" ]; then
    exec air -c .air.worker.toml
else
    exec ../cmd/worker
fi