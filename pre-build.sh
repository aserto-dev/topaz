#!/bin/sh

docker pull ghcr.io/aserto-dev/self-hosted-console:v3
id=$(docker create ghcr.io/aserto-dev/self-hosted-console:v3)
docker cp $id:/app/build ./pkg/app/console
docker rm -v $id
