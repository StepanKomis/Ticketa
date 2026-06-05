FROM node:22-alpine AS frontend-builder
RUN apk add --no-cache git
WORKDIR /client
RUN git clone https://github.com/StepanKomis/Ticketa-client.git .
RUN npm install
RUN npm run build

FROM golang:1.25.0-alpine AS builder
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /client/build ./src/www/static
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /ticketa ./src/cmd \
    && chmod +x /ticketa \
    && mkdir -p /var/log/ticketa

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /ticketa /ticketa
COPY --from=builder /var/log/ticketa /var/log/ticketa
EXPOSE 8080
ENTRYPOINT ["/ticketa"]
