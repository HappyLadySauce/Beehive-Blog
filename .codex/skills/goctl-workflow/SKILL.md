---
name: goctl-workflow
description: Scaffold and maintain go-zero services with goctl. Use when creating or updating go-zero API or RPC services, generating code from `.api` or `.proto`, planning a go-zero monorepo layout, or explaining how generated handler/logic/svc/types files should map to domain code.
---

# goctl Workflow

## Overview

Use `goctl` to generate go-zero service skeletons after the service boundary, API shape, and domain model are already defined. Keep generated transport code thin and place business rules outside generated `handler` scaffolding.

## Workflow

1. Confirm the service boundary first.
2. Write the `.api` or `.proto` contract before generating code.
3. Run `goctl` to generate the transport skeleton.
4. Keep generated `logic` focused on use-case orchestration, not raw domain sprawl.
5. Add hand-written `domain`, `repository`, `searchengine`, or `providers` directories when the service needs them.
6. Regenerate only from source contracts; do not manually fork generated wire-up files unless there is a clear reason.

## API Services

For HTTP-facing go-zero services:

- Put the contract in `api/<service>.api`.
- Generate into the service root so `etc/`, `internal/`, and `cmd/` stay together.
- Use generated `types` for transport DTOs only.

Read [references/goctl.md](references/goctl.md) for concrete command patterns.

## RPC Services

For internal service-to-service calls:

- Put the contract in `rpc/<service>.proto`.
- Generate stubs and keep proto as the source of truth.
- Expose RPC only for stable internal capabilities, not every table operation.

## Generated Code Boundaries

Treat goctl output as transport and wiring code.

- `handler/`: request entry only
- `logic/`: use-case orchestration
- `svc/`: dependency wiring
- `types/`: request and response DTOs

Do not bury domain rules in `handler/`.

Prefer adding hand-written directories when needed:

- `internal/domain/`
- `internal/repository/`
- `internal/searchengine/`
- `internal/providers/`

## Regeneration Rules

- Regenerate from `.api` or `.proto`, not from edited generated files.
- Avoid editing generated routing and type glue unless necessary.
- If a generated file must be customized, isolate the customization and document the reason in code comments.

## Monorepo Fit

Use this skill with a layout like:

- `apps/<service>/api`
- `apps/<service>/rpc`
- `apps/<service>/cmd`
- `apps/<service>/internal`
- `shared/`
- `sql/`
- `scripts/codegen/`

## Validation

After generation:

1. Verify config file names and ports are consistent.
2. Verify the generated API or RPC package names match the service name.
3. Verify domain code lives outside generated transport glue.
4. Build the service before adding more features.
