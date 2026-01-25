---
title: "via S3"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 2
---

This chapter describes the use of pgBackRest in combination with AWS S3 or S3-compatible storage such as MinIO, Cloudian HyperStore, or SwiftStack. While it is not mandatory to operate Kubernetes on the AWS Cloud Platform, the efficiency and duration of a backup depend on the network connection to your storage provider.

{{< hint type=important >}} Precondition: A S3 bucket and a privileged role/user with valid credentials are required before proceeding. {{< /hint >}}

1. Create the Authentication Secret

The operator needs access to your S3 bucket. The credentials and the encryption passphrase are stored in a Kubernetes Secret. This is most easily done by creating a file named s3.conf:
```
[global]
repo1-s3-key=YOUR_S3_ACCESS_KEY
repo1-s3-key-secret=YOUR_S3_KEY_SECRET
repo1-cipher-pass=YOUR_ENCRYPTION_PASSPHRASE
```
{{< hint type=info >}} repo1-cipher-pass is only required if you want to use the backup encryption feature of pgBackRest. {{< /hint >}}

Then, create the secret using `kubectl`:

```
# Create the secret in the same namespace as your cluster
kubectl create secret generic cluster-1-s3-credentials --from-file=s3.conf=s3.conf
```

2. Modifying the Cluster Manifest

Once the secret is created, the cluster manifest must be adapted. This involves defining the repository settings, the backup schedule, and the S3-specific parameters.
S3 Addressing Styles (Host vs. Path)

A critical parameter for S3 compatibility is the repo1-s3-uri-style.

    host: (Default) Accesses the bucket via https://bucket-name.s3.endpoint.com. Used by standard AWS S3.

    path: Accesses the bucket via https://s3.endpoint.com/bucket-name. Often required for MinIO, Ceph, or other on-premise S3 implementations.

    {{< hint type=info >}} The default value is host, so it does not necessarily have to be set unless path is required. {{< /hint >}}


```
  apiVersion: cpo.opensource.cybertec.at/v1
  kind: postgresql
  metadata:
    name: cluster
    namespace: cpo
  spec:
    backup:
      pgbackrest:
        image: 'docker.io/cybertecpostgresql/cybertec-pg-container:pgbackrest-18.1-1'
        repos:
          - endpoint: 's3.eu-central-1.amazonaws.com'
            name: repo1
            region: eu-central-1
            resource: cpo-cluster-bucket
            schedule:
              full: 30 2 * * *
              incr: '*/30 * * * *'
            storage: s3
        configuration:
          secret: cluster-1-s3-credential
        global:
          repo1-path: /cluster/repo1/
          repo1-retention-full: '7'
          repo1-retention-full-type: count
          repo1-s3-uri-style: host
```

{{< hint type=info >}} Each pgBackRest parameter can be used by adding it to the global section. See [pgbackrest documentation](https://pgbackrest.org/configuration.html). {{< /hint >}}

An [example](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/tree/main/cluster-tutorials/pgbackrest_with_s3) with a secret generator is also available in the tutorials. Enter your access data in the s3.conf file and transfer the tutorial to your Kubernetes with kubectl apply -k cluster-tutorials/pgbackrest_with_s3/.
