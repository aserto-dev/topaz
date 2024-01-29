FROM alpine

RUN apk add --no-cache bash tzdata

RUN mkdir /data && \
    mkdir /config && \
    mkdir /certs && \
    mkdir /db && \
    mkdir /decisions
VOLUME ["/data", "/config", "/certs", "/db", "/decisions"]

WORKDIR /app

COPY dist/topaz*_linux_amd64_v1/topaz* /app/

ENV TOPAZ_DIR=/config

ENTRYPOINT ["sh", "-c", "./topazd"]
CMD ["run", "-c", "${TOPAZ_DIR}/config.yaml"]
