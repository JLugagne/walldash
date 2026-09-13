---
title: "Installation"
description: "Install Walldash as a Home Assistant add-on, then open it on your first device."
weight: 10
---

## As a Home Assistant add-on (recommended)

The add-on is the easiest way to install Walldash. It connects to your Home Assistant automatically, so there is nothing to configure.

1. In Home Assistant, open **Settings** → **Add-ons** → **Add-on Store**.
2. Open the menu (**⋮**) → **Repositories** and add `https://github.com/JLugagne/ha-addons`.
3. Find **Walldash** in the store, install it, then start it.
4. Open the web address shown on the add-on page in a browser on your tablet or touchscreen.

Your data is kept inside Home Assistant and survives updates and restarts. The tablets connect straight to Walldash, so they never need a Home Assistant login. Home Assistant snapshots include Walldash's data and the key that signs device sessions, so treat snapshots like passwords (see [Settings](/reference/configuration/)).

To add more tablets later, see [Devices & access](/getting-started/sign-in/).

## With Docker (optional)

If you prefer to run Walldash yourself, a ready-made container image and the Docker instructions are available in the [repository README](https://github.com/JLugagne/walldash).

Open Walldash on your first device — it becomes the owner.
