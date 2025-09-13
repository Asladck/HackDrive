FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-w -s" -o out ./cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/out .

CMD ["./out"]
