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

- [ ] `docs(payment-sandbox): mark level 1 refactor complete`
- [ ] `test(payment-sandbox): go test ./...`

## Notas

- `UnitOfWork`, `outbox` y bus externo quedan fuera de este checklist.
- Si un paso crece demasiado, dividirlo en dos commits antes de seguir.
