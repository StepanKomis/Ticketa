# Ticketa Client

React frontend for [Ticketa](https://github.com/StepanKomis/Ticketa) — a ticketing system built for schools. Students raise requests, staff triage and resolve them, and maintainers manage users.

## Features

- **Console dashboard** — greeting, status summary cards (Otevřené / Řeší se / Vyřešené), the signed-in user's ticket list with status filters, a report-a-problem CTA, and a recent-activity feed derived from real ticket data.
- **Ticket list** — all tickets with the same filtering and quick-actions, plus a new-ticket button and mobile FAB.
- **Ticket detail** — full description, status/priority badges, metadata, an activity timeline, and a comment composer. Staff can change status and use quick actions (Vyřešit / Uzavřít tiket).
- **New ticket flow** — a modal (opened from the header, the CTA, or the mobile FAB) that creates a ticket through the API with client-side validation.
- **Admin** — a maintainer-only profile settings page and a user-management table (role changes, approve pending users, activate/deactivate).
- **Auth** — login and registration wired to the backend, with an HttpOnly session cookie and route guards that react to server-side session revocation.

The UI is in Czech and responsive down to mobile (sidebar collapses to a top bar + bottom navigation).

## Design system

The interface follows a shared set of CSS custom properties defined in `src/index.css` — a green brand scale, ink/neutral palette, functional status colours (new, open, in progress, resolved, closed, urgent), a deterministic avatar palette, spacing and radius scales, and shadows. Typography pairs **Hanken Grotesk** for UI with **Geist Mono** for ticket IDs. Colour is reserved for status and priority; everything else stays neutral.

## Tech stack

- React 19 + React Router 7
- TanStack Query 5 for server state
- TypeScript, Create React App (`react-scripts`)
- Plain CSS with custom properties (no UI framework)

## Architecture

This repository is the frontend half of the Ticketa stack. In production it is not deployed independently — the server's Docker build clones this repo, runs `npm run build`, and embeds the compiled output directly into the Go binary via `go:embed`. See the [server repo](https://github.com/StepanKomis/Ticketa) for deployment instructions.

## Local development

```shell
npm install
npm start
```

The app runs at `http://localhost:3000`. Requests to `/api/*` are proxied to the backend at `http://localhost:8080` (configured via the `proxy` field in `package.json`), so start a Ticketa backend on that port for live data. Without a backend the UI still renders and degrades gracefully (loading and error states).

## API

The client integrates with the following backend endpoints. Errors are returned as `{ "code", "status", "msg" }` JSON and surfaced to the user.

| Method & path                       | Used for                                              |
| ----------------------------------- | ----------------------------------------------------- |
| `POST /api/login`                   | Authenticate; sets the HttpOnly `session_token` cookie |
| `POST /api/register`                | Create an account                                     |
| `GET /api/tickets`                  | List tickets                                          |
| `POST /api/tickets`                 | Create a ticket                                       |
| `GET /api/tickets/{id}`             | Ticket detail                                         |
| `PUT /api/tickets/{id}`             | Update a ticket (full replace — see note)             |
| `DELETE /api/tickets/{id}`          | Delete a ticket                                       |
| `GET /api/tickets/{id}/comments`    | Ticket comments *(pending backend support)*           |
| `POST /api/tickets/{id}/comments`   | Add a comment *(pending backend support)*             |
| `GET /api/admin/ticket-statuses`    | Configured statuses (staff/maintainer)                |
| `GET /api/admin/config`             | Server configuration                                  |
| `GET /api/admin/users`              | List users (maintainer)                               |
| `PATCH /api/admin/users/{id}`       | Update a user's name, role, or active state           |

### POST /api/login

```json
{ "email": "john.doe@example.com", "password": "securedPassword123." }
```

| Field      | Type   | Notes                         |
| ---------- | ------ | ----------------------------- |
| `email`    | string | Must be a valid email address |
| `password` | string |                               |

Returns `200 OK` and sets the session cookie on success.

### POST /api/register

```json
{
  "email": "john.doe@example.com",
  "password": "securedPassword123.",
  "first_name": "John",
  "last_name": "Doe",
  "user_type": "staff"
}
```

| Field        | Type   | Notes                                      |
| ------------ | ------ | ------------------------------------------ |
| `email`      | string | Must be a valid email address              |
| `password`   | string | 8–72 characters, ≥1 digit, ≥1 special char |
| `first_name` | string |                                            |
| `last_name`  | string |                                            |
| `user_type`  | string | `student`, `staff`, or `maintainer`        |

Returns `201 Created` with the new user ID on success.

### Notes on backend behaviour

- **`PUT /api/tickets/{id}` is a full replace, not a patch.** Status changes always send the current `title` and `body` alongside the new `status_id`, so other fields are not blanked.
- **Statuses are configurable**; the client maps server statuses to its UI status set by title (with a position-based fallback) rather than hardcoding IDs. A ticket with no status is treated as newly created.
- **Comments and a `/api/me` endpoint are not yet implemented.** Comment views degrade to a friendly "not available yet" notice, and the signed-in role defaults to `student` until `/api/me` exists — so the maintainer-only admin pages are reachable once the backend can report a maintainer role.

## Scripts

| Command         | Description                   |
| --------------- | ----------------------------- |
| `npm start`     | Run dev server on port 3000   |
| `npm run build` | Production build to `./build` |
| `npm test`      | Run the test suite            |

## Testing

Tests use Jest and React Testing Library (via `react-scripts test`). Run the full suite once with:

```shell
CI=true npm test
```

Coverage spans the console and tickets pages, ticket detail and actions, the new-ticket modal, the admin pages and route guard, and the shared console components.
