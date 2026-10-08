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

## Trivy Security en Golang


Remediación SCA: Trivy reportó 11 hallazgos en dependencias. go mod why -m y go mod graph mostraron que eran dependencias transitivas. Go rechazó la primera combinación de versiones porque x/mod requería una versión más nueva de x/net, que a su vez requería una más nueva de x/sys. Al actualizar x/mod, Go resolvió el grafo y elevó también x/net, x/sys, x/tools, x/sync y la directiva mínima de Go a 1.25.0. go mod tidy -diff quedó limpio; tests, build y govulncheck pasaron, y gosec reportó cero issues. Pendiente: revisar el nuevo informe de Trivy en GitHub.