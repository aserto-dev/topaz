#!/usr/bin/env bash
echo "CONTAINER_TAG=0.0.0-test-$(git rev-parse --short HEAD)-$(uname -m)"
