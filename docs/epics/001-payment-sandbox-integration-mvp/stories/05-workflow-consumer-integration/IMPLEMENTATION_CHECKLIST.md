# Checklist de Implementacion por Commits - Workflow Consumer Integration

## Objetivo

Implementar la integracion del consumidor de workflow contra `payment-sandbox` con inbox, idempotencia y reconciliacion, usando contrato `v1`.

## Regla

- Un commit por paso logico.
- Cada paso debe dejar evidencia verificable en el lado que corresponda.
- El contrato wire-level vive en OpenAPI `v1`.

## 0. Prerequisitos Confirmados

- [x] `Sandbox API base` disponible. [Evidence](../01-sandbox-api-base/README.md)
- [x] `Webhook delivery` disponible. [Evidence](../03-webhook-delivery/README.md)
- [ ] OpenAPI `v1` validado.

## 1. Contracto y Cliente

- [ ] `docs(payment-sandbox): define workflow consumer contract`
  - [ ] Definir el consumidor agnostico (`OrderProcessor` u otro workflow owner).
  - [ ] Definir los endpoints `v1` consumidos.
  - [ ] Definir reglas de idempotencia y versionado.
  - [ ] Definir errores estables de integracion.

## 2. Inbox y Estado Local

- [ ] `refactor(payment-sandbox): add workflow consumer inbox`
  - [ ] Persistir cada webhook antes de aplicar cambios de negocio.
  - [ ] Rechazar duplicados sin duplicar efectos.
  - [ ] Mantener trazabilidad de delivery e inbox.

## 3. Reconciliacion

- [ ] `refactor(payment-sandbox): add workflow reconciliation loop`
  - [ ] Comparar estado local con sandbox v1.
  - [ ] Registrar snapshots de reconciliacion.
  - [ ] Reportar mismatches sin mutar negocio.

## 4. OpenAPI y Vistas

- [ ] `docs(payment-sandbox): publish openapi v1 and browser views`
  - [ ] Validar `docs/openapi/payment-sandbox.v1.yaml`.
  - [ ] Publicar vista navegable con Swagger UI.
  - [ ] Publicar vista navegable con Redoc.

## 5. Tests

- [ ] `test(payment-sandbox): cover workflow consumer integration contract`
  - [ ] Test de inbox antes de side effects.
  - [ ] Test de idempotencia por delivery.
  - [ ] Test de reconciliacion.

## 6. Cierre

- [ ] `docs(payment-sandbox): mark workflow consumer integration complete`
  - [ ] Marcar evidencia en README.
  - [ ] Cerrar el checklist cuando el flujo este listo.
