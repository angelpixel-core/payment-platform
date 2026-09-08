---
id: IMPLEMENTATION_CHECKLIST
aliases: []
tags: []
status: done
---

# Checklist de Implementacion por Commits - Ledger and Reporting

## Objetivo

Completar la capa de ledger y reporting para que Rails pueda reconciliar estado local contra sandbox truth con entradas inmutables, balances derivados y snapshots exportables.

## Regla

- Un commit por paso logico.
- Las proyecciones de reporte deben seguir siendo read-only.
- El ledger debe ser append-only e inmutable.

## 0. Evidencias Ya Confirmadas

- [x] `Transaction report endpoint`. [Evidence](./TRANSACTION_REPORT_CONTRACT.md)
- [x] `Balance projection`. [Evidence](./BALANCE_PROJECTION_CONTRACT.md)
- [x] `Daily settlement projection`. [Evidence](./SETTLEMENT_PROJECTION_CONTRACT.md)

## 1. `feat(payment-sandbox): add immutable ledger entries`

- [x] Definir entradas contables append-only para movimientos financieros. [Evidence](./README.md)
- [x] Registrar create, confirm, capture y refund como eventos/entradas trazables. [Evidence](../../../../../apps/payment-sandbox/internal/application/commands/payments/service.go) [Evidence](../../../../../apps/payment-sandbox/internal/application/commands/refunds/service.go)
- [x] Mantener la derivacion de la proyeccion separada de la persistencia del ledger.
- [x] Archivos objetivo:
  - `apps/payment-sandbox/internal/domain/*.go`
  - `apps/payment-sandbox/internal/adapters/persistence/*`
  - `apps/payment-sandbox/internal/application/*`

## 2. `feat(payment-sandbox): derive balances from ledger entries`

- [x] Derivar `available`, `reserved` y `liquidable` desde las entradas. [Evidence](../../../../../apps/payment-sandbox/internal/application/queries/payments/reports.go)
- [x] Mantener `BalanceProjectionView` como read model read-only.
- [x] Cubrir las reglas de refund y capture manual/automatic. [Evidence](../../../../../apps/payment-sandbox/internal/application/commands/payments/service.go) [Evidence](../../../../../apps/payment-sandbox/internal/application/commands/refunds/service.go)
- [x] Archivos objetivo:
  - `apps/payment-sandbox/internal/application/queries/payments/reports.go`
  - `apps/payment-sandbox/internal/application/queries/payments/service_test.go`
  - `apps/payment-sandbox/internal/sandbox/service_test.go`

## 3. `feat(payment-sandbox): represent fees and refunds in reporting`

- [x] Hacer visibles fees y refunds en el transaction report. [Evidence](../../../../../../apps/payment-sandbox/internal/application/queries/payments/reports.go)
- [x] Mantener trazabilidad hasta charge/payment intent.
- [x] Ajustar contratos de reporte si hace falta. [Evidence](./TRANSACTION_REPORT_CONTRACT.md)
- [x] Archivos objetivo:
  - `docs/epics/001-payment-sandbox-integration-mvp/stories/04-ledger-reporting/TRANSACTION_REPORT_CONTRACT.md`
  - `apps/payment-sandbox/internal/application/queries/payments/reports.go`
  - `apps/payment-sandbox/internal/application/queries/payments/service_test.go`

## 4. `feat(payment-sandbox): add reconciliation snapshot export`

- [x] Definir snapshot exportable y comparable para Rails. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/inbound/http/server.go)
- [x] Garantizar que el snapshot sea estable y diff-friendly. [Evidence](../../../../../apps/payment-sandbox/internal/application/queries/payments/reports.go)
- [x] Exponer suficiente data para validar against sandbox truth. [Evidence](./TRANSACTION_REPORT_CONTRACT.md)
- [x] Archivos objetivo:
  - `docs/epics/001-payment-sandbox-integration-mvp/stories/04-ledger-reporting/README.md`
  - `apps/payment-sandbox/internal/application/queries/payments/reports.go`
  - `apps/payment-sandbox/internal/docs/openapi/*` si se necesita contrato

## 5. `docs(payment-sandbox): close ledger reporting story`

- [x] Marcar los checkboxes de la story 04 solo cuando haya evidencia real.
- [x] Actualizar README de la story con el estado final.
- [x] Cerrar el checklist cuando el flujo quede listo.
- [x] Archivos objetivo:
  - `docs/epics/001-payment-sandbox-integration-mvp/stories/04-ledger-reporting/README.md`
  - `docs/epics/001-payment-sandbox-integration-mvp/README.md` si hace falta

## Notas

- El transaction report ya actua como read model de reconciliation.
- `balance_projection` y `settlement_projection` ya existen; el trabajo pendiente es hacerlos depender de un ledger más formal si queremos llegar al modelo completo.
- `fees` y `refunds` deben quedar traceables y no romper la estabilidad del reporte.
