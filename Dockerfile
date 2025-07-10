FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/vk-segments ./cmd/app/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/vk-segments /app/vk-segments

CMD ["/app/vk-segments"]