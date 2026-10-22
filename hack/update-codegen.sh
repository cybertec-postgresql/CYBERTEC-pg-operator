#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

OPERATOR_PACKAGE_ROOT="github.com/cybertec-postgresql/cybertec-pg-operator"
SCRIPT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo "${GOPATH}"/src/k8s.io/code-generator)}
CODEGEN_PKG=$(cd "${SCRIPT_ROOT}" && cd "${CODEGEN_PKG}" && pwd)

echo "${OPERATOR_PACKAGE_ROOT} - ${CODEGEN_PKG}"

source "${CODEGEN_PKG}/kube_codegen.sh"

# zz_generated.deepcopy.go
kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/apis"

# clientset, listers, informers
kube::codegen::gen_client \
    --with-watch \
    --output-dir "${SCRIPT_ROOT}/pkg/generated" \
    --output-pkg "${OPERATOR_PACKAGE_ROOT}/pkg/generated" \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/apis"