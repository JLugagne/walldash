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

> When you expose Walldash publicly, make sure your Home Assistant long-lived token is scoped and that your proxy does not cache the API responses.
