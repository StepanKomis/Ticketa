# Ticketa

Helpdesk pro školy. Studenti a zaměstnanci zakládají tikety, údržbáři je řeší a správci spravují celý systém za běhu bez restartu serveru.

## Architektura

Monorepo se dvěma hlavními částmi:

| Adresář | Technologie | Popis |
|---|---|---|
| `src/` | Go 1.25 | REST API server, databáze, konfigurace |
| `client/` | React (CRA), TypeScript | SPA frontend |

Go binárka obsluhuje API i zkompilovaný React frontend — frontend je vložen přes `//go:embed` přímo do binárky. Výsledná Docker image je postavená na `scratch`.

```
Fáze Docker buildu
  1. frontend-builder  — node:22-alpine   — npm install + npm run build
  2. builder           — golang:1.25-alpine — go:embed frontendu, go build
  3. runtime           — scratch           — pouze staticky linkovaná binárka
```

Databáze je PostgreSQL. Migrace se spustí automaticky při startu — žádný externí nástroj není potřeba.

## Role uživatelů

| Role | Popis |
|---|---|
| `student` | Vytváří a spravuje vlastní tikety, hlasuje |
| `staff` | Totéž co student + schvaluje/zamítá žádosti o vysokou prioritu a nové uživatele |
| `maintainer` | Řeší tikety, přijímá přiřazení, přistupuje ke všem tiketům |
| `admin` | Plný přístup — správa uživatelů, stavů tiketů, pozvánky, runtime konfigurace |
| `pending` | Čeká na schválení (přechodný stav po registraci pro staff/maintainer) |

## API

Živá dokumentace (Swagger UI) běží na `/docs` hned po spuštění serveru. Níže je přehled skupin endpointů.

Autentizace funguje přes HTTP-only cookie `session_token` (TTL 7 dní). Všechny chyby mají jednotný tvar:

```json
{ "code": 404, "status": "Not Found", "msg": "tiket nenalezen" }
```

### Skupiny endpointů

| Skupina | Middleware | Endpointy |
|---|---|---|
| **Veřejné** | — | `POST /api/register`, `POST /api/login`, `POST /api/auth/invite/accept`, `GET /api/setup-status` |
| **Přihlášený uživatel** | `auth` | `/api/me`, `PATCH /api/me/password`, `PATCH /api/me/email`, `POST /api/logout` |
| **Tikety** | `auth + mustChangePw` | CRUD na `/api/tickets/{id}`, hlasování, history, claim |
| **Komentáře** | `auth + mustChangePw` | CRUD na `/api/tickets/{id}/comments`, `/api/comments/{id}` |
| **Priorita** | `staff nebo admin` | `POST /api/tickets/{id}/approve-priority`, `reject-priority` |
| **Stavy** | `auth + mustChangePw` | `GET /api/ticket-statuses` |
| **Activity log** | `admin` / `auth` | `GET /api/activity`, `GET /api/users/{id}/activity` |
| **Admin** | `admin` | Konfigurace, stavy tiketů, správa uživatelů, pozvánky |

Middleware `mustChangePw` blokuje uživatele s příznakem `must_change_pw = true` — takový uživatel musí nejprve změnit heslo přes `/api/me/password`.

## Konfigurace

### `.env` — infrastruktura a přístupy

```shell
cp .env.example .env
```

| Proměnná | Výchozí | Popis |
|---|---|---|
| `PG_HOST` | `database` | Hostname Postgres |
| `PG_PORT` | `5432` | Port Postgres |
| `PG_USER` | — | Uživatel Postgres **(povinné)** |
| `PG_PASSWORD` | — | Heslo Postgres **(povinné)** |
| `PG_DATABASE` | `ticketa` | Název databáze |
| `PG_SSLMODE` | `disable` | SSL režim (`disable`, `require`, `verify-full`) |
| `SERVER_PORT` | `8080` | HTTP port serveru |
| `LOG_LEVEL` | — | Přepíše úroveň logování z YAML (`debug`, `info`) |
| `COOKIE_SECURE` | `false` | Nastavit `true` při nasazení za HTTPS proxy |

