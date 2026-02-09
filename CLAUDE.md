# Development Guide for Claude

## Project Overview

kube-imds is a centralized IMDS-like service that mints Kubernetes ServiceAccount tokens for external entities (VMs) based on client IP identity. It uses the Kubernetes TokenRequest API to mint short-lived tokens on demand.

Related project: `kubevirt-imds` — similar concept but KubeVirt-specific (sidecar with veth pair, projected volumes). kube-imds is the generic, centralized alternative.

## Kubernetes Context

Use `kind-kind` for all kubectl and skaffold commands.

## Quick Commands

```bash
# Build
go build ./...

# Test
go test ./...

# E2E tests (requires kind cluster)
bats test/e2e/

# Deploy to kind
skaffold run --kube-context kind-kind

# Run locally (requires kubeconfig)
go run ./cmd/server --config config/config.example.yaml --kubeconfig ~/.kube/config
```

## Design Decisions

- **Token minting**: Kubernetes TokenRequest API (centralized, no sidecar)
- **Identity model**: One IP = one ServiceAccount
- **API style**: `GET /api/v1/token` — Kubernetes-style TokenRequest response
- **Deployment**: Single centralized Deployment
- **Auth**: IP-based trust (assumes corporate intranet)

## API

- `GET /api/v1/token` — Returns minted SA token for the caller (identified by client IP)
- `GET /healthz` — Health check

## Project Structure

```
cmd/server/main.go              # Entry point (--config, --kubeconfig flags)
internal/
  config/config.go              # YAML config parsing (identities, defaults)
  identity/resolver.go          # IP → ServiceAccount lookup
  handler/
    token.go                    # GET /api/v1/token (TokenRequest API call)
    health.go                   # GET /healthz
  server/server.go              # HTTP server setup with logging middleware
config/config.example.yaml      # Example configuration
```

## Configuration Format

```yaml
identities:
  - ip: "10.0.1.10"
    serviceAccount:
      name: "vm-worker-1"
      namespace: "external-identities"
    tokenSpec:                    # optional per-identity override
      audiences: ["api"]
      expirationSeconds: 3600

defaults:
  tokenSpec:
    audiences: ["api"]
    expirationSeconds: 3600
```

## RBAC Requirements

The service needs `create` permission on `serviceaccounts/token` in target namespaces.

## Implementation Status

### Done
- Config parsing with validation and defaults merging
- Identity resolver (IP → ServiceAccount map)
- Token handler (TokenRequest API, K8s-style response/errors)
- Health handler
- HTTP server with logging middleware
- Entry point with graceful shutdown
- Dockerfile (multi-stage, distroless)
- Makefile (build, test, test-e2e, image)
- Example config file
- go.mod with dependencies resolved
- Unit tests (config, identity, handler)
- K8s deployment manifests (deploy/kube-imds/)
- RBAC manifests
- Skaffold config
- BATS e2e test suite

### Not Yet Done
- Helm chart

## Git Commit Convention

```
<type>: <short description>

[optional body]
```

Types: feat, fix, refactor, chore, docs, build, test

No "Generated with Claude" footer or Co-Authored-By lines.
