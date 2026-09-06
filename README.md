# Walldash

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](go.mod)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)](frontend/package.json)
[![Three.js](https://img.shields.io/badge/Three.js-r185-black?logo=threedotjs)](frontend/package.json)

**Walldash** brings your house to life in live 3D. The touch-first home automation dashboard for Home Assistant — zero Blender, zero YAML. Built from the ground up for a seamless, tactile experience on wall-mounted touchscreens and tablets.

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
  - Ultra-lightweight Docker image (~30 MB Alpine runtime).

---

## 🚀 Quick Start

### Using Docker Compose (Recommended)

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

## 🏗️ Architecture & Tech Stack

```
walldash/
├── cmd/             # Application entrypoint
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
