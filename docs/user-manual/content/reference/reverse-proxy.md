---
title: "Publishing on the internet"
description: "Reach Walldash from outside your home through a reverse proxy with HTTPS."
weight: 20
---

By default Walldash is meant for your home network. If you want to reach it from the internet, put it behind a reverse proxy that provides HTTPS. Any reverse proxy works; **Nginx Proxy Manager** is a common, beginner-friendly choice.

1. Point a domain name at the machine running Walldash.
2. Create a new proxy host for that domain.
3. Set the forward address to your Walldash server and its port (`8080` by default).
4. Turn on **WebSockets support** so live device updates keep flowing.
5. Tell the proxy to pass through the original connection type, so Walldash knows the visitor arrived over HTTPS.
6. In the SSL settings, request a certificate and turn on **Force SSL**.

Keep the proxy connected directly to Walldash. These steps work for both a container install and the Home Assistant add-on.

For the complete technical reference, see the [repository docs](https://github.com/JLugagne/walldash/blob/main/docs/reverse-proxy.md).
