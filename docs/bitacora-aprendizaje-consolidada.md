# Bitácora consolidada: Go DevSecOps Pipeline

**Proyecto:** API de inventario en Go con arquitectura hexagonal, PostgreSQL, Docker Compose y Prometheus  
**Actualización de esta consolidación:** 2026-10-08  
**Propósito:** registrar qué problema apareció, cómo se investigó, qué solución se adoptó, por qué funciona y qué evidencia la validó.

> Esta bitácora distingue cambios realizados de trabajo planificado. Los comandos y resultados que aparecen como evidencia son los que se compartieron durante el recorrido; no significan que se hayan vuelto a ejecutar al crear esta consolidación.

## 1. Forma de trabajo y objetivos

El aprendizaje se organizó por incrementos: primero entender y mejorar el comportamiento de la API; después convertir los controles de calidad y seguridad en gates de CI; más adelante asegurar la imagen y el despliegue. El usuario implementa los cambios y opera GitHub; la guía explica conceptos, diagnóstico, alternativas y validación.

La meta Enterprise-Ready se dividió en estas etapas:

| Etapa | Resultado esperado | Estado según el recorrido |
|---|---|---|
| 0. Decisiones y línea base | Alcance, riesgos, flujo de la app y baseline reproducible | Parcial; se identificaron arquitectura, herramientas y riesgos iniciales |
| 1. Preparación de la aplicación | API validada, segura, configurable y operable | En curso; se trabajaron validaciones, errores HTTP, DB y timeouts |
| 2. CI y controles de seguridad | Gates de SAST, vulnerabilidades, SCA, secretos, tests y lint | Parcial; SAST, govulncheck, Trivy en modo baseline, tests, build y golangci-lint trabajados; Gitleaks aún pendiente |
| 3. Imagen y cadena de suministro | Imagen mínima, escaneada, firmada y referenciada por digest | Pendiente |
| 4. Kubernetes y GitOps | Manifiestos/Helm y sincronización con Argo CD | Pendiente |
| 5. Seguridad de runtime y operación | Falco, NetworkPolicies, logs, métricas, dashboards y alertas | Pendiente |

## 2. Arquitectura de la aplicación que sirve de contexto

El flujo de una solicitud es:

1. `cmd/api/main.go` registra las rutas y configura el servidor HTTP.
2. `cmd/api/handler.go` interpreta el protocolo HTTP, decodifica el cuerpo, valida datos de entrada y traduce resultados a códigos HTTP.
3. `internal/product` aplica casos de uso y reglas de aplicación.
4. `internal/domain` define el modelo y errores del dominio, además del puerto `ProductRepository`.
5. `internal/platform/postgres` implementa ese puerto con `database/sql` y PostgreSQL.

La separación hexagonal permite que el dominio exprese “producto no encontrado” sin depender de `sql.ErrNoRows`, HTTP o PostgreSQL. El adaptador traduce los detalles de infraestructura al contrato del dominio; el handler traduce los errores del dominio a respuestas HTTP.

## 3. Cambios de aplicación trabajados antes del pipeline

### 3.1 Errores y consistencia en el repositorio PostgreSQL

**Problema y diagnóstico:** el acceso a SQL puede fallar al ejecutar la consulta, al escanear una fila, al iterar resultados o al cerrar el cursor. Además, un resultado vacío no es un error genérico: en operaciones individuales significa que el producto no existe.

**Decisiones implementadas o trabajadas:**

- `GetAll` usa `QueryContext` y propaga el contexto de la solicitud.
- Se comprueban errores de `Scan` y `rows.Err()` después del bucle; un driver puede descubrir un error durante la iteración.
- `GetOne` traduce `sql.ErrNoRows` a `domain.ErrProductNotFound` usando `errors.Is`.
- `Update` y `Delete` usan placeholders `$1`, `$2`, etc. Los valores del usuario se envían como argumentos y no se concatenan al SQL.
- `Update` y `Delete` inspeccionan `RowsAffected()`; cero filas afectadas se traduce a `ErrProductNotFound`.
- Los errores de infraestructura se envuelven con `%w` para conservar su causa y permitir que capas superiores los clasifiquen con `errors.Is`.

