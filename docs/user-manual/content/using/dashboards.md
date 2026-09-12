---
title: "Dashboards"
description: "Compose grids of widgets for quick actions, device toggles, sensors and favorite automations."
weight: 40
---

**Dashboards** are customizable widget grids, designed for touch. Open them from the top bar → **Dashboards**. On a fresh install, Walldash creates a default **Home** dashboard with a 5-day weather forecast, so there is always something useful on screen; rename, extend or delete it like any other dashboard.

{{< figure src="/images/dashboard.png" alt="Walldash dashboard with sensor, weather, device toggle, automation switch and automation list widgets" caption="A dashboard: sensor cards, weather, a device toggle, an automation switch and a list of favorite automations. When more than one dashboard exists, its switcher appears in the top bar." >}}

## Widgets

A **widget** is a single block anchored in the dashboard grid. It is bound to one or more devices or automations, and rendered according to a **display mode**:

- **Number** — a large numeric reading.
- **Arc** — a downward-opening dial between two bounds (great for temperature).
- **Bar** — a horizontal progress-style reading.
- **Toggle** — a tap-to-toggle device.
- **Switch** — a two-position automation switch: one automation runs when you turn it on, another when you turn it off.
- **List** — a list of automations with their last run time.

Because the display mode is independent of the data, you can show the same sensor as a number, an arc or a bar.

## Automation switches

An **Automation switch** is the quickest way to expose a pair of frequently used automations as a single control. Configure one automation for the **On** position and one for the **Off** position:

- turn the switch **on** → Walldash triggers the **On** automation;
- turn the switch **off** → Walldash triggers the **Off** automation.

The switch position reflects the **most recently triggered** of the two automations, so it stays in sync even when an automation is started from Home Assistant rather than from Walldash. Its caption shows the widget title and the automation matching the current position.

{{< figure src="/images/dashboard-edit.png" alt="Editing a dashboard in Setup mode" caption="Editing a dashboard in **Setup → Dashboards**: add widgets, drag them to rearrange the grid, and change the background." >}}

## Viewing vs editing

The dashboard view is **display-only**. A tap on a widget only triggers its **primary action**: toggling the device for an actuator, or nothing at all for a sensor.

Composition happens in **Setup → Dashboards**. The **Setup** button on the dashboard view opens the editor for the dashboard you are currently viewing:

1. **Edit layout** or **Add widget** to compose the dashboard, then drag widgets to rearrange the grid. The grid fills exactly the screen, so layouts stay consistent across tablets.
2. **Change background** to pick a picture **for this dashboard** — every dashboard has its own background.
3. **Rename** or **Delete** the current dashboard, or use **+** in the Setup banner to create another one.

While editing, the Setup banner also shows the dashboard switcher so you can jump from one dashboard to another.

## Stale values

A measurement is marked **stale** when its entity is unreachable or has not been refreshed for more than thirty minutes. Stale widgets stay visible but are visually de-emphasized so you never trust an out-of-date reading.
