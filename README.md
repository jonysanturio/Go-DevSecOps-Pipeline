# GoStock: Hexagonal Inventory API

[![Go DevSecOps Pipeline](https://github.com/jonysanturio/Go-DevSecOps-Pipeline/actions/workflows/ci.yml/badge.svg)](https://github.com/jonysanturio/Go-DevSecOps-Pipeline/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonysanturio/Go-DevSecOps-Pipeline)](https://goreportcard.com/report/github.com/jonysanturio/Go-DevSecOps-Pipeline)
[![Security Rating](https://img.shields.io/badge/Security-A+-green)](https://github.com/securego/gosec)

## Objetivo del Proyecto
Este proyecto no es solo un gestor de inventario; es un **Blueprint (Plantilla de Arquitectura)** diseñado para demostrar cómo construir sistemas de software modernos, escalables y seguros en Go.

El objetivo principal es implementar una **API RESTful** robusta siguiendo las mejores prácticas:
1.  **Desacoplamiento:** Uso de Arquitectura Hexagonal para aislar el dominio de la infraestructura.
2.  **Observabilidad:** Instrumentación completa con Prometheus y Grafana para métricas en tiempo real.
3.  **Seguridad:** Pipeline de DevSecOps con escaneo de vulnerabilidades y contenedores distroless.

## Cómo levantar el proyecto

Toda la infraestructura (API, Base de Datos, Sistema de Métricas) está orquestada con **Docker Compose**. No es necesirio instalar Go ni Postgres localmente.

### Prerrequisitos
* Docker & Docker Compose instalados.

### Pasos
1.  **Clonar el repositorio:**
    ```bash
    git clone [https://github.com/jonysanturio/Go-DevSecOps-Pipeline.git](https://github.com/jonysanturio/Go-DevSecOps-Pipeline.git)
    cd Go-DevSecOps-Pipeline
    ```

2.  **Levantar la infraestructura:**
    Este comando compila la API, crea la base de datos y levanta el sistema de monitoreo.
    ```bash
    docker compose up --build
    ```

3.  **Acceder a los servicios:**
    * **Documentación API (Swagger):** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
    * **Tablero de Control (Grafana):** [http://localhost:3000](http://localhost:3000) *(User: admin / Pass: secret)*
    * **Métricas (Prometheus):** [http://localhost:9090](http://localhost:9090)

## ¿Qué medimos con las métricas?
Gracias a la integración con Prometheus, el sistema monitorea automáticamente las **Golden Signals** para asegurar la fiabilidad:
* **Latencia HTTP:** Tiempo de respuesta de cada endpoint.
* **Saturación:** Consumo de Memoria Heap y Goroutines activas (detección de memory leaks).
* **Tráfico:** Cantidad de peticiones por segundo (RPS) y códigos de estado (200 vs 500).

## Stack Tecnológico

* **Lenguaje:** Golang 1.21+
* **Arquitectura:** Puertos y Adaptadores (Hexagonal).
* **Database:** PostgreSQL 15 (Driver nativo `lib/pq` o `pgx`).
* **Infraestructura:** Docker & Docker Compose (Multi-stage builds ~15MB).
* **Observabilidad:** Prometheus (Metrics), Grafana (Dashboard), Swagger (Docs).
* **CI/CD:** GitHub Actions (Govulncheck + Unit Tests + Linter).

## Arquitectura del Sistema

```mermaid
graph TD
    User((Usuario)) -->|HTTP Request| API[Go API REST]
    API -->|Metrics| Prom{Prometheus}
    Prom -->|Data| Graf[Grafana Dashboard]
    
    subgraph "Hexagonal Core"
        API -->|Handler| Service(Service / Business Logic)
        Service -->|Port| Repo{Repository Interface}
        Repo -->|Adapter| DB[(PostgreSQL)]
    end
