# goctl Reference

## Typical API generation

Create a new API service skeleton:

```powershell
goctl api new gateway
```

Generate server code from an existing `.api` contract:

```powershell
goctl api go -api .\api\gateway.api -dir .
```

Use this pattern inside a service directory such as `apps/gateway/`.

## Typical RPC generation

Generate RPC code from `.proto`:

```powershell
goctl rpc protoc .\rpc\content.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

Keep `.proto` under `rpc/` and generate from the service root.

## Recommended service bootstrap order

1. Create the service directory.
2. Write `.api` or `.proto`.
3. Run `goctl`.
4. Add `internal/domain` and `internal/repository` manually.
5. Wire infra in `svc`.
6. Implement use cases in `logic`.

## Suggested service mapping for Beehive v2

- `gateway`: API only
- `identity-service`: API + RPC
- `content-service`: API + RPC
- `review-service`: API + RPC
- `search-service`: API + RPC
- `agent-service`: API + RPC
- `indexer-worker`: hand-written worker, no goctl required

## Rules

- Do not generate service boundaries before the domain model is stable.
- Do not use goctl as a substitute for service design.
- Do not put database models directly into transport DTOs unless intentionally flattening a response.
- Regenerate from contracts when interfaces change.
