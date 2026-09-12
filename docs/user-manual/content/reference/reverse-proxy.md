---
title: "Custom domain & HTTPS"
description: "Reach Walldash at your own domain with automatic HTTPS through Nginx Proxy Manager."
weight: 20
---

To reach Walldash at `https://walldash.domain.tld`, put a reverse proxy in front of it. The steps below use **Nginx Proxy Manager**, but any reverse proxy that supports WebSockets works.

See [`docs/reverse-proxy.md`](https://github.com/JLugagne/walldash/blob/main/docs/reverse-proxy.md) in the repository for the complete reference.

## Requirements

- A DNS record pointing `walldash.domain.tld` to your proxy host.
- **WebSocket support** enabled on the proxy host. Walldash keeps a WebSocket open at `/api/ws` for real-time device updates; without it, states stop refreshing.

## Nginx Proxy Manager

1. Create a new **Proxy Host**.
2. Set the **Domain Names** to `walldash.domain.tld`.
3. Set the **Forward Hostname / IP** and **Forward Port** to your Walldash server (`8080` by default).
4. Enable **Websockets Support**, **Block Common Exploits** and **Access List** as you see fit.
5. In the **SSL** tab, request a new certificate and enable **Force SSL**.

This works for both the Docker Compose deployment and the Home Assistant add-on.

## Trusting the public hostname

Walldash checks the origin of authenticated requests, and it must recognize that TLS terminates at the proxy. Two things matter:

- **Forward the original scheme.** The proxy must pass `X-Forwarded-Proto: https`. Nginx Proxy Manager does this automatically when **Force SSL** is on. Without it, Walldash sees a plain HTTP request and the `Secure` auth cookies are dropped — the device loops back to the sign-in screen.
- **Set `DOMAIN` if the `Host` header is rewritten.** Walldash compares the request `Origin` against the host it received. If the proxy forwards a different `Host` than the browser used, set `DOMAIN=walldash.domain.tld` (the add-on `domain` option also works) so the CORS and same-origin checks trust the public name. A proxy that preserves the original `Host` needs no extra configuration; extra origins can still be listed in `ALLOWED_ORIGINS`.

See [Configuration](/reference/configuration/) for `DOMAIN` and `ALLOWED_ORIGINS`.

> When you expose Walldash publicly, make sure your Home Assistant long-lived token is scoped and that your proxy does not cache the API responses.
