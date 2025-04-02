---
title: "Clone Cluster"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 2050
---
The function of a cluster clone was implemented to create the possibility of duplicating the current status of a cluster in order to carry out tests such as a major upgrade.
It creates an autonomous and independent cluster based on an existing local cluster or from a cloud storage via pgBackRest (S3, gcs or Azure Blob)

### Preconditions:
The primary cluster must either:
- be accessible from the standby cluster via streaming replication 
- the backup storage used by the standby cluster (S3, GCS or Azure Blob) must be accessible for the standby cluster

The passwords for the Postgres user, the replication user and the exporter user (if monitoring is active) must be created as a secret for the standby cluster. Otherwise connection problems will occur

### Clone a cluster via pvc

```yaml 
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-1-clone
spec:
  dockerImage: 'docker.io/cybertecpostgresql/cybertec-pg-container:postgres-17.4-1'
  numberOfInstances: 1
  postgresql:
    version: '17'
  resources:
    limits:
      cpu: 500m
      memory: 500Mi
    requests:
      cpu: 500m
      memory: 500Mi
  teamId: acid
  volume:
    size: 5Gi
  clone:
    cluster: cluster-1
    pgbackrest:
      configuration:
        secret: cluster-1-pvc-configuration
      repo:
        storage: pvc
```

### Clone a cluster via s3

```yaml
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-1-clone
spec:
  dockerImage: 'docker.io/cybertecpostgresql/cybertec-pg-container:postgres-17.4-1'
  numberOfInstances: 1
  postgresql:
    version: '17'
  resources:
    limits:
      cpu: 500m
      memory: 500Mi
    requests:
      cpu: 500m
      memory: 500Mi
  teamId: acid
  volume:
    size: 5Gi
  clone:
    cluster: cluster-1 # A random cluster name can be used if the source cluster is not present on the k8s.
    pgbackrest:
      configuration:
        secret: cluster-1-s3-credentials
      options:
        repo1-path: /YOUR_PATH_INSIDE_THE_BUCKET_TO_THE_SOURCE_STANZA/repo1/
      repo:
        endpoint: YOUR_SOURCE_S3_ENDPOINT
        name: repo1
        region: YOUR_SOURCE_S3_REGION
        resource: YOUR_SOURCE_BUCKET_NAME
        storage: s3
```

### Limitations
A primary cluster cannot be demoted to a standby cluster. 
If necessary, the recommendation is to create a new cluster as a standby cluster.