# 2. Ségrégation stricte des actions et filtrage des mutations

Date: 2026-09-02

## Status

Accepted

## Context

L'interface frontend communique uniquement avec le backend Go et ne doit jamais avoir d'accès direct ni d'accès sans restriction à Home Assistant.
Le mode utilisateur ne doit avoir aucune capacité de modifier la configuration, ni d'exécuter des actions arbitraires sur Home Assistant (ex: changer des paramètres, envoyer des payloads non contrôlés).

## Decision

Le backend applique une politique stricte de whitelisting des commandes :
1. Les seules actions actionnables permises pour le mode utilisateur sont :
   - `Toggle / TurnOn / TurnOff` pour les actionneurs (`light`, `switch`, `media_player` basique pour marche/arrêt, relais d'arrosage).
   - `Trigger` pour les automatisations (`automation.trigger`).
2. Aucune action de mise à jour arbitraire de configuration (dimmer, couleurs, renommage, mise à jour d'entité) n'est acceptée par l'API backend.
3. Les modifications structurelles (création de plans, ajout/déplacement de devices, favoris) sont cantonnées à l'espace d'administration.

## Consequences

- Sécurité garantie même si un client malveillant sur le réseau local tente d'émettre des messages forgés sur le WebSocket ou l'API.
- Modèle de service simplifié dans l'architecture hexagonale : les commandes du domaine (`service.Commands`) n'exposent que des méthodes ciblées (`ToggleDevice`, `TriggerAutomation`, etc.).
