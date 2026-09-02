# Architecture Plan

## Goal

Evolucionar `payment-sandbox` hacia una base profesional con `Clean Architecture`, `Hexagonal`, `DDD` ligero, `Modular Monolith` y `CQRS` parcial, sin perder la simplicidad del MVP.

## Current Baseline

- `internal/server` actua como adaptador HTTP.
- `internal/application/commands` concentra los casos de uso de escritura.
- `internal/application/queries/payments` concentra los casos de uso de lectura.
- `internal/application/support/observability` concentra el soporte transversal de observabilidad.
- `internal/sandbox` concentra la orquestacion del sandbox y el motor de escenarios.
- `internal/ports` define los contratos de la aplicacion.
- `internal/adapters` contiene HTTP, messaging, persistence y time.
- El comportamiento es determinista via `ScenarioEngine`.

## Target Shape

### Domain

- Entidades y agregados con invariantes propias.
- `PaymentIntent` como agregado principal probable.
- Value objects para `Amount`, `Currency`, `Scenario`, `Status`, `IdempotencyKey`.
- Errores de dominio tipados.

### Application

- Casos de uso de escritura en `internal/application/commands/payments` y `internal/application/commands/refunds`.
- Casos de uso de lectura en `internal/application/queries/payments`.
- Soporte transversal en `internal/application/support/observability`.
- La capa aplica orquestacion y coordina puertos.
- Sin JSON, HTTP ni detalles de persistencia.

### Ports

- Contratos de entrada y salida para la aplicacion.
- `Clock`, `Store`, `ScenarioResolver`, `EventPublisher` y `UnitOfWork`.

### Adapters

- `adapters/inbound/http` como entrada HTTP.
- `adapters/messaging/{inprocess,outbox}` para eventos.
- `adapters/persistence/{memory,postgres}` para almacenamiento.
- `adapters/time/system` para el reloj del sistema.
- `adapters/observability/*` para integraciones de telemetria cuando aparezcan.

## Modular Monolith

Dividir por capacidad de negocio, no por tecnologia.

- `internal/application/commands/payments`
- `internal/application/commands/refunds`
- `internal/application/queries/payments`
- `sandbox` para el motor de escenarios y wiring de demo
- `shared` solo para utilidades realmente neutras

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
- Gate de entrada: solo integrar bus externo cuando exista una integracion real que lo justifique.
- Ejemplos de casos que si lo justificarian: `Kafka` para fanout o integracion de eventos, `SQS` para colas durables con reintentos, `RabbitMQ` para ruteo y workers, o un partner externo custom con contrato real de webhook/EDI/API.
- Si no hay un requerimiento cross-process concreto, mantener el flujo in-process y/o dentro de la base de outbox actual.
- La meta es evitar una integracion traumática mas adelante sin pagar todavia el costo completo de distribucion.

## Unit of Work

- Probablemente util cuando haya multiples escrituras que deban ser atomicas.
- Hoy no es estrictamente necesario con el MemoryStore, pero si lo sera cuando aparezca PostgreSQL real o mas de un agregado por caso de uso.
- Cuando llegue ese punto, la UoW deberia vivir como puerto de aplicacion, no dentro del dominio.
- Caso futuro tipico: registrar una operacion financiera y su contrapartida contable/ledger en la misma transaccion.
- Regla practica: si una escritura no puede quedar pendiente para reconciliacion o consistencia eventual, debe entrar en la UoW.

## Non-Functional Concerns

- Logging estructurado.
- Metrics y tracing.
- Benchmarks y profiling con `pprof`.
- Race detector en CI.
- Limites de concurrencia y timeouts.
- Reintentos y circuit breaking solo donde haya dependencias externas.

## Migration Strategy

1. Separar comandos, queries y soporte transversal.
2. Mantener puertos de repositorio, reloj, escenarios y eventos como contratos.
3. Mover adapters concretos bajo `adapters/*`.
4. Formalizar queries y comandos si la lectura crece.
5. Introducir foundation event-driven antes de escalar a bus externo.
6. Agregar PostgreSQL detras de los mismos puertos.
7. Mantener UoW cuando exista necesidad real de atomicidad multi-repositorio.

## Principles

- Preferir cambios pequenos y verificables.
- No introducir patrones por teoria si no hay dolor medido.
- La direccion es arquitectura robusta, no complejidad gratuita.
