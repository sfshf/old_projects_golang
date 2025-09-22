#!/usr/bin/env bash

while getopts ":k:o:" opt; do
  case $opt in
    k) system_key="$OPTARG";;
    o) old_system_key="$OPTARG";; 
    \?) echo "unknown param $OPTARG" >&2;;
  esac
done
 
shift $((OPTIND-1))

echo "OLD_SYSTEM_KEY=${old_system_key}"
echo "SYSTEM_KEY=${system_key}"

docker build \
  -t nextsurfer/keystore:latest \
  -f build/keystore/Dockerfile .

docker stop -t 60 -s SIGTERM keystore && docker rm keystore || echo 0

docker run -d \
  --name keystore \
  --restart always \
  -e "OLD_SYSTEM_KEY=${old_system_key}" \
  -e "SYSTEM_KEY=${system_key}" \
  -v /var/lib/keystore:/var/lib/keystore \
  -v /:/host:ro,rslave \
  -p '1111:1111' \
  nextsurfer/keystore:latest

docker system prune -f