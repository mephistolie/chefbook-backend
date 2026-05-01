# ChefBook Services Agents Guide

This directory contains ChefBook backend microservices.

## Service Conventions

- Each service is an isolated Go module with its own API, runtime, migrations, and deployment assets.
- Keep changes within the owning service unless a shared contract or library truly needs to change.
- Reuse the service template structure: `api`, `cmd`, `internal`, `migrations`, `deployments`, `scripts`, and optional `pkg`.

## Before Editing

- Read the local `AGENTS.md` in the target service directory.
- Check whether the task affects only that service or also the API gateway, another service, or `common-lib`.

## Cross-Service Changes

- Make contract changes intentionally and update all affected consumers.
- Do not assume shared persistence or hidden side effects across services.
