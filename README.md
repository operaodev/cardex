# Cardex

Cardex es una aplicacion de catalogo y busqueda de cartas de Trading Card Games (TCG), diseñada bajo una arquitectura de tabla unica ("Single Table") en PostgreSQL para optimizar el rendimiento y facilitar consultas complejas sin el uso de JOINs relacionales. La aplicacion esta construida en Go y ofrece tanto una API RESTful como herramientas de linea de comandos para la sincronizacion masiva de datos.

## Arquitectura

El sistema esta dividido en varios componentes principales dentro del directorio `internal`:

*   **cards**: Maneja el modelo de datos principal (`Card`), repositorios interactuando con la base de datos local usando GORM, y logica de negocio interna para listar y filtrar cartas. La arquitectura es de "Single Table", donde cada combinacion fisica de carta e idioma representa una fila unica en la base de datos, identificada por un `unique_id` (ExternalID-Code-Lang-Rarity).
*   **search**: Se encarga de la comunicacion con APIs externas de proveedores de TCG (actualmente YGOPRODeck y Yugipedia). Define contratos abstractos (`TCGProvider`) para facilitar la integracion de multiples origenes de datos.
*   **sync**: Orquesta la descarga masiva de datos desde los proveedores externos (`search`) y su insercion o actualizacion idempotente en la base de datos local (`cards`) mediante operaciones de Upsert.

## Componentes Ejecutables

El proyecto contiene dos puntos de entrada principales en el directorio `cmd`:

### 1. Servidor API (`cmd/api`)

Es el servidor web principal basado en Gin. Proporciona los endpoints para consultas de usuarios finales y administracion:

**Catalogo Local:**
*   `GET /cards` - Listado de cartas con paginacion y filtros opcionales (name, tcg, lang, type, archetype, subtype, set_code, rarity, page, limit).
*   `GET /cards/search` - Busqueda rapida autocompletada por nombre. Requiere el parametro `name`.
*   `GET /cards/:id` - Obtiene el detalle completo de una carta local por su ID numerico.

**Proveedores Externos (Live Search):**
*   `GET /cards/search/:provider/all` - Obtiene todas las cartas de un proveedor especifico.
*   `GET /cards/search/:provider` - Busca cartas por nombre en la API del proveedor externo.
*   `GET /cards/search/:provider/:id` - Busca una carta especifica por ID externo.

**Sincronizacion (Admin):**
*   `GET /sync/status` - Consulta si hay un proceso de sincronizacion en curso.
*   `POST /sync/:tcg` - Dispara un proceso asincrono de sincronizacion masiva para el TCG indicado (ej. `/sync/ygo`). Retorna inmediatamente 202 Accepted.

### 2. Herramienta CLI de Sincronizacion (`cmd/sync`)

Permite ejecutar la sincronizacion masiva de cartas de forma aislada, ideal para tareas programadas (cron jobs) o inicializacion de datos sin depender del servidor HTTP.

Uso basico:
```bash
go run cmd/sync/main.go --tcg=ygo --env=.env
```

## Requisitos

*   Go 1.21 o superior
*   PostgreSQL 15 o superior
*   Docker y Docker Compose (opcional, para entorno local)

## Configuracion

La aplicacion se configura mediante variables de entorno. Puedes basarte en el archivo `.env.example` para crear tu propio archivo `.env` en la raiz del proyecto:

```env
DB_HOST=localhost
DB_USER=cardex
DB_PASSWORD=cardex_secret
DB_NAME=cardex
DB_PORT=5432
DB_SSLMODE=disable
DB_TIMEZONE=America/Mexico_City
```

## Ejecucion Local

### Usando Docker

El proyecto incluye configuracion para levantar rapidamente la aplicacion web y la base de datos en contenedores:

```bash
docker-compose up -d
```

### Ejecucion Directa

Para desarrollo local, inicia PostgreSQL, configura tu archivo `.env` y ejecuta el servidor:

```bash
# Instalar dependencias
go mod tidy

# Iniciar el servidor API
go run cmd/api/main.go
```

## Pruebas (Testing)

El proyecto incluye pruebas unitarias para validar la logica de negocio, los mapeos de datos y la interaccion con repositorios falsos (mocks).

Para ejecutar todas las pruebas en el proyecto:

```bash
go test -v ./...
```

Para ejecutar pruebas con covertura de codigo:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Escribiendo Pruebas

Al escribir nuevas pruebas, manten las siguientes reglas:
*   Usa el paquete `testing` estandar de Go.
*   Implementa interfaces y utiliza mocks (por ejemplo usando `go.uber.org/mock`) para abstraer las llamadas a la base de datos o APIs externas.
*   Verifica tanto los casos de exito ("Happy Path") como los escenarios de error.
