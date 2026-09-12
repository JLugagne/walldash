---
title: "Sign in & device access"
description: "Enroll a tablet with a one-time code, understand the owner, admin and device roles, and keep sessions secure over HTTPS."
weight: 15
---

Walldash does not use your Home Assistant login on wall tablets. Instead, **each device enrolls itself** as its own anonymous account with a one-time code. No Home Assistant credentials ever reach the tablet.

## First open: the sign-in screen

The first time a device opens Walldash, the app asks the server for a one-time code and shows the **sign-in screen** with a six-digit code field.

{{< figure src="/images/login.png" alt="Walldash sign-in screen with a six-digit one-time code field" caption="The sign-in screen: Walldash requests a code for this device, then waits for you to enter it." >}}

Walldash calls `POST /api/auth/connect` in the background when the screen opens. This creates the device's account and issues a code bound to it. The API response never contains the code itself.

## Where to read the code

The code is single-use and expires after **15 minutes**. It is delivered through two channels:

- **The server log** — the intended delivery channel. Walldash logs an `otp_issued` line containing the code, the device label and the client IP. In the Home Assistant add-on, open the add-on page and check its log; with Docker, use `docker compose logs`.
- **Setup → Access** — once an owner or admin device is enrolled, any pending code is listed there with its device label and expiry.

> The **first device** has no authenticated Setup tab yet, so its only channel is the log. Read the first `otp_issued` code from the add-on or container log to bootstrap the installation.

## Enrolling the device

1. Open Walldash on the tablet over **HTTPS**.
2. Read the current code from the log or from **Setup → Access**.
3. Type the six digits into the sign-in screen and submit.
4. Walldash calls `POST /api/auth/verify` with the code. On success the device is enrolled, signed in, and the app shell opens. The device keeps a **refresh token** so it stays signed in on later visits.

The first device to verify becomes **owner**; every later device becomes **device**.

## Roles

| Role | What it can do |
| --- | --- |
| **owner** | Everything an admin can do, plus grant or remove the **owner** role. |
| **admin** | See pending codes, list devices, change a device's role (except to/from owner) and revoke any device. |
| **device** | Use the dashboard and the 3D views only. No access to device management. |

Only the **owner** may grant or remove the **owner** role, which prevents an admin from escalating itself.

## Sessions and revocation

Signing in sets two cookies:

- an **access token** valid for **15 minutes**, used to authenticate API calls and the WebSocket;
- a **refresh token** valid for **60 days**, rotated on every refresh, which keeps the device signed in silently.

**Revocation is not instant.** Revoking a device (or logging it out) deletes its refresh tokens immediately, so it can no longer renew its session. However, an access token that was already issued stays valid until it expires — up to **15 minutes**. There is no deny-list, so this delay is by design. Walldash itself refreshes tokens as needed and reconnects the WebSocket with the new access token.

You can revoke a device at any time from **Setup → Access** (see [Setup mode](/using/setup-mode/)).

## HTTPS is mandatory

Authentication cookies are named `__Host-` and marked `Secure`. Browsers silently drop them over plain HTTP with no visible error, which looks like a **login loop**: the device appears to sign in, then immediately returns to the sign-in screen. Always reach Walldash through HTTPS.

- Terminate TLS at a reverse proxy and make sure it forwards `X-Forwarded-Proto: https` so Walldash knows the original request was secure. See [Custom domain & HTTPS](/reference/reverse-proxy/).
- If the proxy rewrites the `Host` header, set `DOMAIN` to the browser-facing hostname so the CORS and same-origin checks trust it. Additional origins can be listed in `ALLOWED_ORIGINS`.
- `TOKEN_SECRET` is the signing key for the tokens. Leave it empty to generate and persist one on first boot, or set it explicitly (at least 32 bytes) and keep it stable — changing it signs out every device.

See [Configuration](/reference/configuration/) for all three options.
