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

BIN_DIR            := ./bin
EXT_DIR            := ./.ext
EXT_BIN_DIR        := ${EXT_DIR}/bin
EXT_TMP_DIR        := ${EXT_DIR}/tmp

GO_VER             := 1.23
VAULT_VER	         := 1.8.12
SVU_VER 	         := 1.12.0
GOTESTSUM_VER      := 1.11.0
GOLANGCI-LINT_VER  := 1.61.0
GORELEASER_VER     := 2.3.2
WIRE_VER	         := 0.6.0
BUF_VER            := 1.34.0
CHECK2DECISION_VER := 0.1.0
SYFT_VER           := 1.13.0

BUF_USER           := $(shell ${EXT_BIN_DIR}/vault kv get -field ASERTO_BUF_USER kv/buf.build)
BUF_TOKEN          := $(shell ${EXT_BIN_DIR}/vault kv get -field ASERTO_BUF_TOKEN kv/buf.build)
BUF_REPO           := "buf.build/aserto-dev/directory"
BUF_LATEST         := $(shell BUF_BETA_SUPPRESS_WARNINGS=1 ${EXT_BIN_DIR}/buf beta registry label list buf.build/aserto-dev/directory --format json --reverse | jq -r '.results[0].name')
BUF_DEV_IMAGE      := "../pb-directory/bin/directory.bin"

RELEASE_TAG        := $$(${EXT_BIN_DIR}/svu)

.DEFAULT_GOAL      := build

.PHONY: deps
deps: info install-vault install-buf install-svu install-goreleaser install-golangci-lint install-gotestsum install-wire install-check2decision install-syft
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"

.PHONY: build
build: 
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@(go env GOVERSION | grep "go${GO_VER}") || (echo "go version check failed expected go${GO_VER} got $$(go env GOVERSION)"; exit 1)
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
	@GOBIN=${PWD}/${EXT_BIN_DIR} go generate ./...

.PHONY: lint
lint:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/golangci-lint run --config ${PWD}/.golangci.yaml

.PHONY: test-snapshot
test-snapshot:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
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
test: test-snapshot run-test
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

ASSETS = "assets/api-auth/test/api-auth_" "assets/gdrive/test/gdrive_" "assets/github/test/github_" "assets/multi-tenant/test/multi-tenant_" "assets/slack/test/slack_"
.PHONY: update-assets
update-assets: $(ASSETS)
$(ASSETS): install-check2decision
	@echo -e "$(ATTN_COLOR)==> github.com/aserto-dev/topaz/assets/$@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/check2decision -i "$@assertions.json" -o "$@decisions.json"

TEMPLATES = "assets/api-auth.json" "assets/gdrive.json" "assets/github.json" "assets/multi-tenant.json" "assets/peoplefinder.json" "assets/simple-rbac.json" "assets/slack.json" "assets/todo.json"
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

.PHONY: buf-login
buf-login:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@echo ${BUF_TOKEN} | ${EXT_BIN_DIR}/buf registry login --username ${BUF_USER} --token-stdin

.PHONY: buf-lint
buf-lint:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf lint proto

.PHONY: buf-breaking
buf-breaking:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf breaking proto --against "https://github.com/aserto-dev/pb-directory.git#branch=main"

.PHONY: buf-build
buf-build: ${BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf build proto --output ${BIN_DIR}/directory.bin

.PHONY: buf-push
buf-push:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf push proto --label ${RELEASE_TAG}

.PHONY: buf-mod-update
buf-mod-update:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf mod update proto

.PHONY: buf-generate
buf-generate:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf mod update .
	@${EXT_BIN_DIR}/buf generate ${BUF_REPO}:${BUF_LATEST}

.PHONY: buf-generate-dev
buf-generate-dev:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf mod update .
	@${EXT_BIN_DIR}/buf generate "../pb-directory/bin/directory.bin"

.PHONY: info
info:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@echo "GOOS:        ${GOOS}"
	@echo "GOARCH:      ${GOARCH}"
	@echo "BIN_DIR:     ${BIN_DIR}"
	@echo "EXT_DIR:     ${EXT_DIR}"
	@echo "EXT_BIN_DIR: ${EXT_BIN_DIR}"
	@echo "EXT_TMP_DIR: ${EXT_TMP_DIR}"
	@echo "RELEASE_TAG: ${RELEASE_TAG}"
	@echo "BUF_REPO:    ${BUF_REPO}"
	@echo "BUF_LATEST:  ${BUF_LATEST}"

.PHONY: install-vault
install-vault: ${EXT_BIN_DIR} ${EXT_TMP_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@curl -s -o ${EXT_TMP_DIR}/vault.zip https://releases.hashicorp.com/vault/${VAULT_VER}/vault_${VAULT_VER}_${GOOS}_${GOARCH}.zip
	@unzip -o ${EXT_TMP_DIR}/vault.zip vault -d ${EXT_BIN_DIR}/  &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/vault
	@${EXT_BIN_DIR}/vault --version 

.PHONY: install-buf
install-buf: ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${BUF_VER} --repo https://github.com/bufbuild/buf --pattern "buf-$$(uname -s)-$$(uname -m)" --output "${EXT_BIN_DIR}/buf" --clobber
	@chmod +x ${EXT_BIN_DIR}/buf
	@${EXT_BIN_DIR}/buf --version

.PHONY: install-svu
install-svu: install-svu-${GOOS}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@chmod +x ${EXT_BIN_DIR}/svu
	@${EXT_BIN_DIR}/svu --version

.PHONY: install-svu-darwin
install-svu-darwin: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download --repo https://github.com/caarlos0/svu --pattern "svu_*_darwin_all.tar.gz" --output "${EXT_TMP_DIR}/svu.tar.gz" --clobber
	@tar -xvf ${EXT_TMP_DIR}/svu.tar.gz --directory ${EXT_BIN_DIR} svu &> /dev/null

.PHONY: install-svu-linux
install-svu-linux: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download --repo https://github.com/caarlos0/svu --pattern "svu_*_linux_${GOARCH}.tar.gz" --output "${EXT_TMP_DIR}/svu.tar.gz" --clobber
	@tar -xvf ${EXT_TMP_DIR}/svu.tar.gz --directory ${EXT_BIN_DIR} svu &> /dev/null

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
	@GOBIN=${PWD}/${EXT_BIN_DIR} go install github.com/google/wire/cmd/wire@v${WIRE_VER}

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
