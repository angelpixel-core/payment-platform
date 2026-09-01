# Checklist de Implementacion por Commits - Nivel 3

## Objetivo

Elevar `payment-sandbox` al Nivel 3 sin perder los contratos ya cerrados.

## Referencias

- [Roadmap](./ROADMAP.md)
- [Roadmap Coverage](./ROADMAP_COVERAGE.md)
- [Implementation Checklist](./IMPLEMENTATION_CHECKLIST.md)

## Regla

- Un commit por paso logico.
- Antes de trabajar un item de Nivel 3, dejar claros sus prerequisitos y la evidencia que los respalda.

## 0. Prerequisitos Confirmados

- [x] Monolito modular y limites por capacidad de negocio. [Evidence](./LEVEL2_MODULE_BOUNDARIES.md)
- [x] DDD ligero con agregado principal y reglas de transicion. [Evidence](./LEVEL2_DDD_RULES.md)
- [x] CQRS parcial con commands y queries separados. [Evidence](./LEVEL2_MODULE_BOUNDARIES.md)
- [x] Base event-driven in-process con eventos tipados. [Evidence](./IMPLEMENTATION_CHECKLIST.md)
- [x] Observabilidad con logs, metrics y traces. [Evidence](./OBSERVABILITY_METRICS.md)
- [x] Benchmarks y profiling disponibles. [Evidence](./BENCHMARKING.md)
- [x] Unit of Work disponible. [Evidence](./OBSERVABILITY_METRICS.md)
- [x] Outbox foundation disponible. [Evidence](./OBSERVABILITY_METRICS.md)

## 1. Clean Architecture Estricta

- [ ] `refactor(payment-sandbox): make boundaries strict`
  - [x] Prereq: el split domain/application/adapters ya existe. [Evidence](./LEVEL2_MODULE_BOUNDARIES.md)
  - [x] Prereq: HTTP ya es un adaptador delgado. [Evidence](./IMPLEMENTATION_CHECKLIST.md)
  - [x] Prereq: los puertos base ya estan definidos. [Evidence](./ARCHITECTURE_PLAN.md)
  - [ ] Definir reglas de dependencia estrictas para la forma final de Nivel 3.
  - [ ] Mover cualquier helper restante a la capa correcta.

## 2. Hexagonal Completa

- [ ] `refactor(payment-sandbox): harden hexagonal boundaries`
  - [x] Prereq: ports y adapters ya estan separados. [Evidence](./LEVEL2_MODULE_BOUNDARIES.md)
  - [x] Prereq: los adapters concretos ya implementan IO y tiempo. [Evidence](./ARCHITECTURE_PLAN.md)
  - [ ] Auditar fugas entre application y adapters.
  - [ ] Forzar que toda dependencia inbound y outbound pase por un puerto.

## 3. DDD Formal

- [ ] `refactor(payment-sandbox): formalize aggregates and domain events`
  - [x] Prereq: `PaymentIntent` ya tiene reglas de agregado. [Evidence](./LEVEL2_DDD_RULES.md)
  - [x] Prereq: los eventos tipados ya salen del flujo in-process. [Evidence](./IMPLEMENTATION_CHECKLIST.md)
  - [ ] Promover el modelo de eventos a una forma de dominio mas formal.
  - [ ] Endurecer boundaries de agregado y services solo donde haga falta.

## 4. CQRS Completo

- [ ] `refactor(payment-sandbox): introduce read models and projections`
  - [x] Prereq: commands y queries ya estan separados. [Evidence](./LEVEL2_MODULE_BOUNDARIES.md)
  - [x] Prereq: existen modelos de lectura para los flujos actuales. [Evidence](./LEVEL2_MODULE_BOUNDARIES.md)
  - [ ] Agregar read models o proyecciones explicitas donde el modelo actual no alcance.
  - [ ] Mantener la side write aislada de optimizaciones de lectura.

## 5. UoW y Atomicidad Multi-Repo

- [x] `refactor(payment-sandbox): introduce unit of work port`
  - [x] Prereq: la UoW ya figura como implementada en Nivel 2. [Evidence](./IMPLEMENTATION_CHECKLIST.md)
  - [x] Prereq: existe contrato de metricas para UoW. [Evidence](./OBSERVABILITY_METRICS.md)
  - [x] Prereq: la implementacion memory/postgres ya esta prevista por contrato. [Evidence](./OBSERVABILITY_METRICS.md)
  - [ ] Extender la UoW a cualquier flujo nuevo que requiera atomicidad multi-repositorio.

## 6. Outbox, Bus Externo y Consumidores Asincronos

- [ ] `refactor(payment-sandbox): extend outbox to external delivery`
  - [x] Prereq: la base de outbox ya existe. [Evidence](./IMPLEMENTATION_CHECKLIST.md)
  - [x] Prereq: las metricas de outbox ya existen. [Evidence](./OBSERVABILITY_METRICS.md)
  - [ ] Integrar bus externo solo cuando exista una integracion real que lo justifique.
  - [ ] Agregar consumidores asincronos solo cuando el trabajo cross-process sea necesario.

## 7. Acceptance

- [ ] `test(payment-sandbox): keep level 3 contract stable`
  - [ ] Verificar que los cambios de Nivel 3 no rompen el contrato HTTP publico.
  - [ ] Verificar que `go test ./...` sigue verde.
  - [ ] Verificar que `go test -race ./...` y los comandos de profiling siguen usables.
