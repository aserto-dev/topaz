#!/usr/bin/env bash

convertCheck2Decision() {
  check2decision -i "${asset}assertions.json" -o "${asset}decisions.json"
}

assets=("assets/api-auth/test/api-auth_" "assets/gdrive/test/gdrive_" "assets/github/test/github_" "assets/multi-tenant/test/multi-tenant_" "assets/slack/test/slack_" )

for asset in ${assets[@]}; do
  convertCheck2Decision $asset
done
