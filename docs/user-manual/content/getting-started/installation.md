---
title: "Installation"
description: "Install Walldash as a Home Assistant add-on or with Docker Compose, and connect it to your Home Assistant instance."
weight: 10
---

## As a Home Assistant add-on (recommended)

The add-on is the simplest option: it connects to Home Assistant through the Supervisor API, so **no token setup is required**.

1. In Home Assistant, open **Settings** → **Add-ons** → **Add-on Store**.
2. Open the menu (**⋮**) → **Repositories** and add:
   ```
   https://github.com/JLugagne/ha-addons
   ```
3. Find **Walldash** in the store, install it, then start it.
4. Serve Walldash over **HTTPS** (put a TLS reverse proxy in front of it) and open the web UI on your tablets. Plain HTTP cannot hold the authentication cookies.
5. Enroll the first device with the one-time code from the add-on log — see [Sign in & device access](/getting-started/sign-in/).

Data is stored in the add-on `/data` volume and survives updates and reboots. The wall tablets connect directly to Walldash, so **no Home Assistant login is needed** on them — each device signs in once with its own one-time code, and the first device becomes the owner.

## With Docker Compose

Use this option if you run Home Assistant Core in Docker or want Walldash on a separate host.

1. Clone the repository:
   ```bash
   git clone https://github.com/JLugagne/walldash.git
   cd walldash
   ```
2. Create your configuration file:
   ```bash
   cp .env.example .env
   ```
3. Fill in your Home Assistant details:
   ```env
   HA_URL=http://homeassistant.local:8123
   HA_TOKEN=your_long_lived_access_token_here
   PORT=8080
   DB_PATH=/app/data/walldash.db
   ```
4. Start it:
   ```bash
   docker compose up -d
   ```
5. Put a TLS reverse proxy in front of it and open `https://<your-domain>` on your tablets. Plain HTTP cannot hold the authentication cookies.
6. Read the first `otp_issued` code from `docker compose logs` and enter it on the sign-in screen — see [Sign in & device access](/getting-started/sign-in/).

## Generating an access token

The token is only needed for Docker deployments and manual setups. The add-on uses the Supervisor token automatically.

1. Open your Home Assistant web interface.
2. Click your user profile icon (bottom left).
3. Scroll to the **Security** tab → **Long-Lived Access Tokens**.
4. Click **Create Token**, name it `Walldash`, and copy the value into your `.env` file.

> Walldash only ever performs an allow-listed set of actions against Home Assistant: **on/off toggles** and **automation triggering**. It never forwards arbitrary service calls.

When `HA_URL` or `HA_TOKEN` is missing, Walldash starts in **demo mode** and shows a set of built-in example devices. This is useful to explore the interface before connecting a real instance.
