---
title: "Settings"
description: "The few Walldash settings you might change — and why most people never need to."
weight: 10
---

Most people never change a single Walldash setting. Install it, open it and it works.

If you do need to adjust something, the options live in the Home Assistant add-on under **Settings → Add-ons → Walldash → Configuration**. The ones you are most likely to touch are:

- **Log verbosity** — how much detail Walldash writes to its log. Turn it up only while diagnosing a problem, then turn it back down.
- **Rescue mode** — a recovery switch. Turn it on, restart, then open Walldash on the device that should become the new owner. Turn it back off when you are done. See [Devices & access](/getting-started/sign-in/).
- **Token secret / key-encryption key file** — advanced options for protecting session keys in backups. Most people leave them empty; see **Backups and secrets** below.

There are also a few advanced network options for unusual setups, such as when a reverse proxy changes the address Walldash sees. If you are not running a reverse proxy, you can ignore them. See [Publishing on the internet](/reference/reverse-proxy/).

## Backups and secrets

Walldash signs every device session with a key that it keeps, encrypted, in its data directory. Home Assistant snapshots and any copy of that directory therefore contain the key material needed to impersonate a signed-in device: **treat them like passwords.**

Walldash protects the key by default: it is encrypted with a key-encryption key generated next to the database. To also protect against backups that leave the machine (off-site copies, files shared for support), store a 32-character key in a file **outside the backed-up volume** and point the advanced **Key-encryption key file** (`secret_key_file` / `SECRET_KEY_FILE`) option at it. Keep that file somewhere safe: if it is lost, Walldash refuses to start until you set the **Token secret** (`token_secret` / `TOKEN_SECRET`) again, which starts fresh and signs out every device.

For the complete technical reference, see the [repository README](https://github.com/JLugagne/walldash).
