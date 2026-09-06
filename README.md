# HifzApp API

Backend REST API for **HifzApp**, a Quran memorization (Hifz) application. Handles Google-powered
authentication, per-surah/ayah memorization progress, practice sessions, daily goals, and dashboard
aggregation.

- **Base URL:** `/api/v1`
- **Tech stack:** Go 1.25 · [Fiber](https://gofiber.dev) v3 · GORM + PostgreSQL · JWT (HS256) · Google OAuth ID tokens
- **Deployment:** standalone server locally (`:8080`) and serverless on Vercel (`api/index.go`)

> 📖 **Full API reference:** see [`docs/API.md`](docs/API.md)

---

## Getting started

### 1. Environment

Copy `.env.example` to `.env` and fill it in:

```bash
cp .env.example .env
```

Required: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`,
`GOOGLE_CLIENT_ID`, and `JWT_SECRET` (**≥ 32 chars**). Set `APP_ENV=production` in production.

### 2. Run locally

```bash
cd server
go run main.go        # serves http://localhost:8080
```

On startup the app connects to PostgreSQL and **auto-migrates** the `User`, `HifzProgress`,
`DailyProgress`, and `HifzSession` tables.

### 3. Deploy to Vercel

`vercel.json` routes all traffic to `api/index.go` via the `@vercel/go` build.

---

## Authentication

- **Login:** `POST /api/v1/auth/google` with a Google ID token → returns a JWT and sets a `token`
  cookie (HttpOnly).
- **Authenticate:** send `Authorization: Bearer <JWT>` **or** the `token` cookie on protected routes.
- **Logout:** `POST /api/v1/auth/logout` invalidates all sessions (bumps `session_version`) and
  clears the cookie.

---

## Endpoints at a glance

All protected routes are under `/api/v1/app`.

| Method   | Path                                  | Auth | Description                      |
| -------- | ------------------------------------- | ---- | -------------------------------- |
| `POST`   | `/api/v1/auth/google`                 | 🔓    | Google login                     |
| `POST`   | `/api/v1/auth/logout`                 | 🔒    | Logout                           |
| `GET`    | `/api/v1/app/me/`                     | 🔒    | Get my profile                   |
| `POST`   | `/api/v1/app/user/onboarding`         | 🔒    | Complete onboarding              |
| `PATCH`  | `/api/v1/app/user/name`               | 🔒    | Update name                      |
| `GET/POST`     | `/api/v1/app/hifz/progress/`    | 🔒    | List / create progress           |
| `GET/PATCH/DELETE` | `/api/v1/app/hifz/progress/:id` | 🔒    | Get / update / delete progress   |
| `GET/POST`     | `/api/v1/app/sessions/`          | 🔒    | List / create session            |
| `GET/PATCH/DELETE` | `/api/v1/app/sessions/:id`   | 🔒    | Get / update / delete session    |
| `POST`   | `/api/v1/app/sessions/:id/complete`   | 🔒    | Complete session                 |
| `GET/POST`     | `/api/v1/app/daily-progress/`   | 🔒    | List / create daily progress     |
| `GET/PATCH/DELETE` | `/api/v1/app/daily-progress/:date` | 🔒 | Get / update / delete by date    |
| `GET`    | `/api/v1/app/dashboard`               | 🔒    | Dashboard summary                |

🔓 = public · 🔒 = requires authentication.

---

## Project layout

```
api/
├── api/index.go        # Vercel serverless entrypoint
├── config/db.go        # DB connection + auto-migration
├── handlers/           # Request handlers (auth, me, progress, sessions, …)
├── middleware/         # Protected() JWT guard
├── models/             # GORM models (User, HifzProgress, DailyProgress, HifzSession)
├── routes/             # Route registration per module
├── server/main.go      # Standalone server (port 8080)
└── utils/              # Shared helpers (GetUserID)
```

---

## Documentation

- [`docs/API.md`](docs/API.md) — full reference: setup, authentication, every endpoint (request/response
  schemas, validation, error codes), data models, and conventions.