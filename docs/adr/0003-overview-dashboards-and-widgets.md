# 3. Summary dashboards ("Overview") and widget system

Date: 2026-09-02

## Status

Accepted. Point 2 is amended by ADR 0004: the type and display mode of a Widget are two separate axes, not a single composite identifier.

## Context

The user wishes to be able to define, in administration mode, "Overview" dashboards (summary views outside the 3D view) in which they configure and arrange custom widgets (notably to group sensors and precisely select which favorite automations to display and control).

## Decision

Model a widget composition subsystem:
1. An `OverviewDashboard` entity has an identifier, a name, and an ordered collection of `Widget`.
2. Each `Widget` has a type describing what it is linked to and what a tap triggers (`sensor`, `actuator`, `automation_list`), a Display Mode describing its rendering, and a dedicated configuration (e.g. explicit list of selected automation IDs).
3. Administration allows creating/editing Overviews and their widgets.
4. The user can switch between the building's 3D view and the Overview Dashboards.

## Consequences

- Great display flexibility for the touch tablet without overloading the 3D view of the house.
- Extensible data model to add future widget types without altering the 3D geometric core.