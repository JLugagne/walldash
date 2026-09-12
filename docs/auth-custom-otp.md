# Custom OTP authentication — integration plan

Status: implemented (see §15 for deviations found during implementation)
Decisions source: `.questions/2026-09-12-auth-custom-otp.md`
Libraries: `github.com/JLugagne/egauth/otp` + `github.com/JLugagne/egauth/tokens` (+ `tokens/basic`, `tokens/memory`)

## 1. Goal

Replace Home Assistant identification with a self-hosted flow:

- No auth cookie → device goes through an **OTP** login.
- Each **device** is an anonymous account (its own subject UUID) that can be listed, assigned
  a role, and revoked independently.
- OTP is **per-device**, single-use, TTL 15 min, exposed in **logs and the setup tab**.
- Access + refresh tokens are **cookies** (HTTPS only, `__Host-`/Secure defaults).
- Device accounts + refresh tokens persist in SQLite; pending state and OTP live in memory and
  are lost on restart (a new code is issued).

### Decisions locked

| # | Decision |
|---|----------|
| Q1 | OTP per device: `connect` creates a consumer (subject) and an OTP bound to it, single-use |
| Q2 | Code exposed in logs **and** setup tab |
| Q3 | Access TTL 15 min, **no deny-list**; disconnect applies at next refresh (≤15 min) |
| Q4 | First device = `owner`; owner can promote devices to `admin` |
| Q5 | Per-device revoke + "log out this device" |
| Q6 | All clients HTTPS → egauth `__Host-` Secure cookies |
| Q7 | Per-IP rate limit on issue/verify |
| Q8 | `/api/ws` requires the access cookie; client refreshes then reconnects |

## 2. Why not a "pending access token"

`IssueTokenPair` also mints a **refresh-token family**. A pending device would get a rotatable
refresh token before proving the OTP, and `RequireAuth` would accept the JWT, forcing a global
"pending" gate on every other route. A JWT is a bad correlation handle.

Instead: `connect` sets an **opaque `pending` cookie** (HttpOnly, Secure, Path=/, 15 min) whose
value is a random key mapped **in memory** to the device subject UUID. No JWT before auth.

Likewise, there is **no "disconnected" claim**. A stateless JWT cannot be revoked; the only
enforcement point is `ClaimsProvider` during `Rotate`. Account `status` lives in SQLite and is
read there. A status claim would only be observability, not enforcement.

## 3. Package layout

Follows the existing hexagonal shape (`domain` → `app` → `inbound`/`outbound`).

```
internal/dashboard/
  domain/
    auth.go                                  # Account, Role, Status, device labels
    repositories/accounts/accounts.go        # AccountRepository interface
    repositories/accounts/accountstest/contract.go
  outbound/sqlite/
    migrations/012_auth_accounts.sql         # accounts + refresh_tokens (+ api_keys)
    accounts.go                              # sqlite AccountRepository
    tokensstore.go                           # tokens.Store[struct{}] over SQLite
  app/
    auth.go                                  # App implements auth use-cases
  inbound/
    middleware/auth.go                       # RequireAuth wiring + role gates
    commands/auth.go                         # connect / verify / refresh / logout
    queries/auth.go                          # me / pending codes / devices
    websocket/                               # guard ServeWS with ContextMiddleware
internal/dashboard/init.go                   # composition root wiring
```

Frontend:

```
frontend/src/components/auth/LoginScreen.tsx
frontend/src/components/setup/AuthPanel.tsx   # pending codes + device list/roles/revoke
frontend/src/useAuth.ts                        # bootstrap, 401 handling, refresh+retry
```

## 4. Data model (SQLite)

`migrations/012_auth_accounts.sql`:

