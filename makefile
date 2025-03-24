SHELL              := $(shell which bash)

NO_COLOR           := \033[0m
OK_COLOR           := \033[32;01m
ERR_COLOR          := \033[31;01m
WARN_COLOR         := \033[36;01m
ATTN_COLOR         := \033[33;01m

GOOS               := $(shell go env GOOS)
GOARCH             := $(shell go env GOARCH)
GOPRIVATE          := "github.com/aserto-dev"
DOCKER_BUILDKIT    := 1

EXT_DIR            := ${PWD}/.ext
EXT_BIN_DIR        := ${EXT_DIR}/bin
EXT_TMP_DIR        := ${EXT_DIR}/tmp

GO_VER             := 1.24
VAULT_VER	         := 1.8.12
SVU_VER 	         := 3.1.0
GOTESTSUM_VER      := 1.11.0
GOLANGCI-LINT_VER  := 1.64.5
GORELEASER_VER     := 2.3.2
WIRE_VER	         := 0.6.0
CHECK2DECISION_VER := 0.1.0
SYFT_VER           := 1.13.0

RELEASE_TAG        := $$(${EXT_BIN_DIR}/svu current)

.DEFAULT_GOAL      := build

.PHONY: deps
deps: info install-vault install-svu install-goreleaser install-golangci-lint install-gotestsum install-wire install-check2decision install-syft
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"

.PHONY: gover
gover:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@(go env GOVERSION | grep "go${GO_VER}") || (echo "go version check failed expected go${GO_VER} got $$(go env GOVERSION)"; exit 1)

.PHONY: build
build: gover
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/goreleaser build --clean --snapshot --single-target

PHONY: go-mod-tidy
go-mod-tidy:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@go work edit -json | jq -r '.Use[].DiskPath' | xargs -I{} bash -c 'cd {} && echo "${PWD}/go.mod" && go mod tidy -v && cd -'

.PHONY: release
release:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/goreleaser release --clean

.PHONY: snapshot
snapshot:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/goreleaser release --clean --snapshot

.PHONY: generate
generate:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@GOBIN=${EXT_BIN_DIR} go generate ./...

.PHONY: lint
lint: gover
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/golangci-lint config path
	@${EXT_BIN_DIR}/golangci-lint config verify
	@${EXT_BIN_DIR}/golangci-lint run --config ${PWD}/.golangci.yaml

.PHONY: test-snapshot
test-snapshot:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@docker image rm ghcr.io/aserto-dev/topaz:0.0.0-test-$$(git rev-parse --short HEAD)-$$(uname -m) || true
	@${EXT_BIN_DIR}/goreleaser release --config .goreleaser-test.yml --clean --snapshot --skip archive

.PHONE: container-tag
container-tag:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@scripts/container-tag.sh > .container-tag.env

.PHONY: run-test-snapshot
run-test-snapshot:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@echo "topaz run $$(${PWD}/dist/topaz_${GOOS}_${GOARCH}/topaz config info | jq '.runtime.active_configuration_file')"
	@${PWD}/dist/topaz_${GOOS}_${GOARCH}/topaz run --container-tag=0.0.0-test-$$(git rev-parse --short HEAD)-$$(uname -m)

.PHONY: start-test-snapshot
start-test-snapshot:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@echo "topaz start $$(${PWD}/dist/topaz_${GOOS}_${GOARCH}/topaz config info | jq '.runtime.active_configuration_name')"
	@${PWD}/dist/topaz_${GOOS}_${GOARCH}/topaz start --container-tag=0.0.0-test-$$(git rev-parse --short HEAD)-$$(uname -m)

.PHONY: test
test: gover test-snapshot run-test
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"

.PHONY: run-test
run-test:
	@echo -e "$(ATTN_COLOR)==> run-test github.com/aserto-dev/topaz/pkg/app/tests/... $(NO_COLOR)"
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- -count=1 -timeout 120s -parallel=1 -v -coverprofile=cover.out -coverpkg=github.com/aserto-dev/topaz/pkg/app/tests/... github.com/aserto-dev/topaz/pkg/app/tests/...

.PHONY: run-tests
run-tests:
	@echo -e "$(ATTN_COLOR)==> run-tests github.com/aserto-dev/topaz/pkg/app/tests/... $(NO_COLOR)"
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- -count=1 -timeout 120s -parallel=1 -v -coverprofile=cover.out -coverpkg=github.com/aserto-dev/topaz/pkg/app/tests/authz/... github.com/aserto-dev/topaz/pkg/app/tests/authz/...
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- -count=1 -timeout 120s -parallel=1 -v -coverprofile=cover.out -coverpkg=github.com/aserto-dev/topaz/pkg/app/tests/builtin/... github.com/aserto-dev/topaz/pkg/app/tests/builtin/...
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- -count=1 -timeout 120s -parallel=1 -v -coverprofile=cover.out -coverpkg=github.com/aserto-dev/topaz/pkg/app/tests/ds/... github.com/aserto-dev/topaz/pkg/app/tests/ds/...
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- -count=1 -timeout 120s -parallel=1 -v -coverprofile=cover.out -coverpkg=github.com/aserto-dev/topaz/pkg/app/tests/manifest/... github.com/aserto-dev/topaz/pkg/app/tests/manifest/...
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- -count=1 -timeout 120s -parallel=1 -v -coverprofile=cover.out -coverpkg=github.com/aserto-dev/topaz/pkg/app/tests/policy/... github.com/aserto-dev/topaz/pkg/app/tests/policy/...
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- -count=1 -timeout 120s -parallel=1 -v -coverprofile=cover.out -coverpkg=github.com/aserto-dev/topaz/pkg/app/tests/query/... github.com/aserto-dev/topaz/pkg/app/tests/query/...
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- -count=1 -timeout 120s -parallel=1 -v -coverprofile=cover.out -coverpkg=github.com/aserto-dev/topaz/pkg/app/tests/template/... github.com/aserto-dev/topaz/pkg/app/tests/template/...
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- -count=1 -timeout 120s -parallel=1 -v -coverprofile=cover.out -coverpkg=github.com/aserto-dev/topaz/pkg/app/tests/template-no-tls/... github.com/aserto-dev/topaz/pkg/app/tests/template-no-tls/...

.PHONY: write-version
write-version:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@git describe --tags > ./VERSION.txt

ASSETS = "assets/v32/api-auth/test/api-auth_" "assets/v32/gdrive/test/gdrive_" "assets/v32/github/test/github_" "assets/v32/multi-tenant/test/multi-tenant_" "assets/v32/slack/test/slack_" \
		 "assets/v33/api-auth/test/api-auth_" "assets/v33/gdrive/test/gdrive_" "assets/v33/github/test/github_" "assets/v33/multi-tenant/test/multi-tenant_" "assets/v33/slack/test/slack_"
.PHONY: update-assets
update-assets: $(ASSETS)
$(ASSETS): install-check2decision
	@echo -e "$(ATTN_COLOR)==> github.com/aserto-dev/topaz/$@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/check2decision -i "$@assertions.json" -o "$@decisions.json"

TEMPLATES = "assets/v32/api-auth.json" "assets/v32/gdrive.json" "assets/v32/github.json" "assets/v32/multi-tenant.json" "assets/v32/peoplefinder.json" "assets/v32/simple-rbac.json" "assets/v32/slack.json" "assets/v32/todo.json" \
			"assets/v33/api-auth.json" "assets/v33/gdrive.json" "assets/v33/github.json" "assets/v33/multi-tenant.json" "assets/v33/peoplefinder.json" "assets/v33/simple-rbac.json" "assets/v33/slack.json" "assets/v33/todo.json"
.PHONY: test-templates
test-templates: $(TEMPLATES)
$(TEMPLATES): test-snapshot
	@echo -e "$(ATTN_COLOR)==> github.com/aserto-dev/topaz/assets/$@ $(NO_COLOR)"
	@echo topaz templates install $@ --force --no-console --container-tag=test-$$(git rev-parse --short HEAD)-${GOARCH}
	@./dist/topaz_${GOOS}_${GOARCH}/topaz templates install $@ --force --no-console --container-tag=test-$$(git rev-parse --short HEAD)-${GOARCH}
	@./dist/topaz_${GOOS}_${GOARCH}/topaz stop --wait

.PHONY: vault-login
vault-login:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@vault login -method=github token=$$(gh auth token)

.PHONY: info
info:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@echo "GOOS:        ${GOOS}"
	@echo "GOARCH:      ${GOARCH}"
	@echo "EXT_DIR:     ${EXT_DIR}"
	@echo "EXT_BIN_DIR: ${EXT_BIN_DIR}"
	@echo "EXT_TMP_DIR: ${EXT_TMP_DIR}"
	@echo "RELEASE_TAG: ${RELEASE_TAG}"

.PHONY: install-vault
install-vault: ${EXT_BIN_DIR} ${EXT_TMP_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@curl -s -o ${EXT_TMP_DIR}/vault.zip https://releases.hashicorp.com/vault/${VAULT_VER}/vault_${VAULT_VER}_${GOOS}_${GOARCH}.zip
	@unzip -o ${EXT_TMP_DIR}/vault.zip vault -d ${EXT_BIN_DIR}/  &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/vault
	@${EXT_BIN_DIR}/vault --version

.PHONY: install-svu
install-svu: ${EXT_BIN_DIR} ${EXT_TMP_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${SVU_VER} --repo https://github.com/caarlos0/svu --pattern "*${GOOS}_all.tar.gz" --output "${EXT_TMP_DIR}/svu.tar.gz" --clobber
	@tar -xvf ${EXT_TMP_DIR}/svu.tar.gz --directory ${EXT_BIN_DIR} svu &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/svu
	@${EXT_BIN_DIR}/svu --version

.PHONY: install-gotestsum
install-gotestsum: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${GOTESTSUM_VER} --repo https://github.com/gotestyourself/gotestsum --pattern "gotestsum_${GOTESTSUM_VER}_${GOOS}_${GOARCH}.tar.gz" --output "${EXT_TMP_DIR}/gotestsum.tar.gz" --clobber
	@tar -xvf ${EXT_TMP_DIR}/gotestsum.tar.gz --directory ${EXT_BIN_DIR} gotestsum &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/gotestsum
	@${EXT_BIN_DIR}/gotestsum --version

.PHONY: install-golangci-lint
install-golangci-lint: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${GOLANGCI-LINT_VER} --repo https://github.com/golangci/golangci-lint --pattern "golangci-lint-${GOLANGCI-LINT_VER}-${GOOS}-${GOARCH}.tar.gz" --output "${EXT_TMP_DIR}/golangci-lint.tar.gz" --clobber
	@tar --strip=1 -xvf ${EXT_TMP_DIR}/golangci-lint.tar.gz --strip-components=1 --directory ${EXT_TMP_DIR} &> /dev/null
	@mv ${EXT_TMP_DIR}/golangci-lint ${EXT_BIN_DIR}/golangci-lint
	@chmod +x ${EXT_BIN_DIR}/golangci-lint
	@${EXT_BIN_DIR}/golangci-lint --version

.PHONY: install-goreleaser
install-goreleaser: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${GORELEASER_VER} --repo https://github.com/goreleaser/goreleaser --pattern "goreleaser_$$(uname -s)_$$(uname -m).tar.gz" --output "${EXT_TMP_DIR}/goreleaser.tar.gz" --clobber
	@tar -xvf ${EXT_TMP_DIR}/goreleaser.tar.gz --directory ${EXT_BIN_DIR} goreleaser &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/goreleaser
	@${EXT_BIN_DIR}/goreleaser --version

.PHONY: install-wire
install-wire: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@GOBIN=${EXT_BIN_DIR} go install github.com/google/wire/cmd/wire@v${WIRE_VER}

.PHONY: install-check2decision
install-check2decision: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${CHECK2DECISION_VER} --repo https://github.com/aserto-dev/check2decision --pattern "check2decision-${GOOS}-${GOARCH}.zip" --output "${EXT_TMP_DIR}/check2decision.zip" --clobber
	@unzip -o ${EXT_TMP_DIR}/check2decision.zip check2decision -d ${EXT_BIN_DIR}/  &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/check2decision
	@${EXT_BIN_DIR}/check2decision --version

.PHONY: install-syft
install-syft: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${SYFT_VER} --repo https://github.com/anchore/syft --pattern "syft_${SYFT_VER}_${GOOS}_${GOARCH}.tar.gz" --output "${EXT_TMP_DIR}/syft.tar.gz" --clobber
	@tar -xvf ${EXT_TMP_DIR}/syft.tar.gz --directory ${EXT_BIN_DIR} syft &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/syft
	@${EXT_BIN_DIR}/syft --version

.PHONY: clean
clean:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@rm -rf ${EXT_DIR}
	@rm -rf ${BIN_DIR}
	@rm -rf ./dist
	@rm -rf ./test

${BIN_DIR}:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@mkdir -p ${BIN_DIR}

${EXT_BIN_DIR}:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@mkdir -p ${EXT_BIN_DIR}

${EXT_TMP_DIR}:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@mkdir -p ${EXT_TMP_DIR}
