FROM golang:1.23-alpine AS base

RUN apk add --no-cache \
    ca-certificates \
    git \
    wget \
    curl

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

FROM base AS builder
ARG SERVICE_NAME
RUN echo "Building service: $SERVICE_NAME"
RUN if [ "$SERVICE_NAME" = "api-gateway" ]; then \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./api-gateway/; \
    else \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/${SERVICE_NAME}/; \
    fi

FROM alpine:latest AS final

RUN apk --no-cache add ca-certificates curl
RUN wget -O /usr/local/bin/grpc_health_probe \
    https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.19/grpc_health_probe-linux-amd64 && \
    chmod +x /usr/local/bin/grpc_health_probe

WORKDIR /root/

COPY --from=builder /app/main .

CMD ["./main"]
