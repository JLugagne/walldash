# Home Assistant 3D Dashboard (ha-dash)

Manager et visualiseur de dashboard domotique 3D isométrique pour Home Assistant, optimisé pour tablette tactile.

## Language

### Espaces & Plans

**Level**:
Représente un étage physique du bâtiment ou une zone extérieure (ex: jardin avant, jardin arrière).
_Avoid_: Floor (trop restrictif pour les extérieurs), Layer, Stage

**Plan**:
Le tracé 2D vectoriel constitutif d'un Level, composé de murs, pièces, ouvertures et délimitations de zones extérieures.
_Avoid_: Blueprint, Map, Drawing

**Zone**:
Une subdivision délimitée au sein d'un Plan (pièce intérieure ou sous-zone de jardin).
_Avoid_: Room (les jardins ne sont pas des pièces), Area

### Équipements & Entités

**Device Placement**:
L'ancrage d'un équipement Home Assistant sur un Plan à des coordonnées 2D (x, y).
_Avoid_: Device Mapping, Pin, Widget

**Device**:
Un équipement domotique physique ou virtuel rattaché à Home Assistant (ex: lampe, prise, TV, capteur, relais arrosage), filtré et normalisé par le backend.
_Avoid_: Entity (terme interne HA), Item

**Sensor**:
Un Device passif en lecture seule dont les mesures (température, humidité, etc.) sont affichées en permanence.
_Avoid_: Meter, Gauge

**Actuator**:
Un Device pilotable en écriture (on/off, switch, relais, power) sans contrôle de nuanceur (dimmer/couleurs) dans la version actuelle.
_Avoid_: Switch, Controller

**Light Halo**:
L'effet visuel lumineux affiché dans la vue 3D autour d'un Device de type lampe lorsqu'il est allumé.
_Avoid_: Glow, Flare

### Interactions & Automatismes

**Automation**:
Une automatisation définie dans Home Assistant pouvant être déclenchée à la demande et suivie (état d'exécution, heure de dernière exécution).
_Avoid_: Scenario, Script, Routine

**Favorite Automation**:
Une Automation sélectionnée par l'utilisateur pour figurer dans la vue rapide du dashboard.
_Avoid_: Quick Action, Shortcut

### Vues & Rendu

**Isometric View**:
La vue 3D à projection isométrique fixe avec caméra orientée de sorte que le bas du Plan 2D corresponde à la façade avant de la maison (sans rotation libre).
_Avoid_: 3D Orbit, Free Cam, Perspective

### Sécurité & Opérations

**Action Whitelist**:
L'ensemble restreint d'opérations exécutables par le dashboard vers Home Assistant (exclusivement commutations marche/arrêt et déclenchement d'automatisations).
_Avoid_: Command Pass-through, Generic Service Call

### Tableaux de bord & Vues synthétiques

**Overview Dashboard**:
Une vue synthétique personnalisée composée d'une grille de widgets configurables (état des capteurs, interrupteurs d'appareils, listes d'automatisations sélectionnées).
_Avoid_: Dashboard View, Summary Panel

**Widget**:
Un bloc interactif unitaire positionné sur un Overview Dashboard (ex: Widget Automations filtré, Widget Capteur Climat, Widget Éclairage).
_Avoid_: Tile, Card, Component
