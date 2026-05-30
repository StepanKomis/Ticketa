build:
	go build -o ./build/ticketa ./src/cmd/main.go

run-local:
	./build/ticketa

docker-build:
	docker buildx build .

docker-up:
	docker compose up -d