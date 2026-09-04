# Checklist de Implementacion por Commits - PaymentPlatformSDK

## Objetivo

Publicar un gem Ruby reutilizable que consuma `payment-sandbox` v1 con una API pequeña, estable y versionada.

## Regla

- Un commit por paso logico.
- El gem debe poder publicarse y reutilizarse sin exponer detalles internos del sandbox.
- El contrato wire-level vive en OpenAPI `v1`.

## 1. Contrato del SDK

- [ ] `docs(payment-platform-sdk): define public ruby gem contract`
  - [ ] Definir clase principal `PaymentPlatform::Client`.
  - [ ] Definir configuracion publica.
  - [ ] Definir errores estables de la gem.
  - [ ] Definir endpoints `v1` soportados.

## 2. Transporte y Autenticacion

- [ ] `feat(payment-platform-sdk): add http transport and auth`
  - [ ] Soportar base URL configurable.
  - [ ] Soportar API key y tenant scope.
  - [ ] Soportar timeout y retries.

## 3. Operaciones de Pago

- [ ] `feat(payment-platform-sdk): implement payment operations`
  - [ ] `create_payment_intent`
  - [ ] `confirm_payment_intent`
  - [ ] `capture_payment_intent`
  - [ ] `create_refund`
  - [ ] `get_payment_intent`

## 4. Idempotencia y Errores

- [ ] `feat(payment-platform-sdk): normalize idempotency and errors`
  - [ ] Generar idempotency key para mutaciones.
  - [ ] Mapear respuestas de error a excepciones Ruby.
  - [ ] Preservar status code, error code y message.

## 5. Tests y Contrato Vivo

- [ ] `test(payment-platform-sdk): validate gem against openapi contract`
  - [ ] Test de requests con stubs.
  - [ ] Test de responses y error mapping.
  - [ ] Test de compatibilidad contra OpenAPI `v1`.

## 6. Publicacion

- [ ] `release(payment-platform-sdk): prepare public gem packaging`
  - [ ] Definir gemspec.
  - [ ] Definir versionado semver.
  - [ ] Definir changelog/release process.
  - [ ] Documentar instalacion y uso desde Rails.
