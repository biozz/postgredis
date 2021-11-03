FROM golang:1.17.2-alpine AS builder
ENV GO111MODULE=on \
    CGO_ENABLED=0
WORKDIR /src
ADD . .
RUN go build -o ./bin/postgredis

FROM alpine:3.14
COPY --from=builder /src/bin/postgredis /app/postgredis
WORKDIR /app
ENTRYPOINT ["./postgredis"]
