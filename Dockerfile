FROM alpine

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG BUILDPLATFORM

RUN echo "BUILDPLATFORM=$BUILDPLATFORM" \
 && echo "TARGETPLATFORM=$TARGETPLATFORM" \
 && echo "TARGETOS=$TARGETOS" \
 && echo "TARGETARCH=$TARGETARCH"

RUN apk add --no-cache bash tzdata ca-certificates

RUN mkdir /config && \
    mkdir /certs && \
    mkdir /db && \
    mkdir /decisions

VOLUME ["/config", "/certs", "/db", "/decisions"]

WORKDIR /app

COPY \
${TARGETPLATFORM}/topaz \
${TARGETPLATFORM}/topazd \
${TARGETPLATFORM}/topaz-backup \
/app/

ENTRYPOINT ["./topazd"]
CMD ["run", "-c", "/config/config.yaml"]
