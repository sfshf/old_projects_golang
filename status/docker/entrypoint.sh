#!/usr/bin/env sh
set -Eeo pipefail

# backend server
cd /app/ui && node server.js &

exec "$@"
