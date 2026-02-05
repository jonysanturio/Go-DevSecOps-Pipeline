# GOSTOCK: Hexagonal Inventory API

Este proyecto es una API RESTful completa para la gestion de inventarios de productos, construida desde cero con **Go (Golang)**, **PostgreSQL** y **Docker**

Implementa una **Arquitectura Hexagonal (Puertos y Adaptadores)** para desacoplar la lógica de negocio de la base de datos y el transporte HTTP.

# TECNOLOGÍAS

* **Lenguaje:** Go 1.23+
* **Base de Datos:** PostgreSQL
* **Infraestructura:** Docker & Docker Compose
* **Librerías:** `net/http` (Standard Librery), `lib/pq` (Driver Postgres).

El proyecto sigue una estructura por capas para asegurar mantenibilidad:

1.  **Handler (`cmd/api/handler.go`):** Capa de transporte. Recibe HTTP, valida JSON y recorta URLs.
2.  **Service (`internal/product/service.go`):** Capa de negocio. Contiene la lógica pura y validaciones (ej: precio no negativo).
3.  **Repository (`internal/platform/postgres`):** Capa de datos. Ejecuta sentencias SQL puras (`Query`, `Exec`).
4.  **Domain (`internal/domain`):** El corazón. Define los structs (`Product`) y las interfaces (`Repository`).

## Como correr el proyecto

1. Clonar el repositorio
2. Ejecutar con Docker Compose:
    ```bash
    docker compose up --build 
3. La API estará disponible en http://localhost:8080