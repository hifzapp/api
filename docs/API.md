# HifzApp API — Documentation

Backend REST API for **HifzApp**, a Quran memorization (Hifz) application. The API handles user
authentication via Google, tracks memorization progress per surah/ayah, records practice sessions,
and aggregates daily progress and dashboard statistics.

- **Base URL:** `/api/v1`
- **Local server:** `http://localhost:8080`
- **Production (Vercel):** handled by the `api/index.go` serverless entrypoint
- **Format:** JSON only

---

## 1. Tech Stack

| Layer      | Technology                                                          |
| ---------- | ------------------------------------------------------------------- |
| Language   | Go 1.25                                                            |
| Web server | [Fiber](https://gofiber.dev) v3 (`github.com/gofiber/fiber/v3`)     |
| ORM        | GORM v1 + `gorm.io/driver/postgres` (jackc/pgx)                     |
| Database   | PostgreSQL (connection via DSN string)                              |
| Auth       | JWT (HS256, `golang-jwt/jwt/v5`) + Google ID token (`idtoken`)      |
| Config     | `github.com/joho/godotenv` for `.env` loading                       |
| Deploy     | Vercel (`@vercel/go`) for serverless; standalone binary otherwise   |

There are two runnable entrypoints:

- **`api/index.go`** — Vercel serverless handler (`Handler` function), declared in `vercel.json`.
- **`server/main.go`** — standalone local/server process that listens on `:8080`.

Both build the **same** Fiber app (CORS → route registration → middleware) and only differ in how
they are invoked and which CORS origin they allow (localhost:3000 locally, the Netlify frontend in prod).

---

## 2. Project Structure

```
api/
├── api/
│   └── index.go          # Vercel serverless entrypoint (Handler)
├── config/
│   └── db.go             # DB connection + auto-migration (Config.Db(), Config.DB)
├── handlers/             # Request handlers (one file per resource/concern)
│   ├── auth.go           # Google login, logout, JWT issuing
│   ├── me.go             # GET /me
│   ├── name.go           # Update name
│   ├── onboarding.go     # Complete onboarding
│   ├── hifz_progress.go  # Surah/ayah progress CRUD
│   ├── hifz_session.go   # Practice session CRUD + complete
│   ├── daily_progress.go # Daily progress CRUD
│   └── dashboard.go      # Dashboard aggregation
├── middleware/
│   └── middleware.go     # Protected() auth guard (JWT/cookie)
├── models/               # GORM models (auto-migrated on startup)
│   ├── user.go
│   ├── progress.go
│   ├── session.go
│   └── dailyProgress.go
├── routes/               # Route registration per module (Router groups)
│   ├── auth/  me/  user/  hifz/  sessions/  daily/  dashboard/
├── server/
│   └── main.go           # Standalone server entrypoint (port 8080)
└── utils/
    └── utils.go          # GetUserID(context) helper used by handlers
```

Route registration happens in `api/index.go` / `server/main.go`:

- `POST /api/v1/auth/*` — public (protected only for logout)
- Everything else is grouped under `/api/v1/app` and guarded by `middleware.Protected()`.

---
## 3. Setup & Running

### 3.1 Environment variables

Copy `.env.example` to `.env` and fill in the values. The API reads the following variables:

| Variable              | Required | Description                                                        |
| --------------------- | -------- | ------------------------------------------------------------------ |
| `DB_HOST`             | yes      | PostgreSQL host                                                    |
| `DB_PORT`             | yes      | PostgreSQL port                                                    |
| `DB_USER`             | yes      | Database user                                                      |
| `DB_PASSWORD`         | yes      | Database password                                                  |
| `DB_NAME`             | yes      | Database name                                                      |
| `DB_SSLMODE`          | yes      | SSL mode for the connection (e.g. `require`, `disable`)            |
| `PORT`                | no       | Port for the standalone server (default `8080` in `server/main.go`)| 
| `GOOGLE_CLIENT_ID`    | yes      | Google OAuth client ID used to verify ID tokens                    |
| `JWT_SECRET`          | yes      | Secret used to sign JWTs (must be **≥ 32 characters**)             |
| `GOOGLE_CLIENT_SECRET`| no*      | Provided in `.env.example` (reserved for OAuth flows)              |
| `APP_ENV`             | no       | `development` or `production` (drives cookie `Secure`/`SameSite`)  |
| `FRONTEND_URL`        | no*      | Provided in `.env.example` (CORS origin reference)                 |

> **\*** `GOOGLE_CLIENT_SECRET` and `FRONTEND_URL` are listed in `.env.example` for completeness; the
> current code obtains CORS origins from `api/index.go` / `server/main.go` and does not yet read them.

Startup **fails fast** (panic / `log.Fatal`) if `JWT_SECRET` or `GOOGLE_CLIENT_ID` are not configured.

### 3.2 Local development

```bash
# from the repo root
cd server
go run main.go        # starts a Fiber server on :8080
```

On startup the app:

1. Loads `.env` (godotenv),
2. Connects to PostgreSQL using a DSN built from the `DB_*` variables,
3. Tunes the pool (`max_open_conns = 5`, `max_idle_conns = 2`),
4. **Auto-migrates** the `User`, `HifzProgress`, `DailyProgress`, and `HifzSession` tables,
5. Logs `Successfully connected to the database.`

### 3.3 Deployment (Vercel)

`vercel.json` routes all traffic (`/(.*)`) to `api/index.go` using the `@vercel/go` build. The
serverless `Handler` reuses the same Fiber app factory (`getApp()`), initializing it once per
process via a `sync.Once` guard.

---

## 4. Authentication

HifzApp uses **Google OAuth2 ID token** login, then issues a **signed JWT** (HS256) for subsequent
requests. Login also sets an `HttpOnly` cookie named `token`.

### 4.1 Token issuance

- Issued by `POST /api/v1/auth/google` after the Google ID token is validated with the configured
  `GOOGLE_CLIENT_ID`.
- **Claims:** custom `user_id` + `session_version`, plus standard registered claims
  (`iss = "hifz-api"`, `aud = "hifz-client"`, `exp` 72 h, `sub` = Google ID, `jti` random 32-byte hex).
- The token is returned in the response body **and** set as an `HttpOnly` cookie.

### 4.2 Sending the token

Protected routes accept the token in **either** of these ways (cookie checked first):

```http
Authorization: Bearer <JWT>
```

or the `token` HttpOnly cookie.

### 4.3 Session invalidation

- **Logout** (`POST /app/auth/logout`) increments the user's `session_version`, invalidating all
  previously issued tokens across **all** devices, then clears the cookie.
- The `Protected()` middleware rejects a token when its `session_version` no longer matches the
  user's current `session_version`.

### 4.4 Cookie flags

| Env        | `Secure` | `SameSite` |
| ---------- | -------- | ---------- |
| production | `true`   | `None`     |
| otherwise  | `false`  | `Lax`      |

---

## 5. API Reference

### Legend
- 🔓 **Public** — no authentication required.
- 🔒 **Protected** — requires a valid JWT (cookie `token` or `Authorization: Bearer`).

All protected routes live under the `/api/v1/app` group.

---

### 5.1 Authentication

#### `POST /api/v1/auth/google` 🔓 — Google login

Validates a Google ID token, creates/fetches the user, and returns a session JWT.

**Request body**
| Field   | Type   | Required | Description            |
| ------- | ------ | -------- | ---------------------- |
| `token` | string | yes      | Google ID token (OAuth)|

**Example request**
```http
POST /api/v1/auth/google
Content-Type: application/json

{ "token": "<GOOGLE_ID_TOKEN>" }
```

**Response `200 OK`**
```json
{
  "message": "Logged in successfully.",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "Aisha",
    "avatar": "https://...",
    "hearts": 5,
    "streak": 0
  },
  "token": "<JWT>"
}
```

**Notes:** Also sets a `token` cookie. The Google email must be verified. Errors return `400`
(invalid body), `401` (invalid/expired token or unverified email), or `500` (config/service failure).

---

#### `POST /api/v1/auth/logout` 🔒 — Logout

Invalidates the current user's session on all devices and clears the token cookie.

**Example request**
```http
POST /api/v1/auth/logout
Authorization: Bearer <JWT>
```

**Response `200 OK`**
```json
{ "message": "Logged out successfully." }
```

---

### 5.2 Current user

#### `GET /api/v1/app/me/` 🔒 — Get my profile

Returns the authenticated user's profile.

**Example request**
```http
GET /api/v1/app/me/
Authorization: Bearer <JWT>
```

**Response `200 OK`**
```json
{
  "user": {
    "id": 1,
    "name": "Aisha",
    "avatar": "https://...",
    "daily_goal": 5,
    "onboarded": true,
    "hearts": 5,
    "streak": 2,
    "created_at": "2026-01-01T10:00:00Z"
  }
}
```

---

### 5.3 User settings

#### `POST /api/v1/app/user/onboarding` 🔒 — Complete onboarding

Sets the daily goal and marks the user as onboarded.

**Request body**
| Field        | Type | Required | Validation          |
| ------------ | ---- | -------- | ------------------- |
| `daily_goal` | int  | yes      | between `1` and `1000` |

**Response `200 OK`**
```json
{
  "message": "Onboarding completed successfully.",
  "daily_goal": 10,
  "onboarded": true
}
```

---

#### `PATCH /api/v1/app/user/name` 🔒 — Update display name

**Request body**
| Field  | Type   | Required | Validation                  |
| ------ | ------ | -------- | --------------------------- |
| `name` | string | yes      | `1`–`14` characters (UTF-8) |

**Response `200 OK`**
```json
{ "message": "Name updated successfully.", "name": "Aisha" }
```

---
### 5.4 Hifz progress

Common validation for body fields in this group:

| Field             | Type | Validation                            |
| ----------------- | ---- | ------------------------------------- |
| `surah_number`    | int  | between `1` and `114`                  |
| `start_ayah`      | int  | ≥ `1`                                  |
| `end_ayah`        | int  | ≥ `start_ayah`                         |
| `mastery_level`   | int  | between `1` and `5`                    |
| `repetitions`     | int  | ≥ `0`                                  |

A yah range (`surah_number`, `start_ayah`–`end_ayah`) may **not overlap** an existing progress
record for the same user (`409 Conflict`).

---

#### `GET /api/v1/app/hifz/progress/` 🔒 — List my progress

Returns all of the user's progress records, ordered by `surah_number ASC, start_ayah ASC`.

**Response `200 OK`**
```json
{
  "progress": [
    {
      "id": 1,
      "user_id": 1,
      "surah_number": 1,
      "start_ayah": 1,
      "end_ayah": 7,
      "mastery_level": 2,
      "repetitions": 3,
      "last_reviewed_at": "2026-01-02T08:00:00Z",
      "created_at": "2026-01-01T09:00:00Z",
      "updated_at": "2026-01-02T08:00:00Z"
    }
  ]
}
```

---

#### `POST /api/v1/app/hifz/progress/` 🔒 — Create progress

**Request body** (all fields required)
```json
{
  "surah_number": 1,
  "start_ayah": 8,
  "end_ayah": 14,
  "mastery_level": 1,
  "repetitions": 0
}
```

**Response `201 Created`**
```json
{ "message": "Hifz progress created successfully.", "progress": { "...": "..." } }
```

**Errors:** `400` invalid body/validation, `409` overlapping range, `500` failure.

---

#### `GET /api/v1/app/hifz/progress/:id` 🔒 — Get one progress record

Path param `id` (positive integer). Returns the record **only if it belongs to the user**.

**Response `200 OK`** — `{ "progress": { "...": "..." } }`  ·  `404` if not found.

---

#### `PATCH /api/v1/app/hifz/progress/:id` 🔒 — Update progress

Path param `id`; body uses the same fields as create. If the `surah_number`/ayah range changes, it
is checked against overlapping records (excluding itself).

**Response `200 OK`**
```json
{ "message": "Hifz progress updated successfully.", "progress": { "...": "..." } }
```

---

#### `DELETE /api/v1/app/hifz/progress/:id` 🔒 — Delete progress

Path param `id`. Response `200 OK`:
```json
{ "message": "Hifz progress deleted successfully." }
```
`404` if not found.

---
### 5.5 Hifz sessions

`session_type` must be one of: `memorization`, `review` (case-insensitive, trimmed/forced lowercase).

---

#### `GET /api/v1/app/sessions/` 🔒 — List my sessions (paginated)

**Query params**

| Param   | Type | Default | Validation        |
| ------- | ---- | ------- | ----------------- |
| `page`  | int  | `1`     | ≥ `1`             |
| `limit` | int  | `10`    | `1`–`100`         |

Sessions are ordered by `created_at DESC, id DESC`.

**Response `200 OK`**
```json
{
  "sessions": [ { "...": "..." } ],
  "meta": { "total": 37, "page": 1, "limit": 10, "total_pages": 4 }
}
```

---

#### `POST /api/v1/app/sessions/` 🔒 — Create a session

**Request body**
| Field          | Type   | Required | Validation         |
| -------------- | ------ | -------- | ------------------ |
| `session_type` | string | yes      | `memorization`/`review` |
| `surah_number` | int    | yes      | between `1` and `114` |
| `start_ayah`   | int    | yes      | ≥ `1`              |
| `end_ayah`     | int    | yes      | ≥ `start_ayah`     |

**Example request**
```json
{ "session_type": "memorization", "surah_number": 2, "start_ayah": 1, "end_ayah": 10 }
```

**Response `201 Created`**
```json
{
  "message": "Session created successfully.",
  "session": {
    "id": 5,
    "user_id": 1,
    "session_type": "memorization",
    "duration": 0,
    "surah_number": 2,
    "start_ayah": 1,
    "end_ayah": 10,
    "score": 0,
    "mistakes": 0,
    "hearts_lost": 0,
    "is_complete": false,
    "created_at": "2026-01-03T12:00:00Z"
  }
}
```

**Errors:** `400` invalid body/validation, `500` failure.

---

#### `GET /api/v1/app/sessions/:id` 🔒 — Get one session

Path param `id`. Returns the session **only if it belongs to the user**. `404` if not found.

**Response `200 OK`** — `{ "session": { "...": "..." } }`

---

#### `PATCH /api/v1/app/sessions/:id` 🔒 — Update a session

Path param `id`. Updates in-progress session metrics.

**Request body**
| Field          | Type   | Required | Validation                          |
| -------------- | ------ | -------- | ----------------------------------- |
| `session_type` | string | yes      | `memorization`/`review`             |
| `duration`     | int    | yes      | `1`–`86400` (24h)                  |
| `score`        | int    | yes      | `0`–`100`                            |
| `mistakes`     | int    | yes      | `0`–`10000`                          |
| `hearts_lost`  | int    | yes      | `0`–`10000`                          |

**Response `200 OK`**
```json
{ "message": "Session updated successfully.", "session": { "...": "..." } }
```

---

#### `DELETE /api/v1/app/sessions/:id` 🔒 — Delete a session

Path param `id`. **Response `200 OK`** — `{ "message": "Session deleted successfully." }` · `404` if not found.

---

#### `POST /api/v1/app/sessions/:id/complete` 🔒 — Complete a session

Path param `id`. Marks the session complete and atomically updates the user — **subtracts
`hearts_lost` from the user's hearts** and **increments the streak by 1**. Runs inside a DB
transaction. Cannot complete an already-completed session.

**Request body**
| Field         | Type | Required | Validation   |
| ------------- | ---- | -------- | ------------ |
| `mistakes`    | int  | yes      | ≥ `0`        |
| `hearts_lost` | int  | yes      | ≥ `0`        |

**Response `200 OK`**
```json
{ "message": "Hifz session completed successfully." }
```

**Errors:** `400` invalid body/negative values, `404` session not found **or already completed**,
`500` failure.

---
### 5.6 Daily progress

Common validation for body fields in this group:

| Field               | Type | Validation                |
| ------------------- | ---- | ------------------------- |
| `new_ayahs_count`   | int  | `0`–`1000`               |
| `reviewed_count`    | int  | `0`–`1000`               |
| `goal_target`       | int  | `1`–`1000`               |
| `is_goal_achieved`  | bool | —                         |

Only **one** daily-progress record may exist per user per date (`409 Conflict` on duplicate).

---

#### `GET /api/v1/app/daily-progress/` 🔒 — List my daily progress (paginated)

**Query params**

| Param   | Type | Default | Validation |
| ------- | ---- | ------- | ---------- |
| `page`  | int  | `1`     | ≥ `1`      |
| `limit` | int  | `30`    | `1`–`100`  |

Ordered by `date DESC, id DESC`.

**Response `200 OK`**
```json
{
  "progress": [ { "...": "..." } ],
  "meta": { "total": 12, "page": 1, "limit": 30, "total_pages": 1 }
}
```

---

#### `POST /api/v1/app/daily-progress/` 🔒 — Create daily progress

**Request body**
```json
{
  "date": "2026-01-03",
  "new_ayahs_count": 7,
  "reviewed_count": 4,
  "goal_target": 10,
  "is_goal_achieved": false
}
```

`date` uses the `YYYY-MM-DD` format. **Response `201 Created`**:
```json
{ "message": "Daily progress created successfully.", "progress": { "...": "..." } }
```

**Errors:** `400` invalid body / bad date / validation, `409` record already exists for that date, `500` failure.

---

#### `GET /api/v1/app/daily-progress/:date` 🔒 — Get progress by date

Path param `date` (`YYYY-MM-DD`). Returns the record only if it belongs to the user.

**Response `200 OK`** — `{ "progress": { "...": "..." } }` · `404` if not found.

---

#### `PATCH /api/v1/app/daily-progress/:date` 🔒 — Update progress by date

Path param `date` (`YYYY-MM-DD`). Body (same validation as create, without `date`):
```json
{ "new_ayahs_count": 9, "reviewed_count": 5, "goal_target": 10, "is_goal_achieved": true }
```

**Response `200 OK`**
```json
{ "message": "Daily progress updated successfully.", "progress": { "...": "..." } }
```

---

#### `DELETE /api/v1/app/daily-progress/:date` 🔒 — Delete progress by date

Path param `date` (`YYYY-MM-DD`). **Response `200 OK`** — `{ "message": "Daily progress deleted successfully." }` · `404` if not found.

---

### 5.7 Dashboard

#### `GET /api/v1/app/dashboard` 🔒 — Get dashboard summary

Aggregates the user's current stats, today's goal progress, all hifz progress, and the 10 most
recent sessions for the dashboard home screen.

**Response `200 OK`**
```json
{
  "user": { "...": "..." },
  "daily": {
    "date": "2026-01-03",
    "goal": 10,
    "completed": 7,
    "remaining": 3,
    "percentage": 70,
    "goal_achieved": false
  },
  "progress": [ { "...": "..." } ],
  "sessions": [ { "...": "..." } ]
}
```

**Notes on `daily`:** `completed` = the day's `new_ayahs_count`; `percentage` is
`completed / goal × 100` capped at `100`; `remaining` never drops below `0`. If no daily record
exists for today, a zeroed-out placeholder is returned using the user's `daily_goal` as the target.

---
## 6. Data Models

Tables are created on startup via GORM auto-migration. Below are the JSON object shapes.

### 6.1 `User`

| Field             | Type   | Default   | JSON key          |
| ----------------- | ------ | --------- | ----------------- |
| `id`              | uint   | —         | `id`              |
| `google_id`       | string | —         | `google_id`       |
| `email`           | string | —         | `email`           |
| `name`            | string | —         | `name`            |
| `avatar`          | string | —         | `avatar`          |
| `hearts`          | int    | `5`       | `hearts`          |
| `streak`          | int    | `0`       | `streak`          |
| `daily_goal`      | int    | `5`       | `daily_goal`      |
| `onboarded`       | bool   | `false`   | `onboarded`       |
| `session_version` | int    | `1`       | `session_version` |
| `created_at`      | time   | —         | `created_at`      |
| `updated_at`      | time   | —         | `updated_at`      |

### 6.2 `HifzProgress`

| Field             | Type | Default   | JSON key         |
| ----------------- | ---- | --------- | ---------------- |
| `id`              | uint | —         | `id`             |
| `user_id`         | uint | —         | `user_id`        |
| `surah_number`    | int  | —         | `surah_number`   |
| `start_ayah`      | int  | —         | `start_ayah`     |
| `end_ayah`        | int  | —         | `end_ayah`       |
| `mastery_level`   | int  | `1`       | `mastery_level`  |
| `repetitions`     | int  | `0`       | `repetitions`    |
| `last_reviewed_at`| time | nullable  | `last_reviewed_at`|
| `created_at`      | time | —         | `created_at`     |
| `updated_at`      | time | —         | `updated_at`     |

### 6.3 `DailyProgress`

| Field             | Type | Default   | JSON key            |
| ----------------- | ---- | --------- | ------------------- |
| `id`              | uint | —         | `id`                |
| `user_id`         | uint | —         | `user_id`           |
| `date`            | date | —         | `date`              |
| `new_ayahs_count` | int  | `0`       | `new_ayahs_count`   |
| `reviewed_count`  | int  | `0`       | `reviewed_count`    |
| `goal_target`     | int  | `20`      | `goal_target`       |
| `is_goal_achieved`| bool | `false`   | `is_goal_achieved`  |
| `created_at`      | time | —         | `created_at`        |
| `updated_at`      | time | —         | `updated_at`        |

### 6.4 `HifzSession`

| Field          | Type   | Default   | JSON key      |
| -------------- | ------ | --------- | ------------- |
| `id`           | uint   | —         | `id`          |
| `user_id`      | uint   | —         | `user_id`     |
| `session_type` | string | —         | `session_type`|
| `duration`     | int    | —         | `duration`    |
| `surah_number` | int    | —         | `surah_number`|
| `start_ayah`   | int    | —         | `start_ayah`  |
| `end_ayah`     | int    | —         | `end_ayah`    |
| `score`        | int    | —         | `score`       |
| `mistakes`     | int    | `0`       | `mistakes`    |
| `hearts_lost`  | int    | `0`       | `hearts_lost` |
| `is_complete`  | bool   | `false`   | `is_complete` |
| `created_at`   | time   | —         | `created_at`  |

`User` has a one-to-many relationship with `HifzProgress`, `DailyProgress`, and `HifzSession`
(`OnDelete: CASCADE`). Related objects are omitted from user JSON output.

---
## 7. Errors & Conventions

### 7.1 Error payload

All error responses share a single shape:

```json
{ "error": "<human-readable message>" }
```

### 7.2 Status codes used

| Code    | Meaning                                                            |
| ------- | ------------------------------------------------------------------ |
| `200`   | Success (GET/PATCH/DELETE)                                         |
| `201`   | Created (POST create endpoints)                                    |
| `400`   | Invalid request body, bad path/query param, or validation failure  |
| `401`   | Missing/invalid/expired token, invalid session, failed Google auth |
| `404`   | Resource not found (or not owned by the user)                      |
| `409`   | Conflict — overlapping hifz range or duplicate daily record        |
| `500`   | Server/database failure                                            |

### 7.3 CORS

| Environment | Allowed origin                |
| ----------- | ----------------------------- |
| Local       | `http://localhost:3000`       |
| Production  | `https://hifzapp.netlify.app` |

Allowed methods: `GET, POST, PUT, PATCH, DELETE, OPTIONS`. Allowed headers: `Origin,
Content-Type, Accept, Authorization, X-Requested-With`. Credentials are allowed (cookies are used).

### 7.4 Pagination

List endpoints for sessions and daily progress return a consistent `meta` block:
`{ "total", "page", "limit", "total_pages" }`.

---

## 8. Endpoint Summary

| Method  | Path                                | Auth | Description                       |
| ------- | ----------------------------------- | ---- | --------------------------------- |
| `POST`  | `/api/v1/auth/google`               | 🔓    | Google login / session            |
| `POST`  | `/api/v1/auth/logout`               | 🔒    | Logout (invalidates all sessions) |
| `GET`   | `/api/v1/app/me/`                   | 🔒    | Get my profile                    |
| `POST`  | `/api/v1/app/user/onboarding`       | 🔒    | Complete onboarding               |
| `PATCH` | `/api/v1/app/user/name`             | 🔒    | Update display name               |
| `GET`   | `/api/v1/app/hifz/progress/`        | 🔒    | List progress                     |
| `POST`  | `/api/v1/app/hifz/progress/`        | 🔒    | Create progress                   |
| `GET`   | `/api/v1/app/hifz/progress/:id`     | 🔒    | Get progress                      |
| `PATCH` | `/api/v1/app/hifz/progress/:id`     | 🔒    | Update progress                   |
| `DELETE`| `/api/v1/app/hifz/progress/:id`     | 🔒    | Delete progress                   |
| `GET`   | `/api/v1/app/sessions/`             | 🔒    | List sessions (paginated)         |
| `POST`  | `/api/v1/app/sessions/`             | 🔒    | Create session                    |
| `GET`   | `/api/v1/app/sessions/:id`          | 🔒    | Get session                       |
| `PATCH` | `/api/v1/app/sessions/:id`          | 🔒    | Update session                    |
| `DELETE`| `/api/v1/app/sessions/:id`          | 🔒    | Delete session                    |
| `POST`  | `/api/v1/app/sessions/:id/complete` | 🔒    | Complete session                  |
| `GET`   | `/api/v1/app/daily-progress/`       | 🔒    | List daily progress (paginated)   |
| `POST`  | `/api/v1/app/daily-progress/`       | 🔒    | Create daily progress             |
| `GET`   | `/api/v1/app/daily-progress/:date`  | 🔒    | Get daily progress by date        |
| `PATCH` | `/api/v1/app/daily-progress/:date`  | 🔒    | Update daily progress by date     |
| `DELETE`| `/api/v1/app/daily-progress/:date`  | 🔒    | Delete daily progress by date     |
| `GET`   | `/api/v1/app/dashboard`             | 🔒    | Get dashboard summary             |