---
title: "Settings"
description: "The few Walldash settings you might change — and why most people never need to."
weight: 10
---

Most people never change a single Walldash setting. Install it, open it and it works.

If you do need to adjust something, the options live in the Home Assistant add-on under **Settings → Add-ons → Walldash → Configuration**. The two you are most likely to touch are:

- **Log verbosity** — how much detail Walldash writes to its log. Turn it up only while diagnosing a problem, then turn it back down.
- **Rescue mode** — a recovery switch. Turn it on, restart, then open Walldash on the device that should become the new owner. Turn it back off when you are done. See [Devices & access](/getting-started/sign-in/).

There are also a few advanced network options for unusual setups, such as when a reverse proxy changes the address Walldash sees. If you are not running a reverse proxy, you can ignore them. See [Publishing on the internet](/reference/reverse-proxy/).

For the complete technical reference, see the [repository README](https://github.com/JLugagne/walldash).
