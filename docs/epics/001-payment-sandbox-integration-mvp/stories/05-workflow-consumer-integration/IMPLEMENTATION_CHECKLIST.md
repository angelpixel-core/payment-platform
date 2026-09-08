---
id: IMPLEMENTATION_CHECKLIST
aliases: []
tags: []
status: in_progress
---

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
- [x] `Ledger and reporting` disponible. [Evidence](../04-ledger-reporting/README.md)
- [x] OpenAPI `v1` validado. [Evidence](../../../../openapi/payment-sandbox.v1.yaml)

## 1. Contracto y Cliente

- [x] `docs(payment-sandbox): define workflow consumer contract`
  - [x] Definir el consumidor agnostico como un workflow owner generico. [Evidence](./CONTRACT.md)
  - [x] Definir los endpoints `v1` consumidos. [Evidence](./CONTRACT.md)
  - [x] Definir reglas de idempotencia y versionado. [Evidence](./CONTRACT.md)
  - [x] Definir errores estables de integracion. [Evidence](./CONTRACT.md)

## 2. Inbox y Estado Local

- [x] `refactor(payment-sandbox): add workflow consumer inbox`
  - [x] Persistir cada webhook antes de aplicar cambios de negocio. [Evidence](./CONTRACT.md)
  - [x] Rechazar duplicados sin duplicar efectos. [Evidence](./CONTRACT.md)
  - [x] Mantener trazabilidad de delivery e inbox. [Evidence](./CONTRACT.md)

## 3. Reconciliacion

- [x] `refactor(payment-sandbox): add workflow reconciliation loop`
  - [x] Comparar estado local con sandbox v1. [Evidence](./workflow_reconciliation.rb)
  - [x] Registrar snapshots de reconciliacion. [Evidence](./workflow_reconciliation.rb)
  - [x] Reportar mismatches sin mutar negocio. [Evidence](./workflow_reconciliation.rb)

## 4. OpenAPI y Vistas

- [x] `docs(payment-sandbox): publish openapi v1 and browser views`
  - [x] Validar `docs/openapi/payment-sandbox.v1.yaml`.
  - [x] Publicar vista navegable con Swagger UI.
  - [x] Publicar vista navegable con Redoc.

## 5. Tests

- [x] `test(payment-sandbox): cover workflow consumer integration contract`
  - [x] Test de inbox antes de side effects. [Evidence](./workflow_inbox_test.rb)
  - [x] Test de idempotencia por delivery. [Evidence](./workflow_inbox_test.rb)
  - [x] Test de reconciliacion. [Evidence](./workflow_reconciliation_test.rb)

## 6. Cierre

- [ ] `docs(payment-sandbox): mark workflow consumer integration complete`
  - [x] Marcar evidencia en README. [Evidence](./README.md)
  - [ ] Cerrar el checklist cuando el flujo este listo.

## Pendiente

- La integración productiva Rails queda pendiente. Las bases pre-implementadas para el handoff están documentadas en [Implementation Evidence](./README.md#implementation-evidence).
