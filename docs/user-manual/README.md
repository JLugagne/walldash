# Walldash User Manual

The user manual, built with [Hugo](https://gohugo.io/) and published to GitHub
Pages.

- **Content** lives in `content/`, one Markdown file per page.
- **Layout & theme** are a small in-repo theme under `layouts/` and `assets/`.
- **Screenshots** live in `static/images/` and are generated, not hand-made.
- **Navigation** is declared in `data/menu.yaml`.

## Preview locally

```bash
cd docs/user-manual
hugo server
```

Then open <http://localhost:1313/walldash/>.

## Build

```bash
cd docs/user-manual
hugo --gc --minify
```

The static site is written to `public/` (git-ignored).

## Regenerating the screenshots

Screenshots are produced by the Playwright harness in `tools/`. It boots the
Walldash backend in **demo mode** (no `HA_URL` / `HA_TOKEN`, so it uses the
built-in example devices), imports a Sweet Home 3D plan, seeds a few device
placements and a dashboard, then drives the UI at an iPad-landscape
resolution (1194×834).

```bash
cd docs/user-manual/tools
npm install
npx playwright install chromium       # if not already present

SH3D_FILE=/path/to/plan.sh3d npm run generate
```

- `SH3D_FILE` — any `.sh3d` plan to use as the demo house. The screenshot set in
  this repository was produced from a generic example model.
- `MANUAL_PORT` — backend port for the demo server (default `18080`).
- Pass `--skip-build` to reuse the existing backend binary and
  `frontend/dist` build.
- Pass `--reuse-db` to reuse the already-seeded `.work/manual.db` instead of
  importing a `.sh3d` plan: it rebuilds the showcase dashboards and re-captures
  every screenshot except the onboarding wizard. Useful when the demo plan is
  not at hand.

The generated PNGs are written to `../static/images/`. **No real Home Assistant
installation is ever contacted** — the backend runs from an isolated working
directory with an empty Home Assistant configuration, so the repository `.env`
is never read.

## Deployment

`.github/workflows/manual.yml` builds the site and deploys it to GitHub Pages on
every push to `main` that touches this folder. Enable **Settings → Pages →
Build and deployment → Source: GitHub Actions** once in the repository.