**Teoría:** una consulta parametrizada separa el código SQL de los datos y es la defensa estándar frente a inyección SQL. Los errores envueltos conservan causalidad entre capas. El adaptador no debe filtrar texto SQL al cliente; la API devuelve un mensaje genérico y el diagnóstico detallado debe quedar en logs internos.

**Cuidado con `Rows.Close`:** golangci-lint encontró que ignorar el error de cierre deja una ruta de error sin tratar. La solución ensayada usa resultados nombrados y une el error de cierre al error de retorno con `errors.Join`. Esto exige que el `defer` modifique el resultado nombrado `retErr`. Al usar un retorno nombrado `products`, no se debe volver a declarar `var products`: eso provocó `products redeclared`. Se quitó la declaración redundante y `go test -mod=readonly ./...` volvió a pasar.

**Evidencia compartida:** después de quitar la declaración duplicada, todos los paquetes compilaban en `go test -mod=readonly ./...`; `cmd/api`, `internal/product` pasaron y los demás paquetes sin pruebas reportaron `[no test files]`.

### 3.2 Validación de DTO, ID y límites HTTP

**Validación del cuerpo:** se incorporó `Validate()` también a la operación de actualización, igual que en creación. Un cuerpo JSON mal formado o que viola validaciones recibe `400 Bad Request` y no llega al servicio.

**Validación del ID:** se extrajo la función común `parseProductID(path string)`. Comprueba el prefijo `/products/`, extrae el sufijo, lo convierte con `strconv.Atoi` y exige un entero positivo. GET, PUT y DELETE reutilizan el mismo parser. Un ID con formato inválido se responde como 400; un ID válido que no exista se transforma en 404 desde el dominio/servicio.

**Mensajes 500:** se cambiaron respuestas al cliente para no revelar errores internos de PostgreSQL. La especificación es: error técnico detallado para observabilidad interna, mensaje seguro y estable hacia el cliente.

**Reglas de negocio:** se detectó una posible inconsistencia entre permitir `stock == 0` y un mensaje que decía “el stock debe ser mayor a cero”. Si la regla real permite inventario agotado, el texto correcto debe expresar “no puede ser negativo”; si el negocio prohíbe cero, la validación y los tests deben rechazarlo explícitamente. El mensaje no reemplaza la decisión de negocio.

**Métodos HTTP pendientes de completar:** la bitácora anterior señaló que `/products` dirigía cualquier método distinto de POST a la lectura y que algunos métodos no contemplados podían producir una respuesta incorrecta. En el código inspeccionado, `/products/` ya tiene `default` con 405; aún conviene aplicar el contrato por ruta y el header `Allow` a todos los dispatchers. La teoría HTTP es que `405 Method Not Allowed` indica que el método existe pero no está soportado para el recurso; `Allow` enumera los métodos válidos.

### 3.3 Hardening inicial del servidor HTTP por Gosec G114

**Hallazgo:** `gosec ./...` reportó G114 en `http.ListenAndServe("0.0.0.0:8080", nil)`, porque esa forma no permite configurar timeouts del servidor.

**Solución:** pasar de la función abreviada a un `http.Server` explícito con `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` y `MaxHeaderBytes`; iniciar con `server.ListenAndServe()`.

**Teoría:** los timeouts limitan cuánto tiempo puede consumir una conexión lenta y reducen exposición a agotamiento de recursos, como slow-client/slowloris. Son límites de transporte HTTP, no sustituyen límites de negocio ni el timeout de consultas a la base. `ReadHeaderTimeout` limita la lectura de cabeceras; `ReadTimeout` limita la lectura de la solicitud; `WriteTimeout` limita la escritura; `IdleTimeout` limita conexiones keep-alive ociosas.

