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

## Contract Update

- For HTTP gateway changes, edit `api/gateway.api`.
- For internal RPC changes, edit `proto/*.proto`.
- Keep field names and semantics aligned across API DTO and proto messages.

## Regeneration

- Gateway API generation:
  - `goctl api go --api api/gateway.api --dir services/gateway`
- RPC generation:
  - `scripts/codegen/gen-rpc.ps1 -ProtoFile proto/<service>.proto -OutputDir services/<service>`

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

## Verification Checklist

1. Confirm every route in `api/gateway.api` exists in `services/gateway/internal/handler/routes.go`.
2. Confirm route request/response types match `services/gateway/internal/types/types.go`.
3. Confirm gateway logic calls matching RPC methods in `services/*/*` clients.
4. Run compile/tests (`go test ./...` or scoped packages).

If build-cache permissions fail in local environment, still verify generation logs and compile scope as far as possible.
