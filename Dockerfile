FROM golang:1.23.3-alpine3.20 AS builder

WORKDIR /app

RUN apk add --no-cache netcat-openbsd=1.226-r0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/auction cmd/auction/main.go

FROM scratch
WORKDIR /app
COPY --from=builder /app/auction /app/auction
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/app/auction"]
