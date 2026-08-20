FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /app/bin/short-link \
    ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && \
    adduser -S -G app app

WORKDIR /app

COPY --from=builder /app/bin/short-link /app/short-link

RUN chown app:app /app/short-link

USER app

EXPOSE 8080

ENTRYPOINT [ "/app/short-link" ]