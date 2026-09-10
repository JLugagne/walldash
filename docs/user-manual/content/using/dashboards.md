---
title: "Overview dashboards"
description: "Compose grids of widgets for quick actions, device toggles, sensors and favorite automations."
weight: 40
---

**Overview dashboards** are customizable widget grids, designed for touch. Open them from the view switcher → **Overviews**.

{{< figure src="/images/dashboard.png" alt="Walldash overview dashboard with sensor, switch and automation widgets" caption="An overview dashboard: sensor cards, a device toggle and a list of favorite automations." >}}

## Widgets

A **widget** is a single block anchored in the dashboard grid. It is bound to one or more devices or automations, and rendered according to a **display mode**:

- **Number** — a large numeric reading.
- **Arc** — a downward-opening dial between two bounds (great for temperature).
- **Bar** — a horizontal progress-style reading.
- **Toggle** — a tap-to-toggle device.
- **List** — a list of automations with their last run time.

Because the display mode is independent of the data, you can show the same sensor as a number, an arc or a bar.

{{< figure src="/images/dashboard-edit.png" alt="Adding a widget in edit mode" caption="In **edit mode**, add widgets and drag them to rearrange the grid." >}}

## Edit mode

Only administrators compose dashboards. Open the **`● NAME · LIVE` menu** at the top-right of the Overviews view and pick:

1. **Admin mode** to unlock editing.
2. **Edit layout** or **Add widget** to compose the dashboard, then drag widgets to rearrange the grid. The grid fills exactly the screen, so layouts stay consistent across tablets.
3. **Change background** to pick an image for this dashboard (also available from **Setup → Settings**).

Outside edit mode, a tap on a widget only triggers its **primary action**: toggling the device for an actuator, or nothing at all for a sensor.

## Stale values

A measurement is marked **stale** when its entity is unreachable or has not been refreshed for more than thirty minutes. Stale widgets stay visible but are visually de-emphasized so you never trust an out-of-date reading.
