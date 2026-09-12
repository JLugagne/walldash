---
title: "Configuration"
description: "Every setting Walldash understands, how settings resolve, the add-on option mapping, and the security-critical options."
weight: 10
---

Walldash is configured with a small set of settings. The same names are used everywhere: as
environment variables (Docker, manual runs), as Home Assistant add-on options (stored in
`/data/options.json`), and — for `TOKEN_SECRET` / `SECRET_KEY` — as signing keys persisted in the
database.

## How settings resolve

Every setting is resolved with the same precedence:

**environment variable** → **add-on options file** (`/data/options.json`) → **default**

- The **environment variable** layer wins. This covers variables injected by your shell, Docker
  Compose, the add-on runtime, and a local `.env` file. Walldash loads `.env` at startup and only
  uses a value from it when the variable is not already present in the real environment.
- The **add-on options file** layer is read from `/data/options.json` when the file exists. The
  Supervisor writes it from the add-on configuration UI. Keys are matched case-insensitively
  (`log_level`, `token_secret`, …). Only flat scalar values are used.
- The **default** is applied when both layers are empty.

### Worked example: `LOG_LEVEL`

1. **Environment variable wins.** Start the container with `LOG_LEVEL=debug` and no add-on option.
   Walldash resolves `debug`, even if `/data/options.json` says otherwise.
2. **Add-on option is used when the environment is empty.** With no `LOG_LEVEL` in the environment
   but `/data/options.json` containing `{"log_level": "warn"}`, Walldash resolves `warn`.
3. **Default applies when both are empty.** With neither set, Walldash resolves `info`.

The same three layers apply to every setting below. For example, to configure the public hostname
you set `DOMAIN` (Docker) or the `domain` add-on option (Home Assistant) — both end up on the same
setting.

## Settings

### `PORT`

- **What it is.** The TCP port the HTTP server listens on (`:PORT`).
- **Default.** `8080`.
- **Format.** An integer, e.g. `8080` or `9090`. Environment variable `PORT`; add-on option key
  `port` (not exposed in the add-on UI).
- **When to set it.** When `8080` is already in use, or when your proxy/container maps a different
  host port.
- **Security.** This listener serves plain HTTP. Never expose it directly to the internet — put a
  TLS-terminating reverse proxy in front (see [Custom domain & HTTPS](/reference/reverse-proxy/)).
  The authentication cookies are dropped over plain HTTP.

### `DB_PATH`

- **What it is.** The path to the single SQLite database file that stores every Walldash object:
  plans, dashboards, devices, accounts, refresh tokens and the persisted secrets store.
- **Default.** `/data/walldash.db` when the `/data` directory exists (the add-on volume);
  otherwise `walldash.db` relative to the working directory.
- **Format.** A filesystem path, e.g. `/app/data/walldash.db`. Environment variable `DB_PATH`;
  add-on option key `db_path` (not exposed in the add-on UI).
- **When to set it.** Docker/manual deployments should point it at a persistent volume so data
  survives container recreation.
- **Security.** The database contains device accounts, refresh-token records and — when
  `SECRET_KEY` is unset — the JWT signing key. Treat the file as a secret, restrict its
  permissions, and keep it out of source control. SQLite uses write-ahead logging, so the
  `-wal` and `-shm` sidecar files are part of the same data.

### `HA_URL`

- **What it is.** The base URL of the Home Assistant API that Walldash reads and controls.
- **Default.** `http://homeassistant.local:8123`. When running as a Home Assistant add-on with no
  `HA_URL`/`ha_url` and a `SUPERVISOR_TOKEN` present, Walldash uses the Supervisor API proxy
  (`http://supervisor/core`) instead.
- **Format.** A URL with scheme and port, e.g. `http://homeassistant.local:8123` or
  `https://ha.example.com`. Environment variable `HA_URL`; add-on option key `ha_url`.
- **When to set it.** Docker and manual installs. As an add-on you normally leave it empty and let
  the Supervisor proxy handle it.
- **Security.** Prefer `https://`. If you set `HA_TOKEN` while `HA_URL` uses `http://`, Walldash
  logs a warning because the token is sent in cleartext — only acceptable on a trusted management
  network. The Supervisor token is never attached to a user-supplied `HA_URL`.

### `HA_TOKEN`

- **What it is.** A Home Assistant **long-lived access token** used to authenticate API calls and
  the WebSocket.
- **Default.** Empty.
- **Format.** The `eyJ…` token string. Environment variable `HA_TOKEN`; add-on option key
  `ha_token`.
