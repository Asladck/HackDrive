FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# если main.go лежит в cmd/api/
RUN go build -ldflags="-w -s" -o out ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/out .

CMD ["./out"]
