---
title: "Setup mode"
description: "Edit your home from a dedicated Setup workspace, kept apart from everyday use."
weight: 25
---

Walldash separates **using the house** from **setting up the house**.

The everyday shell — what you keep on a wall tablet — only offers the views you use day to day. Editing tools are not peers of the 3D view in the top bar; they live behind a deliberate **Setup** button.

## Everyday shell

{{< figure src="/images/floor-controls.png" alt="The everyday top bar with the 3D View and Overviews segments and the Setup button" caption="The everyday bar: **3D View** and **Overviews**, the connection status, and a quiet **Setup** button on the right." >}}

The everyday bar has two view segments (3D View / Overviews), the connection status, and the **Setup** button. On the Overviews view it also shows the active dashboard's date and a `● NAME · LIVE` menu.

## Entering and leaving Setup mode

Tap **Setup** to switch the whole app into Setup mode. The shell changes unmistakably:

- a **SETUP MODE** banner reminds you that you are configuring, not viewing;
- a task-oriented nav lists the four configuration areas;
- **Exit setup** returns to the everyday shell. You can also press `Esc`.

## Floors & Plans

{{< figure src="/images/plan-editor.png" alt="Setup mode with the Floors & Plans task active and the 2D editor open" caption="**Floors & Plans** opens the 2D editor for the selected floor. This replaces the old top-level *2D Editor*." >}}

The **2D editor** now lives here. Trace walls and openings, define rooms and outdoor zones, place devices, and manage the layers of each level. See [Plan editor (2D)](/using/plan-editor/) for the details.

## Devices

{{< figure src="/images/setup-devices.png" alt="Setup mode listing every supported Home Assistant device" caption="**Devices** lists every supported device discovered from Home Assistant: name, domain and current state." >}}

A read-only inventory of the devices Walldash exposes, useful to confirm that an entity is supported before placing it on a plan.

## Dashboards

{{< figure src="/images/setup-dashboards.png" alt="Setup mode listing the overview dashboards with widget counts" caption="**Dashboards** lists your overview dashboards and how many widgets each holds." >}}

Manage your overview dashboards: open one to edit it, or create a new one. Widget composition (add, layout, background) is done on the [Overviews](/using/dashboards/) view itself, from the `● NAME · LIVE` menu.

## Settings

{{< figure src="/images/setup-settings.png" alt="Setup mode settings with the dashboard background entry point" caption="**Settings** gathers dashboard-wide options such as the background image." >}}

Dashboard-wide options live here, including the [background image](/using/dashboards/).

## Why two shells?

- **The everyday surface stays clean and safe.** On a shared wall panel there is no risk of tapping into an editor by accident.
- **Editing is deliberate.** It is a labelled mode with a clear exit, not a hidden toggle.
- **One clear mental model.** The old overlap between *2D Editor*, *admin mode* and *layout edit mode* is gone: you are either using Walldash or setting it up.
