build:
	go build ./src/cmd/main.go

docker-build:
	docker buildx build .

docker-up:
	docker compose up -d