**Errores de edición que aparecieron:** se reportaron `undefined: html` y `ReadHandlerTimeout` desconocido. `html` no es una variable predeclarada: si se necesita `html.EscapeString`, corresponde importar `html`; si no se está escapando contenido HTML, no se debe añadir ese uso. El nombre correcto del campo de `http.Server` es `ReadHeaderTimeout`, no `ReadHandlerTimeout`. La lección fue validar nombres contra la API estándar y recompilar después del cambio.

## 4. Pipeline de CI: diagnóstico y soluciones

### 4.1 Primera versión de la barrera de calidad

El workflow de GitHub Actions se fue armando con estos controles:

1. Checkout del commit.
2. Instalación de la versión de Go fijada para el proyecto.
3. Verificación de consistencia de módulos con `go mod tidy -diff`.
4. Instalación y ejecución de `govulncheck`.
5. Instalación y ejecución de `gosec`.
6. `go test -mod=readonly ./...`.
7. `go build -mod=readonly ./...`.
8. Trivy de filesystem para dependencias.
9. `golangci-lint`.

`-mod=readonly` impide que tests/build modifiquen silenciosamente `go.mod` o `go.sum`. `go mod tidy -diff` compara la forma actual con la que Go derivaría del código, sin reescribir automáticamente los archivos. Son controles distintos: uno detecta drift de módulos; el otro verifica que código y dependencias sigan compilando.

### 4.2 Error de instalación/ubicación de Gosec en GitHub Actions

**Síntoma:** el paso intentó ejecutar `/home/runner/go/bin/gosec`, pero el archivo no existía (`No such file or directory`, exit code 127).

**Causa técnica:** instalar una herramienta con `go install` no garantiza que una ruta asumida manualmente coincida con el destino configurado. `GOBIN` puede estar definido; si está vacío, Go usa `GOPATH/bin`. Además, modificar `PATH` dentro de un paso no lo hace persistir automáticamente en pasos siguientes.

**Solución aplicada:** el paso calcula `go env GOBIN`; si está vacío usa `go env GOPATH` más `/bin`; agrega ese directorio a `$GITHUB_PATH`; y comprueba que el binario exista con `test -x`. En el paso siguiente se invoca `gosec ./...` por nombre.

**Teoría de GitHub Actions:** escribir una ruta en el archivo indicado por `GITHUB_PATH` agrega ese directorio al `PATH` de los pasos posteriores del mismo job. No cambia retroactivamente el `PATH` del paso que escribió la ruta. La separación instalación/ejecución deja el binario disponible de forma predecible.

**Ejecución antigua:** si se edita `ci.yml` localmente pero GitHub muestra el error anterior, ese run fue disparado para otro commit SHA. Cada ejecución evalúa el workflow de la revisión correspondiente. Hay que confirmar que el cambio se guardó, se incluyó en un commit y se publicó en la rama del PR; luego abrir el run nuevo asociado al SHA más reciente. Volver a abrir un log viejo no reejecuta el workflow.

### 4.3 Versionado de `govulncheck` y toolchains

**Primer problema:** `go install golang.org/x/vuln/cmd/govulncheck@latest` seleccionó una versión reciente de `x/vuln` cuyo `go.mod` requería Go 1.26. El runner del workflow se configuró con Go 1.25. Go intentó descargar automáticamente una toolchain 1.26.9 que todavía no estaba disponible en el origen de toolchains; el fetch devolvió 404.

**Solución adoptada:** fijar la herramienta a `golang.org/x/vuln/cmd/govulncheck@v1.7.0`, compatible con la línea Go 1.25 usada en ese momento. La ejecución compartida mostró que se instaló y que el análisis pudo cargar paquetes después de corregir primero un error de compilación en el repositorio.

**Teoría:** `@latest` es una consulta móvil. Una herramienta de seguridad puede aumentar su versión mínima de Go y romper reproducibilidad aun cuando la aplicación no haya cambiado. Fijar una versión hace el build repetible; debe revisarse y actualizarse de forma deliberada. La versión de la herramienta (`govulncheck`) y la versión de Go requerida por el proyecto son decisiones relacionadas, pero no son el mismo componente.

### 4.4 Fallos de linter como feedback de calidad

La actualización del linter convirtió problemas antes tolerados en errores de CI. Se trabajaron estos resultados:

