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

## Regla Arquitectonica

- `cmd/` contiene solo entrypoints.
- `internal/bootstrap/` ensambla la aplicacion y no contiene reglas de negocio.
- `internal/domain/` contiene el modelo y las invariantes del negocio.
- `internal/application/` orquesta casos de uso y depende solo de `domain` y `ports`.
- `internal/ports/` define contratos, sin dependencias de adaptadores.
- `internal/adapters/` implementa IO concreto y depende de `domain` y `ports`.
- `internal/server/` y `internal/sandbox/` no deben ser capas de negocio; si existen, quedan limitadas a wiring temporal o desaparecen a favor de `internal/bootstrap/`.
- Ningun paquete de produccion debe importar internals de otro paquete hermano fuera de `ports` y contratos publicos.

## Matriz de Dependencias

| Capa | Puede depender de | No puede depender de |
| --- | --- | --- |
| `cmd/` | `internal/bootstrap/` | `domain`, `application`, `ports`, `adapters`, `server`, `sandbox` |
| `internal/bootstrap/` | `application`, `ports`, `adapters`, `server` si solo actua como wiring | `domain` para reglas, `adapter` concretos con logica de negocio |
| `internal/domain/` | stdlib y si misma | `application`, `ports`, `adapters`, `server`, `sandbox`, `bootstrap` |
| `internal/application/` | `domain`, `ports`, `application/support` | `adapters`, `server`, `sandbox`, `bootstrap` |
| `internal/ports/` | `domain` y stdlib | `application`, `adapters`, `server`, `sandbox`, `bootstrap` |
| `internal/adapters/` | `domain`, `ports`, stdlib, libs externas | `application`, `server`, `sandbox`, `bootstrap` |
| `internal/server/` | `bootstrap`, `application`, `ports`, `adapters` solo para wiring | reglas de negocio, dependencias ciclicas con `sandbox` |
| `internal/sandbox/` | `bootstrap`, `application`, `ports`, `adapters` solo para wiring temporal | reglas de negocio, importaciones horizontales entre capas |

## Criterios de Aplicacion

- Si un paquete decide una regla de negocio, pertenece a `internal/domain/`.
- Si un paquete orquesta un caso de uso, pertenece a `internal/application/`.
- Si un paquete traduce o conecta IO, pertenece a `internal/adapters/`.
- Si un paquete solo arma dependencias, pertenece a `internal/bootstrap/` o `cmd/`.
- Si un paquete existe solo por compatibilidad historica, debe ser eliminado o degradado a wiring sin logica.

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

- [x] `refactor(payment-sandbox): make boundaries strict`
  - [x] Prereq: el split domain/application/adapters ya existe. [Evidence](./LEVEL2_MODULE_BOUNDARIES.md)
  - [x] Prereq: HTTP ya es un adaptador delgado. [Evidence](./IMPLEMENTATION_CHECKLIST.md)
  - [x] Prereq: los puertos base ya estan definidos. [Evidence](./ARCHITECTURE_PLAN.md)
  - [x] Definir reglas de dependencia estrictas para la forma final de Nivel 3.
  - [x] Convertir las reglas en un boundary test automatico. [Evidence](../../../../../apps/payment-sandbox/internal/boundaries_test.go)
  - [x] Mover cualquier helper restante a la capa correcta.
    - [x] Helpers HTTP relocados a `internal/adapters/inbound/http`. [Evidence](../../../../../apps/payment-sandbox/internal/adapters/inbound/http/json.go)
    - [x] Wiring HTTP relocado a `internal/bootstrap`. [Evidence](../../../../../apps/payment-sandbox/internal/bootstrap/http.go)
    - [x] Scenario engine relocado a `internal/application/support/scenarios`. [Evidence](../../../../../apps/payment-sandbox/internal/application/support/scenarios/engine.go)

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

- [x] `test(payment-sandbox): keep level 3 contract stable`
  - [x] Verificar que los cambios de Nivel 3 no rompen el contrato HTTP publico. [Evidence](../../../../../apps/payment-sandbox/internal/server/server_test.go)
  - [x] Verificar que `go test ./...` sigue verde.
  - [x] Verificar que `go test -race ./...` y los comandos de profiling siguen usables. [Evidence](../../../../../apps/payment-sandbox/internal/sandbox/bench_test.go)
