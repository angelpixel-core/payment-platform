# Architecture Plan

## Goal

Evolucionar `payment-sandbox` hacia una base profesional con `Clean Architecture`, `Hexagonal`, `DDD` ligero, `Modular Monolith` y `CQRS` parcial, sin perder la simplicidad del MVP.

## Current Baseline

- `internal/server` actua como adaptador HTTP.
- `internal/sandbox` concentra la logica de aplicacion y dominio hoy.
- `Store` ya abstrae persistencia en memoria.
- El comportamiento es determinista via `ScenarioEngine`.

## Target Shape

### Domain

- Entidades y agregados con invariantes propias.
- `PaymentIntent` como agregado principal probable.
- Value objects para `Amount`, `Currency`, `Scenario`, `Status`, `IdempotencyKey`.
- Errores de dominio tipados.

### Application

- Casos de uso por comando: `CreatePaymentIntent`, `ConfirmPaymentIntent`, `CapturePaymentIntent`, `CreateRefund`, `FinalizeProcessingPaymentIntent`.
- La capa aplica orquestacion y coordina puertos.
- Sin JSON, HTTP ni detalles de persistencia.

### Ports

- Entrada: comandos y queries.
- Salida: repositorios, clock, scenario resolver, idempotency, event publisher si aparece.

### Adapters

- HTTP como adaptador de entrada.
- MemoryStore y futuro PostgreSQL como adaptadores de salida.
- Observabilidad, logging y metrics como infraestructura transversal.

## Modular Monolith

Dividir por capacidad de negocio, no por tecnologia.

- `payments`
- `refunds`
- `scenarios`
- `shared` o `platform` solo para infraestructura comun

Regla: un modulo no importa internals de otro modulo; solo contracts puertos o APIs internas bien definidas.

## DDD

- Modelar lenguaje ubicuo claro.
- Mantener invariantes dentro del agregado.
- Evitar anemizar entidades con setters sin reglas.
- Introducir domain services solo cuando la regla no pertenezca a una entidad.
- Usar domain events si hay reacciones internas reales.

## Hexagonal

- El dominio no conoce HTTP, JSON, SQL ni memoria.
- La aplicacion no conoce detalles concretos de IO.
- Los adaptadores se enchufan por interfaces.

## CQRS

- Separar comandos y queries a nivel de uso.
- Mantener consultas simples al principio.
- Introducir read models solo cuando haga falta.
- No forzar CQRS completo si la lectura es trivial.

## Event-Driven

- Introducir eventos de dominio tempranamente como contrato interno.
- Exponer un `EventPublisher` como puerto en Nivel 2.
- Empezar con dispatcher in-process.
- Reservar `outbox`, bus externo y consumidores asincronos para una etapa posterior.
- La meta es evitar una integracion traumática mas adelante sin pagar todavia el costo completo de distribucion.

## Unit of Work

- Probablemente util cuando haya multiples escrituras que deban ser atomicas.
- Hoy no es estrictamente necesario con el MemoryStore, pero si lo sera cuando aparezca PostgreSQL real o mas de un agregado por caso de uso.
- Cuando llegue ese punto, la UoW deberia vivir como puerto de aplicacion, no dentro del dominio.

## Non-Functional Concerns

- Logging estructurado.
- Metrics y tracing.
- Benchmarks y profiling con `pprof`.
- Race detector en CI.
- Limites de concurrencia y timeouts.
- Reintentos y circuit breaking solo donde haya dependencias externas.

## Migration Strategy

1. Separar dominios y casos de uso del paquete actual.
2. Introducir puertos de repositorio y clock.
3. Mover `MemoryStore` a infraestructura.
4. Formalizar queries y comandos si la lectura crece.
5. Introducir foundation event-driven antes de escalar a bus externo.
6. Agregar PostgreSQL detras de los mismos puertos.
7. Introducir UoW cuando exista necesidad real de atomicidad multi-repositorio.

## Principles

- Preferir cambios pequenos y verificables.
- No introducir patrones por teoria si no hay dolor medido.
- La direccion es arquitectura robusta, no complejidad gratuita.
