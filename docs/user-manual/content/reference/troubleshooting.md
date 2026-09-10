---
title: "Troubleshooting"
description: "Common issues with Walldash and how to solve them."
weight: 30
---

## The dashboard shows example devices

Walldash is running in **demo mode**, which means `HA_URL` or `HA_TOKEN` is empty. Set both and restart the container. When running as a Home Assistant add-on, the token is provided automatically — check the add-on logs for the resolved Home Assistant configuration.

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
