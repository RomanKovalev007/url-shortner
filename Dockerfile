FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/


FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./server"]
