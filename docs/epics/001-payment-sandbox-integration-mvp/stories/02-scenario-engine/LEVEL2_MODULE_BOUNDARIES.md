# Level 2 Module Boundaries

## Goal

Definir los limites del monolito modular para que `payment-sandbox` pueda crecer sin convertir el paquete en un bloque acoplado.

## Application Boundaries

### `internal/application/commands/payments`

- Casos de uso y dominio del ciclo de vida de pagos.
- `PaymentIntent`, captura, confirmacion, finalizacion y reglas de estado.

### `internal/application/commands/refunds`

- Casos de uso y dominio de reembolsos.
- Politicas de refund, validaciones y estados relacionados.

### `internal/application/queries/payments`

- Modelos de lectura para `PaymentIntent`, `PaymentAttempt`, `Charge` y `Refund`.
- Consultas read-only alineadas con el contrato HTTP.

### `internal/application/support/observability`

- Registro interno de eventos para tests y diagnostico.
- No es logica de negocio ni un caso de uso.

### `sandbox`

- Resolucion deterministica de escenarios de sandbox.
- Mapping de header/token a resultado esperado.

### `ports`

- Contratos para `Clock`, `Store`, `ScenarioResolver`, `EventPublisher` y `UnitOfWork`.

### `adapters`

- Implementaciones concretas de entrada, messaging, persistencia y tiempo.

### `shared`

- Solo utilidades realmente transversales y agnosticas del dominio.
- Value objects genéricos, helpers puros, errores base o contratos reutilizables.

## Why `application/support` vs `shared`

### Elegir `application/support` cuando:

- el contenido es soporte transversal de la aplicacion
- el codigo acompana a comandos y queries pero no es un caso de uso
- hay cosas como observabilidad interna o helpers de flujo
- queremos evitar que `shared` se vuelva un cajon de sastre

### Elegir `adapters` cuando:

- el contenido es infraestructura concreta o capacidades tecnicas
- el codigo implementa un puerto
- hay cosas como HTTP, messaging, storage o tiempo del sistema

### Elegir `shared` cuando:

- el contenido es verdaderamente neutro y reutilizable
- no pertenece ni a pagos ni a reembolsos ni a escenarios
- son helpers puros o contratos comunes sin semantica de infraestructura
- el codigo no depende de detalles del entorno ni de adaptadores

## Decision Rule

- Si algo pertenece al negocio, va en `internal/application/commands` o `internal/application/queries/payments`.
- Si algo es soporte transversal de la aplicacion, va en `internal/application/support`.
- Si algo es un contrato, va en `ports`.
- Si algo implementa un contrato, va en `adapters`.
- Si algo es puramente comun y sin semantica de runtime, va en `shared`.

## Import Rules

- Un modulo no importa internals de otro modulo.
- Los modulos solo se relacionan por contratos publicos o puertos.
- `application/support` puede ser consumido por comandos y queries, pero no al reves.
- `shared` debe permanecer libre de referencias a modulos de negocio.

## Suggested Location

- `docs/epics/001-payment-sandbox-integration-mvp/stories/02-scenario-engine/LEVEL2_MODULE_BOUNDARIES.md`

## Outcome

Este documento sirve como referencia para el item:
- `docs(payment-sandbox): define level 2 module boundaries`
