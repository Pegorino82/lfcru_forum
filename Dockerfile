FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /forum ./cmd/forum

# Runtime image
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /forum /app/forum
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/templates /app/templates

EXPOSE 8080

CMD ["/app/forum"]
