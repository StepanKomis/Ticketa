CLIENT_REPO        := https://github.com/StepanKomis/Ticketa-client.git
CLIENT_TMP         := /tmp/ticketa-client
STATIC_DIR         := ./src/www/static
DOCS_DIR           := ./src/www/docs
SWAGGER_UI_VERSION := 5.18.2
SWAGGER_CDN        := https://cdn.jsdelivr.net/npm/swagger-ui-dist@$(SWAGGER_UI_VERSION)

.PHONY: test build build-frontend build-full run-local docker-build docker-build-nc deploy sqlc clean swag swagger-ui

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

# Download pinned Swagger UI dist assets into docs embed directory
swagger-ui:
	mkdir -p $(DOCS_DIR)
	for f in swagger-ui.css swagger-ui-bundle.js swagger-ui-standalone-preset.js; do \
		curl -sL $(SWAGGER_CDN)/$$f -o $(DOCS_DIR)/$$f; \
	done

# Regenerate swagger.yaml from swag annotations in handler source files.
# Vyžaduje: go install github.com/swaggo/swag/cmd/swag@latest
swag:
	swag init \
		--generalInfo cmd/main.go \
		--dir ./src \
		--output $(DOCS_DIR) \
		--outputTypes yaml \
		--parseInternal

# Remove local build artifacts and static embed directory
clean:
	rm -rf ./build $(STATIC_DIR)
