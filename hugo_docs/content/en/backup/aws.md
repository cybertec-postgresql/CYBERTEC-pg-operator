---
title: "via S3"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 2
---

This chapter describes the use of pgBackRest in combination with with AWS S3 or S3-compatible storage such as MinIO, Cloudian HyperStore or SwiftStack. It is not absolutely necessary to operate a Kubernetes on the AWS  Cloud Platform. However, as with any cloud storage, the efficiency and therefore the duration of a backup depends on the connection.

This Chapter will use AWS S3 for the example, the usage of different s3-compatible Storage is similiar.

{{< hint type=important >}} Precondition: a S3-bucket and a priviledged role with credentials is needed for this chapter. {{< /hint >}}

### Create a s3-bucket on the AWS console

### Create a priviledged service-role

### Modifying the Cluster 
As soon as all requirements are met:

- A S3 bucket
- Access-Token and Secret-Access-Key for the service role with the required authorisations for the bucket

the cluster can be modified. Firstly, a secret containing the Credentials is created and the cluster manifest is adapted accordingly.

The first step is to create the required secret. This is most easily done storing the needed data in a file called s3.conf and using a `kubectl` command.

```
# Create a file with name s3.conf and add the following infos. Please replace the placeholder by the credentials
[global]
repo1-s3-key=YOUR_S3_ACCESS_KEY
repo1-s3-key-secret=YOUR_S3_KEY_SECRET
repo1-cipher-pass=YOUR_ENCRYPTION_PASSPHRASE

# Create the secret with the credentials
kubectl create secret generic cluster-1-s3-credentials --from-file=s3.conf=s3.conf
```

In the next step, the secret name ais stored in the secret in the cluster manifest. In addition, global settings, such as the retention time of the backups in the global object, are defined, the image for `pgBackRest` is specified and the necessary information for the repository is added. This includes both the desired storage path in the bucket and the times for automatic backups based on the cron syntax.

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
        - endpoint: 'https://s3-zurich.cyberlink.cloud:443'
          name: repo1
          region: zurich
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
```

This example creates a backup in the defined S3 bucket. In addition to the above configurations, a secret is also required which contains the access data for the S3 storage. The name of the secret must be stored in the `spec.backup.pgbackrest.configuration.secret` object and the secret must be located in the same namespace as the cluster.
Information required to address the S3 bucket:
- `Endpoint`: S3 api endpoint
- `Region`: Region of the bucket
- `resource`: Name of the bucket

An [example](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/tree/main/cluster-tutorials/pgbackrest_with_s3) with a sercret generator is also available in the tutorials. Enter your access data in the s3.conf file and transfer the tutorial to your Kubernetes with kubectl apply -k cluster-tutorials/pgbackrest_with_s3/.
