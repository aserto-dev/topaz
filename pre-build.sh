#!/bin/sh

docker pull ghcr.io/aserto-dev/self-hosted-console:$1
id=$(docker create ghcr.io/aserto-dev/self-hosted-console:$1)
rm -rf ./pkg/app/console
docker cp $id:/app/build/. ./pkg/app/console
docker rm -v $id
