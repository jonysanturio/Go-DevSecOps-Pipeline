# GoStock: Hexagonal Inventory API

[![Go DevSecOps Pipeline](https://github.com/jonysanturio/Go-DevSecOps-Pipeline/actions/workflows/ci.yml/badge.svg)](https://github.com/jonysanturio/Go-DevSecOps-Pipeline/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonysanturio/Go-DevSecOps-Pipeline)](https://goreportcard.com/report/github.com/jonysanturio/Go-DevSecOps-Pipeline)
[![Security Rating](https://img.shields.io/badge/Security-A+-green)](https://github.com/securego/gosec)

## Project Objective
This project is not just an inventory manager; it is a **Software Architecture Blueprint** designed to demonstrate how to build modern, scalable, and secure software systems in Go.

The main goal is to implement a robust **RESTful API** following industry best practices:
1.  **Decoupling:** Using **Hexagonal Architecture** (Ports and Adapters) to isolate the domain from the infrastructure.
2.  **Observability:** Full instrumentation with **Prometheus** and **Grafana** for real-time metrics.
3.  **Security:** DevSecOps pipeline including vulnerability scanning and distroless containers.

## Getting Started

The entire infrastructure (API, Database, Monitoring System) is orchestrated with **Docker Compose**. Local installation of Go or Postgres is not required.

### Prerequisites
* Docker & Docker Compose installed.

### Steps
1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/jonysanturio/Go-DevSecOps-Pipeline.git](https://github.com/jonysanturio/Go-DevSecOps-Pipeline.git)
    cd Go-DevSecOps-Pipeline
    ```

2.  **Start the infrastructure:**
    This command builds the API, creates the database, and spins up the monitoring system.
    ```bash
    docker compose up --build
    ```

3.  **Access the services:**
    * **API Documentation (Swagger):** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
    * **Monitoring Dashboard (Grafana):** [http://localhost:3000](http://localhost:3000) *(User: admin / Pass: secret)*
    * **Metrics (Prometheus):** [http://localhost:9090](http://localhost:9090)

## What are we measuring?
Thanks to the Prometheus integration, the system automatically monitors the **Golden Signals** to ensure reliability:
* **HTTP Latency:** Response time for each endpoint.
* **Saturation:** Heap Memory usage and active Goroutines (to detect memory leaks).
* **Traffic:** Requests per second (RPS) and status codes (200 vs 500).

## Tech Stack

* **Language:** Golang 1.25+
* **Architecture:** Ports and Adapters (Hexagonal).
* **Database:** PostgreSQL 15 (Native driver `lib/pq`).
* **Infrastructure:** Docker & Docker Compose (Multi-stage builds ~15MB).
* **Observability:** Prometheus (Metrics), Grafana (Dashboard), Swagger (Docs).
* **CI/CD:** GitHub Actions (Govulncheck + Unit Tests + Linter).

## System Architecture

```mermaid
graph TD
    User((User)) -->|HTTP Request| API[Go API REST]
    API -->|Metrics| Prom{Prometheus}
    Prom -->|Data| Graf[Grafana Dashboard]
    
    subgraph "Hexagonal Core"
        API -->|Handler| Service(Service / Business Logic)
        Service -->|Port| Repo{Repository Interface}
        Repo -->|Adapter| DB[(PostgreSQL)]
    end