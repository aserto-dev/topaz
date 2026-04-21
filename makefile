SHELL              := $(shell which bash)

NO_COLOR           := \033[0m
OK_COLOR           := \033[32;01m
ERR_COLOR          := \033[31;01m
WARN_COLOR         := \033[36;01m
ATTN_COLOR         := \033[33;01m

REGISTRY           := ghcr.io
ORG                := aserto-dev
REPO               := topaz
DESCRIPTION        := "Topaz Authorization Service"
LICENSE            := Apache-2.0

GOOS               := $(shell go env GOOS)
GOARCH             := $(shell go env GOARCH)
TOPAZ_DIST         := ${PWD}/$(shell cat dist/artifacts.json | jq -r '.[] | select(.name == "topaz").path')
GOPRIVATE          := "github.com/aserto-dev"
DOCKER_BUILDKIT    := 1

BIN_DIR            := ${PWD}/bin
EXT_DIR            := ${PWD}/.ext
EXT_BIN_DIR        := ${EXT_DIR}/bin
EXT_TMP_DIR        := ${EXT_DIR}/tmp

GO_VER             := 1.26
SVU_VER            := 3.3.0
GOTESTSUM_VER      := 1.13.0
GOLANGCI-LINT_VER  := 2.11.4
GORELEASER_VER     := 2.14.1
SYFT_VER           := 1.13.0
BUF_VER            := 1.66.1
MERGE-JSON_VER     := 0.1.6

BUF_REPO           := "buf.build/topaz/directory"
BUF_LATEST         := $(shell ${EXT_BIN_DIR}/buf registry module label list ${BUF_REPO} --format json | jq -r '.labels[0].name')
BUF_DEV_IMAGE      := "directory.bin"

RELEASE_TAG        := $$(${EXT_BIN_DIR}/svu current)

.DEFAULT_GOAL      := build

.PHONY: deps
deps: info install-svu install-goreleaser install-golangci-lint install-gotestsum install-syft install-buf install-openapi-spec-converter install-merge-json
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"

.PHONY: gover
gover:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@(go env GOVERSION | grep "go${GO_VER}") || (echo "go version check failed expected go${GO_VER} got $$(go env GOVERSION)"; exit 1)

.PHONY: build
build: gover
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/goreleaser build --config .goreleaser.yml --clean --snapshot --single-target

.PHONY: docker-build
docker-build:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@docker buildx build \
  --platform=linux/arm64,linux/amd64 \
  --tag ${REGISTRY}/${ORG}/${REPO}:$$(svu current) \
	--tag ${REGISTRY}/${ORG}/${REPO}:latest  \
	--progress=plain \
  --build-arg BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
	--build-arg TITLE=${REPO} \
  --build-arg VCS_REF=$$(git rev-parse HEAD) \
  --build-arg VERSION=$$(svu current) \
  --build-arg REPO_URL="https://github.com/${ORG}/${REPO}" \
  --build-arg DESCRIPTION=${DESCRIPTION} \
  --build-arg LICENSE=${LICENSE} \
  --push .

.PHONY: docker-build-test
docker-build-test:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@docker buildx build \
  --platform=linux/${GOARCH} \
  --tag ${REGISTRY}/${ORG}/${REPO}:0.0.0-test-$$(git rev-parse --short HEAD)-$(GOARCH) \
	--progress=plain \
  --build-arg BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
	--build-arg TITLE=${REPO} \
  --build-arg VCS_REF=$$(git rev-parse HEAD) \
  --build-arg VERSION=$$(svu current) \
  --build-arg REPO_URL="https://github.com/${ORG}/${REPO}" \
  --build-arg DESCRIPTION=${DESCRIPTION} \
  --build-arg LICENSE=${LICENSE} \
  .

PHONY: go-mod-tidy
go-mod-tidy:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@go work edit -json | jq -r '.Use[].DiskPath' | xargs -I{} bash -c 'cd {} && echo "${PWD}/go.mod" && go mod tidy -v && cd -'

.PHONY: release
release: gover
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/goreleaser release --config .goreleaser.yml --clean

.PHONY: snapshot
snapshot: gover
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/goreleaser release --config .goreleaser.yml --clean --snapshot --skip archive,homebrew,sbom

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

.PHONY: test
test: gover test-snapshot
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- $$(go list ./... | grep -v topazd/tests)                     -count=1 -timeout 120s --race -parallel=1 -v -coverprofile=cover.out -coverpkg=./...
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- $$(go list ./topazd/tests/... | grep -v tests/template)      -count=1 -timeout 120s --race -parallel=1 -v -coverprofile=cover.out -coverpkg=./...
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- github.com/${ORG}/${REPO}/topazd/tests/template-no-tls/...   -count=1 -timeout 120s --race -parallel=1 -v -coverprofile=cover.out -coverpkg=./...
	@${EXT_BIN_DIR}/gotestsum --format short-verbose -- github.com/${ORG}/${REPO}/topazd/tests/template-with-tls/... -count=1 -timeout 120s --race -parallel=1 -v -coverprofile=cover.out -coverpkg=./...
	
.PHONY: test-snapshot
test-snapshot:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@docker image rm ${REGISTRY}/${ORG}/${REPO}:0.0.0-test-$$(git rev-parse --short HEAD)-$$(uname -m) || true
	@${EXT_BIN_DIR}/goreleaser release --config .goreleaser-test.yml --clean --snapshot --skip archive,homebrew,sbom

.PHONE: container-tag
container-tag:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@scripts/container-tag.sh > .container-tag.env

.PHONY: write-version
write-version:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@git describe --tags > ./VERSION.txt

.PHONY: topaz-run-test-snapshot
topaz-run-test-snapshot:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@echo "topaz run $$(${TOPAZ_DIST} config info | jq '.runtime.active_configuration_file')"
	@${TOPAZ_DIST} run --container-tag=0.0.0-test-$$(git rev-parse --short HEAD)-$$(uname -m)

.PHONY: topaz-start-test-snapshot
topaz-start-test-snapshot:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@echo "topaz start $$(${TOPAZ_DIST} config info | jq '.runtime.active_configuration_name')"
	@${TOPAZ_DIST} start --container-tag=0.0.0-test-$$(git rev-parse --short HEAD)-$$(uname -m)

.PHONY: buf-login
buf-login:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@echo ${BUF_TOKEN} | ${EXT_BIN_DIR}/buf registry login --token-stdin

.PHONY: buf-dep-update
buf-dep-update:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf dep update

.PHONY: buf-format
buf-format:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf format -w proto

.PHONY: buf-build
buf-build: ${BIN_DIR} buf-format 
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf build --output ${BIN_DIR}/${BUF_DEV_IMAGE}

.PHONY: buf-lint
buf-lint:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf lint

.PHONY: buf-breaking
buf-breaking:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf breaking --against "${GIT_ORG}/${PROTO_REPO}.git#branch=main"

.PHONY: buf-push
buf-push:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/buf push --label ${RELEASE_TAG}

.PHONY: buf-generate
buf-generate:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@find . -name 'buf.gen*.yaml' -exec ${EXT_BIN_DIR}/buf generate --template {} \;
	@${PWD}/scripts/upd-openapi.sh

.PHONY: buf-generate-dev
buf-generate-dev:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@find . -name 'buf.gen*.yaml' -exec ${EXT_BIN_DIR}/buf generate --template {} ${PWD}/bin/${BUF_DEV_IMAGE} \;
	@${PWD}/scripts/upd-openapi.sh

.PHONY: info
info:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@echo "GOOS:          ${GOOS}"
	@echo "GOARCH:        ${GOARCH}"
	@echo "BIN_DIR:       ${BIN_DIR}"
	@echo "EXT_DIR:       ${EXT_DIR}"
	@echo "EXT_BIN_DIR:   ${EXT_BIN_DIR}"
	@echo "EXT_TMP_DIR:   ${EXT_TMP_DIR}"
	@echo "RELEASE_TAG:   ${RELEASE_TAG}"
	@echo "TOPAZ_DIST:    ${TOPAZ_DIST}"
	@echo "REGISTRY:      ${REGISTRY}"
	@echo "ORG:           ${ORG}"
	@echo "REPO:          ${REPO}"
	@echo "BUF_REPO:      ${BUF_REPO}"
	@echo "BUF_LATEST:    ${BUF_LATEST}"
	@echo "BUF_DEV_IMAGE: ${BUF_DEV_IMAGE}"
	@echo "PROTO_REPO:    ${PROTO_REPO}"

.PHONY: install-svu
install-svu: ${EXT_BIN_DIR} ${EXT_TMP_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@GOBIN=${EXT_BIN_DIR} go install github.com/caarlos0/svu/v3@v${SVU_VER}
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

.PHONY: install-syft
install-syft: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${SYFT_VER} --repo https://github.com/anchore/syft --pattern "syft_${SYFT_VER}_${GOOS}_${GOARCH}.tar.gz" --output "${EXT_TMP_DIR}/syft.tar.gz" --clobber
	@tar -xvf ${EXT_TMP_DIR}/syft.tar.gz --directory ${EXT_BIN_DIR} syft &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/syft
	@${EXT_BIN_DIR}/syft --version

.PHONY: install-buf
install-buf: ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@GOBIN=${EXT_BIN_DIR} go install github.com/bufbuild/buf/cmd/buf@v${BUF_VER}
	@${EXT_BIN_DIR}/buf --version

.PHONY: install-openapi-spec-converter
install-openapi-spec-converter: ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@GOBIN=${EXT_BIN_DIR} go install github.com/dense-analysis/openapi-spec-converter/cmd/openapi-spec-converter@latest

.PHONY: install-merge-json
install-merge-json: ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@GOBIN=${EXT_BIN_DIR} go install github.com/topaz-authz/merge-json@v${MERGE-JSON_VER}

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
