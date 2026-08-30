# Level 2 DDD Rules: PaymentIntent Aggregate

## Goal

Formalizar `PaymentIntent` como el agregado principal del dominio de pagos para que las transiciones de estado vivan en un solo lugar.

## Aggregate Boundary

### Aggregate Root

- `PaymentIntent` es la raiz del agregado principal.
- Las transiciones del flujo de pago deben pasar por el agregado.
- El agregado es responsable de mantener el estado coherente.

### Related Entities

- `PaymentAttempt` se trata como entidad asociada al flujo del agregado.
- `Charge` se crea y actualiza dentro del contexto de `PaymentIntent`.
- `Refund` pertenece a su propio ciclo, pero sigue reglas derivadas del estado de `Charge`.

## Invariants

- Un `PaymentIntent` no puede confirmarse si no esta en `requires_payment_method` o `requires_confirmation`.
- Un `PaymentIntent` no puede capturarse si no esta en `requires_capture`.
- Un `PaymentIntent` en `processing` solo puede finalizarse si el escenario es `processing_then_succeeded`.
- Un `Charge` solo puede capturarse cuando esta en estado `authorized`.
- Un `Refund` solo puede emitirse sobre un `Charge` capturado.

## Responsibility Split

### Domain

- Decide si una transicion es valida.
- Conserva las invariantes.
- Modela estados, escenarios y reglas de negocio.

### Application

- Orquesta casos de uso.
- Coordina persistencia, eventos e idempotencia.
- No decide reglas de dominio que pertenezcan al agregado.

### Adapters

- Traducen entrada/salida.
- No contienen reglas de transicion.

## Recommended Direction

- Introducir metodos en `PaymentIntent` para validar transiciones.
- Mover las comprobaciones de estado fuera de `payments.Service`.
- Mantener errores de dominio tipados para transiciones invalidas.

## Suggested Location

- `docs/epics/001-payment-sandbox-integration-mvp/stories/02-scenario-engine/LEVEL2_DDD_RULES.md`

## Related Checklist Item

- `refactor(payment-sandbox): formalize payment intent aggregate rules`
