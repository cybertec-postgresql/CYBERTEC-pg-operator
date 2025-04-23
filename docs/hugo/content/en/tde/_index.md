---
title: "TDE"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 2200
---
## What is Transparent Data Encryption (TDE)?

Transparent Data Encryption (TDE) is a technology for encrypting databases at file level. The data is automatically encrypted before it is stored on the storage medium and decrypted transparently for authorised applications and users if required. This ensures that the data is protected at rest without the need for changes to existing applications. TDE is used by various database vendors such as Microsoft, Oracle and IBM to increase the security of database files.

### Difference between hard disk encryption and TDE

Hard disk encryption, also known as Full Disk Encryption (FDE), encrypts the entire hard disk or individual partitions to prevent unauthorised access to sensitive data. This method protects all data on a system, including the operating system, but only when the system is switched off. As soon as an authorised user logs on, the encryption is removed and the data is accessible to anyone who can access the computer while the user is logged on.

In contrast, TDE specifically encrypts the database files at file level. Encryption is transparent to the applications accessing the database and protects the data at rest, regardless of the status of the operating system or hardware. This provides an additional protection mechanism, especially in scenarios where hard disk encryption is not sufficient or not implemented.


### Advantages of TDE

- **Protection of data at rest**: Data on the storage medium is encrypted, reducing the risk of data leaks.
- **Transparency for applications**: Encryption is done directly at database level, so no changes to existing applications are required.
- **Integration with PGEE**: Full support in Kubernetes environments and other modern IT infrastructures.
- **Fulfilment of regulatory requirements**: Support for compliance requirements such as GDPR, HIPAA and other data protection standards.
- **Additional security features**: In combination with other PGEE features such as data masking and obfuscation, comprehensive protection of sensitive data is ensured.

Further information on TDE and PGEE can be found here: [CYBERTEC TDE](https://www.cybertec-postgresql.com/en/products/postgresql-transparent-data-encryption/).

## Securing clusters with TDE

The CYBERTEC pg operator, together with Patroni, takes over the setup and administration of the TDE functionality in conjunction with the cost-effective PGEE containers

### Preconditions
- CYBERTEC-pgee-container
- Valid licence agreement for PGEE

### Deploy a TDE-Cluster

Setting up a TDE cluster is basically the same as setting up a conventional cluster. 
The only difference is the defined Postgres. container and the object TDE.enabled: true, which instructs the operator to initialise the database with the TDE functionality.

```yaml
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: tde-cluster-1
  namespace: cpo
spec:
  dockerImage: 'containers.cybertec.at/cybertec-pgee-container/postgres:rocky9-17.4-1'
  numberOfInstances: 1
  postgresql:
    version: '17'
  resources:
    limits:
      cpu: 250m
      memory: 500Mi
    requests:
      cpu: 250m
      memory: 500Mi
  tde:
    enable: true
  teamId: acid
  volume:
    size: 5Gi
```
- `dockerImage` - Must contain a PostgreSQL image of the pgee container suite
- `tde.enabled`- initialises the DB with TDE

{{< hint type=important >}} Please note that the activation of TDE is only possible when creating new clusters. Subsequent activation is not possible. {{< /hint >}}

### Check TDE-Status

```sh
[postgres@tde-cluster-1-0 ~]$ psql
psql (17.4 EE 1.4.1)
 ____   ____ _____ _____
|  _ \ / ___| ____| ____|
| |_) | |  _|  _| |  _|
|  __/| |_| | |___| |___
|_|    \____|_____|_____|
PostgreSQL EE by CYBERTEC
Type "help" for help.

postgres=# show data_encryption;
 data_encryption 
-----------------
 on
(1 row)
```