```sql
CREATE TABLE IF NOT EXISTS auth_accounts (
    id           TEXT PRIMARY KEY,              -- device subject UUID
    status       TEXT NOT NULL DEFAULT 'active',-- 'active' | 'revoked'
    role         TEXT NOT NULL DEFAULT 'device',-- 'owner' | 'admin' | 'device'
    label        TEXT NOT NULL DEFAULT '',      -- User-Agent / human name
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP,
    revoked_at   TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auth_refresh_tokens (
    hash        TEXT PRIMARY KEY,               -- SHA-256 hex (egauth stores only the hash)
    family_id   TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    tenant_id   TEXT NOT NULL DEFAULT '',
    auth_time   TIMESTAMP NOT NULL,
    expires_at  TIMESTAMP NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    consumed_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_auth_rt_user   ON auth_refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_rt_family ON auth_refresh_tokens(family_id);
CREATE INDEX IF NOT EXISTS idx_auth_rt_exp    ON auth_refresh_tokens(expires_at);

-- Present so tokens.Store[struct{}] is fully implemented; walldash never issues API keys.
CREATE TABLE IF NOT EXISTS auth_api_keys (
    id         TEXT PRIMARY KEY,
    tenant_id  TEXT NOT NULL DEFAULT '',
    hash       TEXT NOT NULL UNIQUE,
    prefix     TEXT NOT NULL DEFAULT '',
    type       INTEGER NOT NULL DEFAULT 0,
    created_by TEXT NOT NULL,
    claims     BLOB NOT NULL,
    expires_at TIMESTAMP,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

OTP is **not** persisted: use `otp/memory.NewStore()`.

Domain types (`domain/auth.go`):

```go
type Role   string // owner | admin | device
type Status string // active | revoked

