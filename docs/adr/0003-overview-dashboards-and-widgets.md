# 3. Tableaux de bord synthétiques ("Overview") et système de widgets

Date: 2026-09-02

## Status

Accepted. Le point 2 est amendé par l'ADR 0004 : le type et le mode de rendu d'un Widget sont deux axes distincts, et non un identifiant composé unique.

## Context

L'utilisateur souhaite pouvoir définir en mode administration des tableaux de bord "Overview" (vues synthétiques hors vue 3D) dans lesquels il configure et dispose des widgets personnalisés (notamment pour grouper des capteurs et cocher précisément les automatisations favorites à afficher et piloter).

## Decision

Modéliser un sous-système de composition de widgets :
1. Une entité `OverviewDashboard` possède un identifiant, un nom et une collection ordonnée de `Widget`.
2. Chaque `Widget` possède un type décrivant ce à quoi il est lié et ce qu'un appui déclenche (`sensor`, `actuator`, `automation_list`), un Display Mode décrivant son rendu, et une configuration dédiée (ex: liste explicite des IDs d'automatisations sélectionnées).
3. L'administration permet la création/édition des Overviews et de leurs widgets.
4. L'utilisateur peut basculer entre la vue 3D du bâtiment et les Overview Dashboards.

## Consequences

- Grande flexibilité d'affichage pour la tablette tactile sans surcharger la vue 3D de la maison.
- Modèle de données extensible pour ajouter de futurs types de widgets sans altérer le coeur géométrique 3D.