### `config/ticketa.yaml` — runtime nastavení

Logování a stavy tiketů. Soubor je volume-mountován z hostu — změny provedené přes admin API jsou zapsány zpět do souboru a přežijí restart kontejneru.

```shell
cp config/ticketa.yaml.example config/ticketa.yaml
```

```yaml
logging:
  level: info          # debug | info | warn | error
  dir: /var/log/ticketa

ticket_statuses:
  - title: "Otevřeno"
    color: "#3498db"
  - title: "Probíhá"
    color: "#f39c12"
  - title: "Vyřešeno"
    color: "#2ecc71"
    is_closed: true    # tikety v tomto stavu jsou považovány za uzavřené
```

`ticket_statuses` je synchronizován obousměrně — DB je autoritativní zdroj, YAML je persistentní záloha pro případ přepsání kontejneru.

## Nasazení

```shell
git clone https://github.com/StepanKomis/Ticketa.git
cd Ticketa
cp .env.example .env
cp config/ticketa.yaml.example config/ticketa.yaml
# doplňte hodnoty v .env
make docker-build
make deploy
```

Aplikace: `http://localhost:8080`  
API dokumentace: `http://localhost:8080/docs`

## Vývoj

Potřebujete: Go 1.25+, Node.js 22+, Docker, `sqlc` (pro změny SQL dotazů).

```shell
# 1. spustit jen databázi
docker compose up -d database

# 2a. sestavit Go binárku bez frontendu (API je plně funkční)
make build
./build/ticketa

# 2b. sestavit kompletně (frontend + backend)
make build-full
./build/ticketa

# frontend samostatně (pro react dev server)
cd client && npm install && npm start
```

Při vývoji frontendu proxy v `client/package.json` přeposílá `/api` požadavky na Go server — stačí spustit obojí paralelně.

### Make cíle

| Cíl | Popis |
|---|---|
| `build` | Zkompiluje Go binárku do `./build/ticketa` |
| `build-frontend` | `npm install + npm run build` v `client/`, zkopíruje výstup do embed adresáře |
| `build-full` | `build-frontend` + `build` |
| `run-local` | Spustí databázi přes Docker a pak lokální binárku |
| `docker-build` | Sestaví Docker image (s cache) |
| `docker-build-nc` | Sestaví Docker image bez cache |
| `deploy` | `docker compose up -d` |
| `test` | `go test ./...` |
| `sqlc` | Regeneruje Go kód z SQL dotazů v `src/database/postgres/queries/` |
| `swag` | Regeneruje `swagger.yaml` ze swag anotací v handlerech |
| `swagger-ui` | Stáhne Swagger UI assets do `src/www/docs/` |
| `clean` | Smaže `./build` a embed adresář frontendu |

### Změny SQL

Po úpravě libovolného `.sql` souboru v `src/database/postgres/queries/` je nutné regenerovat Go kód:

```shell
make sqlc
```

Nové migrace patří do `src/database/postgres/migrations/up/` a musí být zaregistrovány v `migrations.go`.

## Databáze

Migrace jsou vloženy do binárky (`//go:embed`) a spustí se automaticky při startu. Aktuální schéma (14 migrací):

| Tabulka | Popis |
|---|---|
| `users` | Účty s rolí (`student`, `staff`, `maintainer`, `admin`, `pending`), příznakem aktivity a schvalovatelem |
| `local_login` | Lokální přihlášení — bcrypt hash hesla, příznak `must_change_pw` |
| `ldap_login` | LDAP/AD přihlášení — DN uživatele |
| `sessions` | HTTP-only cookie sessions (TTL 7 dní, soft-delete při odhlášení/deaktivaci) |
| `ticket_statuses` | Seřazené stavy tiketů synchronizované s YAML konfigurací; příznak `is_closed` |
| `tickets` | Tikety — titulek, tělo, autor, priorita, stav, hlasy, řešitel, `is_closed`, `resolution_note` |
| `ticket_votes` | Hlasy uživatelů na tiketech (unikátní per user+ticket) |
| `ticket_comments` | Komentáře k tiketům, podpora odpovědí (`parent_id`), soft-delete |
| `ticket_history` | Auditní log změn tiketu (změna stavu, priority, přiřazení…) |
| `invitations` | Pozvánkové tokeny (e-mail + role + expirace) |
| `activity_log` | Systemový audit log — actor, typ události, payload JSONB |

