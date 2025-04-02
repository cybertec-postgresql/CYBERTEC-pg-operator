---
title: "Standby Cluster"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 2080
---

A standby cluster is an independent PostgreSQL cluster that consists of a standby leader and optionally further replicas (if `numberOfInstances` > 1). The standby leader runs in read-only mode and does not accept any write operations. A standby cluster can be promoted to a primary cluster if required, whereby the standby leader becomes a fully-fledged leader and allows write operations.

### Preconditions:
The primary cluster must either:
- be accessible from the standby cluster via streaming replication 
- the backup storage used by the standby cluster (S3, GCS or Azure Blob) must be accessible for the standby cluster

The passwords for the Postgres user, the replication user and the exporter user (if monitoring is active) must be created as a secret for the standby cluster. Otherwise connection problems will occur

### Create standby cluster

The `standby` object in the cluster manifest is required to create a standby cluster.

```yaml 
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: standby-cluster-1
spec:
  standby:
    standby_host: "cluster-1.cpo"
    standby_port: "5432"
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
```

The primary cluster must be accessible from the standby cluster. It can be located in the same Kubernetes cluster or in a different one. 

- `standby_host`: Corresponds to the endpoint via which the primary pod can be reached. It can be a kubernetes-internal DNS name or an IP or DNS name that can be reached in the network. 
- `standby_port`: Corresponds to the PostgreSQL port used (default 5432)


### Promoting cluster

To promote a cluster, it is only necessary to remove the standby object. 
The cluster is then promoted to a primary cluster.


### Limitations
A primary cluster cannot be demoted to a standby cluster. 
If necessary, the recommendation is to create a new cluster as a standby cluster.