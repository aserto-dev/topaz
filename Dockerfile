FROM alpine

RUN apk add --no-cache bash tzdata

EXPOSE 8282
EXPOSE 8383
EXPOSE 8484
EXPOSE 8585
EXPOSE 9292
EXPOSE 8080

WORKDIR /app

COPY dist/topazd_linux_amd64_v1/topazd /app/

ENTRYPOINT ["./topazd"]
