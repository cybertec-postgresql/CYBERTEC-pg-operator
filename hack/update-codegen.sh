#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

GENERATED_PACKAGE_ROOT="github.com"
OPERATOR_PACKAGE_ROOT="${GENERATED_PACKAGE_ROOT}/cybertec-postgresql/cybertec-pg-operator"
SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
TARGET_CODE_DIR=${1-${SCRIPT_ROOT}/pkg}
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo "${GOPATH}"/src/k8s.io/code-generator)}

cleanup() {
    rm -rf "${GENERATED_PACKAGE_ROOT}"
}
trap "cleanup" EXIT SIGINT
echo "${OPERATOR_PACKAGE_ROOT} - ${CODEGEN_PKG}"
bash "${CODEGEN_PKG}/generate-groups.sh" all \
  "${OPERATOR_PACKAGE_ROOT}/pkg/generated" "${OPERATOR_PACKAGE_ROOT}/pkg/apis" \
  "cpo.opensource.cybertec.at:v1 zalando.org:v1" \
  --go-header-file "${SCRIPT_ROOT}"/hack/custom-boilerplate.go.txt \
  -o ./

cp -r "${OPERATOR_PACKAGE_ROOT}"/pkg/* "${TARGET_CODE_DIR}"

cleanup
