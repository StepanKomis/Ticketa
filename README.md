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

Interactive documentation is served at `/docs` once the server is running.

Authentication is cookie-based (`session_token`, HTTP-only). Authenticated and admin routes return `401` without a valid session cookie and `403` when the role requirement is not met. All error responses share the same shape:

```json
{ "code": 404, "status": "Not Found", "msg": "ticket not found" }
```

### Public

No authentication required.

#### `POST /api/register`

Create a local account.

**Request body**

| Field | Type | Required | Notes |
|---|---|---|---|
| `email` | string | yes | Valid e-mail address |
| `password` | string | yes | Min. 8 chars, must include uppercase, digit, and special character |
| `user_type` | string | yes | `student`, `staff`, or `maintainer` |
| `first_name` | string | no | |
| `last_name` | string | no | |

**Responses**

| Status | Description |
|---|---|
| `201` | `{ "id": 42 }` |
| `400` | Missing field, weak password, or unknown user type |
| `500` | Internal error (e.g. duplicate e-mail) |

---

#### `POST /api/login`

Authenticate and receive a session cookie.

**Request body**

| Field | Type | Required |
|---|---|---|
| `email` | string | yes |
| `password` | string | yes |

**Responses**

| Status | Description |
|---|---|
| `200` | Sets `session_token` HTTP-only cookie (7-day TTL) |
| `400` | Invalid request body |
| `401` | Wrong credentials or inactive account |

---

### Authenticated (any active user)

Requires a valid `session_token` cookie.

#### `POST /api/tickets` — Create ticket

**Request body**

| Field | Type | Required | Notes |
|---|---|---|---|
| `title` | string | yes | Short summary |
| `body` | string | yes | Full description |
| `status_id` | integer | no | ID of the initial status |

**Responses:** `201` ticket object · `400` bad body · `401` no session · `422` missing title or body · `500`

---

#### `GET /api/tickets` — List tickets

Returns all tickets ordered newest first. Empty result returns `[]`.

**Responses:** `200` array of ticket objects · `401` · `500`

---

#### `GET /api/tickets/{id}` — Get ticket

**Responses:** `200` ticket object · `400` bad ID · `401` · `404` not found · `500`

---

#### `PUT /api/tickets/{id}` — Update ticket

Only the author can update. All fields are optional in the body.

**Request body:** `title`, `body`, `status_id` (all optional)

**Responses:** `200` updated ticket · `400` · `401` · `403` not the author · `404` · `500`

---

#### `DELETE /api/tickets/{id}` — Delete ticket

Only the author can delete.

**Responses:** `204` deleted · `400` · `401` · `403` not the author · `404` · `500`

---

### Admin (maintainer only)

Requires a valid `session_token` cookie **and** `user_type = maintainer`.

#### `GET /api/admin/config` — Get runtime config

**Responses:** `200` config object · `401` · `403`

---

#### `PATCH /api/admin/config` — Update runtime config

Changes are written atomically to `/config/ticketa.yaml` on the host and take effect immediately without a restart. Providing `ticket_statuses` requires at least 3 entries.

**Request body (all optional)**

```json
{
  "logging": { "level": "debug", "dir": "/var/log/ticketa" },
  "ticket_statuses": [
    { "title": "Otevřeno", "color": "#3498db" },
    { "title": "Probíhá",  "color": "#f39c12" },
    { "title": "Vyřešeno", "color": "#2ecc71" }
  ]
}
```

**Responses:** `200` updated config · `400` bad body or disk write failure · `401` · `403`

---

#### `GET /api/admin/ticket-statuses` — List statuses

Returns statuses ordered by position. Empty result returns `[]`.

**Responses:** `200` array · `401` · `403` · `500`

---

#### `POST /api/admin/ticket-statuses` — Create status

Also appends the new status to the YAML config.

**Request body**

| Field | Type | Required | Notes |
|---|---|---|---|
| `title` | string | yes | |
| `color` | string | no | HEX format, e.g. `#9b59b6`. Defaults to `#808080` |
| `position` | integer | no | Must be unique |

