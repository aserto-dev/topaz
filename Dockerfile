FROM alpine

RUN apk add --no-cache bash tzdata

RUN mkdir /config
COPY config/config.yaml /config/

WORKDIR /app

COPY dist/topaz*_linux_amd64_v1/topaz* /app/

ENTRYPOINT ["./topazd"]
CMD ["run", "-c", "/config/config.yaml"]
