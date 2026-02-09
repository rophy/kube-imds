# Development Guide for Claude

## Kubernetes Context

Use `kind-kind` for all kubectl and skaffold commands.

## Quick Commands

```bash
go build ./...          # Build
go test ./...           # Unit tests
bats test/e2e/          # E2E tests (requires kind cluster)
skaffold run --kube-context kind-kind   # Deploy to kind
```

## Project Structure

```
cmd/server/main.go              # Entry point (--config, --kubeconfig flags)
internal/
  config/config.go              # YAML config parsing (identities, defaults, validation)
  identity/resolver.go          # IP/CIDR → ServiceAccount lookup (exact map + CIDR scan)
  handler/
    token.go                    # POST /api/v1/token + writeStatusError helper
    selfsubjectreview.go        # POST /api/v1/selfsubjectreviews
    clientip.go                 # resolveClientIP (supports clientIPHeader)
    health.go                   # GET /healthz
  server/server.go              # HTTP server setup with logging middleware
deploy/
  kube-imds/                    # K8s manifests (namespace, deployment, service, rbac, configmap)
  e2e/test-resources.yaml       # E2E test pods and ServiceAccounts
test/e2e/                       # BATS e2e tests
```

## Design Decisions

- IP/CIDR → ServiceAccount mapping, multiple IPs per identity
- Resolver: exact IP map (O(1)) then CIDR scan (O(n))
- RBAC: per-namespace Role with `resourceNames` (not ClusterRole)
- `clientIPHeader` config for ingress gateway support
- All API responses use `apiVersion: kube-imds/v1`

## Testing Patterns

- Handler tests: `fake.NewSimpleClientset` + `PrependReactor` for token minting
- Config tests: `writeTemp()` helper for temp YAML files
- E2E: BATS with 3 curl pods (`curl`, `curl-unauthorized`, `curl-unknown`)
- E2E: `--field-selector=status.phase=Running` to avoid picking up terminated pods

## Git Commit Convention

```
<type>: <short description>
```

Types: feat, fix, refactor, chore, docs, build, test

No "Generated with Claude" footer or Co-Authored-By lines.

## Implementation Status

### Not Yet Done
- Helm chart
