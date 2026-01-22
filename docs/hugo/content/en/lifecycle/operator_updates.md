---
title: "Updating the Operator"
date: 2025-12-28T14:26:51+01:00
draft: false
weight: 2100
---

This chapter describes the recommended process for updating the CYBERTEC PostgreSQL Operator (CPO). To ensure a smooth transition and compatibility with new features, updates should be performed using our official Helm repository.

{{< hint type=important >}} CRD Update Requirement: Due to how Helm handles the crds/ directory, helm upgrade will not automatically update or patch existing Custom Resource Definitions (CRDs). You must manually apply the updated CRDs before upgrading the Helm release. {{< /hint >}}

## Using Helm-Chart

1. Update the Helm Repository

First, ensure your local Helm chart cache is up to date with the latest versions from the CYBERTEC repository:
```
helm repo update cpo
```

2. Update the Custom Resource Definitions (CRDs)

Before upgrading the Helm release, you must manually apply the latest CRDs from the CYBERTEC-operator-tutorials repository. This is a safety measure because Helm does not touch existing CRDs to prevent accidental data loss.

Apply the definitions for the Postgres clusters and the operator configuration directly from the source:

```
# Update the PostgreSQL Cluster CRD
kubectl apply -f https://raw.githubusercontent.com/cybertec-postgresql/CYBERTEC-operator-tutorials/refs/heads/main/setup/helm/operator/crds/postgresql.crd.yaml

# Update the Operator Configuration CRD
kubectl apply -f https://raw.githubusercontent.com/cybertec-postgresql/CYBERTEC-operator-tutorials/refs/heads/main/setup/helm/operator/crds/operatorconfiguration.crd.yaml
```

3. Execute the Helm Upgrade

Once the CRDs are up to date, you can upgrade the operator deployment. This process replaces the operator pod with the new version and updates the necessary RBAC roles and service accounts.
```
# Upgrade the CPO release in the 'cpo' namespace
helm upgrade cpo cpo/cybertec-pg-operator \
  --namespace cpo \
  --reuse-values
```

## Using CPO-Tutorial-Repository

1. Clone or Update the Tutorial Repo
```
git clone https://github.com/$GITHUB_USER/CYBERTEC-operator-tutorials.git
cd CYBERTEC-operator-tutorials
```

2. Patch CRDS

```
# Update the PostgreSQL Cluster CRD
kubectl apply -f setup/helm/operator/crds/postgresql.crd.yaml

# Update the Operator Configuration CRD
kubectl apply -f setup/helm/operator/crds/operatorconfiguration.crd.yaml
```

3. Execute the Helm Upgrade
```
# Upgrade the CPO release in the 'cpo' namespace
helm upgrade cpo setup/helm/operator/. \
  --namespace cpo \
  --reuse-values
```

## Verification

To ensure the update was successful, perform the following checks:

1. Pod Status: Verify that the new operator pod is running.
```
kubectl get pods -n cpo -l app.kubernetes.io/name=cybertec-pg-operator
```

2. Version Check: Check the logs to see the version string during startup.
```
kubectl logs -n cpo deployment/cybertec-pg-operator | grep "Starting operator"
```

3. CRD Integrity: Ensure the new CRD fields are recognized by the Kubernetes API.
```
kubectl describe crd postgresqls.cpo.opensource.cybertec.at
```

## Why manual CRD patching? ##

The CRDs are located in the helm/operator/crds/ folder. By design, Helm only installs these during the initial helm install. During an upgrade, Helm ignores this folder to protect the cluster from unintended schema changes. Therefore, manual application via kubectl apply is the standard and safest path for CPO updates.

## Compatibility ##

Always ensure that your postgresql manifests are compatible with the new operator version by checking the [Release Notes](release_notes) in the Documentation.