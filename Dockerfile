FROM golang:1.17-alpine AS build-dev
RUN apk add --no-cache bash build-base git tree curl protobuf openssh
WORKDIR /src

# make sure git ssh is properly setup so we can access private repos
RUN mkdir -p $HOME/.ssh && umask 0077 \
	&& git config --global url."git@github.com:".insteadOf https://github.com/ \
	&& ssh-keyscan github.com >> $HOME/.ssh/known_hosts

ENV GOBIN=/bin
ENV GOPRIVATE=github.com/aserto-dev
ENV ROOT_DIR=/src

# generate & build
ARG VERSION
ARG COMMIT

COPY go.mod go.sum Depfile mage.go magefiles/* ./
RUN --mount=type=cache,target=/go/pkg/mod \
		--mount=type=cache,target=/root/.cache/go-build \
		--mount=type=ssh \
		go run mage.go deps

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    	--mount=type=cache,target=/root/.cache/go-build \
    	--mount=type=ssh \
    	go run mage.go build

FROM alpine
ARG VERSION
ARG COMMIT

LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.source=https://github.com/aserto-dev/topaz-private
LABEL org.opencontainers.image.title="Topaz"
LABEL org.opencontainers.image.revision=$COMMIT
LABEL org.opencontainers.image.url=https://aserto.com

RUN apk add --no-cache bash tzdata
WORKDIR /app
COPY --from=build-dev /src/dist/topaz_linux_amd64/topaz /app/

EXPOSE 8282
EXPOSE 8383
EXPOSE 8484
EXPOSE 8585

ENTRYPOINT ["./topaz"]
