#!/usr/bin/env bash

KUBE_CONTEXT="${KUBE_CONTEXT:-kind-kind}"
NAMESPACE="kube-imds"

get_project_root() {
    cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd
}

# Override kubectl to always use the correct context
kubectl() {
    command kubectl --context "$KUBE_CONTEXT" "$@"
}

# Wait for a deployment to be ready
wait_for_deployment() {
    local name="$1"
    local timeout="${2:-60s}"
    kubectl -n "$NAMESPACE" rollout status deployment/"$name" --timeout="$timeout"
}

# Wait for a pod to be ready by label
wait_for_pod_ready() {
    local label="$1"
    local timeout="${2:-60}"
    local attempts=0
    while [[ $attempts -lt $timeout ]]; do
        local phase
        phase=$(kubectl -n "$NAMESPACE" get pod -l "$label" -o jsonpath='{.items[0].status.phase}' 2>/dev/null || true)
        if [[ "$phase" == "Running" ]]; then
            return 0
        fi
        sleep 1
        attempts=$((attempts + 1))
    done
    echo "# Timed out waiting for pod with label $label to be Running"
    return 1
}

# Get the IP of a pod by name
get_pod_ip() {
    local name="$1"
    kubectl -n "$NAMESPACE" get pod "$name" -o jsonpath='{.status.podIP}'
}

# Wait for a file to exist in a pod
wait_for_file() {
    local pod="$1"
    local path="$2"
    local timeout="${3:-60}"
    local attempts=0
    while [[ $attempts -lt $timeout ]]; do
        if kubectl -n "$NAMESPACE" exec "$pod" -- test -f "$path" 2>/dev/null; then
            return 0
        fi
        sleep 1
        attempts=$((attempts + 1))
    done
    echo "# Timed out waiting for $path in pod $pod"
    return 1
}

# Update the kube-imds configmap with given client IPs
# $1 = mapped client IP (authorized SA)
# $2 = optional unauthorized client IP (SA not in RBAC resourceNames)
# $3 = optional client-test IP (maps to same SA as $1)
update_config_ip() {
    local ip="$1"
    local unauthorized_ip="${2:-}"
    local client_ip="${3:-}"
    local identities="
identities:
  - ips: [\"${ip}\"]
    serviceAccount:
      name: \"vm-worker-1\"
      namespace: \"${NAMESPACE}\""

    if [[ -n "$unauthorized_ip" ]]; then
        identities="${identities}
  - ips: [\"${unauthorized_ip}\"]
    serviceAccount:
      name: \"vm-worker-unauthorized\"
      namespace: \"${NAMESPACE}\""
    fi

    if [[ -n "$client_ip" ]]; then
        identities="${identities}
  - ips: [\"${client_ip}\"]
    serviceAccount:
      name: \"vm-worker-1\"
      namespace: \"${NAMESPACE}\""
    fi

    kubectl -n "$NAMESPACE" create configmap kube-imds \
        --from-literal=config.yaml="${identities}

defaults:
  tokenSpec:
    audiences: [\"https://kubernetes.default.svc.cluster.local\"]
    expirationSeconds: 3600
" \
        --dry-run=client -o yaml | kubectl apply -f -
}
