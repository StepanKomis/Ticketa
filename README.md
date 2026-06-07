# Ticketa

Ticketing system built for schools. Staff and students submit tickets; maintainers manage them and configure the deployment at runtime.

## Architecture

Ticketa is a single binary that serves both the REST API and the compiled React frontend. The frontend ([Ticketa-client](https://github.com/StepanKomis/Ticketa-client)) is cloned and built inside the Docker image — no pre-built assets are committed here. The final image is based on `scratch` and contains only the binary.

```
Docker build stages
  1. frontend-builder  — node:22-alpine  — clones Ticketa-client, runs npm build
  2. builder           — golang:1.24-alpine — embeds compiled frontend via //go:embed, builds binary
  3. runtime           — scratch — contains only the statically linked binary
```

The database is PostgreSQL. Migrations run automatically at startup; no external migration tool is required.

## User roles

| Role | Description |
|---|---|
| `student` | Can create and manage their own tickets |
| `staff` | Can create and manage their own tickets |
| `maintainer` | Full access — manages all tickets, users, statuses, and runtime config |

## API

### Public

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/register` | Register a new account |
| `POST` | `/api/login` | Authenticate; sets an HTTP-only session cookie |

### Authenticated (any active user)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/tickets` | Create a ticket |
| `GET` | `/api/tickets` | List all tickets |
| `GET` | `/api/tickets/{id}` | Get a single ticket |
| `PUT` | `/api/tickets/{id}` | Update a ticket (author only) |
| `DELETE` | `/api/tickets/{id}` | Delete a ticket (author only) |

### Admin (maintainer only)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/admin/config` | Get the current runtime config |
| `PATCH` | `/api/admin/config` | Update runtime config (persisted to YAML immediately) |
| `GET` | `/api/admin/ticket-statuses` | List ticket statuses |
| `POST` | `/api/admin/ticket-statuses` | Create a ticket status |
| `PUT` | `/api/admin/ticket-statuses/{id}` | Update a ticket status |
| `DELETE` | `/api/admin/ticket-statuses/{id}` | Delete a ticket status |
| `GET` | `/api/admin/users` | List all users |
| `GET` | `/api/admin/users/{id}` | Get a single user |
| `PATCH` | `/api/admin/users/{id}` | Change a user's active state or type |

Authentication is cookie-based (`ticketa_session`). All authenticated and admin routes return `401` without a valid session cookie and `403` when the role requirement is not met.

## Configuration

Configuration is split into two layers:

### `.env` — secrets and infrastructure

Copy `.env.example` and fill in your values:

```shell
cp .env.example .env
```

| Variable | Default | Description |
|---|---|---|
| `PG_HOST` | `database` | Postgres hostname |
| `PG_PORT` | `5432` | Postgres port |
| `PG_USER` | — | Postgres user (required) |
| `PG_PASSWORD` | — | Postgres password (required) |
| `PG_DATABASE` | `ticketa` | Postgres database name |
| `SERVER_PORT` | `8080` | HTTP port the server listens on |
| `LOG_LEVEL` | `info` | Overrides the log level from YAML (`info` or `debug`) |

### `config/ticketa.yaml` — runtime config

Controls logging and ticket statuses. The file is volume-mounted from the host so changes written via the admin API persist through container restarts.

Copy the example and adjust as needed:

```shell
cp config/ticketa.yaml.example config/ticketa.yaml
```

```yaml
logging:
  level: info          # debug | info
  dir: /var/log/ticketa

ticket_statuses:
  - title: "Otevřeno"
    color: "#3498db"
  - title: "Probíhá"
    color: "#f39c12"
  - title: "Vyřešeno"
    color: "#2ecc71"
```

Rules for `ticket_statuses`:
- Minimum three statuses required
- First status = open state, last status = resolved state, any middle statuses = in-progress states
- Order in the array determines the `position` stored in the database
- Changes made via `PATCH /api/admin/ticket-statuses` are written back to this file automatically

## Deployment

```shell
git clone https://github.com/StepanKomis/Ticketa.git
cd Ticketa
cp .env.example .env
cp config/ticketa.yaml.example config/ticketa.yaml
# edit both files with your values
make docker-build
make deploy
```

The app is available at `http://localhost:8080` once the containers are healthy.

## Development

Requirements: Go 1.24+, Docker, `sqlc` (for query regeneration).

```shell
# Start only the database
docker compose up -d database

# Build the Go binary and run it locally (frontend not embedded)
make build
./build/ticketa

# Or build frontend + binary together
make build-full
./build/ticketa
```

## Make targets

| Target | Description |
|---|---|
| `build` | Compile Go binary to `./build/ticketa` |
| `build-frontend` | Clone Ticketa-client, build it, copy `dist/` into the embed directory |
| `build-full` | `build-frontend` + `build` — full local build including the frontend |
| `run-local` | Start database via Docker, then run the local binary |
| `docker-build` | Build Docker image (with layer cache) |
| `docker-build-nc` | Build Docker image without cache |
| `deploy` | Start all services via `docker compose up -d` |
| `test` | Run the Go test suite |
| `sqlc` | Regenerate database query code via sqlc |
| `clean` | Remove `./build` and the frontend embed directory |

## Database

Migrations are embedded in the binary and run automatically on startup. The schema includes:

- `users` — accounts with role (`student`, `staff`, `maintainer`) and active flag
- `sessions` — HTTP-only cookie sessions with expiry
- `ticket_statuses` — ordered list of statuses (position-aware, synced with YAML config)
- `tickets` — support tickets with title, body, author reference, and optional status

## Project structure

```
config/
  ticketa.yaml.example   runtime config template
src/
  cmd/
    main.go              entrypoint — loads config, starts server
    server/
      logs/              structured file logger
      startup/           server bootstrap (DB connect, migrate, listen)
  config/                YAML config types, loader, atomic writer, thread-safe Store
  database/postgres/
    migrations/          embedded SQL migrations (UP_000N.sql)
    queries/             sqlc-generated type-safe query functions
  internal/
    ctxkeys/             shared context key types (avoids import cycles)
    security/            session store, token generation, cookie helpers
  www/
    midleware/           AuthMiddleware, MaintainerMiddleware
    router/
      handlers/          UserHandler, TicketHandler, AdminHandler, StaticHandler
      router.go          route registration
    embed.go             //go:embed for static frontend assets
```
