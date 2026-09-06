# 1. Use of SQLite for local storage

Date: 2026-09-02

## Status

Accepted

## Context

The application must persist level definitions (interior levels, outdoor gardens, ordering), plan layouts (walls, vector zones), device placements, and the list of favorite automations.
A PostgreSQL database was considered but would add a heavy external dependency not required for an embedded local dashboard.

## Decision

Use SQLite for local persistence via a dedicated outbound adapter. Schemas and migrations will be applied automatically at backend startup.

## Consequences

- Single-binary deployment or standalone container without an additional PostgreSQL service.
- Simplified backup and restore (a single database file).
- Limited write concurrency but more than sufficient for home dashboard configuration usage.