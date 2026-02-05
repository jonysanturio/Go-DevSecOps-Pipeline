# GoStock: Hexagonal Inventory API

![Build Status](https://github.com/TU_USUARIO/TU_REPO/actions/workflows/ci.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/TU_USUARIO/TU_REPO)](https://goreportcard.com/report/github.com/TU_USUARIO/TU_REPO)
[![Security Rating](https://img.shields.io/badge/Security-A+-green)](https://github.com/securego/gosec)

API de alto rendimiento para gestión de inventario, diseñada con principios de **Arquitectura Hexagonal**, **DevSecOps** y **Clean Code**.

## Características Técnicas

* **Arquitectura:** Puertos y Adaptadores (Hexagonal) para desacople total.
* **Database:** PostgreSQL nativo.
* **Infraestructura:** Docker & Docker Compose (Multi-stage builds ~10MB).
* **CI/CD:** GitHub Actions con pipeline de seguridad (Govulncheck + Unit Tests).
* **Seguridad:** Escaneo estático de vulnerabilidades y contenedores `scratch` (Distroless).

## Arquitectura del Sistema

```mermaid
graph TD
    A[Cliente / HTTP] -->|Handler| B(Service / Core Logic)
    B -->|Port| C{Repository Interface}
    C -->|Adapter| D[Postgres Implementation]
    D --> E[(Database)]
