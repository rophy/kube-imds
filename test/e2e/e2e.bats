#!/usr/bin/env bats

setup_file() {
    load test_helper

    export BATS_TEST_TIMEOUT="${BATS_TEST_TIMEOUT:-120}"
    local project_root
    project_root="$(get_project_root)"

    echo "# Deploying kube-imds with skaffold..." >&3
    skaffold run --kube-context "$KUBE_CONTEXT" -f "$project_root/skaffold.yaml" >&3 2>&3

    echo "# Waiting for deployment to be ready..." >&3
    wait_for_deployment kube-imds 60s

    echo "# Waiting for curl pods to be ready..." >&3
    wait_for_pod_ready "role=mapped-client" 60
    wait_for_pod_ready "role=unauthorized-client" 60
    wait_for_pod_ready "role=unmapped-client" 60

    # Get pod IPs and update the configmap
    local curl_ip
    curl_ip="$(get_pod_ip curl)"
    echo "# Curl pod IP: $curl_ip" >&3

    local curl_unauthorized_ip
    curl_unauthorized_ip="$(get_pod_ip curl-unauthorized)"
    echo "# Curl-unauthorized pod IP: $curl_unauthorized_ip" >&3

    echo "# Updating configmap with pod IPs..." >&3
    update_config_ip "$curl_ip" "$curl_unauthorized_ip"

    echo "# Restarting deployment to pick up new config..." >&3
    kubectl -n "$NAMESPACE" rollout restart deployment/kube-imds
    wait_for_deployment kube-imds 60s

    echo "# Setup complete." >&3
}

teardown_file() {
    load test_helper
    # Leave resources for inspection; skaffold delete can clean up
}

setup() {
    load test_helper
}

@test "kube-imds pod is running" {
    local phase
    phase=$(kubectl -n "$NAMESPACE" get pod -l app=kube-imds --field-selector=status.phase=Running -o jsonpath='{.items[0].status.phase}')
    echo "# Pod phase: $phase"
    [[ "$phase" == "Running" ]]
}

@test "healthz endpoint returns ok" {
    local result
    result=$(kubectl -n "$NAMESPACE" exec curl -- curl -sf http://kube-imds/healthz)
    echo "# healthz response: $result"
    [[ "$result" == "ok" ]]
}

@test "token endpoint returns valid TokenRequest" {
    local response
    response=$(kubectl -n "$NAMESPACE" exec curl -- curl -sf -X POST http://kube-imds/api/v1/token)
    echo "# token response (truncated): ${response:0:120}..."

    local kind
    kind=$(echo "$response" | python3 -c "import sys,json; print(json.load(sys.stdin)['kind'])")
    echo "# kind: $kind"
    [[ "$kind" == "TokenRequest" ]]

    local api_version
    api_version=$(echo "$response" | python3 -c "import sys,json; print(json.load(sys.stdin)['apiVersion'])")
    echo "# apiVersion: $api_version"
    [[ "$api_version" == "authentication.k8s.io/v1" ]]
}

@test "minted token has correct subject" {
    local response
    response=$(kubectl -n "$NAMESPACE" exec curl -- curl -sf -X POST http://kube-imds/api/v1/token)

    local subject
    subject=$(echo "$response" | python3 -c "
import sys, json, base64
token = json.load(sys.stdin)['status']['token']
payload = token.split('.')[1]
payload += '=' * (4 - len(payload) % 4)
claims = json.loads(base64.b64decode(payload))
print(claims['sub'])
")
    echo "# Token subject: $subject"
    [[ "$subject" == "system:serviceaccount:kube-imds:vm-worker-1" ]]
}

@test "minted token has correct audience" {
    local response
    response=$(kubectl -n "$NAMESPACE" exec curl -- curl -sf -X POST http://kube-imds/api/v1/token)

    local audience
    audience=$(echo "$response" | python3 -c "
import sys, json, base64
token = json.load(sys.stdin)['status']['token']
payload = token.split('.')[1]
payload += '=' * (4 - len(payload) % 4)
claims = json.loads(base64.b64decode(payload))
print(','.join(claims['aud']))
")
    echo "# Token audience: $audience"
    [[ "$audience" == "api" ]]
}

@test "RBAC-denied SA gets 500" {
    # curl-unauthorized is mapped to vm-worker-unauthorized, which is NOT
    # in the Role's resourceNames, so the K8s API will deny the TokenRequest
    local http_code
    http_code=$(kubectl -n "$NAMESPACE" exec curl-unauthorized -- \
        curl -s -o /dev/null -w '%{http_code}' -X POST http://kube-imds/api/v1/token)
    echo "# HTTP code from unauthorized SA: $http_code"
    [[ "$http_code" == "500" ]]

    local response
    response=$(kubectl -n "$NAMESPACE" exec curl-unauthorized -- curl -s -X POST http://kube-imds/api/v1/token)
    local status
    status=$(echo "$response" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])")
    echo "# Status: $status"
    [[ "$status" == "Failure" ]]
}

@test "unknown IP gets 403 Forbidden" {
    # curl-unknown pod has an IP that is NOT in the configmap
    local http_code
    http_code=$(kubectl -n "$NAMESPACE" exec curl-unknown -- \
        curl -s -o /dev/null -w '%{http_code}' -X POST http://kube-imds/api/v1/token)
    echo "# HTTP code from unmapped client: $http_code"
    [[ "$http_code" == "403" ]]

    # Verify the response body is a proper K8s Status error
    local response
    response=$(kubectl -n "$NAMESPACE" exec curl-unknown -- curl -s -X POST http://kube-imds/api/v1/token)
    local status
    status=$(echo "$response" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])")
    echo "# Status: $status"
    [[ "$status" == "Failure" ]]
}
