# Checklist de Refactor del Payment Sandbox

## Objetivo

Dividir la implementación del sandbox en unidades más pequeñas sin cambiar el contrato HTTP.

## Fase 1: Extraer el Motor de Escenarios

- [x] Crear `internal/sandbox/scenario_engine.go`.
- [x] Mover la resolución de escenarios a un tipo dedicado `ScenarioEngine`.
- [x] Mantener `X-Sandbox-Scenario` con mayor prioridad que `payment_method_token`.
- [x] Mantener el mapeo de token a escenario dentro del motor o su configuración.
- [x] Agregar tests para prioridad del header, fallback por token y errores de escenario desconocido.

## Fase 2: Extraer la Interfaz de Store

- [x] Definir una interfaz `Store` para payment intents, attempts, charges y refunds.
- [x] Sacar los mapas en memoria y el locking fuera de `Service`.
- [x] Implementar `MemoryStore` como primer backend de store.
- [x] Dejar `Service` responsable solo de la orquestación.
- [x] Agregar tests que demuestren que el comportamiento HTTP no cambió.

## Fase 3: Simplificar el Servidor HTTP

- [x] Extraer helpers compartidos de parseo JSON en `internal/server`.
- [x] Extraer helpers compartidos de mapeo de errores en `internal/server`.
- [x] Mantener los handlers delgados y de propósito único.
- [x] Preservar las formas existentes de request/response.

## Fase 4: Normalizar Errores

- [x] Consolidar todos los errores de API en una sola forma tipada.
- [x] Mantener estables `code`, `message` y `status`.
- [x] Asegurar que la serialización de errores siga siendo idéntica para los clientes.
- [x] Agregar tests para errores de validación y de escenarios.

## Fase 5: Convertir Escenarios y Workflows en Table-Driven Tests

- [ ] Convertir los tests de escenarios en tests table-driven.
- [ ] Convertir los tests de workflow en tests table-driven donde tenga sentido.
- [ ] Agregar cobertura explícita para la finalización de `processing_then_succeeded`.
- [ ] Mantener cada test enfocado en un solo comportamiento.

## Orden

1. Motor de escenarios.
2. Interfaz de store.
3. Simplificación del servidor HTTP.
4. Normalización de errores.
5. Tests table-driven.

## Terminado Cuando

- [ ] El contrato HTTP público no cambió.
- [ ] `go test ./...` pasa.
- [ ] El sandbox sigue siendo demostrable manualmente.
- [ ] El storage futuro en PostgreSQL puede agregarse detrás de la nueva interfaz de store.
