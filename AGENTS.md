# ChefBook Backend Agents Guide

This directory contains the ChefBook backend platform.

## Architecture

- Go microservice architecture
- `api-gateway` is the external entrypoint for client traffic
- `services/` contains domain microservices
- `common-lib/` contains reusable infrastructure libraries
- each service owns its own storage and deployment configuration

## Backend Rules

- Keep business logic inside the owning service; do not move domain behavior into `common-lib`.
- Treat `common-lib` as reusable infrastructure or utility code, not a dumping ground for service-specific logic.
- Preserve service isolation: avoid coupling services through direct database assumptions.
- Prefer explicit contracts through API modules, gRPC, or MQ integration points already present in the codebase.
- Keep deployment and migration changes close to the affected service.

## Directory Guidance

- `api-gateway/` for client-facing routing, REST/gRPC bridging, auth enforcement, and edge concerns
- `services/<name>/` for domain logic of a specific microservice
- `common-lib/` for shared backend primitives reused by multiple services
- `secrets/` for secret-management artifacts; treat as sensitive and do not modify casually

## Validation

- Run Go commands from the smallest affected module.
- Prefer service-local build or test commands before broader backend-wide checks.
- When changing contracts between services, verify both the provider and the consumer.

## Local Overrides

- Read `services/*/AGENTS.md` before editing a specific microservice.
- The nearest `AGENTS.md` takes precedence over this backend-wide file.
- Use `API_ARCHITECTURE.md` as the first-stop map for gateway, service ownership, and inter-service dependencies.
