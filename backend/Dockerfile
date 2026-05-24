FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
ENV GOFLAGS=-mod=mod
RUN go mod download || true

COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
