# SETP1
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o pod-creator-demo ./main.go

# SETP2
FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/pod-creator .

EXPOSE 8080

ENTRYPOINT ["./pod-creator"]