CLIENT_REPO := https://github.com/StepanKomis/Ticketa-client.git
CLIENT_TMP  := /tmp/ticketa-client
STATIC_DIR  := ./src/www/static

.PHONY: test build build-frontend build-full run-local docker-build docker-build-nc deploy sqlc clean

test:
	go test ./...

build:
	go build -o ./build/ticketa ./src/cmd/main.go

# Clone client repo, build it, copy dist into embed directory
build-frontend:
	rm -rf $(CLIENT_TMP)
	git clone $(CLIENT_REPO) $(CLIENT_TMP)
	npm --prefix $(CLIENT_TMP) install
	npm --prefix $(CLIENT_TMP) run build
	rm -rf $(STATIC_DIR)
	cp -r $(CLIENT_TMP)/build $(STATIC_DIR)
	rm -rf $(CLIENT_TMP)

# Full local build: frontend + Go binary
build-full: build-frontend build

# Start database in background, then run the binary locally
run-local:
	docker compose up -d database && ./build/ticketa

# Docker build (uses cached layers)
docker-build:
	docker buildx build -t ticketa:latest .

# Docker build without cache
docker-build-nc:
	docker buildx build --no-cache -t ticketa:latest .

# Run full stack via docker compose
deploy:
	docker compose up -d

# Generate sqlc code from queries
sqlc:
	sqlc generate

# Remove local build artifacts and static embed directory
clean:
	rm -rf ./build $(STATIC_DIR)
