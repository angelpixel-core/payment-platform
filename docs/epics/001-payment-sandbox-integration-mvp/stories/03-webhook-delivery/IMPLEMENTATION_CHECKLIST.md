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

- [ ] `docs(payment-sandbox): define webhook signature configuration`
  - [ ] Definir `shared secret` y su fuente de configuracion.
  - [ ] Definir como se calcula y valida la firma.
  - [ ] Acordar el comportamiento ante firma invalida.

## 3. Emisor Go con Retries

- [ ] `refactor(payment-sandbox): add webhook dispatcher with retries`
  - [ ] Enviar el mismo payload en todos los reintentos.
  - [ ] Reintentar con backoff simple.
  - [ ] Registrar historial de intentos y estado final.
  - [ ] Hacer idempotente el delivery por `delivery_id`.

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
