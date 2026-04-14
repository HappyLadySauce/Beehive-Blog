# Beehive Blog v2

Beehive Blog v2 is being rebuilt as a go-zero microservice project.

Current architecture direction:

- `api/`: external API contracts for gateway
- `proto/`: internal RPC contracts
- `services/gateway`: only public HTTP entry
- `services/identity`: RPC service
- `services/content`: RPC service
- `services/search`: RPC service
- `services/indexer`: async worker
- `pkg/`: reusable shared packages

The repository is in active restructuring. RPC contracts are in place, but full RPC code generation still depends on `protoc`.
