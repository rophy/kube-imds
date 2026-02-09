# kube-imds

A centralized Instance Metadata Service (IMDS) for Kubernetes that mints ServiceAccount tokens for external entities based on client IP identity.

In public clouds, IMDS (like AWS EC2 Instance Metadata Service or GCP Metadata Server) allows VMs to obtain credentials without pre-provisioned secrets. **kube-imds** brings this pattern to Kubernetes — external machines (VMs, bare-metal servers) on a trusted network can request short-lived ServiceAccount tokens simply by calling an HTTP endpoint.

## How It Works

1. An admin maps client IPs to Kubernetes ServiceAccounts in a config file
2. An external machine calls `POST /api/v1/token`
3. kube-imds identifies the caller by source IP, then uses the Kubernetes TokenRequest API to mint a short-lived token
4. The caller receives the token and can authenticate against Kubernetes or any system that trusts K8s ServiceAccount tokens

## Quick Start

```bash
# Build
make build

# Run (requires kubeconfig with tokenrequest permissions)
./bin/kube-imds --config config/config.example.yaml --kubeconfig ~/.kube/config
```

## Configuration

```yaml
identities:
  # Single IP
  - ips: ["10.0.1.10"]
    serviceAccount:
      name: "vm-worker-1"
      namespace: "external-identities"

  # Multiple IPs mapped to the same ServiceAccount
  - ips:
      - "10.0.1.20"
      - "10.0.1.21"
      - "10.0.1.22"
    serviceAccount:
      name: "vm-pool"
      namespace: "external-identities"

  # CIDR range -- all IPs in this subnet get the same identity
  - ips: ["10.0.2.0/24"]
    serviceAccount:
      name: "vm-subnet"
      namespace: "external-identities"
    tokenSpec:                    # optional per-identity override
      audiences: ["custom-audience"]
      expirationSeconds: 7200

defaults:
  tokenSpec:
    audiences: ["api"]
    expirationSeconds: 3600

# When behind an ingress gateway, set this to the header containing the real client IP
# clientIPHeader: "X-Forwarded-For"
```

## API

All responses use `apiVersion: kube-imds/v1`. Error responses follow the Kubernetes `Status` envelope.

### `POST /api/v1/token`

Mint a ServiceAccount token for the caller (identified by client IP).

**200 OK** — token minted successfully:
```json
{
  "kind": "TokenRequest",
  "apiVersion": "kube-imds/v1",
  "status": {
    "token": "eyJhbGciOiJSUzI1NiIs...",
    "expirationTimestamp": "2025-01-01T01:00:00Z"
  }
}
```

**403 Forbidden** — no identity mapped for the caller's IP:
```json
{
  "kind": "Status",
  "apiVersion": "kube-imds/v1",
  "status": "Failure",
  "message": "no identity mapped for IP 10.0.1.99",
  "code": 403
}
```

**500 Internal Server Error** — token minting failed (e.g. RBAC denied):
```json
{
  "kind": "Status",
  "apiVersion": "kube-imds/v1",
  "status": "Failure",
  "message": "failed to mint token",
  "code": 500
}
```

### `POST /api/v1/selfsubjectreviews`

Look up the identity mapped to the caller's IP without minting a token.

**200 OK** — identity found:
```json
{
  "kind": "SelfSubjectReview",
  "apiVersion": "kube-imds/v1",
  "status": {
    "userInfo": {
      "username": "system:serviceaccount:external-identities:vm-worker-1"
    }
  }
}
```

**403 Forbidden** — no identity mapped:
```json
{
  "kind": "Status",
  "apiVersion": "kube-imds/v1",
  "status": "Failure",
  "message": "no identity mapped for IP 10.0.1.99",
  "code": 403
}
```

### `GET /healthz`

Health check endpoint.

**200 OK:**
```
ok
```

## RBAC

The service needs a per-namespace `Role` with `create` permission on `serviceaccounts/token`, scoped to specific ServiceAccount names via `resourceNames`:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kube-imds
  namespace: external-identities
rules:
  - apiGroups: [""]
    resources: ["serviceaccounts/token"]
    verbs: ["create"]
    resourceNames: ["vm-worker-1", "vm-pool", "vm-subnet"]
```

Create a separate Role + RoleBinding in each namespace referenced by identities in the config.

## Security Model

kube-imds uses **IP-based trust** — it assumes the network between clients and the service is trusted (e.g., corporate intranet, private VPC). No additional authentication is required. Tokens are short-lived and scoped to specific audiences.

## Testing

```bash
# Unit tests
make test

# E2E tests (requires a kind cluster)
make test-e2e
```

## License

[MIT](LICENSE)
