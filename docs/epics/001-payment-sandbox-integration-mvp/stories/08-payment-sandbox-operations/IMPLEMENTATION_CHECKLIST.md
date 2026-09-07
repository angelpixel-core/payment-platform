---
id: IMPLEMENTATION_CHECKLIST
aliases: []
tags: []
status: in_progress
---

# Checklist de Implementacion por Commits - Payment Sandbox Operations

## Objetivo

Crear tooling operativo alrededor de `payment-sandbox` para escenarios, carga, observabilidad y entorno local reproducible.

## Regla

- Un commit por paso logico.
- El CLI vive dentro del binario `payment-sandbox`.
- El runtime de desarrollo usa PostgreSQL por defecto; los tests usan memory.

## Secuencia Propuesta

### 1. `docs(payment-sandbox): define operations runtime contract`

- [x] Alinear la story 08 con el vocabulario de artefactos: `sandbox`, `database`, `telemetry`, `dashboard`, `scenario-runner`. [Evidence](./README.md)
- [x] Fijar la postura del CLI: subcomandos bajo `payment-sandbox`, no binario hermano. [Evidence](./README.md)
- [x] Dejar claro que la app publica métricas, pero Prometheus/Grafana consumen esas métricas por scraping/visualización. [Evidence](./README.md)
- [x] Acordar que desarrollo usa PostgreSQL y tests usan memory. [Evidence](./README.md)
- [x] Archivos objetivo:
  - `docs/epics/001-payment-sandbox-integration-mvp/stories/08-payment-sandbox-operations/README.md`
  - `docs/epics/001-payment-sandbox-integration-mvp/stories/08-payment-sandbox-operations/IMPLEMENTATION_CHECKLIST.md`

### 2. `feat(payment-sandbox): add cobra-based scenario runner shell`

- [x] Crear el esqueleto Cobra dentro de `cmd/payment-sandbox`. [Evidence](../../../../../../apps/payment-sandbox/internal/cli/root.go)
- [x] Añadir subcomandos `simulate`, `replay`, `burst` y `seed`. [Evidence](../../../../../../apps/payment-sandbox/internal/cli/root.go)
- [x] Conectar cada subcomando con una interfaz de caso de uso, sin acoplarlo al transporte HTTP. [Evidence](../../../../../../apps/payment-sandbox/internal/application/operations/runner.go)
- [x] Archivos objetivo:
  - `apps/payment-sandbox/cmd/payment-sandbox/main.go` [Evidence](../../../../../../apps/payment-sandbox/cmd/payment-sandbox/main.go)
  - `apps/payment-sandbox/internal/bootstrap/*.go` si hace falta wiring
  - `apps/payment-sandbox/internal/*` para los casos de uso invocados

### 3. `docs(payment-sandbox): define load testing strategy`

- [x] Documentar `k6` como herramienta para secuencias de negocio. [Evidence](./README.md)
- [x] Scaffold de `k6` para secuencias de negocio. [Evidence](../../../../../apps/payment-sandbox/load-tests/k6/business-sequences.js)
- [x] Documentar `vegeta` como herramienta para presión HTTP reproducible. [Evidence](./README.md)
- [x] Scaffold de `vegeta` para presión HTTP reproducible. [Evidence](../../../../../apps/payment-sandbox/load-tests/vegeta/pressure.sh)
- [ ] Reservar goroutines Go para escenarios de dominio puntuales, no como reemplazo de load testing.
- [ ] Archivos objetivo:
  - `docs/epics/001-payment-sandbox-integration-mvp/stories/08-payment-sandbox-operations/README.md`
  - `apps/payment-sandbox/load-tests/k6/README.md`
  - `apps/payment-sandbox/load-tests/k6/business-sequences.js`
  - `apps/payment-sandbox/load-tests/vegeta/README.md`
  - `apps/payment-sandbox/load-tests/vegeta/pressure.sh`
  - (si se necesita) un documento nuevo de estrategia bajo la story 08

### 4. `feat(payment-sandbox): expose operational metrics`

- [ ] Publicar requests, latency, error rate, CPU y memoria.
- [ ] Publicar outbox/webhook retries y payloads procesados.
- [ ] Publicar métricas de negocio por minuto y reconciliation mismatches.
- [ ] No acoplar la app a dashboards concretos; solo exponer el surface de métricas.
- [ ] Archivos objetivo:
  - `apps/payment-sandbox/internal/adapters/observability/metrics/*`
  - `apps/payment-sandbox/internal/bootstrap/*`
  - `apps/payment-sandbox/internal/server/*` si hace falta wiring

### 5. `docs(payment-sandbox): define dashboards for infra and business metrics`

- [ ] Separar dashboards técnicos y de negocio.
- [ ] Usar Prometheus como colector y Grafana como visualización.
- [ ] Nombrar los dashboards por dominio, no por tecnología.
- [ ] Archivos objetivo:
  - `docs/epics/001-payment-sandbox-integration-mvp/stories/08-payment-sandbox-operations/README.md`
  - (si hace falta) docs adicionales de story 08

### 6. `feat(payment-sandbox): wire postgres as default dev runtime`

- [ ] Hacer que el runtime por defecto en desarrollo use PostgreSQL.
- [ ] Mantener memory como adaptador de tests.
- [ ] Dejar el wiring explícito en bootstrap/configuración.
- [ ] Archivos objetivo:
  - `apps/payment-sandbox/internal/sandbox/service.go`
  - `apps/payment-sandbox/internal/bootstrap/*.go`
  - `apps/payment-sandbox/cmd/payment-sandbox/main.go`

### 7. `feat(payment-sandbox): add docker compose and stack targets`

- [ ] Definir `sandbox`, `database`, `telemetry`, `dashboard`.
- [ ] Agregar `make stack/up` y `make stack/down`.
- [ ] Soportar detached mode y limpieza de volúmenes.
- [ ] Usar una imagen Go multi-stage pequeña y `postgres:17-alpine`.
- [ ] Archivos objetivo:
  - `docker-compose.yml`
  - `Makefile`
  - `apps/payment-sandbox/Dockerfile` o equivalente

### 8. `docs(payment-sandbox): mark operations story complete`

- [ ] Cerrar los checkboxes de la story 08 cuando la implementación esté lista.
- [ ] Actualizar el README del epic si corresponde.
- [ ] Archivos objetivo:
  - `docs/epics/001-payment-sandbox-integration-mvp/stories/08-payment-sandbox-operations/README.md`
  - `docs/epics/001-payment-sandbox-integration-mvp/README.md` si hace falta

## Notas

- `payment-sandbox` es el dominio del binario y de los subcomandos.
- `k6` y `vegeta` se complementan, no compiten.
- Prometheus/Grafana son para observabilidad; los dashboards de negocio viven encima de esas métricas.
