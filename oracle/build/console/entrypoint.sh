#!/usr/bin/env bash
set -Eeo pipefail

env=$APP_ENV
envString="dev"
tag="dev"
if [[ "$env" == "1" ]]; then
    envString="ppe"
    tag="test"
fi
if [[ "$env" == "2" ]]; then
    envString="prod"
    tag="prod"
fi
id="${envString}-go-${CONSOLE_APP_NAME}"
sed -i "s/ID/$id/g" /etc/filebeat/filebeat.yml
sed -i "s/TAG/$tag/g" /etc/filebeat/filebeat.yml
chmod go-w /etc/filebeat/filebeat.yml

if [[ "$env" == "1" || "$env" == "2" ]]; then
    filebeat &
fi

exec "$@"