## Struktura projektu

```
client/                          React frontend (CRA + TypeScript)
  src/
    components/                  Znovupoužitelné UI komponenty
      console/                   Tikety — TicketCard, FilterBar, ActivityFeed, StatusBadge…
      layout/                    Sidebar, BottomNav, AppShell
      tickets/                   NewTicketModal, PriorityBadge
      auth/                      Formuláře přihlášení a registrace
      admin/                     Admin panely (uživatelé, pozvánky, konfigurace)
    pages/                       Stránky (consolePage, ticketDetailPage, usersPage…)
    hooks/                       React Query hooky pro API volání
    utils/                       Sdílené utility (relativeTime, labels, avatar…)
    types/                       TypeScript typy

config/
  ticketa.yaml.example           Šablona runtime konfigurace

src/
  cmd/
    main.go                      Vstupní bod — načte config, spustí server
    server/
      env/                       Pomocné funkce pro čtení env proměnných
      logs/                      File logger s rotací (lumberjack)
      startup/                   Připojení k DB, migrace, start HTTP serveru
  config/                        YAML typy, loader, atomický writer, thread-safe Store (RWMutex)
  database/
    postgres/
      connection.go              Sestavení DSN a otevření *sql.DB
      migrations/                Vložené SQL migrace (UP_000N.sql) + runner
      queries/                   sqlc-generované dotazy (.sql zdrojové + .sql.go výstup)
  internal/
    activity/                    ActivityLogger — duální zápis do DB a JSONL souboru
    API/users/                   Login a registrační logika (oddělena od handlerů)
    ctxkeys/                     Sdílené klíče kontextu (zabraňuje cyklickým importům)
    security/                    SessionStore, generování tokenů, bcrypt, cookie helpers
  www/
    docs/                        Swagger UI assety + swagger.yaml (generováno přes make swag)
    midleware/                   AuthMiddleware, AdminMiddleware, MustChangePwMiddleware
    router/
      handlers/                  UserHandler, TicketHandler, CommentHandler,
                                 AdminHandler, ActivityHandler, StaticHandler, DocsHandler
      router.go                  Registrace routes a sestavení middleware chain
    embed.go                     //go:embed pro frontend assety
    docs_embed.go                //go:embed pro Swagger UI assety
```

## Čtení kódu

**Kde začít:** `src/www/router/router.go` — vidíte všechny routes a jaký middleware je chrání.

**Request flow:** HTTP požadavek → `router.go` → middleware chain (auth → mustChangePw → role) → handler → sqlc dotaz → PostgreSQL.

**Handler pattern:** Každý handler dostává přes dependency injection `*db.Queries` (sqlc), `*logs.Logger` a `*activity.ActivityLogger`. Handlery nesdílí stav — jsou bezpečné pro souběžný přístup.

**Konfigurace za běhu:** `config.Store` je chráněn `RWMutex`. Změny přes admin API se atomicky zapíší do YAML souboru a okamžitě se projeví — server není nutné restartovat.

**Migrace:** Každá migrace je samostatný SQL soubor vložený přes `//go:embed`. Nová migrace = nový soubor `UP_000N.sql` + záznam v `migrations.go`. Runner spustí jen migrace s vyšším číslem než je aktuální verze v DB.

**Frontend → Backend:** React Query hooky v `client/src/hooks/` volají API. Při lokálním vývoji proxy (`client/package.json`) přeposílá `/api` na Go server. V produkci Go binárka servíruje obojí.
