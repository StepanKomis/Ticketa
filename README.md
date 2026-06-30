# Ticketa

Helpdesk pro školy. Studenti zakládají tikety, učitelé a správci je prioritizují, údržbáři řeší. Správci spravují celý systém za běhu bez restartu serveru.

### Role

| Role | Popis |
|---|---|
| `student` | Zakládá tikety, hlasuje o prioritě |
| `staff` | Schvaluje urgentní priority a nové uživatele |
| `maintainer` | Přijímá přiřazení, řeší tikety |
| `admin` | Plný přístup — uživatelé, stavy, pozvánky, konfigurace |

Nový uživatel žádající o roli `staff` nebo `maintainer` čeká ve stavu `pending` na schválení.

---

## Nasazení

```shell
git clone https://github.com/StepanKomis/Ticketa.git
cd Ticketa
cp .env.example .env          # doplňte PG_USER a PG_PASSWORD
cp config/ticketa.yaml.example config/ticketa.yaml
make docker-build
make deploy
```

Aplikace běží na `http://localhost:8080`, API dokumentace (Swagger) na `http://localhost:8080/docs`.

Migrace proběhnou automaticky při prvním spuštění.

### Konfigurace

`.env` — přístupy k databázi a port serveru (`SERVER_PORT`, `COOKIE_SECURE` pro HTTPS).

`config/ticketa.yaml` — runtime nastavení logování a stavů tiketů. Soubor je volume-mountován; změny přes admin UI jsou zapsány zpět a přežijí restart.

---

## Vývoj

Potřebujete: Go 1.25+, Node.js 22+, Docker, `sqlc`.

```shell
# databáze
docker compose up -d database

# backend (API bez frontendu)
make build && ./build/ticketa

# frontend (React dev server s proxy na Go)
cd client && npm install && npm start
```

Po změně SQL dotazů: `make sqlc`. Nová migrace = soubor `UP_000N.sql` v `src/database/postgres/migrations/up/` + záznam v `migrations.go`.

---

## Roadmap

### Hotovo
- Tikety — CRUD, priority, hlasování, kategorie, lokace, history
- Komentáře s odpověďmi a soft-delete
- Role a schvalovací workflow (`pending` → `staff` / `maintainer`)
- In-app notifikace (vyřešení, smazání, přiřazení, schválení role/priority)
- Soft-delete tiketů (přístupné přes přímou URL, viditelné pro staff/admin)
- Admin panel — uživatelé, pozvánky, stavy tiketů, runtime konfigurace
- Activity log
- Docker nasazení na `scratch` image (~15 MB)

### Plánováno
- [ ] Reporty a statistiky (počty tiketů, průměrná doba řešení, vytížení údržbářů)
- [ ] Adresář uživatelů pro staff
- [ ] E-mailové notifikace
- [ ] LDAP/AD přihlášení (schéma připraveno)
- [ ] Vícejazyčnost (i18n)
