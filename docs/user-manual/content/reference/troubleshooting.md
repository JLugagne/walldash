---
title: "Troubleshooting"
description: "Common issues with Walldash and how to solve them."
weight: 30
---

## The dashboard shows example devices

Walldash is running in **demo mode**, which means `HA_URL` or `HA_TOKEN` is empty. Set both and restart the container. When running as a Home Assistant add-on, the token is provided automatically — check the add-on logs for the resolved Home Assistant configuration.

## I keep returning to the sign-in screen (login loop)

This almost always means the browser is not on HTTPS. Walldash's auth cookies are `__Host-` + `Secure`, and browsers silently drop them over plain HTTP, so the sign-in never sticks. Reach Walldash through HTTPS.

If you already use a reverse proxy, check that it forwards `X-Forwarded-Proto: https` so Walldash sees the original scheme. If the proxy rewrites the `Host` header, set `DOMAIN` to the browser-facing hostname; otherwise the CORS and same-origin checks reject the request. See [Custom domain & HTTPS](/reference/reverse-proxy/) and [Configuration](/reference/configuration/).

## I cannot find the first one-time code

The **first device** has no authenticated **Setup → Access** tab yet, so read the code from the server log instead. Look for an `otp_issued` line in the add-on log or `docker compose logs`; it contains the six-digit code, the device label and the client IP. Enter it on the sign-in screen to bootstrap the installation as **owner**. After that, later codes also appear under **Setup → Access**.

## I revoked a device but it still works

Revocation is not instant. Revoking deletes the device's refresh tokens immediately, but an access token that was already issued remains valid until it expires — up to **15 minutes** — because there is no deny-list. The device stops working when its access token expires and it cannot refresh.

## Device states do not update

Real-time updates use a WebSocket at `/api/ws`. If you access Walldash through a reverse proxy, make sure **WebSockets support** is enabled. States will still update on a manual page refresh even when the WebSocket is blocked.

## A device is listed but cannot be controlled

Walldash only controls **allow-listed** operations: on/off toggles and automation triggering. Domains without a supported control (for example a read-only diagnostic sensor) appear but have no primary action.

## The 3D view is empty or distorted after an import

Very large or degenerate `.sh3d` geometry can produce unusual plans. Open the level in the 2D editor and check that rooms form closed polygons and that walls have a sane thickness. Re-exporting the file from Sweet Home 3D with a simpler model usually fixes it.

## I lost my layout after an update

Data lives in the `/data` volume (add-on) or in the `DB_PATH` file (Docker). Make sure that volume or file is mounted persistently. Use the **export** feature to keep a portable JSON backup.

## Where are the logs?

Set `LOG_LEVEL=debug` for verbose logs. In the add-on, logs are available from the add-on page in Home Assistant; with Docker, use `docker compose logs -f`.
