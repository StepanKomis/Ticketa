# Ticketa

Helpdesk pro školy. Studenti a zaměstnanci zakládají tikety, udržovatelé je řeší a celý systém spravují za běhu bez restartu.

## Architektura

Jedna Go binárka — obsluhuje REST API i zkompilovaný React frontend. Frontend ([Ticketa-client](https://github.com/StepanKomis/Ticketa-client)) se klonuje a builduje přímo v Docker image, žádné předbuilděné assety se do tohoto repozitáře necommitují. Výsledná image je postavená na `scratch` a obsahuje jen samotnou binárku.

```
Fáze Docker buildu
  1. frontend-builder  — node:22-alpine  — klonuje Ticketa-client, spustí npm build
  2. builder           — golang:1.24-alpine — vloží frontend přes //go:embed, sestaví binárku
  3. runtime           — scratch — obsahuje pouze staticky linkovanou binárku
```

Databáze je PostgreSQL. Migrace se spustí automaticky při startu — žádný externí nástroj není potřeba.

## Role uživatelů

| Role | Popis |
|---|---|
| `student` | Vytváří a spravuje vlastní tikety |
| `staff` | Vytváří a spravuje vlastní tikety |
| `maintainer` | Plný přístup — tikety, uživatelé, stavy, runtime konfigurace |

## API

Živá dokumentace běží na `/docs` hned po spuštění serveru.

Přihlašování funguje přes cookie (`session_token`, HTTP-only). Chráněné endpointy vrátí `401` bez platného cookie, `403` při nedostatečné roli. Všechny chyby mají stejný tvar:

```json
{ "code": 404, "status": "Not Found", "msg": "tiket nenalezen" }
```

### Veřejné

#### `POST /api/register`

Vytvoří lokální účet.

**Tělo požadavku**

| Pole | Typ | Povinné | Poznámky |
|---|---|---|---|
| `email` | string | ano | Platná e-mailová adresa |
| `password` | string | ano | Min. 8 znaků, musí obsahovat velké písmeno, číslici a speciální znak |
| `user_type` | string | ano | `student`, `staff` nebo `maintainer` |
| `first_name` | string | ne | |
| `last_name` | string | ne | |

**Odpovědi**

| Status | Popis |
|---|---|
| `201` | `{ "id": 42 }` |
| `400` | Chybí pole, slabé heslo nebo neznámý typ uživatele |
| `500` | Interní chyba (např. duplicitní e-mail) |

---

#### `POST /api/login`

Přihlásí uživatele a nastaví session cookie.

**Tělo požadavku**

| Pole | Typ | Povinné |
|---|---|---|
| `email` | string | ano |
| `password` | string | ano |

**Odpovědi**

| Status | Popis |
|---|---|
| `200` | Nastaví HTTP-only cookie `session_token` platný 7 dní |
| `400` | Neplatné tělo požadavku |
| `401` | Špatné přihlašovací údaje nebo neaktivní účet |

---

### Přihlášení (libovolný aktivní uživatel)

Vyžaduje platný cookie `session_token`.

#### `POST /api/tickets` — Vytvoření tiketu

**Tělo požadavku**

| Pole | Typ | Povinné | Poznámky |
|---|---|---|---|
| `title` | string | ano | Krátký souhrn |
| `body` | string | ano | Úplný popis |
| `status_id` | integer | ne | ID počátečního stavu |

**Odpovědi:** `201` tiket · `400` chybné tělo · `401` chybí session · `422` chybí title nebo body · `500`

---

#### `GET /api/tickets` — Seznam tiketů

Vrátí všechny tikety od nejnovějšího. Prázdný seznam vrátí `[]`.

**Odpovědi:** `200` pole tiketů · `401` · `500`

---

#### `GET /api/tickets/{id}` — Detail tiketu

**Odpovědi:** `200` tiket · `400` chybné ID · `401` · `404` · `500`

---

#### `PUT /api/tickets/{id}` — Úprava tiketu

Upravovat může jen autor. Všechna pole jsou volitelná.

**Tělo požadavku:** `title`, `body`, `status_id` (vše volitelné)

**Odpovědi:** `200` tiket · `400` · `401` · `403` nejste autor · `404` · `500`

---

#### `DELETE /api/tickets/{id}` — Smazání tiketu

Smazat může jen autor.

**Odpovědi:** `204` · `400` · `401` · `403` nejste autor · `404` · `500`

---

### Admin (jen maintainer)

Vyžaduje platný cookie `session_token` **a** `user_type = maintainer`.

#### `GET /api/admin/config` — Aktuální konfigurace

**Odpovědi:** `200` konfigurace · `401` · `403`

---

#### `PATCH /api/admin/config` — Změna konfigurace

Změny se zapíší atomicky do `/config/ticketa.yaml` a projeví se okamžitě. Pokud posíláte `ticket_statuses`, musí mít aspoň 3 položky.

**Tělo požadavku (vše volitelné)**

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

**Odpovědi:** `200` konfigurace · `400` chybné tělo nebo chyba zápisu · `401` · `403`

---

#### `GET /api/admin/ticket-statuses` — Seznam stavů

Seřazeno podle pozice. Prázdný seznam vrátí `[]`.

**Odpovědi:** `200` pole · `401` · `403` · `500`

---

#### `POST /api/admin/ticket-statuses` — Nový stav

Nový stav se automaticky přidá i do YAML konfigurace.

**Tělo požadavku**

| Pole | Typ | Povinné | Poznámky |
|---|---|---|---|
| `title` | string | ano | |
| `color` | string | ne | HEX formát, např. `#9b59b6`. Výchozí `#808080` |
| `position` | integer | ne | Musí být unikátní |

**Odpovědi:** `201` stav · `400` · `401` · `403` · `422` chybí title · `500`

---

#### `PUT /api/admin/ticket-statuses/{id}` — Úprava stavu

Změna se synchronizuje do YAML konfigurace.

**Tělo požadavku:** `title`, `color` (obojí volitelné)

**Odpovědi:** `200` stav · `400` · `401` · `403` · `404` · `500`

---

#### `DELETE /api/admin/ticket-statuses/{id}` — Smazání stavu

Tikety s tímto stavem budou mít `status_id` nastaveno na `null`. YAML se synchronizuje automaticky.

**Odpovědi:** `204` · `400` · `401` · `403` · `500`

---

#### `GET /api/admin/users` — Seznam uživatelů

**Odpovědi:** `200` pole uživatelů · `401` · `403` · `500`

---

#### `GET /api/admin/users/{id}` — Detail uživatele

**Odpovědi:** `200` uživatel · `400` · `401` · `403` · `404` · `500`

---

#### `PATCH /api/admin/users/{id}` — Úprava uživatele

**Tělo požadavku (vše volitelné)**

| Pole | Typ | Poznámky |
|---|---|---|
| `is_active` | boolean | Neaktivní uživatelé se nepřihlásí |
| `user_type` | string | `student`, `staff` nebo `maintainer` |

**Odpovědi:** `200` uživatel · `400` · `401` · `403` · `404` · `500`

---

### Datové struktury

**Tiket**
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

**Stav tiketu**
```json
{ "ID": 1, "Title": "Probíhá", "Color": "#f39c12", "Position": 1 }
```

**Uživatel**
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

**Konfigurace**
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

## Konfigurace

Konfigurace jsou ve dvou souborech:

### `.env` — přístupy a infrastruktura

```shell
cp .env.example .env
```

| Proměnná | Výchozí | Popis |
|---|---|---|
| `PG_HOST` | `database` | Hostname Postgres |
| `PG_PORT` | `5432` | Port Postgres |
| `PG_USER` | — | Uživatel Postgres (povinné) |
| `PG_PASSWORD` | — | Heslo Postgres (povinné) |
| `PG_DATABASE` | `ticketa` | Název databáze |
| `SERVER_PORT` | `8080` | HTTP port serveru |
| `LOG_LEVEL` | `info` | Přepíše úroveň logování z YAML (`info` nebo `debug`) |

### `config/ticketa.yaml` — runtime nastavení

Logování a stavy tiketů. Soubor je volume-mountován z hostu, takže změny provedené přes admin API přežijí restart kontejneru.

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

Pár pravidel pro `ticket_statuses`:
- Minimum jsou tři stavy
- První = otevřeno, poslední = vyřešeno, prostřední = cokoliv mezitím
- Pořadí v poli se uloží jako `position` v databázi
- Změny přes `PATCH /api/admin/config` se automaticky zapíší zpět do souboru

## Nasazení

```shell
git clone https://github.com/StepanKomis/Ticketa.git
cd Ticketa
cp .env.example .env
cp config/ticketa.yaml.example config/ticketa.yaml
# doplňte hodnoty v obou souborech
make docker-build
make deploy
```

Aplikace běží na `http://localhost:8080`, API dokumentace je na `http://localhost:8080/docs`.

## Vývoj

Potřebujete: Go 1.24+, Docker, `sqlc` (pro regeneraci dotazů).

```shell
# jen databáze
docker compose up -d database

# sestavit a spustit lokálně (bez frontendu)
make build
./build/ticketa

# kompletní sestavení včetně frontendu
make build-full
./build/ticketa
```

## Make cíle

| Cíl | Popis |
|---|---|
| `build` | Zkompiluje Go binárku do `./build/ticketa` |
| `build-frontend` | Klonuje Ticketa-client, sestaví ho a zkopíruje `dist/` do embed adresáře |
| `build-full` | `build-frontend` + `build` |
| `run-local` | Spustí databázi přes Docker a pak lokální binárku |
| `docker-build` | Sestaví Docker image (s cache) |
| `docker-build-nc` | Sestaví Docker image bez cache |
| `deploy` | `docker compose up -d` |
| `test` | Spustí testy |
| `sqlc` | Regeneruje kód dotazů přes sqlc |
| `swagger-ui` | Stáhne Swagger UI assety do `src/www/docs/` |
| `swag` | Regeneruje `swagger.yaml` ze swag anotací v handlerech |
| `clean` | Smaže `./build` a embed adresář frontendu |

## Databáze

Migrace jsou součástí binárky a spustí se automaticky. Schéma:

- `users` — účty s rolí (`student`, `staff`, `maintainer`) a příznakem aktivity
- `sessions` — HTTP-only cookie sessions s expirací
- `ticket_statuses` — seřazený seznam stavů synchronizovaný s YAML konfigurací
- `tickets` — tikety s nadpisem, popisem, autorem a volitelným stavem

## Struktura projektu

```
config/
  ticketa.yaml.example     šablona runtime konfigurace
src/
  cmd/
    main.go                vstupní bod — načte konfiguraci, spustí server
    server/
      logs/                file logger
      startup/             připojení k DB, migrace, start HTTP serveru
  config/                  YAML typy, loader, atomický writer, thread-safe Store
  database/postgres/
    migrations/            vložené SQL migrace (UP_000N.sql)
    queries/               sqlc-generované dotazy
  internal/
    ctxkeys/               sdílené klíče kontextu (zamezuje cyklickým importům)
    security/              session store, generování tokenů, cookie helpers
  www/
    docs/                  Swagger UI assety + swagger.yaml (generováno přes make swag)
    midleware/             AuthMiddleware, MaintainerMiddleware
    router/
      handlers/            UserHandler, TicketHandler, AdminHandler, StaticHandler, DocsHandler
      router.go            registrace routes
    embed.go               //go:embed pro frontend assety
    docs_embed.go          //go:embed pro Swagger UI assety
```
