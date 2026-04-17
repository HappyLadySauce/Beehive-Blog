---
name: contract-first-goctl
description: Enforce a contract-first workflow for Beehive Blog go-zero services. Use when adding or modifying HTTP or RPC interfaces so work always follows this order: update `api/*.api` or `proto/*.proto`, regenerate transport code with goctl, then implement business logic and verification.
---

# Contract-First goctl Workflow

Apply this workflow for every interface change.

## Required Order

1. Update contracts first.
2. Regenerate code from contracts.
3. Implement or adjust business logic.
4. Verify compile and route/RPC wiring.

Never start by editing generated transport files first.

This rule also applies to internal services and background workers.  
If a new internal capability is introduced (for example indexer, scheduler, consumer), define a minimal `proto/*.proto` contract first, generate with goctl, then implement business logic on top of generated skeletons.

## Contract Update

- For HTTP gateway changes, edit `api/gateway.api`.
- For internal RPC changes, edit `proto/*.proto`.
- For new internal services (even non-public worker services), create `proto/<service>.proto` first.
- Keep field names and semantics aligned across API DTO and proto messages.

## Regeneration

- Daily command (recommended):
  - `./scripts/codegen/sync-all.ps1`
- Full regeneration including gateway:
  - `./scripts/codegen/sync-all.ps1 -WithGateway`
- Contract check only:
  - `./scripts/codegen/check-contract-sync.ps1`

Script behavior:

- `sync-all.ps1` runs RPC generation + contract sync check by default.
- `sync-all.ps1` skips gateway generation by default to protect hand-maintained middleware wiring in `services/gateway/internal/handler/routes.go`.
- Use `-WithGateway` only when you intentionally regenerate gateway transport files.

Repository caveat:

- `goctl api go` does not overwrite many existing files by default.
- Treat generation output as a sync signal, not as an overwrite guarantee.
- If a handler name changes (for example `StudioQuery`), wire new handlers into `routes.go` and route logic explicitly.

## Business Code Update

After regeneration, update business layers as needed:

- `services/*/internal/logic/`
- `services/*/internal/svc/`
- `services/*/internal/middleware/` (gateway only)

Do not move business rules into request parsing handlers.

For internal worker-style services:

- Keep contract entrypoints minimal (for example `Health`, `Sync/RunOnce`).
- Put loop/poll/retry logic in business layer packages (for example `internal/worker`, `internal/svc`), not in generated server stubs.

## Verification Checklist

1. Confirm every route in `api/gateway.api` exists in `services/gateway/internal/handler/routes.go`.
2. Confirm route request/response types match `services/gateway/internal/types/types.go`.
3. Confirm gateway logic calls matching RPC methods in `services/*/*` clients.
4. Run compile/tests (`go test ./services/...` or scoped packages).
5. For any new service directory under `services/*`, confirm there is a matching contract file in `proto/*.proto` and generated `pb/*` files.

If build-cache permissions fail in local environment, still verify generation logs and compile scope as far as possible.
