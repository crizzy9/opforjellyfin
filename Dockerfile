FROM golang:1.24.4-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o opfor .

FROM alpine:latest

RUN apk --no-cache add ca-certificates git

WORKDIR /root/

COPY --from=builder /app/opfor .

RUN mkdir -p /data /config

ENV HOME=/config

VOLUME ["/data", "/config"]

ENTRYPOINT ["./opfor"]
CMD ["--help"]