| Hallazgo | Qué significa | Solución/criterio |
|---|---|---|
| `errcheck`: resultado de `fmt.Fprintln` ignorado | La escritura al cliente puede fallar, por ejemplo si corta la conexión | Comprobar `err`; en `/health` se registra el fallo del write en el logger del servidor |
| `errcheck`: `rows.Close()` ignorado | El cierre tiene un resultado observable que la herramienta exige tratar | Capturar error de cierre y unirlo al retorno con `errors.Join`, sin ocultar un error de consulta/iteración ya existente |
| `ST1005`: error empieza con mayúscula | Los errores suelen agregarse a otro contexto y deben componerse como fragmentos | Usar minúscula inicial, salvo nombre propio/acrónimo permitido; por ejemplo `ruta inválida` |
| `products redeclared` | Un parámetro de retorno nombrado ya declaró el identificador | Eliminar `var products`; inicializar/apendear directamente al resultado nombrado |

**Teoría de `errors.Join`:** permite conservar simultáneamente el error principal y errores secundarios. Es útil cuando la operación falló al iterar y además falló el cierre. Los consumidores pueden seguir buscando causas mediante `errors.Is`/`errors.As`. Si no existe error primario, el error de cierre pasa a ser el retorno.

**Teoría de ST1005:** un error debe ser un fragmento componible, no una frase formateada para empezar un mensaje independiente. Así, `fmt.Errorf("leyendo archivo: %w", err)` no produce una mayúscula en medio del mensaje. Referencia oficial: Staticcheck ST1005.

**Evidencia local compartida:** después de corregir la redeclaración, `go test -mod=readonly ./...` pasó. Luego el log de CI reportó ST1005 en `cmd/api/handler.go` para `errors.New("Ruta inválida")`; se cambió a minúscula. Se compartió una captura posterior con “All checks have passed”. La última inspección local, sin embargo, detectó `cmd/api/handler.go` modificado sin commit; por eso el check verde documenta el commit que GitHub ejecutó, no necesariamente cualquier edición local posterior.

### 4.5 Incompatibilidad de versiones de golangci-lint

**Síntoma:** la acción v6 con `version: latest` instaló `golangci-lint v1.64.8`, construido con Go 1.24, mientras el proyecto apuntaba a Go 1.25. El linter se negó a cargar el config porque su versión del lenguaje era inferior a la versión objetivo.

**Solución adoptada:** actualizar la acción a `golangci/golangci-lint-action@v9` y fijar `version: v2.14.0`. Esto movió el pipeline a la línea 2.x compatible con el Go target del proyecto y produjo diagnósticos concretos, en lugar de fallar al analizar la configuración.

**Teoría:** una acción GitHub y la herramienta que instala son dos versiones independientes. `@v6` identifica la acción; `version: v2.14.0` identifica `golangci-lint`. `latest` puede cambiar entre ejecuciones y no es reproducible. Pinning reduce sorpresas, aunque también requiere mantenimiento y actualizaciones periódicas.

## 5. SCA: investigar y actualizar dependencias transitivas

### 5.1 Qué reportó Trivy y cómo se investigó

Trivy encontró inicialmente 11 vulnerabilidades (7 altas, 3 medias y 1 de severidad desconocida) entre módulos de Go, incluidos `golang.org/x/mod`, `golang.org/x/net` y `golang.org/x/sys`.

Se utilizaron estas consultas:

```bash
go mod why -m golang.org/x/mod
go mod why -m golang.org/x/net
go mod why -m golang.org/x/sys
go mod graph | grep -E 'golang.org/x/(mod|net|sys)@'
go list -m golang.org/x/mod golang.org/x/net golang.org/x/sys golang.org/x/tools golang.org/x/sync
```

`go mod why -m` explica por qué el módulo forma parte del grafo desde los paquetes del proyecto. `go mod graph` muestra las aristas de requisitos entre módulos; `go list -m` muestra las versiones seleccionadas para este build.

### 5.2 MVS y por qué no funcionó la primera combinación

