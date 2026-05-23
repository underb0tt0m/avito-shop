# docker build -t avito-shop -f cmd/Dockerfile .
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o main ./cmd/main.go

FROM scratch

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080
EXPOSE 9200

CMD ["./main"]