# Level 2 Module Boundaries

## Goal

Definir los limites del monolito modular para que `payment-sandbox` pueda crecer sin convertir el paquete en un bloque acoplado.

## Module Names

### `payments`

- Casos de uso y dominio del ciclo de vida de pagos.
- `PaymentIntent`, captura, confirmacion, finalizacion y reglas de estado.

### `refunds`

- Casos de uso y dominio de reembolsos.
- Politicas de refund, validaciones y estados relacionados.

### `scenarios`

- Resolucion deterministica de escenarios de sandbox.
- Mapping de header/token a resultado esperado.

### `platform`

- Infraestructura comun y capacidades tecnicas compartidas.
- Clock, logging, metrics, tracing, events, storage base, helpers de runtime.

### `shared`

- Solo utilidades realmente transversales y agnosticas del dominio.
- Value objects genéricos, helpers puros, errores base o contratos reutilizables.

## Why `platform` vs `shared`

### Elegir `platform` cuando:

- el contenido es infraestructura o capacidades tecnicas
- el codigo debe servir a varios modulos pero no representa negocio
- hay cosas como observabilidad, bus, storage base o configuracion de runtime
- queremos evitar que `shared` se vuelva un cajon de sastre

### Elegir `shared` cuando:

- el contenido es verdaderamente neutro y reutilizable
- no pertenece ni a pagos ni a reembolsos ni a escenarios
- son helpers puros o contratos comunes sin semantica de infraestructura
- el codigo no depende de detalles del entorno ni de adaptadores

## Decision Rule

- Si algo es tecnico/operacional, va en `platform`.
- Si algo es puramente comun y sin semantica de runtime, va en `shared`.
- Si algo pertenece al negocio, va en su modulo.

## Import Rules

- Un modulo no importa internals de otro modulo.
- Los modulos solo se relacionan por contratos publicos o puertos.
- `platform` puede ser consumido por los modulos, pero no al reves.
- `shared` debe permanecer libre de referencias a modulos de negocio.

## Suggested Location

- `docs/epics/001-payment-sandbox-integration-mvp/stories/02-scenario-engine/LEVEL2_MODULE_BOUNDARIES.md`

## Outcome

Este documento sirve como referencia para el item:
- `docs(payment-sandbox): define level 2 module boundaries`