La combinación tentativa `x/mod@v0.40.0`, `x/net@v0.56.0`, `x/sys@v0.44.0` fue rechazada: `x/mod@v0.40.0` requería una versión superior de `x/net`, y `x/net@v0.56.0` requería una versión superior de `x/sys`.

Go usa Minimal Version Selection (MVS): recorre el grafo de requisitos y elige para cada módulo la versión mínima que satisface el requisito más alto encontrado. Por eso actualizar un módulo puede elevar otros transitivos. No se debe elegir un conjunto de versiones aisladas ignorando los `require` de sus módulos.

Se ejecutó:

```bash
go get golang.org/x/mod@v0.40.0
```

Go actualizó en forma consistente `x/mod` a `v0.40.0`, `x/net` a `v0.58.0`, `x/sys` a `v0.47.0`, `x/tools` a `v0.49.0` y `x/sync` a `v0.22.0`; también elevó la directiva `go` del módulo de `1.24.0` a `1.25.0`. Esto pasó porque los módulos seleccionados exigían esa versión mínima de Go.

### 5.3 Validación y significado de cada comando

```bash
go mod tidy -diff
go test -mod=readonly ./...
go build -mod=readonly ./...
govulncheck ./...
gosec ./...
```

- `go mod tidy -diff` no mostró diferencias al final: `go.mod` y `go.sum` coincidían con las dependencias requeridas por el código.
- `go test -mod=readonly ./...` pasó en todos los paquetes; el paquete `internal/product` tenía pruebas y los demás reportaron que no tenían archivos de test.
- `go build -mod=readonly ./...` terminó sin errores.
- `govulncheck ./...` informó `No vulnerabilities found.`
- Gosec informó `Issues: 0`.
- En el run de Trivy compartido, la tabla de `go.mod` mostró `0` vulnerabilidades.

**Aclaraciones de comandos/Git:** `go diff` no existe; para ver cambios se usa `git diff -- go.mod go.sum`. `git diff --cached` muestra el área staged; `git diff` sin `--cached` muestra cambios no staged. `git diff --check` detecta whitespace problemático. Los avisos LF/CRLF son advertencias de conversión de finales de línea; no equivalen por sí solos a una diferencia funcional. El output de `go mod tidy -diff` debe compararse con el archivo real; la variante sin `-diff` sí escribe cambios.

### 5.4 Lo que significa —y no significa— el baseline de Trivy

La acción se configuró para escanear filesystem (`scan-type: fs`) con `scanners: vuln`, output tabla y `exit-code: '0'`. Esto dio visibilidad del reporte sin bloquear CI durante la remediación inicial. Un run verde con `exit-code: 0` **no** demuestra que el reporte esté vacío: hay que leer la tabla y el conteo.

Una vez que el baseline está limpio, la evolución prevista es convertir el escaneo en gate, elegir umbrales y severidades, y acordar el procedimiento para excepciones temporales con responsable, motivo y fecha de vencimiento. No se debe dejar `exit-code: 0` como gate permanente si el objetivo es bloquear nuevas vulnerabilidades.

## 6. PR, commits y evidencia de CI

El flujo practicado fue editar, ejecutar validaciones locales, revisar cambios staged/no staged, hacer commit, push y abrir/actualizar PR. El PR es una unidad de revisión; el workflow se ejecuta sobre una revisión concreta. La captura más reciente compartida mostró:

- `All checks have passed`;
- un check de `Go DevSecOps Pipeline / Security & Testing Gate` exitoso;
- ausencia de conflictos;
- el PR todavía abierto con botón `Merge pull request`.

La captura acredita que los controles configurados para ese commit pasaron. No acredita controles todavía no añadidos (por ejemplo Gitleaks, escaneo de imagen o despliegue Kubernetes), ni que ediciones locales posteriores estén dentro de ese commit.

## 7. Próximo paso: higiene de secretos y Gitleaks

El `.env` apareció entre los archivos abiertos del IDE, por eso antes de añadir un detector de secretos se planteó verificar su relación con Git, sin mostrar su contenido:

