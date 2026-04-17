# Codegen

Use this directory for repeatable generation commands, such as:

- `goctl api go`
- `goctl rpc protoc`
- contract synchronization checks
- future model generation

Do not run generation ad hoc without updating the corresponding contract files first.

## Commands

- Generate gateway API code:
  - `./scripts/codegen/gen-gateway.ps1`
- Generate RPC code:
  - `./scripts/codegen/gen-rpc.ps1 -ProtoFile proto/<service>.proto -OutputDir services/<service>`
- Verify contract sync:
  - `./scripts/codegen/check-contract-sync.ps1`
- Generate + sync-check in one command:
  - `./scripts/codegen/sync-all.ps1`
  - Optional flags: `-WithGateway`, `-SkipRpc`, `-SkipCheck`
  - Note: gateway generation is skipped by default to avoid overwriting hand-maintained middleware wiring in `routes.go`.

`check-contract-sync.ps1` validates:

- `api/gateway.api` and `services/gateway/internal/handler/routes.go` route consistency.
- `proto/*.proto` and corresponding generated files in `services/*/pb` (including stale file detection by mtime).
