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

The CYBERTEC-pg-operator, together with Patroni, takes over the setup and administration of the TDE functionality in conjunction with the cost-effective PGEE containers

### Preconditions
- CYBERTEC-pgee-container
- Valid licence agreement for PGEE

### Deploy a TDE-Cluster

Setting up a TDE cluster is basically the same as setting up a conventional cluster. 
The only difference is the defined Postgres-container and the object TDE.enabled: true, which instructs the operator to initialise the database with the TDE functionality.

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
    keybits: 256
  teamId: acid
  volume:
    size: 5Gi
```
- `dockerImage` - Must contain a PostgreSQL image of the pgee container suite
- `tde.enabled`- initialises the DB with TDE
- `tde.keybits`- Defines keylength in bits. Possible Values: 128,192,256 Default: 256

{{< hint type=important >}} Please note that the activation of TDE is only possible when creating new clusters. Subsequent activation is not possible. {{< /hint >}}

### Key-Management

For TDE, we or the operator must work with the required encryption key. The key is transferred to the Postgres containers using a secret. There are two basic options for the necessary key management.

#### Automatic key generation (default)
If no existing secret is provided, the operator automatically generates a cryptographically secure key when creating a new cluster. This is stored in the secret in the cluster's namespace.

It is important to note for key management that TDE allows you to choose from the available key lengths (128 bit, 192 bit, 256 bit). By default, the operator chooses 256 bit. The keybits parameter allows you to adjust this if desired. See [CRD](crd/crd-postgresql).
Further information on TDE can also be found in the [PGEE-Documentation](https://repository.cybertec.at/doc/18ee/encryption.html)

#### Use of your own keys (Bring Your Own Key)
You have the option of defining your own encryption key before the cluster is created.
To do this, create a secret with the desired key in advance.
When starting, the operator checks whether a corresponding secret already exists and uses this instead of a newly generated key.
Use case: This enables the integration of external secret store solutions (e.g. HashiCorp Vault) to stream or synchronise keys directly into the Kubernetes secret.

{{< hint type=important >}} If you provide your own key, you must ensure that the length of the key (in bytes) matches the configured keybits exactly. 
| keybits  | Keylength (Byte) | Notes          |
| -------- |:----------------:| --------------:|
| 128      | 16 Byte          |                |
| 192      | 24 Byte	        |                |
| 256      | 32 Byte          | (Default)      |

A discrepancy between the configured bit length and the actual byte length of the provided key will result in errors when starting the database.
{{< /hint >}}

The secret name follows the following fixed naming convention: [CLUSTERNAME]-tde

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: [CLUSTERNAME]-tde
stringData:
  key: [TDE-KEY]
type: Opaque
```

### Check TDE-Status

```sh

sh-5.1$ /usr/pgsql-17/bin/pg_controldata  | grep -i "Encryption"
Data encryption:                      on
Encryption key length:                128

[postgres@tde-cluster-1-0 ~]$ psql
psql (18.1 EE 1.5.0, server 17.7 EE 1.4.4)
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