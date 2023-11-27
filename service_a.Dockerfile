# Build Phase
FROM golang:alpine AS builder

ARG VERSION=1.0.0
ENV VERSION=$VERSION

RUN apk update && apk add gcc make librdkafka-dev openssl-libs-static zlib-static zstd-libs libsasl lz4-dev lz4-static zstd-static libc-dev musl-dev

WORKDIR /app
COPY . /app
ENV GO111MODULE=on
ENV BLOG_CONFIG_PATH=config/blog/prod.yaml
RUN make build_blog

# Execution Phase
FROM alpine:latest

RUN apk --no-cache add ca-certificates \
	&& addgroup -S app \
	&& adduser -S app -G app

WORKDIR /app
COPY --from=builder /app .
RUN chmod -R 777 /app
USER app

# Expose port to the outside world
EXPOSE 8260

# Command to run the executable
CMD ["./service_a"]
