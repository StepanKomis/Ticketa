# Ticketa
Ticketing system built for schools

## Architecture

Ticketa is a single binary that serves both the API and the compiled React frontend. The frontend ([Ticketa-client](https://github.com/StepanKomis/Ticketa-client)) is cloned and built inside the Docker image — no pre-built assets are committed here. The final image is based on `scratch` and contains only the binary.

```
Docker build stages
  1. frontend-builder  — node:22-alpine, clones Ticketa-client, runs npm build
  2. builder           — golang:1.24.4-alpine, embeds compiled frontend via go:embed, builds binary
  3. runtime           — scratch, contains only the binary
```

## Deployment

```shell
git clone https://github.com/StepanKomis/Ticketa.git
cd Ticketa
cp .env.example .env
# edit .env with your values
make docker-build
make deploy
```

The app is available at `http://localhost:8080` once the containers are healthy.

## Environment variables

| Variable      | Default    | Description                        |
| ------------- | ---------- | ---------------------------------- |
| `PG_HOST`     | `database` | Postgres hostname                  |
| `PG_PORT`     | `5432`     | Postgres port                      |
| `PG_USER`     | —          | Postgres user (required)           |
| `PG_PASSWORD` | —          | Postgres password (required)       |
| `PG_DATABASE` | `ticketa`  | Postgres database name             |
| `SERVER_PORT` | `8080`     | HTTP port the server listens on    |
| `LOG_DIR`     | `/var/log/ticketa` | Directory for log files    |
| `LOG_LEVEL`   | `info`     | Log level (`info` or `debug`)      |

## Make targets

| Target           | Description                                                       |
| ---------------- | ----------------------------------------------------------------- |
| `build`          | Compile Go binary to `./build/ticketa`                            |
| `build-frontend` | Clone Ticketa-client, build it, copy `dist/` into embed directory |
| `build-full`     | `build-frontend` + `build` (full local build with frontend)       |
| `run-local`      | Start database via Docker, run local binary                       |
| `docker-build`   | Build Docker image (with layer cache)                             |
| `docker-build-nc`| Build Docker image without cache                                  |
| `deploy`         | Start all services via `docker compose up -d`                     |
| `test`           | Run Go test suite                                                 |
| `sqlc`           | Regenerate database query code via sqlc                           |
| `clean`          | Remove `./build` and the frontend embed directory                 |
