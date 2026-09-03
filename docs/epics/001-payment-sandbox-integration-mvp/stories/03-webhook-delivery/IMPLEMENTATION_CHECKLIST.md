# Checklist de Implementacion por Commits - Webhook Delivery

## Objetivo

Implementar la entrega de webhooks con retries e idempotencia entre Go y Rails, sin mezclarla con reporting ni ledger.

## Regla

- Un commit por paso logico.
- Cada paso debe dejar evidencia verificable en el lado que corresponda.
- El bus externo solo entra si el webhook a Rails existe como integracion real.

## 0. Prerequisitos Confirmados

- [x] `Sandbox API base` disponible. [Evidence](../01-sandbox-api-base/README.md)
- [x] `Scenario engine` disponible. [Evidence](../02-scenario-engine/README.md)

## 1. Contrato del Webhook

- [x] `docs(payment-sandbox): define webhook delivery contract`
  - [x] Definir eventos soportados: `payment.succeeded`, `payment.failed`, `payment.processing`. [Evidence](./WEBHOOK_CONTRACT.md)
  - [x] Definir payload minimo comun. [Evidence](./WEBHOOK_CONTRACT.md)
  - [x] Definir `event_id`, `delivery_id`, `attempt` y metadata de delivery. [Evidence](./WEBHOOK_CONTRACT.md)
  - [x] Definir el formato de firma compartida. [Evidence](./WEBHOOK_CONTRACT.md)

## 2. Firma y Configuracion

- [x] `docs(payment-sandbox): define webhook signature configuration`
  - [x] Definir `shared secret` y su fuente de configuracion. [Evidence](./WEBHOOK_SIGNATURE.md)
  - [x] Definir como se calcula y valida la firma. [Evidence](./WEBHOOK_SIGNATURE.md)
  - [x] Acordar el comportamiento ante firma invalida. [Evidence](./WEBHOOK_SIGNATURE.md)

## 3. Emisor Go con Retries

- [x] `refactor(payment-sandbox): add webhook dispatcher skeleton`
  - [x] Crear `WebhookDispatcher` como componente dedicado. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)
  - [x] Definir la entrada con `event`, `delivery_id`, `attempt` y endpoint. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)
  - [x] Mantener la responsabilidad aislada de la publicacion de webhooks. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)

- [x] `refactor(payment-sandbox): reuse immutable webhook payload bytes`
  - [x] Serializar el payload una sola vez por delivery. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)
  - [x] Reutilizar exactamente los mismos bytes en todos los reintentos. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)
  - [x] Mantener `event_id` y `delivery_id` estables. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)

- [x] `refactor(payment-sandbox): add retry loop with backoff`
  - [x] Implementar reintentos sincronos in-process. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)
  - [x] Definir backoff simple. [Evidence](./WEBHOOK_RETRY_POLICY.md)
  - [x] Definir limite maximo de intentos. [Evidence](./WEBHOOK_RETRY_POLICY.md)
  - [x] Registrar fallo final si se agotan los retries. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)

- [x] `refactor(payment-sandbox): record webhook delivery attempts`
  - [x] Registrar cada intento. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)
  - [x] Registrar estado final por delivery. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)
  - [x] Dejar trazabilidad para debug y reconciliacion. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher.go)

- [ ] `test(payment-sandbox): cover webhook dispatcher retry contract`
  - [x] Test de payload inmutable entre retries. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/messaging/webhook/dispatcher_test.go)
  - [ ] Test de retry por fallo temporal.
  - [ ] Test de fallo final tras agotar intentos.
  - [ ] Test de `delivery_id` estable.

## 4. Inbox Rails

- [ ] `refactor(payment-sandbox): add rails webhook inbox`
  - [ ] Persistir cada webhook antes de aplicar cambios de negocio.
  - [ ] Rechazar duplicados sin duplicar efectos.
  - [ ] Validar la firma antes de mutar estado.
  - [ ] Exponer el historial de entregas para debugging y reconciliacion.

## 5. Proyeccion y Reconciliacion

- [ ] `refactor(payment-sandbox): update payment projection from validated inbox`
  - [ ] Actualizar la proyeccion local solo desde inbox validado.
  - [ ] Mantener consistencia entre inbox y estado de pago.

## 6. Tests y Contractos

- [ ] `test(payment-sandbox): cover webhook delivery and inbox contract`
  - [ ] Test de entrega exitosa.
  - [ ] Test de retry por fallo temporal.
  - [ ] Test de evento duplicado.
  - [ ] Test de firma invalida.
  - [ ] Test de persistencia de inbox antes de side effects.

## 7. Cierre

- [ ] `docs(payment-sandbox): mark webhook delivery story complete`
  - [ ] Marcar evidencia en `README.md`.
  - [ ] Cerrar checkboxes cuando la implementacion este lista.

## Notas

- No incluye ledger reporting.
- No incluye external queue infrastructure.
- No incluye proveedores externos fuera del contrato sandbox.
