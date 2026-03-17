#!/usr/bin/env bash

# Copyright The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

# Configuration
CLUSTER_NAME="kan-conformance"
CONFORMANCE_NAMESPACE="gateway-conformance-infra"
SYSTEM_NAMESPACE="agentic-net-system"

# Find the repository root
REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}"
source dev/ci/lib.sh

# Main execution logic
main() {
  setup_kind_cluster "${CLUSTER_NAME}"
  
  local registry="us-central1-docker.pkg.dev/k8s-staging-images/agentic-net"
  local image_name="agentic-networking-controller"
  local tag=$(git rev-parse HEAD)-dirty-$(date +%s)
  
  build_and_load_controller_image "${CLUSTER_NAME}" "${registry}" "${image_name}" "${tag}"
  
  header "Pre-loading Envoy image"
  docker pull envoyproxy/envoy:v1.36-latest
  kind load docker-image envoyproxy/envoy:v1.36-latest --name "${CLUSTER_NAME}"
  
  install_crds
  setup_agentic_identity "${SYSTEM_NAMESPACE}"
  deploy_controller "${tag}" "${SYSTEM_NAMESPACE}"

  header "Running Conformance tests"
  # Requirements: K8s v1.35+, PodCertificateRequest/ClusterTrustBundle enabled, and KAN Controller running with --enable-agentic-identity-signer=true.
  cd tests && go clean -testcache
  
  local test_args=(-mod=mod -tags conformance -v ./conformance/... -gateway-class=kube-agentic-networking -cleanup-base-resources=false)
  if [ -n "${RUN_TEST:-}" ]; then
    test_args+=(-run-test="$RUN_TEST")
  fi
  
  GOWORK=off CGO_ENABLED=0 go test "${test_args[@]}"
}

# Function to print a prominent header
header() {
  local title=$1
  echo ""
  echo "================================================================================"
  echo "  ${title}"
  echo "================================================================================"
  echo ""
}

# Function to dump logs on failure
cleanup() {
  local status=$?
  if [ "${status}" -ne 0 ]; then
    header "Tests failed, dumping logs..."

    header "Cluster-wide Resources"
    kubectl get all -A || true

    header "Cluster Events"
    kubectl get events -A || true

    header "Controller Description"
    kubectl describe deployment agentic-net-controller -n "${SYSTEM_NAMESPACE}" || true

    header "Controller logs (last 200 lines)"
    kubectl logs deployment/agentic-net-controller -n "${SYSTEM_NAMESPACE}" --all-containers --tail=200 || true

    header "Conformance Test Namespace Resources"
    kubectl get all -n "${CONFORMANCE_NAMESPACE}" || true

    header "Gateway Resources"
    kubectl get gateway -n "${CONFORMANCE_NAMESPACE}" -o yaml || true

    header "Access Policies"
    kubectl get xaccesspolicies -n "${CONFORMANCE_NAMESPACE}" || true

    header "Backend Resources"
    kubectl get xbackends -n "${CONFORMANCE_NAMESPACE}" || true

    header "Pods in Conformance namespace"
    kubectl get pods -n "${CONFORMANCE_NAMESPACE}" -o wide || true

    header "Pod Certificate Requests"
    kubectl get podcertificaterequests -n "${CONFORMANCE_NAMESPACE}" -o yaml || true

    header "Cluster Trust Bundles"
    kubectl get clustertrustbundles || true

    header "Tester Pod YAML"
    kubectl get pod conformance-tester -n "${CONFORMANCE_NAMESPACE}" -o yaml || true

    header "MCP Server Logs"
    kubectl logs -n "${CONFORMANCE_NAMESPACE}" -l app=mcp-everything --tail=100 || true

    header "Envoy Proxy Logs"
    kubectl logs -n "${CONFORMANCE_NAMESPACE}" -l "gateway.networking.k8s.io/gateway-name" --all-containers --tail=100 || true
  fi
  exit "${status}"
}

# Register the cleanup trap and run main
trap cleanup EXIT
main "$@"
