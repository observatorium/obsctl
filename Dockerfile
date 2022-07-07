FROM golang:1.18-alpine3.16 as builder

RUN apk add ca-certificates --no-cache make git && update-ca-certificates

WORKDIR /workspace

COPY . .

RUN make build

FROM alpine:3.16.0

COPY --from=builder /workspace/obsctl /usr/bin/obsctl
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/