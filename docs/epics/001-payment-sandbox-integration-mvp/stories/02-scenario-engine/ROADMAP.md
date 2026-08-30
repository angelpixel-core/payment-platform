# Roadmap

## Nivel 1: Pragmatico

Objetivo: mejorar la estructura sin sobrediseñar.

- Mantener HTTP delgado.
- Separar dominio, aplicacion y adaptadores de forma ligera.
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
- Outbox / event dispatcher si hay integraciones asincronas.

Cuando elegirlo:
- la complejidad del dominio lo justifica.
- hay multiples flujos asynchronos o proyecciones pesadas.
- la consistencia y escalabilidad operativa requieren formalidad extra.

## Recomendacion Para Este Proyecto

- Base actual: Nivel 1.
- Siguiente paso natural: subir a Nivel 2.
- Reservar Nivel 3 para cuando el dominio o la carga lo exijan.

## Decision Notes

- `Modular Monolith` si aplica bien aqui.
- `Unit of Work` probablemente aparecera al introducir PostgreSQL real o multiples repositorios por caso de uso.
- CQRS debe ser parcial al inicio; no merece complejidad completa sin una necesidad real.