- **When to set it.** Docker and manual installs. The add-on authenticates through the Supervisor
  token automatically and does **not** need a manual token.
- **Security.** This token grants broad Home Assistant access; store it in a secret manager, never
  in the repository. It is only needed on the server — it never reaches a wall tablet. Rotate it
  in Home Assistant if it leaks.

### `TOKEN_SECRET`

- **What it is.** The HS256 key used to sign Walldash's access and refresh tokens.
- **Default.** Empty. Walldash generates 32 random bytes on first boot and persists them in the
  database, then reuses them on every later start.
- **Format.** Any string of **at least 32 bytes** (a shorter value aborts startup). Environment
  variable `TOKEN_SECRET`; add-on option `token_secret`.
- **When to set it.** When you want to control the key explicitly — for example to keep it in a
  secret manager or to share it across replicas. Leave it empty for the zero-config default.
- **Security.** **Changing it signs out every device** — existing access and refresh tokens can no
  longer be verified. If you leave it empty, the generated key is stored in the database, so a
  database copy or a Home Assistant snapshot exposes it unless you also set `SECRET_KEY`. Back up
  whichever value you use and keep it stable.

### `SECRET_KEY`

- **What it is.** An optional key-encryption key (KEK) used to envelope-encrypt
  (AES-256-GCM) the auto-generated `TOKEN_SECRET` before it is written to the database.
- **Default.** Empty — the stored signing key is kept as plaintext (zero-config behaviour).
- **Format.** **Exactly 32 bytes** (32 ASCII characters). A different length aborts startup.
  Environment variable `SECRET_KEY`; add-on option `secret_key`.
- **When to set it.** Recommended whenever a copy of the database or a Home Assistant snapshot
  could leave your control. It is what keeps `TOKEN_SECRET` out of database backups.
- **Security.**
  - **Keep it stable and backed up.** Store it in a secret manager, never in the repository.
  - **Removing it after use fails closed.** Once the stored signing secret is encrypted with
    `SECRET_KEY`, starting without `SECRET_KEY` aborts with *"stored token secret is encrypted but
    SECRET_KEY is not set"* rather than falling back to a new key. Losing it means you must clear
    the stored secret and re-enroll every device.
  - Setting `SECRET_KEY` on a database whose secret was stored in plaintext re-wraps it in
    encrypted form on the next boot (existing sessions survive).
  - A directly supplied `TOKEN_SECRET` is used as-is; `SECRET_KEY` only protects the
    auto-generated, database-stored key.

### `DOMAIN`

- **What it is.** The public hostname of this Walldash instance. It is merged into the trust list
  used for the CORS `Access-Control-Allow-Origin` header and the CSRF/same-origin (refresh, logout,
  WebSocket) checks.
- **Default.** Empty.
- **Format.** A bare host, `http://host` or `https://host`. An optional port is kept; any path or
  query is ignored. Environment variable `DOMAIN`; add-on option `domain`.
- **When to set it.** When your reverse proxy rewrites the `Host` header so the forwarded host
  differs from the name in the browser. If the proxy preserves `Host` (the common case), you do not
  need it.
- **Security.** This setting defines which origins are trusted for authenticated requests. Set it
  to your real browser-facing hostname and serve it over HTTPS.

### `ALLOWED_ORIGINS`

- **What it is.** A comma-separated list of extra origins trusted for CORS, CSRF and the
  same-origin checks. It is merged (deduplicated, case-insensitive) with `DOMAIN`.
- **Default.** Empty.
- **Format.** Comma-separated origins, each `https://host` or a bare `host`, e.g.
  `https://walldash.example.com,https://dash.example.net`. Environment variable `ALLOWED_ORIGINS`;
  add-on option `allowed_origins`.
- **When to set it.** When Walldash is reachable under more than one hostname, or when a proxy
  rewrites `Host` and you prefer not to use `DOMAIN`.
- **Security.** This is the trust boundary for cross-origin and same-origin enforcement. List only
  the HTTPS origins you actually use; avoid blank or wildcard entries.

### `LOG_LEVEL`

- **What it is.** The verbosity of the structured JSON logs.
- **Default.** `info`.
- **Format.** One of `debug`, `info`, `warn` (alias `warning`), `error`. Any other value falls back
  to `info`. Environment variable `LOG_LEVEL`; add-on option `log_level`.
- **When to set it.** Use `debug` while diagnosing a problem; keep `info` in normal operation.
- **Security.** Debug logs can include operational detail (hosts, token-presence flags, request
  paths). Lower the level again once you are done.