**Responses:** `201` status object · `400` · `401` · `403` · `422` missing title · `500`

---

#### `PUT /api/admin/ticket-statuses/{id}` — Update status

Syncs change to YAML config.

**Request body:** `title`, `color` (both optional)

**Responses:** `200` updated status · `400` · `401` · `403` · `404` · `500`

---

#### `DELETE /api/admin/ticket-statuses/{id}` — Delete status

Tickets referencing this status will have `status_id` set to `null`. YAML config is synced.

**Responses:** `204` · `400` · `401` · `403` · `500`

---

#### `GET /api/admin/users` — List users

**Responses:** `200` array of user objects · `401` · `403` · `500`

---

#### `GET /api/admin/users/{id}` — Get user

**Responses:** `200` user object · `400` · `401` · `403` · `404` · `500`

---

#### `PATCH /api/admin/users/{id}` — Update user

**Request body (all optional)**

| Field | Type | Notes |
|---|---|---|
| `is_active` | boolean | Inactive users cannot log in |
| `user_type` | string | `student`, `staff`, or `maintainer` |

**Responses:** `200` updated user · `400` · `401` · `403` · `404` · `500`

---

### Data shapes

**Ticket**
```json
{
  "ID": 1,
  "Title": "Nemohu se přihlásit",
  "Body": "Po zadání hesla se nic nestane.",
  "CreatedAt": "2026-06-07T14:22:55Z",
  "AuthorID": 3,
  "StatusID": { "Int32": 0, "Valid": false }
}
```

**TicketStatus**
```json
{ "ID": 1, "Title": "Probíhá", "Color": "#f39c12", "Position": 1 }
```

**User**
```json
{
  "ID": 3,
  "Email": "jan.novak@skola.cz",
  "FirstName": { "String": "Jan", "Valid": true },
  "LastName":  { "String": "Novák", "Valid": true },
  "UserType": "student",
  "Provider": "local",
  "IsActive": true,
  "CreatedAt": "2026-06-07T12:00:00Z",
  "LastLoginAt": { "Time": "0001-01-01T00:00:00Z", "Valid": false }
}
```

**Config**
```json
{
  "Logging": { "Level": "info", "Dir": "/var/log/ticketa" },
  "TicketStatuses": [
    { "Title": "Otevřeno", "Color": "#3498db" },
    { "Title": "Probíhá",  "Color": "#f39c12" },
    { "Title": "Vyřešeno", "Color": "#2ecc71" }
  ]
}
```

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
- First = open state, last = resolved state, any middle = in-progress
- Order in the array determines the `position` stored in the database
- Changes via `PATCH /api/admin/config` are written back to this file automatically

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

The app is available at `http://localhost:8080` once the containers are healthy. Interactive API docs are at `http://localhost:8080/docs`.

## Development

Requirements: Go 1.24+, Docker, `sqlc` (for query regeneration).

```shell
# Start only the database
docker compose up -d database

# Build and run locally (frontend not embedded)
make build
./build/ticketa

# Full local build including frontend
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
| `swagger-ui` | Download pinned Swagger UI dist assets into `src/www/docs/` |
| `docs` | Validate `openapi.yaml` is well-formed YAML |
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
  ticketa.yaml.example     runtime config template
src/
  cmd/
    main.go                entrypoint — loads config, starts server
    server/
      logs/                structured file logger
      startup/             server bootstrap (DB connect, migrate, listen)
  config/                  YAML config types, loader, atomic writer, thread-safe Store
  database/postgres/
    migrations/            embedded SQL migrations (UP_000N.sql)
    queries/               sqlc-generated type-safe query functions
  internal/
    ctxkeys/               shared context key types (avoids import cycles)
    security/              session store, token generation, cookie helpers
  www/
    docs/                  embedded Swagger UI assets + openapi.yaml
    midleware/             AuthMiddleware, MaintainerMiddleware
    router/
      handlers/            UserHandler, TicketHandler, AdminHandler, StaticHandler, DocsHandler
      router.go            route registration
    embed.go               //go:embed for static frontend assets
    docs_embed.go          //go:embed for Swagger UI assets
```
