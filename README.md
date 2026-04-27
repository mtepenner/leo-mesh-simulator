# LEO Satellite Mesh Network Routing Simulator

A high-fidelity simulation engine for modeling Low Earth Orbit (LEO) satellite constellations, dynamic inter-satellite link (ISL) routing, and network topology visualization.

## Table of Contents
- [Features](#features)
- [Architecture](#architecture)
- [Technologies](#technologies)
- [Installation](#installation)
- [License](#license)

## 🚀 Features
- **Orbital Kinematics**: Tracks 500+ satellites using SGP4 propagation.
- **Dynamic Topology**: Real-time raycasting to calculate line-of-sight and link availability.
- **Adaptive Routing**: Implements dynamic Dijkstra pathfinding to calculate optimal routes across the mesh.
- **3D Visualization**: Holographic "Operations Center" interface using React-Three-Fiber to display live satellite nodes and laser links.

## 🏗️ Architecture
The system is orchestrated through a high-speed data pipeline:
1.  **Orbital Engine (Go)**: The "Brain" that calculates satellite positions and publishes the adjacency matrix via Redis at 10Hz.
2.  **Routing Daemon (Python)**: Subscribes to the Redis stream to update graph weights and compute routing paths.
3.  **Mesh Dashboard (React)**: Consumes live data to render the Earth-scale mesh in real-time.

## 🛠️ Technologies
- **Simulation**: Go (SGP4 kinematics, Redis integration).
- **Routing**: Python, FastAPI, NetworkX, NumPy.
- **Frontend**: React, TypeScript, Three.js, React-Three-Fiber.
- **Infrastructure**: Redis Cluster, Docker Compose.

## 📥 Installation
1. Clone the repository: `git clone https://github.com/mtepenner/leo-mesh-simulator.git`
2. Spin up the environment: `docker-compose up`

## ⚖️ License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
