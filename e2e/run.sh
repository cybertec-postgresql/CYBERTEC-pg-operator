#!/usr/bin/env bash

# enable unofficial bash strict mode
set -o errexit
set -o nounset
set -o pipefail
IFS=$'\n\t'

readonly cluster_name="postgres-operator-e2e-tests"
readonly kubeconfig_path="/tmp/kind-config-${cluster_name}"
readonly spilo_image="${SPILO_IMAGE:-containers.cybertec.at/cybertec-pg-container/postgres:rocky9-18.6-3}"
readonly e2e_test_runner_image="registry.opensource.zalan.do/acid/postgres-operator-e2e-tests-runner:0.4"

# Tests that currently pass with the CPO images. Remaining tests still rely on Zalando specific
# roles/images or on a bigger kind cluster; override with E2E_TEST_CASE to run others.
readonly default_tests=(discover -v --start-directory /tests -k test_cross_namespace_secrets -k test_custom_ssl_certificate -k test_enable_load_balancer -k test_ignored_annotations -k test_service_annotations -k test_statefulset_annotation_propagation)

export GOPATH=${GOPATH-~/go}
export PATH=${GOPATH}/bin:$PATH

echo "Clustername: ${cluster_name}"
echo "Kubeconfig path: ${kubeconfig_path}"

# 'kind load docker-image' imports all platforms of a multi-arch image and fails when
# only the host platform is present locally, so load a single-platform archive instead
function kind_load(){
  local arch
  case "$(uname -m)" in aarch64|arm64) arch=arm64 ;; *) arch=amd64 ;; esac
  local archive
  archive=$(mktemp)
  docker save --platform "linux/${arch}" -o "${archive}" "$1"
  kind load image-archive "${archive}" --name ${cluster_name}
  rm -f "${archive}"
}

function pull_images(){
  # test the operator image built by 'make docker' (see Makefile target e2e), not an upstream one
  operator_tag=$(git describe --tags --always --dirty)
  operator_image="docker.io/cybertecpostgresql/cybertec-pg-operator:${operator_tag}"
  if [[ -z $(docker images -q "${operator_image}") ]]
  then
    echo "Operator image ${operator_image} not found, run 'make docker' first"
    exit 1
  fi
}

function start_kind(){
  echo "Starting kind for e2e tests"
  # avoid interference with previous test runs
  if [[ $(kind get clusters | grep "^${cluster_name}*") != "" ]]
  then
    kind delete cluster --name ${cluster_name}
  fi

  export KUBECONFIG="${kubeconfig_path}"
  kind create cluster --name ${cluster_name} --config kind-cluster-postgres-operator-e2e-tests.yaml  
  # images available on the Docker host (e.g. built locally) are used as they are
  [[ -n $(docker images -q "${spilo_image}") ]] || docker pull "${spilo_image}"
  kind_load "${spilo_image}"
}

function load_operator_image() {
  echo "Loading operator image"
  export KUBECONFIG="${kubeconfig_path}"
  kind_load "${operator_image}"
}

function set_kind_api_server_ip(){
  echo "Setting up kind API server ip"
  # use the actual kubeconfig to connect to the 'kind' API server
  # but update the IP address of the API server to the one from the Docker 'bridge' network
  readonly local kind_api_server_port=6443 # well-known in the 'kind' codebase
  readonly local kind_api_server=$(docker inspect --format "{{ .NetworkSettings.Networks.kind.IPAddress }}:${kind_api_server_port}" "${cluster_name}"-control-plane)
  # '-i.bak' works with both GNU and BSD (macOS) sed
  sed -i.bak "s/server.*$/server: https:\/\/$kind_api_server/g" "${kubeconfig_path}"
  rm -f "${kubeconfig_path}.bak"
}

function generate_certificate(){
  openssl req -x509 -nodes -newkey rsa:2048 -keyout tls/tls.key -out tls/tls.crt -subj "/CN=cpo.opensource.cybertec.at"
}

function run_tests(){
  local test_args=("${default_tests[@]}")
  [[ -n ${E2E_TEST_CASE-} ]] && test_args=(${E2E_TEST_CASE})
  echo "Running tests... image: ${e2e_test_runner_image}"
  # tests modify files in ./manifests, so we mount a copy of this directory done by the e2e Makefile

  docker run --rm --network=host -e "TERM=xterm-256color" \
  --mount type=bind,source="$(readlink -f ${kubeconfig_path})",target=/root/.kube/config \
  --mount type=bind,source="$(readlink -f manifests)",target=/manifests \
  --mount type=bind,source="$(readlink -f tls)",target=/tls \
  --mount type=bind,source="$(readlink -f tests)",target=/tests \
  --mount type=bind,source="$(readlink -f exec.sh)",target=/exec.sh \
  --mount type=bind,source="$(readlink -f scripts)",target=/scripts \
  -e OPERATOR_IMAGE="${operator_image}" -e SPILO_IMAGE="${spilo_image}" "${e2e_test_runner_image}" "${test_args[@]}" $@
}

function cleanup(){
  echo "Executing cleanup"
  unset KUBECONFIG
  kind delete cluster --name ${cluster_name}
  rm -rf ${kubeconfig_path}
}

function main(){
  echo "Entering main function..."
  [[ -z ${NOCLEANUP-} ]] && trap "cleanup" QUIT TERM EXIT
  pull_images
  [[ ! -f ${kubeconfig_path} ]] && start_kind
  load_operator_image
  set_kind_api_server_ip
  generate_certificate

  shift
  run_tests $@
  exit 0
}

"$1" $@
