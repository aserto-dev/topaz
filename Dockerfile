FROM alpine

RUN apk add --no-cache bash tzdata

WORKDIR /app

COPY dist/topaz*_linux_amd64_v1/topaz* /app/

ENTRYPOINT ["./topazd"]
