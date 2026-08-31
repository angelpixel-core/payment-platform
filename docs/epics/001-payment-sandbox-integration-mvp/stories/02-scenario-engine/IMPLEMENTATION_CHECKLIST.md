# Checklist de Implementacion por Commits

## Objetivo

Ejecutar el Nivel 1 sobre `payment-sandbox` con cambios pequenos, verificables y ordenados por commit.

## Regla

- Un commit por paso logico.
- Cada commit debe dejar la suite verde o, como minimo, sin romper el contrato publico.
- La base event-driven solo se prepara; no se introduce bus externo todavia.

## 1. Estructura de Capas

- [x] `chore(payment-sandbox): create domain/application/ports folders`
- [x] `refactor(payment-sandbox): move domain models to domain package`
- [x] `refactor(payment-sandbox): move domain errors to domain package`

## 2. Casos de Uso

- [x] `refactor(payment-sandbox): extract payment use cases into application package`
- [x] `refactor(payment-sandbox): keep service orchestration thin`
- [x] `test(payment-sandbox): preserve create confirm capture refund flows`

## 3. Puertos y Adaptadores

- [x] `refactor(payment-sandbox): define repository and clock ports`
- [x] `refactor(payment-sandbox): move memory store behind infrastructure adapter`
- [x] `refactor(payment-sandbox): keep HTTP as thin input adapter`

## 4. Base Event-Driven

- [x] `refactor(payment-sandbox): introduce in-process domain event publisher`
- [x] `refactor(payment-sandbox): emit typed events from use cases`
- [x] `test(payment-sandbox): validate event emission contract`

## 5. Ajuste de Contratos

- [x] `refactor(payment-sandbox): align error mapping with layered packages`
- [x] `refactor(payment-sandbox): keep request and response shapes unchanged`
- [x] `test(payment-sandbox): verify HTTP contract remains stable`

## 6. Cierre del Nivel 1

- [x] `docs(payment-sandbox): mark level 1 refactor complete`
- [x] `test(payment-sandbox): go test ./...`

## 7. Nivel 2

### 7.1 Modular Monolith

- [x] `docs(payment-sandbox): define level 2 module boundaries`
- [x] `refactor(payment-sandbox): split payment sandbox into business modules`
- [x] `refactor(payment-sandbox): enforce module import boundaries`

### 7.2 DDD Ligero

- [x] `refactor(payment-sandbox): formalize payment intent aggregate rules`
- [x] `refactor(payment-sandbox): extract value objects for amount and currency`
- [x] `refactor(payment-sandbox): move transition invariants into the domain`

### 7.3 CQRS Parcial

- [x] `refactor(payment-sandbox): separate command handlers from queries`
- [x] `refactor(payment-sandbox): add read-only query models where needed`
- [x] `test(payment-sandbox): preserve command and query behavior`

### 7.4 Event-Driven Foundation

- [x] `refactor(payment-sandbox): add internal event handlers for payment events`
- [x] `refactor(payment-sandbox): prepare outbox-backed event publication`
- [x] `test(payment-sandbox): validate internal event handler contract`

### 7.5 Unit of Work

- [x] `refactor(payment-sandbox): introduce unit of work port`
- [x] `refactor(payment-sandbox): apply unit of work to multi-repository use cases`
- [x] `test(payment-sandbox): preserve atomicity across repository writes`

### 7.6 Infra y Adaptadores

- [x] `refactor(payment-sandbox): add postgres adapter behind ports`
- [ ] `refactor(payment-sandbox): keep memory adapter for tests and sandbox`
- [ ] `test(payment-sandbox): run repository contract tests against postgres`

### 7.7 Observabilidad y Rendimiento

- [ ] `refactor(payment-sandbox): add structured logging and trace hooks`
- [ ] `refactor(payment-sandbox): expose metrics for payment flows`
- [ ] `test(payment-sandbox): add benchmarks and profiling checks`

### 7.8 Cierre del Nivel 2

- [ ] `docs(payment-sandbox): mark level 2 refactor complete`
- [ ] `test(payment-sandbox): go test ./...`

## Notas

- `UnitOfWork`, `outbox` y bus externo quedan fuera de este checklist.
- Si un paso crece demasiado, dividirlo en dos commits antes de seguir.
