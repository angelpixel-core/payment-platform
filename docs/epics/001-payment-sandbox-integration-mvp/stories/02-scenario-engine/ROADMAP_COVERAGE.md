# Roadmap Coverage

## Purpose

This document maps each `ROADMAP.md` bullet to the current implementation state.

## Legend

- `implemented`: the bullet is covered by the current codebase and story docs.
- `partial`: part of the bullet is covered, but the whole item is not fully closed yet.
- `not covered`: the bullet is still aspirational.

## Nivel 1

| Roadmap bullet | Status | Evidence |
| --- | --- | --- |
| Mantener HTTP delgado | implemented | `IMPLEMENTATION_CHECKLIST.md` section 3, thin HTTP adapter changes |
| Separar dominio, `internal/application/commands`, `internal/application/queries`, `internal/application/support` y `adapters` de forma ligera | implemented | `ARCHITECTURE_PLAN.md`, `LEVEL2_MODULE_BOUNDARIES.md` |
| Definir repositorios y puertos basicos | implemented | `IMPLEMENTATION_CHECKLIST.md` sections 3 and 7.5 |
| Introducir tests por capa | implemented | `IMPLEMENTATION_CHECKLIST.md` across levels 1 and 2 |
| Usar `MemoryStore` como infraestructura | implemented | `LEVEL2_MODULE_BOUNDARIES.md`, memory adapter and tests |

## Nivel 2

| Roadmap bullet | Status | Evidence |
| --- | --- | --- |
| Modular Monolith por capacidad de negocio | implemented | `LEVEL2_MODULE_BOUNDARIES.md`, `ARCHITECTURE_PLAN.md`, checklist 7.1 |
| DDD ligero con agregados claros | implemented | `LEVEL2_DDD_RULES.md`, checklist 7.2 |
| CQRS parcial para separar comandos y consultas de lectura | implemented | `LEVEL2_MODULE_BOUNDARIES.md`, checklist 7.3 |
| Event-driven foundation: domain events tipados, `EventPublisher` como puerto y dispatcher in-process como primer paso | implemented | checklist 4 and 7.4 |
| Observabilidad completa: logs, metrics, traces | implemented | `OBSERVABILITY_METRICS.md`, checklist 7.7, 7.7.1-7.7.6 |
| Benchmarks, profiling y race detector | partial | `BENCHMARKING.md` and benchmark tests are in place; `go test -race` is documented but not yet closed as a tracked implementation step |
| Hardening de errores, idempotencia y timeouts | partial | error mapping and idempotent flows are covered; timeout hardening is still only a roadmap intent |

## Nivel 3

| Roadmap bullet | Status | Evidence |
| --- | --- | --- |
| Clean Architecture estricta | not covered | current structure is layered, but not strict clean architecture |
| Hexagonal completa con puertos bien definidos | partial | ports and adapters exist, but the full Level 3 strictness is not the target shape yet |
| DDD formal con agregados y eventos de dominio | partial | `LEVEL2_DDD_RULES.md` and typed events exist, but not a fully formal Level 3 DDD model |
| CQRS completo con read models o proyecciones | not covered | current CQRS is partial only |
| Unit of Work para operaciones multi-repositorio | implemented | checklist 7.5 and UoW metrics contract |
| Outbox, bus externo y consumidores asincronos si hay integraciones reales | partial | outbox is present; external bus and async consumers are still future work |

## Notes

- This map is a status bridge, not a replacement for `IMPLEMENTATION_CHECKLIST.md`.
- Some Level 3 bullets already have partial or full implementation support even though the roadmap reserves them for later.
