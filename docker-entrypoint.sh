#!/usr/bin/env bash
COMMAND=$(sed 's~@@TOPAZ_DIR@@~'${TOPAZ_DIR}'~' <<<"$@")
./topazd $COMMAND
