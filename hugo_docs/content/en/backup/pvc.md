---
title: "via Blockstorage (pvc)"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 1
---

### Backups on PVC (PersistentVolumeClaim)

When using block storage, the operator creates an additional pod that acts as a repo host. Based on a TLS connection, the repo host obtains the data for the Backup from the current primary of the cluster, which is compressed before being sent.
WAL archives are pushed from the primary pod to the repo host.

```
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster
  namespace: cpo
spec:
  backup:
    pgbackrest:
      image: 'docker.io/cybertecpostgresql/cybertec-pg-container:pgbackrest-16.4-1'
      repos:
        - name: repo1
          schedule:
            full: 30 2 * * *
          storage: pvc
          volume:
            size: 15Gi
            storageClass: default
      global:
        repo1-retention-full: '7'
        repo1-retention-full-type: count
```

This example creates backups based on a repo host with a daily full Backup at 2:30 am. In addition, pgBackRest is instructed to keep a maximum of 7 full Backups. The oldest one is always removed when a new Backup is created. You can increase the pvc-size all time if needed. Therefore you just need to update the `size` value to a higher amount of Gi. Please be aware that shrinking the volume is not possible. 

{{< hint type=info >}} In addition, further configurations for pgBackRest can be defined in the global object. Information on possible configurations can be found in the [pgBackRest documentation](https://pgbackrest.org/configuration.html) {{< /hint >}}


