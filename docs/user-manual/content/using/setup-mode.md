---
title: "Setup mode"
description: "Edit your home from a dedicated Setup workspace, kept apart from everyday use."
weight: 25
---

Walldash separates **using the house** from **setting up the house**.

The everyday shell — what you keep on a wall tablet — only offers the views you use day to day. Editing tools are not peers of the 3D view in the top bar; they live behind a deliberate **Setup** button.

## Everyday shell

{{< figure src="/images/floor-controls.png" alt="The everyday top bar with the 3D View and Dashboards segments, the clock and the Setup button" caption="The everyday bar: **3D View** and **Dashboards**, the live clock and date, and a quiet **Setup** button on the right." >}}

The everyday bar has two view segments (**3D View** / **Dashboards**), the live clock and date, and the **Setup** button. On the Dashboards view a dashboard switcher appears in the bar as soon as there is more than one dashboard.

## Entering and leaving Setup mode

Tap **Setup** to switch the whole app into Setup mode. The button is context-aware: on a floor it opens that floor's 2D editor, and on the Dashboards view it opens the editor for the dashboard you are viewing. The shell changes unmistakably:

- a **SETUP MODE** banner reminds you that you are configuring, not viewing;
- the five configuration areas are listed below the banner;
- **Exit setup** returns to the everyday shell — for dashboards it reopens the one you were editing. You can also press `Esc`.

Focused editors (**Floors & Plans** and a specific **Dashboard**) hide the cross-section nav so they fill the height and preview the final result. Use **Exit setup** to reach the other areas.

## Floors & Plans

{{< figure src="/images/plan-editor.png" alt="Setup mode with the 2D editor open" caption="**Floors & Plans** opens the 2D editor for the selected floor." >}}

The **2D editor** lives here. Trace walls and openings, define rooms and outdoor zones, place devices, and manage the layers of each level. See [Plan editor (2D)](/using/plan-editor/) for the details.

## Devices

{{< figure src="/images/setup-devices.png" alt="Setup mode listing every supported Home Assistant device" caption="**Devices** lists every supported device discovered from Home Assistant: name, domain and current state." >}}

A read-only inventory of the devices Walldash exposes, useful to confirm that an entity is supported before placing it on a plan.

## Dashboards

{{< figure src="/images/setup-dashboards.png" alt="Setup mode listing the dashboards with widget counts" caption="**Dashboards** lists your dashboards and how many widgets each holds." >}}

Manage your dashboards: open one to edit it, or create a new one. Rename, delete, layout, widgets and the per-dashboard background are all handled in the dashboard editor (see [Dashboards](/using/dashboards/)).

## Access

{{< figure src="/images/setup-access.png" alt="Setup Access tab listing pending sign-in codes and enrolled devices with roles and a revoke action" caption="**Access** lists pending one-time codes, then every enrolled device with its role and a **revoke** action." >}}

The **Access** area manages device sign-in and is available to **owner** and **admin** devices only:

- **Pending codes** — devices that have opened the sign-in screen and are waiting for their one-time code. Each entry shows the device label, the code and when it expires, so you can read it out instead of digging through the log.
- **Devices** — every enrolled device with its label, role, status and last-seen time. Change a device's role (`admin` / `device`; only the **owner** can grant or remove `owner`) or **revoke** it.
- **Revoke** deletes the device's refresh tokens immediately, but an access token already issued keeps working for up to 15 minutes. See [Sign in & device access](/getting-started/sign-in/).

## Settings

{{< figure src="/images/setup-settings.png" alt="Setup mode global settings" caption="**Settings** gathers the global Walldash options." >}}

Global application options live here. The dashboard background is **not** one of them: each dashboard carries its own picture, edited in the dashboard editor.

## Why two shells?

- **The everyday surface stays clean and safe.** On a shared wall panel there is no risk of tapping into an editor by accident.
- **Editing is deliberate.** It is a labelled mode with a clear exit, not a hidden toggle.
- **One clear mental model.** There is no admin toggle to find: you either use Walldash or open Setup to configure it.
