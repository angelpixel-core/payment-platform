# Propuesta de Refactor: Base del Nivel 1

## Objetivo

Evolucionar `payment-sandbox` hacia una base pragmatica de `Clean Architecture`, `Hexagonal` y `DDD` ligero, dejando preparada la ruta para `Modular Monolith`, `CQRS` parcial y eventos de dominio sin sobrediseñar el MVP.

## Decision

Implementar primero el **Nivel 1** y, dentro de ese mismo camino, introducir una **base event-driven temprana** como contrato interno.

- Nivel 1 cubre la separacion de capas y puertos.
- La base event-driven entra como fundamento en el siguiente paso natural del Nivel 2.
- El envio asincrono real, `outbox` y consumidores externos quedan para cuando exista necesidad real.

## Por Que Aqui

- Evita que la integracion event-driven aparezca tarde y de forma traumática.
- Permite modelar eventos desde temprano sin pagar el costo completo de infraestructura distribuida.
- Encaja bien con un monolito modular y con un futuro `CQRS` parcial.

## Alcance del Nivel 1

- Separar dominio, aplicacion y adaptadores.
- Mantener `server` como adaptador HTTP delgado.
- Mantener `Store` como puerto de persistencia.
- Formalizar errores de dominio y contratos internos.
- Preparar el terreno para eventos internos sin habilitar todavía mensajeria distribuida.

## Base Event-Driven

La decision es introducirlo en **Nivel 2**, pero como una preparacion que debe nacer desde el final del Nivel 1:

- eventos de dominio tipados
- un `EventPublisher` como puerto
- dispatcher in-process primero
- `outbox` cuando haya persistencia real y consistencia multi-repositorio

## Fuera de Alcance del Nivel 1

- `UnitOfWork`
- `outbox`
- bus externo
- proyecciones de lectura
- handlers asíncronos reales

## Criterios de Exito

- el contrato HTTP sigue igual
- la estructura del codigo ya muestra limites claros
- futuras integraciones event-driven no requieren reescribir el dominio
