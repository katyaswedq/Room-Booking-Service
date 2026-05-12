FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/app ./cmd/app

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/bin/app ./app

EXPOSE 8080

CMD ["./app"]