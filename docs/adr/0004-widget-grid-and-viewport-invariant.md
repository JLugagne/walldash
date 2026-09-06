# 4. Grille de widgets à coordonnées entières et invariant de non-débordement

Date: 2026-09-04

## Status

Accepted

## Context

Le dashboard est destiné à une tablette murale : il doit occuper exactement la fenêtre, sans barre de défilement, et aucun Widget ne doit jamais sortir du viewport. Le modèle issu de l'ADR 0003 ne portait qu'un `Order int` par Widget, insuffisant pour un déplacement et un redimensionnement libres.

Deux alternatives ont été écartées. Des positions absolues normalisées (`x, y, w, h` en flottants de 0 à 1) autorisent un placement au pixel près, mais obligent à revalider le confinement à chaque redimensionnement de fenêtre et produisent des cibles tactiles de tailles arbitraires. Un flux responsive piloté par le CSS ne permet pas de redimensionner un Widget.

## Decision

Ancrer chaque Widget dans une Widget Grid à coordonnées entières (`col`, `row`, `col_span`, `row_span`). Les dimensions de la grille sont persistées sur chaque Overview Dashboard, avec 12 × 8 par défaut et sans écran d'édition en v1.

Le non-débordement devient une propriété du modèle, vérifiée dans le domaine : `col + col_span <= cols` et `row + row_span <= rows`. Chaque Display Mode déclare en outre une taille minimale, appliquée à la création comme au redimensionnement.

Le déplacement se fait par poussée des Widgets voisins dans la direction du geste, appliquée en tout ou rien : le réarrangement complet est calculé avant d'être appliqué, et si son résultat ne tient pas intégralement dans la grille, le dépôt est refusé et rien ne bouge.

Les positions issues d'une poussée sont écrites par une route de layout dédiée, en une seule transaction, distincte de la mise à jour du contenu d'un Widget.

## Consequences

- L'invariant « aucun Widget hors du viewport » ne peut pas être perdu par oubli d'une validation d'interface : il est faux ou vrai dans le domaine.
- La grille étant persistée par dashboard, modifier la valeur par défaut ne réarrange pas silencieusement les dashboards déjà composés.
- Le tout ou rien peut refuser un geste que l'utilisateur croyait réalisable ; c'est le prix de l'invariant, et le refus doit donc être signalé visuellement pendant le survol et non seulement au lâcher.
- Une poussée dirigée par le geste n'est pas commutative : deux trajectoires aboutissant à la même cellule peuvent produire deux dispositions différentes. Accepté au profit du naturel du geste tactile.
- Les cellules ne sont pas carrées et leur rapport varie avec l'orientation de la tablette ; les Display Modes circulaires doivent se dessiner dans un carré centré à l'intérieur de leur cellule plutôt que remplir celle-ci.
