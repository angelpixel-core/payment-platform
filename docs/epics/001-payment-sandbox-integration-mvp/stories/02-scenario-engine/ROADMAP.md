# Roadmap

## Nivel 1: Pragmatico

Objetivo: mejorar la estructura sin sobrediseñar.

- Mantener HTTP delgado.
- Separar dominio, `internal/application/commands`, `internal/application/queries`, `internal/application/support` y `adapters` de forma ligera.
- Definir repositorios y puertos basicos.
- Introducir tests por capa.
- Usar `MemoryStore` como infraestructura.

Cuando elegirlo:
- el sistema sigue pequeño o mediano.
- se prioriza velocidad de entrega.
- las consultas son simples.

## Nivel 2: Empresarial

Objetivo: preparar el sistema para operacion seria y crecimiento.

- Modular Monolith por capacidad de negocio.
- DDD ligero con agregados claros.
- CQRS parcial para separar comandos y consultas de lectura.
- Event-driven foundation: domain events tipados, `EventPublisher` como puerto y dispatcher in-process como primer paso.
- Observabilidad completa: logs, metrics, traces.
- Benchmarks, profiling y race detector.
- Hardening de errores, idempotencia y timeouts.

Cuando elegirlo:
- hay mas equipos o mas trafico.
- se necesitan cambios seguros y faciles de operar.
- la lectura y escritura empiezan a evolucionar distinto.

## Nivel 3: Estricto

Objetivo: maximizar aislamiento y formalidad.

- Clean Architecture estricta.
- Hexagonal completa con puertos bien definidos.
- DDD formal con agregados y eventos de dominio.
- CQRS completo con read models o proyecciones.
- Unit of Work para operaciones multi-repositorio.
- Outbox, bus externo y consumidores asincronos si hay integraciones reales.

Cuando elegirlo:
- la complejidad del dominio lo justifica.
- hay multiples flujos asynchronos o proyecciones pesadas.
- la consistencia y escalabilidad operativa requieren formalidad extra.

## Recomendacion Para Este Proyecto

- Base actual: Nivel 1.
- Siguiente paso natural: subir a Nivel 2.
- Reservar Nivel 3 para cuando el dominio o la carga lo exijan.

## Coverage

- Nivel 1: implementado.
- Nivel 2: implementado para esta story y su checklist de cierre.
- Nivel 3: pendiente.
- Ver detalle en [Roadmap Coverage](./ROADMAP_COVERAGE.md).

## Decision Notes

- `Modular Monolith` si aplica bien aqui.
- `Unit of Work` probablemente aparecera al introducir PostgreSQL real o multiples repositorios por caso de uso.
- Event-driven debe empezar como foundation en Nivel 2, no como bus distribuido completo.
- CQRS debe ser parcial al inicio; no merece complejidad completa sin una necesidad real.
