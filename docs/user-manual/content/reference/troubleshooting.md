---
title: "Troubleshooting"
description: "Common issues with Walldash and how to solve them."
weight: 30
---

## I cannot reach the Walldash web page

- Make sure Walldash is running. For the Home Assistant add-on, open the add-on page and check that it is started.
- Check that the tablet is on the same network as the machine running Walldash.
- Open the web address shown on the add-on page, or the one your installer gave you.
- If you use a reverse proxy, make sure it handles the connection correctly — see [Publishing on the internet](/reference/reverse-proxy/).

## The dashboard shows example devices

Walldash is showing its built-in demo home instead of your real devices. For the Home Assistant add-on, open the add-on page and check that Walldash is started and connected to Home Assistant.

## Some devices are missing or cannot be controlled

Walldash only controls the devices it supports: lights, switches, plugs and automations. A device that is read-only, or that Walldash does not support yet, may appear without a control button. If a device you expect is missing completely, check that it is set up in Home Assistant first.

## A device cannot sign in

A new device waits on the access screen until an owner or admin lets it in.

1. On an owner or admin device, open **Setup → Access**.
2. Either approve the waiting device under **Pending approvals**, or create an invitation and open its link on the new device.

See [Devices & access](/getting-started/sign-in/) for the steps.

## I revoked a device but it still works

A revoked device loses access shortly after — within about **15 minutes**. Wait a few minutes and try again.

## The owner device is lost

You can make another device the owner with Rescue mode:

1. In the add-on configuration, turn on the **Rescue mode** option.
2. Restart the add-on.
3. Open Walldash on the device that should become the new owner.
4. Turn **Rescue mode** back off.

See [Settings](/reference/configuration/) for where the option lives.

## The 3D view looks empty or distorted

A very large or unusual plan can confuse the 3D view. Open the level in the 2D editor and check that rooms are closed and walls look right. Re-exporting a simpler plan from Sweet Home 3D usually fixes it.

## I lost my layout after an update

Use the **export** feature in Walldash to save a copy of your home, and restore it if something goes wrong. When running as an add-on, your data is stored inside Home Assistant and survives updates.

## Still stuck?

Open an issue on the project's GitHub page and describe what you see.
