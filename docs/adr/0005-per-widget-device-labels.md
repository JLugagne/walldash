# 5. Per-widget device renaming rather than globally

Date: 2026-09-04

## Status

Accepted

## Context

Entity names returned by Home Assistant are often too technical or too long to be read at a glance on a Widget. The dashboard must therefore be able to rename them.

A renaming mechanism already exists: `DevicePlacement.CustomName` renames a Device, but **per placement on a Plan**, and applies only to the isometric view.

The serious alternative was a global table associating a label with each Device, shared by Overview Dashboards and the isometric view. It would have allowed renaming once and for all, but required immediately deciding the fate of `CustomName`, migrating existing data, and losing the ability to adapt the label to available space.

## Decision

The rename label lives in the configuration of the Widget that displays the entity. Two Widgets can name the same Device differently, depending on the space available to them.

`DevicePlacement.CustomName` remains unchanged and keeps its scope: the isometric view.

## Consequences

- Two renaming mechanisms deliberately coexist, with distinct scopes; this is not accidental duplication.
- No table or migration is needed for this feature.
- Renaming the same Device across multiple dashboards must be done as many times as there are Widgets, which becomes tedious beyond a few dozen reused entities.
- Switching to global renaming later will require migrating labels scattered across Widget configurations, rather than a single column.