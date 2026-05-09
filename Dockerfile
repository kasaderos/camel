# syntax=docker/dockerfile:1.7

FROM golang:1.26 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/portfolio-manager ./cmd/portfolio-manager

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/portfolio-manager /usr/local/bin/portfolio-manager
COPY --from=builder /src/portfolios /app/portfolios
COPY --from=builder /src/migrations /app/migrations

ENTRYPOINT ["/usr/local/bin/portfolio-manager"]