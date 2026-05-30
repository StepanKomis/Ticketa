FROM golang:1.26.3-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o /ticketa ./src/cmd

# ── Runtime image ────────────────────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /ticketa .

EXPOSE 8080

ENTRYPOINT ["/app/ticketa"]
CMD ./app/ticketa