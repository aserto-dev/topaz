FROM alpine

RUN apk add --no-cache bash tzdata

RUN mkdir /config && \
    mkdir /certs && \
    mkdir /db && \
    mkdir /decisions
VOLUME ["/config", "/certs", "/db", "/decisions"]

WORKDIR /app

COPY dist/topaz*_linux_amd64_v1/topaz* /app/

ENTRYPOINT ["./topazd", "run", "-c"]
CMD ["/config/config.yaml"]
