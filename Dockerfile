ARG GO_VERSION
FROM golang:$GO_VERSION-alpine AS build-dev
RUN apk add --no-cache bash build-base git tree curl protobuf openssh
WORKDIR /src

ENV GOBIN=/bin
ENV ROOT_DIR=/src

# generate & build
ARG VERSION
ARG COMMIT

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    	--mount=type=cache,target=/root/.cache/go-build \
    	--mount=type=ssh \
    	go run mage.go deps build

FROM alpine
ARG VERSION
ARG COMMIT

LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.source=https://github.com/aserto-dev/topaz
LABEL org.opencontainers.image.title="Topaz"
LABEL org.opencontainers.image.revision=$COMMIT
LABEL org.opencontainers.image.url=https://aserto.com

RUN apk add --no-cache bash tzdata
WORKDIR /app
COPY --from=build-dev /src/dist/topazd_linux_amd64_v1/topazd /app/

EXPOSE 8282
EXPOSE 8383
EXPOSE 8484
EXPOSE 8585
EXPOSE 9292

ENTRYPOINT ["./topazd"]
