---
title: "Minor version update"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 2110
---

Minor version updates for PostgreSQL are performed by updating the PostgreSQL container image in use. 
With the update object `spec.dockerImage` of the cluster manifest, the operator takes over the update based on the rolling update strategy. This means that the pods are replaced one after the other, with the replicas being updated first and then the old primary after a switchover. The operational interruption should generally last less than 5 seconds (switchover time), but the clients must still reconnect.

If necessary, the operator also supports the downgrade of minor releases in the same way.

To install minor version updates, PostgreSQL only requires the binaries to be replaced and the database to be restarted. For more information see [PostgreSQL - Versioning Policy](https://www.postgresql.org/support/versioning/)

{{< hint type=info >}}This procedure can also be used for all other containers in a cluster. Whether sidecars, exporter, pooler or backup image{{< /hint >}}


### Preconditions:
1. Check if there is a newer image for the PostgreSQL container - [Check on Docker hub](https://hub.docker.com/repository/docker/cybertecpostgresql/cybertec-pg-container/general)
2. Check - Check that the new `PGVERSION` is larger than the previously used one.
3. Check whether the new `PGVERSION` is larger than the previously used one and the maintenance mode of the cluster must be deactivated. In addition, the replicas should not have a high lag.

### Updating PostgreSQL-Container-Image
Old-Manifest:
```yaml
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-1
  namespace: cpo
spec:
  dockerImage: 'docker.io/cybertecpostgresql/cybertec-pg-container:postgres-17.3-1'
```
New-Manifest:
```yaml
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-1
  namespace: cpo
spec:
  dockerImage: 'docker.io/cybertecpostgresql/cybertec-pg-container:postgres-17.4-1'
```
#### Updating via kubectl/oc-client
```sh
kubectl patch postgresqls.cpo.opensource.cybertec.at cluster-1 --type='merge' -p \
'{"spec":{"dockerImage":"docker.io/cybertecpostgresql/cybertec-pg-container:postgres-17.4-1"}}'
```

### Updating Exporter-Container-Image

#### Updating Cluster-Manifest:
Old-Manifest:
```yaml
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-1
  namespace: cpo
spec:
  monitor:
    image: 'docker.io/cybertecpostgresql/cybertec-pg-container:exporter-17.3-1'
```
New-Manifest:
```yaml
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-1
  namespace: cpo
spec:
  monitor:
    image: 'docker.io/cybertecpostgresql/cybertec-pg-container:exporter-17.4-1'
```

#### Updating via kubectl/oc-client
```sh
kubectl patch postgresqls.cpo.opensource.cybertec.at cluster-1 --type='merge' -p \
'{"spec":{"monitor":{"image":"docker.io/cybertecpostgresql/cybertec-pg-container:exporter-17.4-1"}}}'
```

