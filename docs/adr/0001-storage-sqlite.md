# 1. Utilisation de SQLite pour le stockage local

Date: 2026-09-02

## Status

Accepted

## Context

L'application doit persister la définition des niveaux (niveaux intérieurs, jardins extérieurs, ordonnancement), le tracé des plans (murs, zones vectorielles), les placements de devices et la liste des automatisations favorites.
Une base de données PostgreSQL a été envisagée mais ajouterait une dépendance externe lourde non requise pour un dashboard local embarqué.

## Decision

Utiliser SQLite pour la persistance locale via un adaptateur outbound dédié. Les schémas et migrations seront appliqués automatiquement au démarrage du backend.

## Consequences

- Déploiement mono-binaire ou conteneur autonome sans service PostgreSQL additionnel.
- Sauvegarde et restauration simplifiées (simple fichier de base de données).
- Concurrence en écriture limitée mais largement suffisante pour un usage de configuration de dashboard domestique.
