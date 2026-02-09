# kube-imds

A centralized Instance Metadata Service (IMDS) for Kubernetes that mints ServiceAccount tokens for external entities based on client IP identity.

In public clouds, IMDS (like AWS EC2 Instance Metadata Service or GCP Metadata Server) allows VMs to obtain credentials without pre-provisioned secrets. **kube-imds** brings this pattern to Kubernetes — external machines (VMs, bare-metal servers) on a trusted network can request short-lived ServiceAccount tokens simply by calling an HTTP endpoint.

## How It Works

1. An admin maps client IPs to Kubernetes ServiceAccounts in a config file
2. An external machine calls `GET /api/v1/token`
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

## API

| Endpoint | Description |
|----------|-------------|
| `POST /api/v1/token` | Returns a minted ServiceAccount token for the caller (identified by client IP) |
| `GET /healthz` | Health check |

## RBAC

The service needs `create` permission on `serviceaccounts/token` in each target namespace:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kube-imds
rules:
  - apiGroups: [""]
    resources: ["serviceaccounts/token"]
    verbs: ["create"]
```

## Security Model

kube-imds uses **IP-based trust** — it assumes the network between clients and the service is trusted (e.g., corporate intranet, private VPC). No additional authentication is required. Tokens are short-lived and scoped to specific audiences.
