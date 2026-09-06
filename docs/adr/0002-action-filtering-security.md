# 2. Strict action segregation and mutation filtering

Date: 2026-09-02

## Status

Accepted

## Context

The frontend interface communicates exclusively with the Go backend and must never have direct or unrestricted access to Home Assistant.
User mode must have no ability to modify configuration, nor execute arbitrary actions on Home Assistant (e.g. changing settings, sending uncontrolled payloads).

## Decision

The backend enforces a strict command whitelisting policy:
1. The only permitted actionable operations for user mode are:
   - `Toggle / TurnOn / TurnOff` for actuators (`light`, `switch`, basic `media_player` for on/off, irrigation relay).
   - `Trigger` for automations (`automation.trigger`).
2. No arbitrary configuration update actions (dimmer, colors, renaming, entity updates) are accepted by the backend API.
3. Structural changes (plan creation, device addition/movement, favorites) are confined to the administration space.

## Consequences

- Security guaranteed even if a malicious client on the local network attempts to send forged messages over WebSocket or the API.
- Simplified service model in the hexagonal architecture: domain commands (`service.Commands`) only expose targeted methods (`ToggleDevice`, `TriggerAutomation`, etc.).