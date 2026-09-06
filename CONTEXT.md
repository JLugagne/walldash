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

**Display Layer**:
Une couche d'affichage logique exclusive propre à un Level, permettant de filtrer les Device Placements visibles dans la vue 3D (par défaut : `controls` et `sensors`, extensible par l'utilisateur).
_Avoid_: Layer (sans qualificatif, pour éviter la confusion avec Level), Filter

**Ceiling Display**:
Le marquage d'informations (mesures de capteurs ou nom de zone) projeté sur un plan horizontal 3D à hauteur de plafond (`y = WALL_HEIGHT`) au pôle d'inaccessibilité de chaque Zone, suivant fidèlement les fuyantes et la perspective 3D.
_Avoid_: Floor Print, Ceiling HUD, Floating Tag

**Zone Alert Pulse**:
L'effet visuel de respiration sinusoïdale continue et d'interpolation chromatique progressive (bleu vers le froid, rouge vers le chaud) appliqué au texte du Ceiling Display lorsqu'une mesure franchit les seuils configurés pour la Zone.
_Avoid_: Blink, Flash, Glow

### Sécurité & Opérations

**Action Whitelist**:
L'ensemble restreint d'opérations exécutables par le dashboard vers Home Assistant (exclusivement commutations marche/arrêt et déclenchement d'automatisations).
_Avoid_: Command Pass-through, Generic Service Call

### Tableaux de bord & Vues synthétiques

**Overview Dashboard**:
Une vue synthétique personnalisée composée d'une grille de widgets configurables (état des capteurs, interrupteurs d'appareils, listes d'automatisations sélectionnées).
_Avoid_: Dashboard View, Summary Panel

**Widget**:
Un bloc unitaire ancré dans la Widget Grid d'un Overview Dashboard, lié à un ou plusieurs Devices ou Automations et rendu selon un Display Mode.
_Avoid_: Tile, Card, Component

**Widget Grid**:
La trame fixe de colonnes et de lignes propre à un Overview Dashboard, qui occupe exactement la fenêtre et dans laquelle chaque Widget est ancré.
_Avoid_: Canvas, Layout, Board

**Display Mode**:
La façon dont un Widget dessine la donnée à laquelle il est lié, indépendante de la nature de cette donnée.
_Avoid_: Style, Skin, Renderer, Variant

**Arc**:
Le Display Mode en cadran ouvert vers le bas représentant une mesure située entre deux bornes saisies par l'administrateur.
_Avoid_: Gauge, Jauge, Fer à cheval, Dial

**Primary Action**:
L'unique opération déclenchée par un appui sur un Widget lié à un Actuator. Un Widget lié à un Sensor n'en possède aucune.
_Avoid_: Default Action, Main Command, Tap Action

**Edit Mode**:
L'état d'un Overview Dashboard dans lequel un administrateur compose la disposition de ses Widgets, par opposition à l'usage courant où un appui ne fait que déclencher une Primary Action.
_Avoid_: Admin Mode, Design Mode, Layout Mode

**Stale Value**:
Une mesure qu'un Widget continue d'afficher alors qu'elle n'est plus digne de confiance, soit parce que son entité est déclarée injoignable, soit parce qu'elle n'a pas été rafraîchie depuis plus de trente minutes.
_Avoid_: Outdated, Expired, Obsolete
