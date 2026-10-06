# Bitácora de aprendizaje: GoStock

Este documento registra los cambios del proyecto y la lógica detrás de cada uno. La dinámica acordada es que el usuario modifica el código y el asistente explica, revisa y documenta; el asistente no edita código por su cuenta.

## Flujo de una solicitud

1. `cmd/api/main.go` registra rutas HTTP.
2. `ProductHandler` lee parámetros y cuerpo, valida la entrada HTTP y elige una respuesta.
3. El servicio aplica reglas de negocio y llama al puerto `domain.ProductRepository`.
4. El adaptador PostgreSQL implementa ese puerto y ejecuta SQL.
5. El resultado vuelve por el servicio y el handler lo convierte en una respuesta HTTP.

## Pasos trabajados

### 1. Manejo de errores del repositorio

**Archivo:** `internal/platform/postgres/product_repository.go`

- `GetAll` revisa `rows.Err()` después del bucle porque un error del driver puede aparecer durante la iteración.
- `GetOne` traduce `sql.ErrNoRows` a `domain.ErrProductNotFound`, separando el dominio de detalles de PostgreSQL.
- `Update` usa el ID recibido como argumento y placeholders SQL. Los datos del usuario no se concatenan a la consulta.
- `Update` y `Delete` revisan `RowsAffected()`; cero filas significa que el producto no existe.
- Los handlers convierten `ErrProductNotFound` en 404 y otros errores internos en 500.

### 2. Validación del cuerpo de actualización

**Archivo:** `cmd/api/handler.go`

El usuario agregó `updateReq.Validate()` a `UpdateProduct`, igual que en creación. Una solicitud JSON mal formada o con valores inválidos recibe 400 y no llega al servicio.

El patrón `if err := ...; err != nil` crea un error de alcance local y termina temprano con `return`.

### 3. Validación del ID y función común

**Archivo:** `cmd/api/handler.go`

El usuario agregó validación de ID positivo en las operaciones. Después creó `parseProductID(path string)`, que verifica el prefijo `/products/`, extrae el sufijo, lo convierte con `strconv.Atoi` y devuelve error si no es un entero positivo.

Los tres handlers llaman ahora al helper. Esto mantiene una sola regla para GET, PUT y DELETE. El handler convierte el error de parseo en 400; un ID positivo que no existe sigue resultando en 404 desde el servicio.

### 4. Respuestas 500 sin detalles internos

**Archivo:** `cmd/api/handler.go`

El usuario cambió las respuestas de error de `CreateProduct` y `GetAllProducts` a mensajes genéricos. Así el cliente no recibe detalles internos de la base de datos. Más adelante agregaremos logging del lado del servidor para conservar esos detalles para diagnóstico.

## Siguiente paso: métodos HTTP no permitidos

**Archivo:** `cmd/api/main.go`

El dispatcher de `/products` llama a `GetAllProducts` para cualquier método distinto de POST. En `/products/`, el `switch` no tiene `default`; un método no contemplado puede terminar con una respuesta vacía 200.

El siguiente cambio será responder 405 Method Not Allowed para métodos no admitidos y agregar el header `Allow` con los métodos válidos para cada ruta. Esto hace que el contrato HTTP sea explícito y evita que DELETE, PATCH u otros métodos se interpreten accidentalmente como una lectura.

## Temas para etapas siguientes

- Limitar el tamaño del cuerpo JSON, comprobar Content-Type, rechazar campos desconocidos y contenido JSON adicional.
- Unificar las respuestas de error en JSON; hoy `http.Error` responde texto plano y los éxitos usan JSON.
- Agregar logging estructurado y request IDs sin exponer errores internos al cliente.
- Configurar timeouts del servidor y mantener propagación de `r.Context()` a la base.
- Alinear validación de DTO, servicio y dominio.
- Crear migraciones PostgreSQL, reforzar CI/CD, integrar frontend y preparar pruebas de seguridad locales.

No se ejecutaron pruebas ni compilaciones en estos pasos.

## Registro

- 2026-10-02: documentados el flujo inicial del repositorio, la validación del cuerpo de actualización y la validación de IDs.
- 2026-10-02: documentada la reutilización de `parseProductID` en GET, PUT y DELETE.
- 2026-10-02: documentadas las respuestas genéricas 500 para creación y listado; próximo foco, métodos HTTP no permitidos.

## Incidente: versión de Go y análisis con govulncheck

### 1. Qué estaba configurado

El workflow usaba Go 1.25.7:

```yaml
go-version: '1.25.7'
go install golang.org/x/vuln/cmd/govulncheck@latest

Son dos versiones distintas: go-version selecciona el toolchain para el proyecto; @latest selecciona la versión de la herramienta govulncheck.

2. Por qué apareció Go 1.26.8 durante la instalación
En esa ejecución, @latest resolvió a golang.org/x/vuln@v1.8.0, que requiere Go 1.26 o superior. Por eso el comando go install descargó Go 1.26.8 para instalar govulncheck.
Ese mensaje no significa que el proyecto se haya actualizado a Go 1.26. La descarga ocurrió para ejecutar la instalación de la herramienta. La selección automática de toolchain de Go puede cambiar la versión usada para un comando cuando el módulo de ese comando requiere una versión posterior. Documentación de toolchains de Go

3. Qué significaban las 17 vulnerabilidades
El informe indicaba que las vulnerabilidades de la biblioteca estándar estaban presentes en go1.25.7 y que el análisis encontraba rutas desde el programa hasta las funciones afectadas. Por eso govulncheck devolvió un código de salida distinto de cero. Documentación de govulncheck
El mismo informe mencionó otras vulnerabilidades en paquetes importados y módulos requeridos, pero indicó que no detectó llamadas desde el programa hasta esas vulnerabilidades.

4. Solución aplicada al workflow
Se actualizó la versión de Go configurada en GitHub Actions:
    go-version: '1.25.14'

Esto actualiza la biblioteca estándar usada para analizar y compilar la aplicación. No era un cambio de una dependencia de go.mod; las correcciones de la biblioteca estándar llegan mediante una versión parcheada de Go. Go 1.25.14 incluye correcciones de seguridad en varios de los paquetes mencionados por el informe. Historial oficial de versiones de Go
En una ejecución posterior, el workflow avanzó hasta el paso de gosec. Como los pasos se ejecutan en orden, eso indica que Vulncheck ya había terminado correctamente en esa ejecución.

5. Error local al instalar govulncheck
El comando corto:
    go install govulncheck@latest

falló porque govulncheck no es una ruta completa de paquete Go. El comando con la ruta completa es:
    go install golang.org/x/vuln/cmd/govulncheck@latest

El workflow ya usa esa ruta completa.

6. Pendiente de versionado reproducible
El workflow todavía usa @latest para govulncheck y version: latest para golangci-lint. Eso permite que la herramienta cambie entre ejecuciones. Como siguiente mejora, fijaremos versiones concretas y compatibles para esas herramientas y registraremos cuándo actualizarlas.

La actualización de `go-version` corrigió la versión de la biblioteca estándar que analiza el pipeline. Queda como trabajo separado fijar govulncheck y golangci-lint para que sus versiones también sean reproducibles.