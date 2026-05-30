build:
	go build -o ./build/ticketa ./src/cmd/main.go

run-local:
	docker compose up -d database && ./build/ticketa

docker-build:
	docker buildx build .

deploy:
	docker compose up -d