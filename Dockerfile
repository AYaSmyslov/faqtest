# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux COARCH=amd64 go build -o /faq-api ./cmd/api

# Runtime stage
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /faq-api /app/faq-api

ENV HTTP_ADDR=:8080

EXPOSE 8080

CMD ["/app/faq-api"]
