build:
	go build -o ./build/ticketa ./src/cmd/main.go

run-local:
	docker compose up -d database && ./build/ticketa

docker-build:
	docker buildx build -t ticketa:0.0.1 .

deploy:
	docker compose up -d