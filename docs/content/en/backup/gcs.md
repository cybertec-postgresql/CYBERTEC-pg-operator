---
title: "via GCS"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 3
---

This chapter describes the use of pgBackRest in combination with Google Cloud Storage (gcs). It is not absolutely necessary to operate a Kubernetes on the Google Cloud Platform. However, as with any cloud storage, the efficiency and therefore the duration of a backup depends on the connection.

{{< hint type=important >}} Precondition: a gcs-bucket and a priviledged role is needed for this chapter. {{< /hint >}}

### Create a gcs-bucket on the google cloud console

### Create a priviledged service-role

### Modifying the Cluster 
As soon as all requirements are met:

- A GCS bucket
- A JSON token for the service role with the required authorisations for the bucket

the cluster can be modified. Firstly, a secret containing the JSON token is created and the cluster manifest is adapted accordingly.

The first step is to create the required secret. This is most easily done using a `kubectl` command.

```
kubectl create secret generic cluster-1-gcs-credentials --from-file=gcs.json=fluent.json
```

In the next step, both the secret name and the file name of the JSON token are stored in the secret in the cluster manifest. In addition, global settings, such as the retention time of the backups in the global object, are defined, the image for `pgBackRest` is specified and the necessary information for the repository is added. This includes both the desired storage path in the bucket and the times for automatic backups based on the cron syntax.

```
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-1
spec:
  backup:
    pgbackrest:
      configuration:
        secret: cluster-1-gcs-credentials
      global:
        repo1-path: /cluster-1/repo1/
        repo1-retention-full: '7'
        repo1-retention-full-type: count
      image: docker.io/cybertecpostgresql/cybertec-pg-container:pgbackrest-16.4-1'
      repos:
        - name: repo1
          resource: postgresql-backup-bucket
          key: gcs.json
          keyType: service
          schedule:
            full: 30 2 * * *
          storage: gcs
```
