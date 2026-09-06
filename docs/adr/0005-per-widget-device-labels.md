# 5. Renommage des devices par widget plutôt que globalement

Date: 2026-09-04

## Status

Accepted

## Context

Les noms d'entités remontés par Home Assistant sont souvent trop techniques ou trop longs pour être lus d'un coup d'œil sur un Widget. Le dashboard doit donc pouvoir les renommer.

Un mécanisme de renommage existe déjà : `DevicePlacement.CustomName` renomme un Device, mais **par placement sur un Plan**, et ne concerne que la vue isométrique.

L'alternative sérieuse était une table globale associant un libellé à chaque Device, partagée par les Overview Dashboards et la vue isométrique. Elle aurait permis de renommer une fois pour toutes, mais imposait de trancher immédiatement le sort de `CustomName`, de migrer les données existantes, et de perdre la possibilité d'adapter le libellé à la place disponible.

## Decision

Le libellé de renommage vit dans la configuration du Widget qui affiche l'entité. Deux Widgets peuvent nommer le même Device différemment, selon la taille dont ils disposent.

`DevicePlacement.CustomName` reste inchangé et conserve sa portée : la vue isométrique.

## Consequences

- Deux mécanismes de renommage coexistent délibérément, avec des portées distinctes ; ce n'est pas une duplication accidentelle.
- Aucune table ni migration n'est nécessaire pour cette fonctionnalité.
- Renommer le même Device sur plusieurs dashboards se fait autant de fois qu'il y a de Widgets, ce qui devient pénible au-delà de quelques dizaines d'entités réutilisées.
- Basculer plus tard vers un renommage global obligera à migrer des libellés dispersés dans les configurations de Widgets, et non une colonne unique.
