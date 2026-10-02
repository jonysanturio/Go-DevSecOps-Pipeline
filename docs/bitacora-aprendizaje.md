# Bitácora de aprendizaje: GoStock

Este documento registra los cambios del proyecto y la lógica detrás de cada uno. La dinámica acordada es que el usuario modifica el código y el asistente explica, revisa y documenta; el asistente no edita código por su cuenta.

## Cómo fluye una solicitud

1. `cmd/api/main.go` registra la ruta HTTP.
2. Un método de `ProductHandler` interpreta la ruta y el cuerpo, y convierte errores a códigos HTTP.
3. El servicio valida reglas de negocio y prepara el producto.
4. El servicio usa la interfaz `domain.ProductRepository`.
5. `internal/platform/postgres/product_repository.go` implementa esa interfaz y ejecuta SQL parametrizado contra PostgreSQL.

Esta separación es parte de la arquitectura hexagonal: el servicio depende del puerto (la interfaz), mientras PostgreSQL es un adaptador.

## Cambios revisados

### 1. Errores de lectura y escritura en el repositorio

**Archivo:** `internal/platform/postgres/product_repository.go`

- `GetAll` revisa `rows.Err()` después del bucle. `rows.Next()` recorre los resultados, pero un fallo del driver puede aparecer durante esa iteración.
- `GetOne` traduce `sql.ErrNoRows` a `domain.ErrProductNotFound`. Así el dominio no necesita conocer el detalle de PostgreSQL.
- `Update` usa el `id` recibido como argumento y placeholders SQL (`$1` a `$4`). Los valores del usuario no se concatenan dentro del SQL.
- `Update` y `Delete` revisan `RowsAffected()`. Cero filas afectadas significa que no existía ese producto y se devuelve `ErrProductNotFound`.
- El handler usa `errors.Is` para convertir ese error en HTTP 404. Otros errores de persistencia se traducen a HTTP 500.

**Casos:** producto inexistente → 404; fallo interno de base de datos → 500; operación válida → 200 para lectura/actualización y 204 para eliminación.

### 2. Validación del cuerpo en `UpdateProduct`

**Archivo:** `cmd/api/handler.go`

El usuario agregó `updateReq.Validate()` después de decodificar el JSON. Si la validación falla, el handler responde 400 y hace `return`; por eso la solicitud no llega al servicio.

El patrón `if err := ...; err != nil` declara una variable `err` cuyo alcance se limita a ese `if`. En este caso, la validación comparte el método con el flujo de creación.

**Casos:** JSON mal formado → 400; nombre o valores fuera de las reglas de `Validate` → 400; solicitud válida → continúa al servicio.

### 3. ID positivo en `GetOneProduct`

**Archivo:** `cmd/api/handler.go`

El usuario agregó la condición `id <= 0` después de `strconv.Atoi`. `Atoi` verifica que el texto se pueda convertir a entero; la nueva condición valida que el entero tenga sentido como ID.

**Casos:** texto no numérico → 400; cero o número negativo → 400; ID positivo inexistente → 404; ID positivo existente → 200.

### 4. ID positivo en `UpdateProduct` y `DeleteProduct`

**Archivo:** `cmd/api/handler.go`

El usuario agregó la comprobación `<= 0` después de `strconv.Atoi` en ambas operaciones. La validación ocurre antes de llamar al servicio; en actualización también se hace antes de decodificar el cuerpo.

**Casos:** ID no numérico → 400 por error de conversión; cero o negativo → 400 por ID fuera del rango válido; ID positivo que no existe → 404; ID válido existente → actualización 200 o eliminación 204.

**Revisión:** el mensaje de `UpdateProduct` ya dice `cero`. En `GetOneProduct` todavía falta el espacio de formato: escribir `id <= 0`.

## Siguiente paso

En curso: el usuario convirtió `parseProductID` en función de paquete, pero todavía no comprueba que el path tenga el prefijo esperado ni que el ID sea positivo, y `GetOneProduct` aún no la usa. Hay además un bloque `if id <= 0` suelto fuera de toda función, lo que impide compilar. Ese chequeo debe devolver un error desde el helper; el handler debe traducirlo a HTTP 400. Después hay que integrar el helper en `GetOneProduct` y reutilizarlo en las otras operaciones.

## Temas pendientes para las siguientes etapas

- Alinear las reglas y los mensajes de `CreateProductRequest.Validate`. La implementación actual permite precio y stock iguales a cero porque comprueba si son menores que cero; los mensajes deben describir esa regla con precisión.
- Asegurar las reglas de negocio también en el dominio/servicio, no solo en HTTP.
- Crear el esquema inicial o migraciones de PostgreSQL.
- Mejorar pruebas unitarias e integración cuando se acuerde ejecutarlas.
- Fortalecer CI/CD, agregar el frontend y preparar ejercicios de seguridad aislados para esta API.

## Registro

- 2026-10-02: documentados los cambios iniciales del repositorio, la validación del cuerpo de actualización y la validación de ID en lectura.
- 2026-10-02: registrada la validación de ID positivo que el usuario agregó a actualización y eliminación.
- 2026-10-02: comenzó la extracción de parseo de ID en `parseProductID`; quedan pendientes su validación completa y uso en el handler.
- 2026-10-02: el primer intento del helper quedó incompleto: se detectó un bloque condicional fuera de una función y el helper aún no se conecta a `GetOneProduct`. Se documentó para corregirlo en el siguiente paso.