```bash
git ls-files -- .env
git check-ignore -v .env
```

- Primera salida vacía: `.env` no está versionado en el índice actual.
- Primera salida con `.env`: el archivo está bajo seguimiento de Git. Si contiene credenciales reales que llegaron a un remoto, no basta con borrar el archivo: hay que revocar/rotar esas credenciales y revisar el historial.
- Segunda salida con una regla: muestra qué patrón de `.gitignore` lo excluye. Sin salida, esa ruta no está ignorada por la regla consultada.

**Teoría de Gitleaks:** busca patrones que se parecen a tokens, claves y credenciales en archivos/historial Git. El escáner reduce el tiempo entre filtración y detección, pero no revoca secretos ni prueba que un repositorio esté libre de todas las credenciales. Los secretos reales deben venir de un gestor/secreto del entorno, no quedar en el repo. Los falsos positivos requieren supresión explícita, estrecha y justificada; jamás pegar el secreto en logs o en una conversación.

**Implementación que sigue:** comprobar `.env`/`.gitignore`, ejecutar un scan local controlado, agregar Gitleaks al workflow, observar formato y comportamiento ante hallazgos, luego fijar versión/referencia de action y hacer que el job falle cuando aparezca un secreto no exceptuado.

## 8. Diseño teórico de las etapas siguientes

Estas son especificaciones de diseño, no cambios ya completados.

### 8.1 Endurecimiento de la imagen

**Objetivo:** reducir contenido, privilegios y superficie de ataque del runtime.

1. Mantener un stage de builder con Go y herramientas de compilación.
2. Compilar un binario reproducible con flags definidos; evaluar `CGO_ENABLED=0` según dependencias y requisitos.
3. Copiar solo el binario y los archivos mínimos necesarios a la etapa final.
4. Usar una base `distroless` apropiada o `scratch`; copiar certificados CA si el servicio realiza HTTPS saliente y verificar si necesita zona horaria/archivos del sistema.
5. Ejecutar con UID/GID no privilegiados, por ejemplo `USER 65532:65532` cuando la imagen y ownership lo permiten.
6. En runtime, limitar filesystem (read-only si la app lo soporta), capabilities, escalada de privilegios y recursos.
7. Escanear la imagen final con Trivy; corregir vulnerabilidades aplicables y registrar excepciones justificadas.

**Teoría:** un multi-stage build separa build-time de runtime. La imagen final no necesita compilador, gestor de paquetes ni código fuente. Menos contenido reduce CVEs potenciales y herramientas disponibles tras un compromiso. `USER` establece el usuario del proceso, pero no sustituye seccomp, capabilities, filesystem read-only ni parcheo de la base. `scratch` no trae CA roots, shell, `/etc/passwd` ni timezone data; la app debe funcionar sin ellos o se deben incluir explícitamente. Distroless conserva únicamente runtime y archivos mínimos seleccionados.

### 8.2 Firma y procedencia con Cosign

**Objetivo:** poder verificar que la imagen desplegada es el artefacto que construyó el workflow autorizado.

1. Construir y escanear la imagen en CI.
2. Publicarla en un registry.
3. Referenciarla por digest `sha256:...`, no solo por una etiqueta mutable como `latest`.
4. Firmar el digest con Cosign; considerar firma keyless basada en identidad OIDC del workflow.
5. Verificar firma, identidad del workflow/issuer y digest antes de aceptar despliegue.

**Teoría:** el digest identifica contenido; la firma enlaza ese contenido con una identidad autorizada. Una firma no vuelve segura una imagen vulnerable: demuestra procedencia/integridad. La política de despliegue debe verificar que la identidad firmante sea la esperada.

### 8.3 Migración de Compose a Kubernetes/Helm

**Objetivo:** declarar en Git el estado operativo, sin trasladar literalmente todos los detalles de desarrollo de Compose.