### `FRONTEND_DIR`

- **What it is.** Serve the web UI from a directory on disk instead of the React build embedded in
  the binary.
- **Default.** Empty — the embedded assets are served.
- **Format.** A filesystem path to a built frontend directory. **Environment variable only** (it is
  not read from `/data/options.json` and has no add-on option).
- **When to set it.** Frontend development or a custom build. Regular deployments do not need it.
- **Security.** Only point it at a directory you trust; the static files it serves are public and
  unauthenticated.

## Home Assistant add-on options

As an add-on, these options appear in the configuration UI and are stored in `/data/options.json`.
Each has an environment-variable equivalent, and the environment variable always wins. Edit them in
Home Assistant under **Settings → Add-ons → Walldash → Configuration**.

| Add-on option | Environment variable | Default | Notes |
| --- | --- | --- | --- |
| `log_level` | `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error`. |
| `token_secret` | `TOKEN_SECRET` | _(empty → auto-generated)_ | At least 32 bytes; changing it signs out every device. |
| `secret_key` | `SECRET_KEY` | _(empty)_ | Exactly 32 characters; encrypts the stored signing key. Keep stable and backed up. |
| `domain` | `DOMAIN` | _(empty)_ | Browser-facing hostname; merged into the trust list. |
| `allowed_origins` | `ALLOWED_ORIGINS` | _(empty)_ | Extra comma-separated origins. |

`PORT`, `DB_PATH`, `HA_URL`, `HA_TOKEN` and `FRONTEND_DIR` are not exposed as add-on options; the
add-on sets sensible values (Supervisor API for Home Assistant, `/data` for storage) on its own.

## Security notes

- **HTTPS is mandatory.** Authentication cookies use the `__Host-` prefix and are always set with
  `Secure`, so browsers silently drop them over plain HTTP. The symptom is a login loop with no
  error. Terminate TLS in a reverse proxy in front of Walldash and reach the UI over HTTPS. See
  [Sign in & device access](/getting-started/sign-in/) and [Custom domain & HTTPS](/reference/reverse-proxy/).
- **Forward the original scheme.** The proxy must send `X-Forwarded-Proto: https`, otherwise
  Walldash sees a plain HTTP request and the `Secure` cookies are dropped.
- **Host-rewriting proxies need `DOMAIN` or `ALLOWED_ORIGINS`.** If the forwarded `Host` differs
  from the browser's, the same-origin checks reject authenticated requests. A proxy that preserves
  `Host` needs no extra configuration.
- **`SECRET_KEY` protects the signing key at rest.** It must be exactly 32 characters; keep it
  stable and back it up in a secret manager. Removing it after it has been used fails closed (the
  server refuses to start) instead of silently generating a new key.
- **Changing `TOKEN_SECRET` signs out every device.** Leave it empty to have Walldash generate and
  persist one, or set it explicitly and keep it stable.
- **Never commit secrets.** `HA_TOKEN`, `TOKEN_SECRET` and `SECRET_KEY` belong in a secret manager
  or an untracked `.env` file, not in the repository.

## Demo mode

If `HA_URL` or `HA_TOKEN` is empty and no Supervisor token is available, Walldash starts in **demo
mode**: it serves built-in example devices and automations instead of querying Home Assistant. This
lets you explore the interface (and is how the screenshots in this manual were produced). Set both
`HA_URL` and `HA_TOKEN` — or run as an add-on — to connect a real instance.

## Storage and backup

All state lives in a single SQLite database at `DB_PATH` (the pure-Go `modernc.org/sqlite` driver,
no CGO). In the add-on this is the `/data` volume, which survives updates and reboots.

- **Back up** by copying the database file (with its `-wal` and `-shm` sidecars), or use the built-in
  **export** feature to download a portable JSON snapshot.
- **Protect** the database: without `SECRET_KEY` it contains the signing key in plaintext.
- **Single replica only.** Pending enrollments and one-time codes are held in memory, so run exactly
  one Walldash container per database.

## Related pages

- [Sign in & device access](/getting-started/sign-in/) — HTTPS, cookies and `TOKEN_SECRET`.
- [Setup mode](/using/setup-mode/) — the in-app Settings and Access areas.
- [Custom domain & HTTPS](/reference/reverse-proxy/) — `DOMAIN`, `ALLOWED_ORIGINS` and the proxy.
- [Troubleshooting](/reference/troubleshooting/) — login loops, demo mode and logs.
