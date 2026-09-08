# Exposing Walldash on Its Own Domain (Nginx Proxy Manager)

This guide shows how to reach Walldash at `https://walldash.domain.tld` using
[Nginx Proxy Manager](https://nginxproxymanager.com/) (NPM) with automatic HTTPS.
It works the same whether Walldash runs via Docker Compose or as a Home Assistant
add-on — only the forward host/port differ.

## Prerequisites

- A domain you control (e.g. `domain.tld`), with a DNS record pointing at your server:
  - `A walldash.domain.tld → <server-public-ip>` (or a `CNAME` to your DynDNS name).
  - For LAN-only access, create the same record on your local DNS (router, Pi-hole, AdGuard).
- Ports `80` and `443` reachable on the NPM host (port-forward on your router for
  external access; not needed for LAN-only with DNS challenge, see below).
- Walldash reachable on your LAN and its port known:
  - Docker Compose (this repo): `http://<docker-host-ip>:9090` by default
    (host-mapped port from `compose.yml`).
  - Home Assistant add-on: `http://<home-assistant-ip>:8080` (or the host port set
    in the add-on **Network** settings if you changed it).

## Add the Proxy Host in NPM

1. Open NPM → **Hosts** → **Proxy Hosts** → **Add Proxy Host**.
2. **Details** tab:
   - **Domain Names**: `walldash.domain.tld`
   - **Scheme**: `http`
   - **Forward Hostname / IP**: `<docker-host-ip>` or `<home-assistant-ip>`
   - **Forward Port**: `9090` (Compose default) or `8080` (add-on default)
   - **Cache Assets**: off
   - **Block Common Exploits**: on
   - **Websockets Support**: **on (required)** — Walldash uses a WebSocket for
     real-time device state sync; without it toggles and live updates break.
3. **SSL** tab:
   - **SSL Certificate**: **Request a new SSL Certificate**
   - **Force SSL**: on
   - **HTTP/2 Support**: on
   - **HSTS Enabled**: on (only if you always use HTTPS for this domain)
   - **Email**: your address (Let's Encrypt notifications), then **Save**.
   - LAN-only without open port 80: use a **DNS Challenge** instead
     (NPM → **SSL Certificates** → DNS provider credentials), since the HTTP
     challenge needs port 80 reachable from the internet.
4. Open `https://walldash.domain.tld` — Walldash loads with a valid certificate.

## Optional hardening

- **Access List** (NPM → **Access Lists**): restrict by IP/CIDR (e.g. your LAN only),
  or add HTTP basic auth in front. Note tablets in kiosk mode handle basic auth
  poorly — prefer IP restriction for wall panels.
- **Custom locations**: not needed; Walldash serves everything (UI + API +
  WebSocket) from `/` on a single port.

## Troubleshooting

| Symptom | Likely cause / fix |
| --- | --- |
| `502 Bad Gateway` | NPM cannot reach Walldash: check forward host/port, and that the container/add-on is running. From the NPM host, `curl http://<host>:<port>/api/health` must return `ok`. |
| Toggles work but states never update live | **Websockets Support** is off on the Proxy Host — enable it. |
| Browser mixed-content warnings / redirect loop | **Force SSL** on, and open only the `https://` URL. |
| Certificate fails (HTTP challenge) | Port 80 must reach NPM from the internet — check the router forward, or switch to DNS challenge. |
| Works on LAN but not remotely | Router port-forward `80`/`443` → NPM host missing, or DNS record pointing at the wrong (e.g. CGNAT) IP. |

## Reference

- Compose mapping: see `compose.yml` (`9090:8080`).
- Add-on default port and options: see `docs/home-assistant-add-on.md`.
- Health endpoint for checks: `GET /api/health`.
