#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

while IFS= read -r -d '' file; do
    echo "Processing: ${file} => ${file}l"
    jq -c '.objects[]' "${file}" > "${file}l"
done < <(find . -type f -name "*_objects.json" -print0)

while IFS= read -r -d '' file; do
    echo "Processing: ${file} => ${file}l"
    jq -c '.relations[]' "${file}" > "${file}l"
done < <(find . -type f -name "*_relations.json" -print0)