1. Mapear API, PostgreSQL, Prometheus, puertos, variables, volúmenes y dependencias de Compose.
2. Diseñar `Deployment` + `Service` para API; `ConfigMap` para configuración no sensible y `Secret`/gestor externo para credenciales.
3. Decidir PostgreSQL administrado o un `StatefulSet` con PVC, backups, restore probado, upgrades y monitoreo.
4. Definir probes de startup/readiness/liveness coherentes con la app; requests/limits; réplicas; estrategia de rollout.
5. Aplicar `securityContext`: UID no-root, `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem` cuando proceda, capabilities eliminadas y seccomp apropiado.
6. Usar Helm si se necesitan parámetros/variantes de entorno; separar valores de desarrollo y producción sin almacenar secretos en Git.
7. Validar manifiestos y probar en cluster local antes de producción.

**Teoría:** Kubernetes ofrece controladores declarativos que comparan estado deseado con estado observado. El contenedor no equivale a una VM; persistencia, identidad, red, health, recursos y rollout se configuran explícitamente. No se debe tratar una base de datos con estado como un Deployment sin plan de almacenamiento/recuperación.

### 8.4 GitOps con Argo CD

**Objetivo:** Git como fuente de verdad para despliegues.

1. Guardar manifiestos o Helm values de entorno en un repo/ruta declarativa.
2. Argo CD observa esa fuente y reconcilia el cluster con el commit aprobado.
3. CI prueba, escanea, construye, firma y actualiza la referencia de imagen por digest mediante un cambio revisable.
4. Promover a entornos superiores por PR/commit; no dar al pipeline de build permisos directos amplios sobre el cluster.
5. Definir de forma deliberada auto-sync, self-heal, prune, rollback y políticas de promoción.

**Teoría:** el modelo pull reduce credenciales de despliegue en CI y hace auditable quién cambió el estado deseado. Argo CD compara Git con el estado live y marca drift. `prune` elimina recursos retirados del manifiesto; `selfHeal` revierte cambios manuales al cluster. Son opciones con impacto operativo y deben activarse con una estrategia de protección.

### 8.5 NetworkPolicies para proteger PostgreSQL

**Objetivo:** que solo workloads necesarios puedan hablar con PostgreSQL.

1. Confirmar que el plugin CNI del cluster aplica NetworkPolicy.
2. Etiquetar de forma estable los Pods de API y PostgreSQL.
3. Empezar con política default-deny adecuada al namespace.
4. Permitir ingress a PostgreSQL desde Pods de API únicamente por TCP/5432.
5. Permitir egress de la API a PostgreSQL y DNS a los resolvers del cluster cuando se aplique deny-all egress.
6. Probar tráfico permitido y denegado desde Pods reales antes de promover.

**Teoría:** NetworkPolicy filtra tráfico IP de capa 3/4 según selectores de Pods/namespaces/IPs y puertos. Crear el objeto no garantiza aplicación: el CNI debe soportarlo. Default-deny egress puede bloquear DNS si no hay regla explícita; eso puede hacer que la API no resuelva el Service de DB.

### 8.6 Runtime Security con Falco

**Objetivo:** detectar conductas sospechosas que el análisis de código/imagen no observa.

1. Desplegar Falco en el cluster según modo de captura soportado por nodos/kernel.
2. Elegir reglas estables y revisar ruido antes de activar alertas críticas.
3. Definir casos relevantes para la API: shell inesperado, ejecución de binarios, escritura en rutas sensibles, escalada/privilegios y conexiones fuera de patrón.
4. Enriquecer eventos con metadata de Kubernetes y enviar alertas al canal/plataforma operativa.
5. Probar una detección controlada y documentar triage, dueño y respuesta.

**Teoría:** Falco observa eventos runtime (syscalls y fuentes mediante plugins), los evalúa con reglas y emite alertas con contexto. Un evento detectado no es automáticamente un incidente confirmado; reglas necesitan tuning y un procedimiento de respuesta. Algunas instalaciones requieren acceso privilegiado/nodo, que debe justificarse y limitarse.

### 8.7 Logs, Prometheus y Grafana

**Objetivo:** conectar una alerta cuantitativa con los eventos que explican qué pasó.

