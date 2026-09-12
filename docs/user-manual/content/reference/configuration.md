---
title: "Configuration"
description: "Every environment variable Walldash understands and how settings are resolved."
weight: 10
---

Settings resolve with the following precedence:

**environment variable** → **add-on options file** (`/data/options.json`) → **default**

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port the server listens on. |
| `DB_PATH` | `walldash.db` (or `/data/walldash.db` when the `/data` volume exists) | SQLite database path. |
| `HA_URL` | `http://homeassistant.local:8123` | Home Assistant base URL. |
| `HA_TOKEN` | _(empty)_ | Long-lived access token. Not needed when running as an add-on: `SUPERVISOR_TOKEN` is used automatically via the Supervisor API proxy. The Supervisor token is never attached to a custom `HA_URL`. |
| `TOKEN_SECRET` | _(auto-generated + persisted)_ | HS256 key used to sign access and refresh tokens, at least 32 bytes. Leave empty to generate one on first boot and persist it in the database; keep it stable, because changing it signs out every device. |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error`. |
| `ALLOWED_ORIGINS` | _(empty)_ | Comma-separated trusted origins for CORS, CSRF and the same-origin check, needed when a reverse proxy rewrites the `Host` header. Accepts `https://host` or a bare `host`. Also settable as the add-on option `allowed_origins`. |
| `DOMAIN` | _(empty)_ | Public hostname of this Walldash instance, used for the CORS `Access-Control-Allow-Origin` header and the CSRF/same-origin checks. Accepts a bare host, `http://host` or `https://host`; an optional port is kept and any path or query is ignored. Merged with `ALLOWED_ORIGINS`. Also settable as the add-on option `domain`. |
| `FRONTEND_DIR` | _(embedded assets)_ | Serve the frontend from a directory instead of the embedded build. |

## Demo mode

If `HA_URL` or `HA_TOKEN` are unset, Walldash starts in **demo mode**: it serves a set of built-in example devices and automations instead of querying Home Assistant. This lets you explore the interface, and it is how the screenshots in this manual were produced.

## Storage

Walldash stores everything in a single SQLite database (via the pure-Go `modernc.org/sqlite` driver — no CGO). Back it up by copying the `DB_PATH` file, or use the built-in **export** feature to download a portable JSON snapshot.
