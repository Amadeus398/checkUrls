FROM golang:1.15
WORKDIR /CheckUrls
COPY . /CheckUrls
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o checkUrls CheckUrls/cmd

FROM alpine:3.11
WORKDIR /CheckUrls
COPY --from=0 /CheckUrls/checkUrls /bin
ENV SERVERADDRESS=:8080 LOGLEVEL=loglevel HOST=127.0.0.1 PORT=5432 USER=user PASSWORD=password NAME=dbname SSLMODE=disable
ENTRYPOINT ["checkUrls", "server"]
