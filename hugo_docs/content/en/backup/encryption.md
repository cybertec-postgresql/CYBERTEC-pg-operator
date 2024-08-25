---
title: "Encrypted Backups"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 6
---
pgBackRest also allows you to encrypt your backups on the client side before uploading them. This is possible with any type of storage and is very easy to activate. 

Firstly, we need to define an encryption key. This must be specified separately for each repo and stored in the same secret that is defined in the `spec.backup.pgbackrest.configuration.secret` object.
```
kind: Secret
apiVersion: v1
metadata:
  name: cluster-s3-credential
  namespace: cpo
stringData:
  s3.conf |
    [global]
    repo1-s3-key=YOUR_S3_KEY
    repo1-s3-key-secret=YOUR_S3_KEY_SECRET
    repo1-cipher-pass=YOUR_ENCRYPTION_KEY
```

We also need to configure the type of encryption for pgBackRest. This is done via the cipher-type parameter, which must also be specified for each repo.  You can find the available values for the parameter [here](https://pgbackrest.org/configuration.html#section-repository/option-repo-cipher-type) 

```
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster
  namespace: cpo
spec:
  backup:
    pgbackrest:
      configuration:
        secret: cluster-s3-credential
      global:
        repo1-path: /cluster/repo1/
        repo1-retention-full: '7'
        repo1-retention-full-type: count
        repo1-cipher-type: aes-256-cbc
      image: 'docker.io/cybertecpostgresql/cybertec-pg-container-dev:pgbackrest-16.3-1'
      repos:
        - endpoint: 'https://s3-zurich.cyberlink.cloud:443'
          name: repo1
          region: zurich
          resource: cpo-cluster-bucket
          schedule:
            full: 30 2 * * *
            incr: '*/30 * * * *'
          storage: s3
```