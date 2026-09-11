# Payment Platform

> A backend payment platform sandbox focused on API design, payment workflows,
> domain boundaries, reproducible local environments, and developer-oriented
> documentation.

The project provides a runnable payment sandbox API with an OpenAPI contract,
a local containerized development stack, automated test targets, and a
published Swagger UI for exploring the API.

> This project is currently under active development and is intended as an
> engineering and architecture sandbox rather than a production payment
> processor.

## API Documentation

The API is documented using OpenAPI and exposed through Swagger UI.
**[Open Swagger UI](https://angelpixel-core.github.io/payment-platform/)**
The OpenAPI contract is versioned under:

```sh
apps/payment-sandbox/docs/openapi/v1/
```

The published documentation allows the API contract and available endpoints
to be explored without running the project locally.

---

## Current Status

The current implementation provides:

- Payment sandbox API
- Versioned OpenAPI specification
- Published Swagger UI
- Reproducible local development stack
- Docker-based service orchestration
- Unit, integration, contract, and load-test targets
- Database seeding and smoke-test workflows
- Make-based developer commands

The following areas are planned or still evolving:

- SDK
- Web-based API interface
- Production-oriented deployment
- Further hardening and operational maturity

The project is intentionally being developed incrementally, with the API
contract and backend foundation taking priority before additional clients
and interfaces are introduced.

---

## Quick Start

Requirements

- Docker
- Docker Compose
- Make

Start the local stack

```sh
make stack/up ARGS="--build -d"
```

Once the stack is running, the sandbox API is available at:

```
http://localhost:10201
```

The API can also be explored locally through its Swagger documentation.

---

## Local Stack

The local development environment is operated through Make targets.

Start

```sh
make stack/up ARGS="--build -d"
```

Seed

```sh
make stack/seed
```

Smoke test

```sh
make stack/smoke
```

````
Stop
```sh
make stack/down
````

Reset

```sh
make stack/reset
```

It removes local volumes and can be used when a clean local database is required.

Compose options can be passed through ARGS.

For example:

```sh
make stack/up ARGS="--build -d"
```

To restart the stack:

```sh
make stack/down
make stack/up ARGS="-d"
```

Inside Docker Compose, services should reach the sandbox API using the
service hostname:

```txt
payment-sandbox
```

Local configuration is intended to be provided through environment variables.
Production secrets must never be committed to the repository.

---

## Testing

The project exposes separate targets for different testing concerns.

Unit tests

```sh
make test-unit
```

Integration tests

```sh
make test-integration
```

Contract tests

```sh
make test-contract
```

Business load tests

```sh
make test-load-business
```

Pressure load tests

```sh
make test-load-pressure
```

The separation between unit, integration, contract, and load testing is
intentional: each target exercises a different level of system behavior.

---

## Project Structure

The repository is organized around the payment sandbox application and its
supporting documentation.

```txt
apps/
└── payment-sandbox/
    ├── cmd/
    ├── docs/
    │   ├── openapi/
    │   │   └── v1/
    │   ├── redoc/
    │   └── swagger/
    ├── internal/
    │   ├── adapters/
    │   ├── application/
    │   ├── bootstrap/
    │   ├── cli/
    │   ├── docs/
    │   ├── domain/
    │   ├── ports/
    │   ├── sandbox/
    │   └── server/
    ├── Dockerfile
    ├── go.mod
    └── go.sum
```

The payment-sandbox application contains the API implementation, while the
docs directory contains the API contract and developer-facing API
documentation.

---

## Documentation

The project currently exposes two complementary forms of API documentation:

- OpenAPI — machine-readable API contract
- Swagger UI — interactive API exploration

The published Swagger UI is available here:

🌐 [Payment Platform API](https://angelpixel-core.github.io/payment-platform)

---

## Roadmap

API & Developer Experience

- Payment sandbox API
- OpenAPI specification
- Swagger UI
- Local containerized environment
- Automated testing targets
- SDK
- Additional API examples

Interfaces

- Web-based API interface

Operations

- Production deployment
- Further observability and operational hardening
- Production-oriented configuration management

The roadmap is intentionally incremental. The current priority is establishing
a reliable API and contract foundation before expanding the number of clients
and interfaces.

---

## Project Status

This repository is an evolving engineering project.

The current milestone is focused on making the payment API runnable,
documented, testable, and easy to explore locally or through the published
Swagger documentation.

Future iterations will extend the developer experience through an SDK,
additional interfaces, and more production-oriented infrastructure.

---
