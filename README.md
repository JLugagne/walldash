# Walldash

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](go.mod)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)](frontend/package.json)
[![Three.js](https://img.shields.io/badge/Three.js-r185-black?logo=threedotjs)](frontend/package.json)

**Walldash** brings your house to life in live 3D. The touch-first home automation dashboard for Home Assistant — zero Blender, zero YAML. Built from the ground up for a seamless, tactile experience on wall-mounted touchscreens and tablets.

---

## 📸 Screenshots

![3D Controls view - 1st floor with interactive light toggles](docs/screenshots/controls.png)
*Live 3D Controls view — isometric floor with per-room ambient halos and touch-friendly light / switch toggles.*

![3D Sensors view - temperature and humidity overlays](docs/screenshots/sensors.png)
*Live 3D Sensors view — in-space temperature gauges with comfort status plus humidity readouts.*

![2D Floor Plan Editor - zones, devices and room properties](docs/screenshots/devices.png)
*In-browser 2D Floor Plan Editor — trace walls, define zones across floors, place devices, and tune room color and ambient sensor thresholds.*

---

## ✨ Features

- **🎮 3D Isometric View**:
  - Fixed isometric camera projection built with **Three.js** and **React Three Fiber**.
  - Dynamic room geometry with ceiling status discs and ambient lighting halos around active lights.
  - Interactive device toggles (lights, switches, appliances) and real-time sensor readouts directly inside the 3D space.

- **✏️ 2D Floor Plan Editor**:
  - In-browser vector plan editor to trace walls, define interior rooms and outdoor zones (gardens, patios).
  - Multi-level management (ground floor, upper floors, basements, outdoor areas).
  - Drag-and-drop device placement with real-time coordinate mapping onto the 3D view.

- **📊 Overview Dashboards & Widgets**:
  - Touch-friendly widget grid layouts for quick actions and monitoring.
  - Dedicated widgets for favorite automations, dials/gauges, toggle switches, and sensor statistics.
  - Touch-drag reordering in edit mode.

- **⚡ Real-Time WebSocket Synchronization**:
  - Instant bi-directional communication between Home Assistant, backend, and all connected wall panels.
  - Device states update immediately without manual page refreshes.

- **📦 Zero-Dependency Single Binary**:
  - Pure Go backend using `modernc.org/sqlite` (no CGO required).
  - Frontend assets compiled and embedded directly into the Go binary (`embed.FS`).
  - Home Assistant base image with s6-overlay: the same image runs standalone and as an official-style add-on.

---

## 🚀 Quick Start

### As a Home Assistant Add-on (Recommended)

1. Add the add-on repository to Home Assistant: **Settings** → **Add-ons** → **Add-on Store** → menu (⋮) → **Repositories**, then add `https://github.com/JLugagne/ha-addons`.
2. Install **Walldash** from the store and start it. No token setup is required: the add-on connects to Home Assistant through the Supervisor API automatically.
3. Open the Walldash web UI at `http://<home-assistant-ip>:8080` on your wall tablets. No Home Assistant login is needed on the tablets (direct port access, no Ingress).
4. Data is stored in the add-on `/data` volume and survives updates and reboots.

See [`docs/home-assistant-add-on.md`](docs/home-assistant-add-on.md) for packaging details, the release process, and the manual token fallback.

### Using Docker Compose (Alternative)

1. Clone the repository:
   ```bash
   git clone https://github.com/JLugagne/walldash.git
   cd walldash
   ```

2. Create your `.env` file from the example:
   ```bash
   cp .env.example .env
   ```

3. Fill in your Home Assistant details in `.env`:
   ```env
   HA_URL=http://homeassistant.local:8123
   HA_TOKEN=your_long_lived_access_token_here
   PORT=8080
   DB_PATH=/app/data/walldash.db
   ```

4. Launch with Docker Compose:
   ```bash
   docker compose up -d
   ```

5. Open your browser or tablet display at `http://<server-ip>:9090` (or `8080` depending on port mapping).

---

## 🔑 Home Assistant Setup

Walldash communicates with Home Assistant through its REST API and WebSocket events.

To generate a Long-Lived Access Token:
1. Open your Home Assistant web interface.
2. Click on your user profile icon (bottom left).
3. Scroll down to the **Security** tab > **Long-Lived Access Tokens**.
4. Click **Create Token**, give it a name (e.g., `Walldash`), and copy the generated token into your `.env` file.

---

## ⚙️ Configuration Reference

Settings resolve with the following precedence: **environment variable** → **add-on options file** (`/data/options.json`) → **default**.

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port the server listens on. |
| `DB_PATH` | `walldash.db` (or `/data/walldash.db` when the `/data` volume exists) | SQLite database path. |
| `HA_URL` | `http://homeassistant.local:8123` | Home Assistant base URL. |
| `HA_TOKEN` | _(empty)_ | Long-lived access token. Not needed when running as an add-on: `SUPERVISOR_TOKEN` is used automatically via the Supervisor API proxy. The Supervisor token is never attached to a custom `HA_URL`. |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error`. |
| `ALLOWED_ORIGINS` | _(empty)_ | Comma-separated CORS origins. |
| `FRONTEND_DIR` | _(embedded assets)_ | Serve the frontend from a directory instead of the embedded build. |

---

## 🌐 Custom Domain (Reverse Proxy)

To reach Walldash at `https://walldash.domain.tld` with automatic HTTPS, see
[`docs/reverse-proxy.md`](docs/reverse-proxy.md) (Nginx Proxy Manager setup,
works for both Docker Compose and the add-on — WebSocket support required).

---

## 🏗️ Architecture & Tech Stack

```
walldash/
├── cmd/             # Application entrypoint (+ add-on runtime helpers)
├── rootfs/          # s6-overlay service definitions for the add-on image
├── internal/
│   └── dashboard/   # Hexagonal architecture (Domain, App, Inbound, Outbound)
│       ├── domain/  # Core business models, interfaces, domain errors
│       ├── app/     # Application logic & use-cases
│       ├── inbound/ # HTTP handlers (Gorilla Mux), WebSocket hub, CSRF protection
│       └── outbound/# Adapters for SQLite and Home Assistant API
├── pkg/dashboard/   # Public API contracts and DTO types
└── frontend/        # React 19 SPA (Vite, TypeScript, Three.js, React Three Fiber, Tailwind CSS v4)
```

- **Backend**: Go (Go 1.26+), Hexagonal / Clean Architecture, Gorilla Mux, Gorilla WebSocket, modernc SQLite, Logrus.
- **Frontend**: React 19, TypeScript, Three.js, React Three Fiber, Drei, Tailwind CSS v4, Lucide React, Vitest.

---

## 🛠️ Local Development

### Prerequisites
- [Go](https://go.dev/) (1.26+)
- [Node.js](https://nodejs.org/) (22+) and `npm`

### 1. Run the Frontend (Dev server with HMR)
```bash
cd frontend
npm install
npm run dev
```

### 2. Run the Backend
```bash
# From repository root
go run ./cmd/main.go
```

### 3. Run Tests

**Backend tests:**
```bash
go test -race -failfast ./...
```

**Frontend tests:**
```bash
cd frontend
npm test
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