1. Emitir logs JSON con timestamp, nivel, servicio, entorno, request/correlation ID, ruta normalizada, método, status y duración.
2. No registrar passwords, tokens, Authorization headers, cuerpos sensibles ni DSNs.
3. Instrumentar contadores de requests/errores y histogramas de latencia; agregar métricas de DB/pool y eventos de seguridad pertinentes.
4. Evitar etiquetas de alta cardinalidad (por ejemplo ID de producto, request ID o path completo con ID) en Prometheus.
5. Conservar el mismo request/correlation ID en logs y respuesta/capa de observabilidad, y cruzarlo desde dashboard/alerta.
6. Crear dashboards de tasa de errores, latencia, saturación y eventos de seguridad; definir alertas con umbral, ventana, severidad, dueño y runbook.

**Teoría:** métricas resumen tendencias y permiten alertar con bajo volumen; logs explican eventos individuales con mayor contexto. Las métricas Prometheus son series temporales etiquetadas: etiquetas con valores ilimitados producen explosión de cardinalidad. Correlacionar por ID es útil en logs/traces, pero esos IDs no deben convertirse en labels de métricas. Grafana puede consultar métricas/logs y administrar alertas; la ruta exacta depende del backend de logs que se elija.

## 9. Checklist de aprendizaje/validación para cada próximo cambio

Para cada incremento, conservar el mismo método:

1. Escribir el riesgo o necesidad y el comportamiento esperado.
2. Consultar la documentación oficial de la herramienta/regla.
3. Implementar un cambio pequeño y explicar sintaxis y límites.
4. Ejecutar primero el control específico que motivó el cambio y luego tests/build pertinentes.
5. Revisar reporte completo y confirmar que el check está escaneando lo esperado.
6. Registrar comando, salida relevante, causa raíz, decisión y riesgo residual en esta bitácora.
7. Commit/push/PR, revisar el run correspondiente al último SHA y registrar resultado.

## 10. Referencias oficiales

- Go security y `govulncheck`: <https://go.dev/doc/security/vuln/> y <https://go.dev/doc/tutorial/govulncheck>
- Go Modules, MVS, `go mod graph` y `go mod tidy`: <https://go.dev/ref/mod>
- Go `database/sql`: <https://pkg.go.dev/database/sql>
- `errors.Join`: <https://pkg.go.dev/errors#Join>
- `http.Server`: <https://pkg.go.dev/net/http#Server>
- Gosec: <https://github.com/securego/gosec>
- Staticcheck ST1005: <https://staticcheck.dev/docs/checks#ST1005>
- GitHub Actions `GITHUB_PATH`: <https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#adding-a-system-path>
- Trivy filesystem scan: <https://github.com/aquasecurity/trivy-action> y <https://github.com/aquasecurity/trivy/blob/main/docs/guide/target/filesystem.md>
- Gitleaks: <https://github.com/gitleaks/gitleaks>
- Docker multi-stage builds y `USER`: <https://docs.docker.com/get-started/docker-concepts/building-images/multi-stage-builds/> y <https://docs.docker.com/reference/dockerfile/#user>
- Distroless: <https://github.com/GoogleContainerTools/distroless>
- Cosign/Sigstore: <https://docs.sigstore.dev/cosign/signing/signing_with_containers/>
- Kubernetes NetworkPolicies: <https://kubernetes.io/docs/concepts/services-networking/network-policies/>
- Argo CD automated sync: <https://argo-cd.readthedocs.io/en/stable/user-guide/auto_sync/>
- Falco: <https://falco.org/docs/> y <https://falco.org/docs/concepts/rules/>
- Prometheus instrumentation: <https://prometheus.io/docs/practices/instrumentation/>
- Grafana observability signals: <https://grafana.com/docs/grafana/latest/visualizations/simplified-exploration/metrics/about-metrics/>

## Registro de cambios de esta bitácora

- **2026-10-08:** consolidado el recorrido disponible de la API y la etapa de CI/SAST/SCA; se agregaron fundamentos teóricos, diagnóstico de fallos, evidencias compartidas y especificación de las etapas futuras. No se modificó el código del proyecto al preparar este documento.
