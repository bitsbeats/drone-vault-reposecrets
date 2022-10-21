FROM golang:1.19-alpine as builder

RUN true \
    && apk add --no-cache ca-certificates

ADD . /app
WORKDIR /app

ENV CGO_ENABLED=0

RUN go build -o drone-vault-reposecrets .

# ---

FROM scratch
COPY --from=builder /app/drone-vault-reposecrets /plugin
ENTRYPOINT ["/plugin"]
