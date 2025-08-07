FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go test -v ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/shortener-service ./cmd/shortener-service

FROM scratch


COPY --from=builder /app/bin/shortener-service /shortener-service
COPY --from=builder /app/config/config.yaml /config.yaml

ENTRYPOINT ["/shortener-service"]