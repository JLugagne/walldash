# Home Assistant Add-on Packaging

Walldash is distributed as a Home Assistant add-on with **direct port access** (no Ingress),
so wall-mounted tablets can open the dashboard without a Home Assistant login.
The runtime image follows the official recommendations: Home Assistant base image
(`ghcr.io/home-assistant/base`, pinned) with s6-overlay as PID 1, bashio for option
mapping, and the official builder actions for multi-arch publishing.

## Distribution Model

Two repositories are involved:

| Repository | Role |
| --- | --- |
| `ha-dash` (this repo) | Application source. Builds the image from `Dockerfile` (HA base + s6 service in `rootfs/`) and pushes per-arch images plus a multi-arch manifest to GHCR on every published GitHub Release. |
| `ha-addons` (companion repo) | Add-on store repository. Contains only metadata (`config.yaml`, translations, docs) that points at the GHCR manifest via its `image:` field. Users add this repository URL to their add-on store. |

Pre-built containers are used: nothing is compiled on user machines, installs are fast and reliable.

## Runtime Layout (this repo)

- `Dockerfile`: multi-stage build (Node frontend → Go backend → HA base). Accepts the
  builder-provided `BUILD_ARCH`/`BUILD_VERSION` args (`BUILD_VERSION` is baked into the
  binary via `-X main.Version=...`); `BUILD_FROM` defaults to the pinned base for local builds.
- `rootfs/etc/services.d/walldash/run`: s6 service script (`#!/usr/bin/with-contenv bashio`).
  Maps the `log_level` option to `LOG_LEVEL` and execs `/usr/bin/walldash`.
  The binary itself falls back to `/data/options.json` and the Supervisor API when env vars are unset.
- `io.hass.*` labels are added automatically by the builder actions; the Dockerfile only
  carries OCI labels.

## Runtime Behavior (Add-on Mode)

When the binary detects it runs inside an add-on container, it adapts without any manual setup:

- **Supervisor API**: if `HA_URL`/`ha_url` is unset and `SUPERVISOR_TOKEN` is present, the app uses `http://supervisor/core/api` with the Supervisor token. This requires `homeassistant_api: true` in the add-on `config.yaml`. The Supervisor token is never sent to a user-provided `HA_URL`.
- **Options file**: `/data/options.json` (written by the Supervisor from the add-on configuration) is read as a fallback for every setting. Environment variables always win.
- **Persistent storage**: when the `/data` volume exists, the database defaults to `/data/walldash.db`, surviving updates and reboots.
- **Version**: the binary version is injected at build time (`BUILD_VERSION` build arg → `main.Version`), keeping the `/api/health` version and the add-on `version:` in sync.
- **Health check**: `GET /api/health` serves as the add-on `watchdog` URL.
- **Init system**: s6-overlay is PID 1, so the add-on `config.yaml` must set `init: false` (required since S6 V3, otherwise the add-on will not start).

Full precedence per setting: environment variable → `/data/options.json` → default (see `README.md`).

## Release Process

1. Publish a GitHub Release named `v1.2.3` (leading `v` is stripped for image tags and labels).
2. The `Publish` workflow builds native `amd64` + `aarch64` images
   (`ghcr.io/jlugagne/<arch>-walldash:1.2.3`) and combines them into the multi-arch
   manifest `ghcr.io/jlugagne/walldash:1.2.3` (plus `:latest`), signed with Cosign.
3. In the `ha-addons` repo, bump `walldash/config.yaml` → `version: "1.2.3"` and append to `walldash/CHANGELOG.md`.
4. Test the upgrade on a staging Home Assistant before announcing it.

## Companion Repository Templates

Below are ready-to-copy starting files for the `ha-addons` repository.

### `repository.yaml` (repo root)

```yaml
name: JLugagne Home Assistant Add-ons
url: https://github.com/JLugagne/ha-addons
maintainer: JLugagne
```

### `walldash/config.yaml`

```yaml
name: Walldash
version: "1.0.0"
slug: walldash
description: Touch-first 3D home automation dashboard for Home Assistant.
url: https://github.com/JLugagne/ha-dash
arch:
  - amd64
  - aarch64
init: false
image: ghcr.io/jlugagne/walldash
ports:
  8080/tcp: 8080
ports_description:
  8080/tcp: Walldash web interface (direct access, no Home Assistant login required)
webui: http://[HOST]:[PORT]/
boot: auto
startup: application
homeassistant_api: true
map:
  - type: data
    read_only: false
options:
  log_level: info
schema:
  log_level: list(debug|info|warn|error)
watchdog: http://[HOST]:8080/api/health
panel_icon: mdi:tablet-dashboard
stage: experimental
```

### `walldash/translations/en.yaml`

```yaml
configuration:
  log_level:
    name: Log level
    description: Verbosity of the add-on logs.
network:
  8080/tcp: Walldash web interface (direct access, no Home Assistant login required).
```

> Note: additional translation files (e.g. `fr.yaml`) can be added next to `en.yaml`
> once the add-on is validated.

### `walldash/DOCS.md` (outline)

Cover: prerequisites (HA OS/Supervised with add-on store), adding the repository,
installing and starting Walldash, opening `http://<home-assistant-ip>:8080` on tablets
(kiosk mode tip), configuration options table, data persistence (`/data`), updating,
uninstallation, troubleshooting (port conflict, Supervisor API unreachable, logs).

### `walldash/README.md`, `walldash/CHANGELOG.md`, icons

- `README.md`: one-paragraph summary + link to `DOCS.md`.
- `CHANGELOG.md`: one entry per `version:` bump.
- `icon.png` / `logo.png`: 256x256 PNGs.

## Local Test Checklist

- [ ] Add-on installs from the custom repository on a test instance.
- [ ] Starts with default options; logs show Supervisor API in use, no token configured manually.
- [ ] Dashboard reachable at `http://<ha-ip>:8080` from a device without a Home Assistant session.
- [ ] Device toggle from the dashboard reflects in Home Assistant.
- [ ] Database persists across add-on restart and update.
- [ ] Version reported by `/api/health` matches the `config.yaml` version.
- [ ] Upgrade from the previous version works without data loss.
