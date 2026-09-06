# Walldash

3D isometric smart home dashboard manager and viewer for Home Assistant, optimized for touch tablets.

## Language

### Spaces & Plans

**Level**:
Represents a physical building floor or an outdoor area (e.g. front garden, back garden).
_Avoid_: Floor (too restrictive for outdoor spaces), Layer, Stage

**Plan**:
The 2D vector layout that makes up a Level, composed of walls, rooms, openings and outdoor area boundaries.
_Avoid_: Blueprint, Map, Drawing

**Zone**:
A delimited subdivision within a Plan (interior room or garden sub-zone).
_Avoid_: Room (gardens are not rooms), Area

### Equipment & Entities

**Device Placement**:
The anchor point of a Home Assistant device on a Plan at 2D coordinates (x, y).
_Avoid_: Device Mapping, Pin, Widget

**Device**:
A physical or virtual smart home device attached to Home Assistant (e.g. lamp, plug, TV, sensor, irrigation relay), filtered and normalized by the backend.
_Avoid_: Entity (internal HA term), Item

**Sensor**:
A read-only passive Device whose measurements (temperature, humidity, etc.) are displayed permanently.
_Avoid_: Meter, Gauge

**Actuator**:
A controllable Device (on/off, switch, relay, power) without dimmer/color control in the current version.
_Avoid_: Switch, Controller

**Light Halo**:
The visual light effect displayed in the 3D view around a lamp-type Device when it is on.
_Avoid_: Glow, Flare

### Interactions & Automations

**Automation**:
An automation defined in Home Assistant that can be triggered on demand and monitored (execution status, last execution time).
_Avoid_: Scenario, Script, Routine

**Favorite Automation**:
An Automation selected by the user to appear in the dashboard's quick view.
_Avoid_: Quick Action, Shortcut

### Views & Rendering

**Isometric View**:
The fixed isometric projection 3D view with a camera oriented so that the bottom of the 2D Plan corresponds to the front facade of the house (no free rotation).
_Avoid_: 3D Orbit, Free Cam, Perspective

**Display Layer**:
An exclusive logical display layer specific to a Level, allowing filtering of visible Device Placements in the 3D view (default: `controls` and `sensors`, extensible by the user).
_Avoid_: Layer (without qualifier, to avoid confusion with Level), Filter

**Ceiling Display**:
Information marking (sensor measurements or zone name) projected on a 3D horizontal plane at ceiling height (`y = WALL_HEIGHT`) at the pole of inaccessibility of each Zone, faithfully following the vanishing lines and 3D perspective.
_Avoid_: Floor Print, Ceiling HUD, Floating Tag

**Zone Alert Pulse**:
The continuous sinusoidal breathing visual effect and progressive color interpolation (blue for cold, red for heat) applied to the Ceiling Display text when a measurement crosses the configured thresholds for the Zone.
_Avoid_: Blink, Flash, Glow

### Security & Operations

**Action Whitelist**:
The restricted set of executable operations from the dashboard to Home Assistant (exclusively on/off toggles and automation triggering).
_Avoid_: Command Pass-through, Generic Service Call

### Dashboards & Overview Views

**Overview Dashboard**:
A customized summary view composed of a grid of configurable widgets (sensor states, device switches, selected automation lists).
_Avoid_: Dashboard View, Summary Panel

**Widget**:
A single unit block anchored in an Overview Dashboard's Widget Grid, linked to one or more Devices or Automations and rendered according to a Display Mode.
_Avoid_: Tile, Card, Component

**Widget Grid**:
The fixed column and row grid belonging to an Overview Dashboard, which occupies exactly the window and in which each Widget is anchored.
_Avoid_: Canvas, Layout, Board

**Display Mode**:
The way a Widget renders the data it is linked to, independent of the nature of that data.
_Avoid_: Style, Skin, Renderer, Variant

**Arc**:
The downward-opening dial Display Mode representing a measurement between two bounds set by the administrator.
_Avoid_: Gauge, Jauge, Horseshoe, Dial

**Primary Action**:
The single operation triggered by a tap on a Widget linked to an Actuator. A Widget linked to a Sensor has none.
_Avoid_: Default Action, Main Command, Tap Action

**Edit Mode**:
The state of an Overview Dashboard in which an administrator composes the layout of its Widgets, as opposed to normal use where a tap only triggers a Primary Action.
_Avoid_: Admin Mode, Design Mode, Layout Mode

**Stale Value**:
A measurement that a Widget continues to display even though it is no longer trustworthy, either because its entity is declared unreachable or because it has not been refreshed in over thirty minutes.
_Avoid_: Outdated, Expired, Obsolete