type Account struct {
    ID        string
    Status    Status
    Role      Role
    Label     string
    CreatedAt time.Time
    LastSeen  *time.Time
    RevokedAt *time.Time
}
```

`AccountRepository` (mirrors `repositories/*` conventions + a `accountstest` contract):
`Create`, `FindByID`, `FindAll`, `SetRole`, `Revoke`, `Touch`, plus `CountByRole(owner)` for the
"first device is owner" rule (or `Create` returns owner when the table is empty, done inside a
transaction).

## 5. egauth wiring (`init.go`)

```go
cookies := tokens.DefaultCookies() // __Host-access_token / __Host-refresh_token, Secure, Path=/

otpSvc := otp.NewService(
    otpmemory.NewStore(),                 // in-memory, forgotten on restart (Q1/Q3)
    otp.WithTTL(15*time.Minute),
    otp.WithMaxAttempts(5),
    otp.WithCooldown(30*time.Second),
)

tokenStore := sqlite.NewTokensStore(adapter.DB())

issuerCfg := basic.Config{
    Store:      tokenStore,
    Issuer:     "walldash",
    SecretKey:  cfg.TokenSecret,          // >= 32 bytes; Validate() at startup
    AccessTTL:  15 * time.Minute,         // Q3
    RefreshTTL: 60 * 24 * time.Hour,
    ClaimsProvider: basic.ClaimsProviderFunc(func(ctx context.Context, userID uuid.UUID, tenantID string) (basic.Claims, error) {
        acct, err := accountRepo.FindByID(ctx, userID.String())
        if err != nil || acct.Status != domain.StatusActive {
            return basic.Claims{}, errAuthRevoked // aborts Rotate → device re-enrolls
        }
        return claimsFor(acct), nil           // subject, roles, scopes, AMR ["otp"]
    }),
}
if err := issuerCfg.Validate(); err != nil { return nil, err }
issuer := basic.NewIssuer(issuerCfg)
```

`claimsFor(acct)`:

```go
basic.Claims{
    Subject: uuid.MustParse(acct.ID),
    Roles:   []string{string(acct.Role)},
    Scopes:  scopesFor(acct.Role), // e.g. owner/admin → {"setup:manage"}, all → {"dashboard"}
    AMR:     []string{"otp"},
}
```

egauth treats `Scopes` as opaque; using them for capability gating keeps `tokens/basic` (no
custom claims type) and lets `WithRequiredScopes("setup:manage")` gate setup routes.

Secret handling: `TOKEN_SECRET` env / add-on option. If absent, generate 32 random bytes on first
boot and persist in an `app_secrets` row (new tiny table); reuse thereafter. Never rotate silently.

## 6. Flows and endpoints

All JSON, JSend envelope, mounted on the existing `gorilla/mux` router. Existing global CSRF
middleware stays; unauthenticated clients fetch `GET /api/csrf-token` first (safe method, exempt).

### 6.1 `POST /api/auth/connect` — public, rate-limited
1. Read `User-Agent` → label; client IP from `X-Forwarded-For`/`RemoteAddr`.
2. Generate `pendingID` (32 random bytes hex) + `deviceUUID`.
3. Store `{pendingID → {deviceUUID, label, createdAt}}` in an in-memory pending map (mutex).
4. `ch, _ := otpSvc.Issue(ctx, "", deviceUUID, "device-enroll")`; cache `ch.Code` in the pending
   record for the setup tab; **log** `otp_issued` with `device_id`, `label`, `ip`, `code`.
5. Set `pending` cookie (HttpOnly, Secure, SameSite=Lax, Path=/, Max-Age 900).
6. `200 {status:"pending"}` — never returns the code.

### 6.2 `POST /api/auth/verify` — public, rate-limited
Body `{ "code": "004217" }`, `pending` cookie required.
1. Resolve `pendingID` → `deviceUUID`; missing/expired → `401 invalid_code`.
2. `otpSvc.Verify(ctx, "", deviceUUID, "device-enroll", code)`; any error → `401 invalid_code`.
3. `accountRepo.Create(deviceUUID, label)`: role = `owner` if no account exists, else `device`.
4. `pair, _ := issuer.IssueTokenPair(ctx, claimsFor(acct))`.
5. `cookies.SetAccess(...)`, `cookies.SetRefresh(..., persistent=true)`, clear `pending` cookie,
   drop the pending record and cached code.
6. `200 {data:{device:{id,label,role}}}`.

### 6.3 `POST /api/auth/refresh`
`basic.RefreshHandler(issuer, tokens.WithCookies(cookies), tokens.WithTrustedOrigins(hosts...))`.
Rotates the family and rewrites cookies; `ClaimsProvider` rejects revoked accounts. `204` on
success.

### 6.4 `POST /api/auth/logout`
`basic.LogoutHandler(tokenStore, tokens.WithCookies(cookies), ...)`: revokes the family, clears
cookies, idempotent `204`. The account stays `active`.

### 6.5 `GET /api/auth/me` — protected
Returns `{id, label, role, status}` from the verified claims + DB.

### 6.6 Setup endpoints — protected + `WithRequiredScopes("setup:manage")`
- `GET  /api/setup/auth/pending` → list `{device_id, label, code, expires_at}` (codes from the
  in-memory pending map).
- `GET  /api/setup/auth/devices` → list accounts (id, label, role, status, created, last_seen).
- `POST /api/setup/auth/devices/{id}/revoke` → `accountRepo.Revoke` +
  `tokenStore.RevokeAllRefreshTokensForUser(ctx, "", id)`.
- `POST /api/setup/auth/devices/{id}/role` → owner only, sets `admin`/`device`. Only `owner` may
  grant/revoke `owner` (prevent admin→owner escalation).

## 7. Route protection & middleware

`inbound/middleware/auth.go`:

```go
protected := basic.RequireAuth(issuer, next, 
    tokens.WithCookieAuth(basic.Cookies(cookies)),
    tokens.WithAutoRefresh(issuer, basic.Cookies(cookies)),
)
```

- **Public**: `POST /api/auth/{connect,verify,refresh,logout}`, `GET /api/auth/status`
  (bootstrap probe → `{authenticated:bool}`), `GET /api/csrf-token`, static SPA.
- **Protected**: every existing `/api/*` command/query route + `/api/ws`.
- **Setup**: protected + `tokens.WithRequiredScopes[struct{}]("setup:manage")`.

WebSocket: wrap `wsHub.ServeWS` with
`basic.ContextMiddleware(issuer, http.HandlerFunc(wsHub.ServeWS), tokens.WithCookieAuth(cookies))`.
On invalid/expired access token it returns `401`; the SPA refreshes over HTTP then reconnects.
The handshake cannot auto-refresh.

`/api/auth/status` and `/api/auth/me` let the SPA decide between app and `LoginScreen` without
parsing cookies.

## 8. `tokens.Store[struct{}]` over SQLite (`outbound/sqlite/tokensstore.go`)

Interface (source of truth: `tokens/store.go`) — 12 methods:

| Group | Methods |
|---|---|
| RefreshTokenStore | `SaveRefreshToken`, `FindRefreshToken`, `ConsumeRefreshToken`, `RevokeRefreshToken`, `RevokeFamily`, `RevokeAllRefreshTokensForUser` |
| APIKeyStore | `SaveAPIKey`, `FindAPIKeyByHash`, `RevokeAPIKey`, `ListAPIKeysByCreator`, `RevokeAllAPIKeysForUser` |
| TokenReaper | `DeleteExpired` |

Implementation notes:

- Struct `TokensStore struct { db *sql.DB }`, built from `adapter.DB()` (shared single connection,
  `SetMaxOpenConns(1)` already serializes writes).
- `ConsumeRefreshToken(hash)`:
  `UPDATE auth_refresh_tokens SET consumed_at=CURRENT_TIMESTAMP WHERE hash=? AND consumed_at IS NULL`.
  - `RowsAffected()==1` → nil (this caller consumed it).
  - `0` → `SELECT consumed_at ...`: row exists → `tokens.ErrRefreshTokenReused`; absent →
    `tokens.ErrRefreshTokenNotFound`.
- `RevokeFamily` / `RevokeAllRefreshTokensForUser` → `DELETE`; idempotent (no error on 0 rows).
- `DeleteExpired` → `DELETE ... WHERE expires_at < CURRENT_TIMESTAMP`; return rows deleted.
- API-key methods are implemented against `auth_api_keys` for interface fidelity, but walldash
  issues none: `FindAPIKeyByHash` → `tokens.ErrAPIKeyNotFound`; revoke/list are idempotent/empty.
- Add a test that runs egauth's exported `tokens/storetest` conformance suite against this store,
  plus a race test for concurrent `ConsumeRefreshToken`.
- Schedule GC: a janitor goroutine calling `tokenStore.DeleteExpired(ctx, "")` hourly
  (`egauth/janitor` or a simple ticker).

## 9. Rate limiting

Use `egauth/ratelimit` (or the existing middleware style) with a `TokenBucket` keyed by client IP,
mounted only on `connect` and `verify`. This sits on top of the OTP cooldown (30 s) and 5-attempt
cap. Reject with `429` + JSend.

## 10. Frontend

- `useAuth.ts`: on app mount call `GET /api/auth/me`; `401` → unauthenticated.
  Wrap `apiFetch` so a `401` triggers `POST /api/auth/refresh` then retries once; if refresh
  fails → unauthenticated.
- `LoginScreen.tsx`: `POST /api/auth/connect` on mount → "waiting for approval" state + a 6-digit
  input; submit `POST /api/auth/verify`; on success reload the app shell. Show a friendly
  retry/lock message on `429`/`401`.
- `AuthPanel.tsx` (setup, owner/admin): pending codes list (poll every ~3 s) + device table with
  role selector and revoke button. Hidden/disabled for `device` role.
- Add the route under `SetupShell` (`/setup/access`) in `routes.tsx`.

## 11. Config

| Var | Meaning |
|---|---|
| `TOKEN_SECRET` | HS256 signing key, >= 32 bytes (or auto-generated + persisted in `app_secrets`). |
| `ALLOWED_ORIGINS` | already exists; also fed to `WithTrustedOrigins` (normalized to bare hosts). Also settable as the add-on option `allowed_origins`. |
| `DOMAIN` | Public hostname of this instance (bare `host`, `http://host` or `https://host`; an optional port is kept and path/query are stripped). Normalized to a host and merged with `ALLOWED_ORIGINS` (deduplicated, order stable), then used for the CORS `Access-Control-Allow-Origin` header, the CSRF/same-origin checks and egauth `WithTrustedOrigins`. Also settable as the add-on option `domain`. |
| `AUTH_ENABLED` | not implemented; authentication is always on (local development uses the same OTP flow). |

## 12. Testing

- `accountstest/contract.go` conformance for the SQLite account repository.
- egauth `tokens/storetest` conformance for `TokensStore`.
- `app/auth` unit tests: first-device-is-owner, promote, revoke aborts `ClaimsProvider`,
  consent/label handling.
- httptest integration: `connect` → read code from the in-memory display → `verify` → access a
  protected route → `refresh` → `revoke` → refresh now fails → device re-enrolls.
- Frontend: `LoginScreen` (pending → code → success), `AuthPanel` (revoke, role).
- Ensure `/api/ws` rejects without a valid cookie.

## 13. Risks / gotchas

- **Q6 blocker**: confirm the browser-facing HA URL is `https://` for *all* clients (tablets).
  `__Host-`+`Secure` cookies are dropped on plain HTTP with no visible error → login loop.
  If any client is HTTP, switch to custom cookie names + conditional `Secure`.
- **Reverse proxy**: ensure `Host` is preserved or add the external host to `ALLOWED_ORIGINS` /
  `WithTrustedOrigins`; pass `X-Forwarded-Proto: https` so `IsHTTPS()` sees TLS termination.
- **Double CSRF**: existing CSRF middleware and egauth's Origin check coexist; both must trust the
  same origins.
- **First device bootstrap**: the very first code is only in the logs (no authenticated setup tab
  yet). Document it.
- **Single container only**: the pending map and OTP are in-memory; multiple replicas would break
  the flow. Fine for the HA add-on.
- **Revocation latency**: up to 15 min by design (Q3). Do not claim "instant". `revoke` kills
  refresh immediately; access tokens expire naturally.
- **Pending cleanup**: sweep the in-memory pending map on a timer (TTL 15 min) to avoid leaks.
- **Never log** access/refresh tokens; log OTP codes only (that is the intended channel).
- **Role escalation**: only `owner` may set the `owner` role.

## 14. Suggested phasing (tracer bullets)

1. Migration `012` + `domain.Account` + SQLite account repo + contract test.
2. `TokensStore` + `tokens/storetest` conformance + janitor.
3. egauth wiring in `init.go` + `TOKEN_SECRET` + startup validation.
4. Auth endpoints (connect/verify/refresh/logout/me) + protected middleware + bootstrap probe.
5. `/api/ws` auth.
6. Setup endpoints (pending, devices, revoke, role) + rate limiting.
7. Frontend `LoginScreen` + `useAuth` bootstrap/refresh-retry.
8. Frontend `AuthPanel` + route.
9. Hardening pass: proxy/origin config, logging, docs (README + `.env.example`).

## 15. Implementation notes & deviations

- **Refresh TTL is 60 days.** Access TTL stays 15 min (Q3).
- **First device is `owner` atomically**: `accountRepo.Create` inserts
  `CASE WHEN NOT EXISTS (SELECT 1 FROM auth_accounts) THEN 'owner' ELSE <requested role> END`,
  so no separate `CountByRole` call is needed.
- **Pending-map cleanup is lazy**, not a dedicated ticker: `Auth.sweep()` runs on
  `StartEnrollment`, `VerifyEnrollment` and `ListPending`. The map is in-memory and reset on
  restart, and `connect` is rate-limited, so it cannot grow unbounded.
- **`AUTH_ENABLED` was not implemented** - authentication is always on.
- **`ALLOWED_ORIGINS` is normalized to bare hosts** before being passed to egauth
  `WithTrustedOrigins` (`trustedOriginHosts` in `init.go`). The existing CSRF middleware and the
  WebSocket `CheckOrigin` already accept either a full URL or a bare host; egauth keyed origins by
  raw string and therefore only matched bare hosts. Without this, an `ALLOWED_ORIGINS` value such
  as `https://walldash.example.com` was silently ignored by the refresh/logout same-origin check
  (403 `cross_site_blocked`). A regression test covers it
  (`TestRefreshTrustsAllowedOrigins`).
- **CORS matches the request origin host, not the raw string**: the `CORS` middleware now parses
  the `Origin` header and compares its host (case-insensitively) against each allowed entry,
  accepting a bare host or a URL, so a `DOMAIN` value such as `walldash.domain.tld` matches
  `https://walldash.domain.tld`. The `*` behavior is unchanged.
- **`ALLOWED_ORIGINS` is read through the add-on option loader** (`configValue`), so it can be set
  as `allowed_origins` in `/data/options.json` rather than only as an environment variable.
- **`X-Forwarded-Proto` is honored** by egauth's `IsHTTPS`, so TLS terminated at a reverse proxy
  is recognized. **`X-Forwarded-Host` is not consulted**: each origin check compares the `Origin`
  host with `r.Host`. If a proxy rewrites `Host`, set `ALLOWED_ORIGINS` to the browser-facing
  host(s). The usual reverse proxy (preserving `Host`) needs no extra config.
- **HTTPS is mandatory** (Q6): `tokens.DefaultCookies()` yields `__Host-` + `Secure`. There is no
  conditional `Secure` fallback. Over plain HTTP the browser drops the cookies and the SPA enters
  a login loop with no visible error. Documented in `README.md`, `.env.example` and
  `docs/home-assistant-add-on.md`; there is no plain-HTTP deployment path.
- **Logging**: `otp_issued` intentionally logs the one-time code (the delivery channel). No
  access or refresh token value is logged anywhere in the auth code path.

## 16. Manual smoke checklist

Run against a real (HTTPS) deployment or the `TestAuthEndpointsFlow` integration test:

1. `GET /api/health` without a cookie -> `401`.
2. `POST /api/auth/connect` -> `200 {"status":"pending"}`, no code in the body; the code is in
   the log (`otp_issued`) and in Setup -> Access.
3. `POST /api/auth/verify` with the code -> cookies set; the device is `owner` (first device).
4. `GET /api/auth/me` -> `200` with `role: owner`; `GET /api/health` -> `200`.
5. `POST /api/auth/refresh` -> `204` and a rotated refresh cookie.
6. Connect `/api/ws` with the access cookie -> accepted; without it -> `401`.
7. Setup -> Access: a second device enrolls as `device`; the owner promotes it to `admin`, then
   revokes it.
8. After revoke, refresh fails (`401`); an access token already issued keeps working for at most
   the remaining 15-minute access TTL, then stops.
9. `POST /api/auth/logout` revokes the family and clears cookies; refresh then fails.
