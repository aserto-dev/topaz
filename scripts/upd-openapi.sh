#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

printf 'convert *.swagger.json => *.openapi.json\n'

for FILENAME in $(find . -name '*.swagger.json'); do \
   echo "${FILENAME} => ${FILENAME//swagger/openapi}"; \
  .ext/bin/openapi-spec-converter -t 3.1 -f json -o "${FILENAME//swagger/openapi}" ${FILENAME};\
  rm -f ${FILENAME}
done

printf '\n'

mkdir -p ./api/openapi

.ext/bin/merge-json -output ./api/openapi/directory.openapi.json $(find ./tmp -type f -name "*.json")

rm -rf ./tmp
