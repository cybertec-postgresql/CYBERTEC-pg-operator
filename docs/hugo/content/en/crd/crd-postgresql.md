---
title: "PostgreSQL"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 331
---
#### CRD for kind postgresql

| Name        | Type           | required  | Description        |
| ----------- |:--------------:| ---------:| ------------------:|
| apiVersion  | string         | true      | cpo.opensource.cybertec.at/v1 |
| kind        | string         | true      |                    |
| metadata    | object         | true      |                    |
| [spec](#spec)        | object         | true      |                    |
| [status](#status)      | object         | false     |                    |

---

#### spec

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| [additionalVolumes](#additionalvolumes) | array   | false     | List of additional volumes to mount in each container of the statefulset pod |
| allowedSourceRanges            | array   | false     | The corresponding load balancer is accessible only to the networks defined by this parameter |
| [backup](#backup)              | object  | false     | Enables the definition of a customised backup solution for the cluster |
| [clone](#clone)                | object  | false     | Defines the clone-target for the Cluster |
| [connectionPooler](#connectionpooler) | object  | false     | Defines the configuration and settings for every type of a connectionPoolers (Primary and Replica). |
| databases                      | map     | false     | Defines the name of the database, they are created by the operator. See [tutorial](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/tree/main/cluster-tutorials/configure_users_and_databases) |
| dockerImage                    | string  | true      | Defines the used PostgreSQL-Container-Image for this cluster |
| enableLogicalBackup            | boolean | false     | Enable logical Backups for this Cluster (Stored on S3) - s3-configuration for Operator is needed (Not for pgBackRest) |
| enableConnectionPooler         | boolean | false     | creates a ConnectionPooler for the primary Pod |
| enableReplicaConnectionPooler  | boolean | false     | creates a ConnectionPooler for the replica Pods |
| enableMasterLoadBalancer       | boolean | false     | Define whether to enable the load balancer pointing to the Postgres primary |
| enableReplicaLoadBalancer      | boolean | false     | Define whether to enable the load balancer pointing to the Postgres replicas |
| enableMasterPoolerLoadBalancer | boolean | false     | Define whether to enable the load balancer pointing to the primary ConnectionPooler |
| enableReplicaPoolerLoadBalancer| boolean | false     | Define whether to enable the load balancer pointing to the Replica-ConnectionPooler |
| enableShmVolume                | boolean | false     | Start a database pod without limitations on shm memory. By default Docker limit /dev/shm to 64M (see e.g. the docker issue, which could be not enough if PostgreSQL uses parallel workers heavily. If this option is present and value is true, to the target database pod will be mounted a new tmpfs volume to remove this limitation. |
| [env](#env)                    | array   | false     | Allows to add own Envs to the PostgreSQL containers |
| [initContainers](#initcontainers) | array   | false    | Enables the definition of init-containers |
| logicalBackupSchedule          | string  | false     | Enables the scheduling of logical backups based on cron-syntax. Example: `30 00 * * *` |
| maintenanceWindows             | array   | false     | Enables the definition of maintenance windows for the cluster. Example: `Sat:00:00-04:00` |
| masterServiceAnnotations       | map     | false     | Enables the definition of annotations for the Primary Service |
| [monitor](#monitor)            | map     | false     | Enables monitoring on the basis of the defined image |
| nodeAffinity                   | map     | false     | Enables overwriting of the nodeAffinity |
| numberOfInstances              | int     | true      | Number of nodes of the cluster |
| [patroni](#patroni)             | map     | false     | Enables the customisation of patroni settings |
| podPriorityClassName           | string  | false     | a name of the priority class that should be assigned to the cluster pods. If not set then the default priority class is taken. The priority class itself must be defined in advance |
| podAnnotations                 | map     | false     | A map of key value pairs that gets attached as annotations to each pod created for the database. |
| [postgresql](#postgresql)      | map     | false     | Enables the customisation of PostgreSQL settings and parameters |
| [preparedDatabases](#prepareddatabases) | map     | false     | Allows you to define databases including owner, schemas and extension and have the operator generate them. item See [tutorial](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/tree/main/cluster-tutorials/prepared_databases) |
| replicaServiceAnnotations      | map     | false     | Enables the definition of annotations for the Replica Service |
| [resources](#resources)        | map     | true      | CPU & Memory (Limit & Request) definition for the Postgres container |
| ServiceAnnotations             | map     | false     | A map of key value pairs that gets attached as annotations to each Service created for the database. |
| [sidecars](#sidecars)          | array   | false     | Enables the definition of custom sidecars |
| spiloFSGroup                   | int     | false    |  the Persistent Volumes for the Spilo pods in the StatefulSet will be owned and writable by the group ID specified. This will override the spilo_fsgroup operator parameter |
| spiloRunAsGroup                | int     | false     | sets the group ID which should be used in the container to run the process. |
| spiloRunAsUser                 | int     | false     | Sets the user ID which should be used in the container to run the process. This must be set to run the container without root. |
| [standby](#standby)            | map     | false     | Enables the creation of a standby cluster at the time of the creation of a new cluster |
| [streams](#streams)            | array   | false     | Enables change data capture streams for defined database tables |
| [tde](#tde)                    | map     | false     | Enables the activation of TDE if a new cluster is created |
| teamId                         | string  | true      | name of the team the cluster belongs to. Will be removed soon |
| [tls](#tls)                    | map     | false     | Custom TLS certificate |
| [tolerations](#tolerations)    | array    | false    | a list of tolerations that apply to the cluster pods. Each element of that list is a dictionary with the following fields: 
key, operator, value, effect and tolerationSeconds |
| [topologySpreadConstraints](https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/) | map     | false    | Enables the definition of a topologySpreadConstraint. See [K8s-Documentation](https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/) |
| users                          | map     | false     | a map of usernames to user flags for the users that should be created in the cluster by the operator. See [tutorial](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/tree/main/cluster-tutorials/configure_users_and_databases) |
| usersWithSecretRotation        | list    | false     | list of users to enable credential rotation in K8s secrets. The rotation interval can only be configured globally. |
| usersWithInPlaceSecretRotation | list    | false     | list of users to enable in-place password rotation in K8s secrets. The rotation interval can only be configured globally. |
| [volume](#volume)              | map     | true      | define the properties of the persistent storage that stores Postgres data |


{{< back >}}

---

#### additionalVolumes

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| name                           | string  | true      | Enables the definition of a pgbackrest-setup for the cluster |
| mountPath                      | string  | true      | Enables the definition of a pgbackrest-setup for the cluster |
| targetContainers               | array   | true      | Enables the definition of a pgbackrest-setup for the cluster |
| subPath                        | string  | false     | Enables the definition of a pgbackrest-setup for the cluster |
| isSubPathExpr                  | boolean | false     | Enables the definition of a pgbackrest-setup for the cluster |
| [volumeSource](#volumeSource)  | map     | true      | Enables the definition of a pgbackrest-setup for the cluster |

{{< back >}}

---

#### backup

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| [pgbackrest](#pgbackrest)      | object  | false     | Enables the definition of a pgbackrest-setup for the cluster |

{{< back >}}

---

#### clone

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| cluster                        | string  | true      | Name of the cluster to be cloned. Random value if the cluster does not exist locally.  |
| [pgbackrest](#pgbackrest)      | object  | false     | Enables the definition of a pgbackrest-setup for the cluster |

{{< back >}}

---

#### connectionPooler

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| numberOfInstances              | int     | true      | Number of Pods per Pooler  |
| mode                           | string  | true      | pooling mode for pgBouncer (session, transaction, statement) |
| schema                         | string  | true      | Schema for Pooler (Default: pooler) |
| user                           | string  | true      | Username for Pooler (Default: pooler) |
| maxDBConnections               | string  | true      | maxConnections to the DB-Pod(s) |
| [resources](#resources)        | map     | true      | CPU & Memory (Limit & Request) definition for the Pooler |

{{< back >}}
---

#### env

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| name                           | string  | true      | Keyfield for the ENV-Entry  |
| value                          | string  | true      | Valuefield for the ENV-Entry  |

{{< back >}}

---

#### initContainers

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| name                           | string  | true      | Name for the container  |
| image                          | string  | true      | Docker-Image for container  |
| command                        | string  | false     | to override CMD inside the container  |
| [env](#env)                    | array   | false     | Allows to add own Envs to the container |
| [resources](#resources)        | map     | false     | CPU & Memory (Limit & Request) definition for the container  |
| [ports](ports)                 | array   | false     | Define open ports for the container  |

{{< back >}}

---

#### monitor

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| image                          | string  | true      | Docker-Image for the metric exporter  |

{{< back >}}

---

#### patroni

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| failsafe_mode                  | boolean | false     | Patroni failsafe_mode parameter value. See the [Patroni documentation](https://patroni.readthedocs.io/en/master/dcs_failsafe_mode.html) for more details.  |
| initdb                         | map     | false     | a map of key-value pairs describing initdb parameters  |
| loop_wait                      | string  | false     | Patroni `loop_wait` parameter value, optional. The default is set by the PostgreSQL image.  |
| maximum_lag_on_failover        | string  | false     | Patroni `maximum_lag_on_failover` parameter value, optional. The default is set by the PostgreSQL image.  |
| [multisite](#multisite)        | map     | false     | Multisite configuration - Check the [Documentation](CYBERTEC-pg-operator/multisite/) first  |
| pg_hba                         | array   | false     | list of custom pg_hba lines to replace default ones. One entry per item (example: - hostssl all all 0.0.0.0/0 scram-sha-256)  |
| retry_timeout                  | int     | false     | Patroni `retry_timeout` parameter value, optional. The default is set by the PostgreSQL image.  |
| [slots](#slots)                | map     | false     | permanent replication slots that Patroni preserves after failover by re-creating them on the new primary immediately. after doing a promote. Use preferred slot-name as map-item |
| synchronous_mode               | boolean | false     | DPatroni `synchronous_mode` parameter value, optional. The default is false.  |
| synchronous_mode_strict        | boolean | false     | Patroni `synchronous_mode_strict` parameter value, optional. The default is false.  |
| synchronous_node_count         | int     | false     | Patroni `synchronous_node_count` parameter value, optional. The default is set to 1. Only used if `synchronous_mode_strict` is true  |
| ttl                            | int     | false     | Patroni `ttl` parameter value, optional. The default is set by the PostgreSQL image.  |

{{< back >}}

---

#### PostgreSQL

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| parameters                     | map     | false     | PostgreSQL-Parameter as item (Example: max_connections: "100"). For help check out the [CYBERTEC PostgreSQL Configurator](https://pgconfigurator.cybertec.at)  |
| version                        | string  | false     | a map of key-value pairs describing initdb parameters  |

{{< back >}}

---

#### preparedDatabases

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| defaultUsers                   | boolean | false     | Creates roles with `LOGIN` permission and `_user`suffix. Default: false |
| extensions                     | map     | false     | Includes the Extensions as items (key:value). Key is the Name of the Extension and value the schema. Example: pgcrypto: public |
| [schemas](#schemas)            | map     | false     | Includes the schemanames as items. |

{{< back >}}

---

#### resources

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| [requests](#requests)          | map     | true      | cpu and memory definitons (request.cpu / request.memory) |
| [limits](#limits)              | map     | true      | cpu and memory definitons (limits.cpu / limits.memory) |

{{< back >}}

---

#### sidecars

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| name                           | string  | true      | Name for the container  |
| image                          | string  | true      | Docker-Image for container  |
| command                        | string  | false     | to override CMD inside the container  |
| [env](#env)                    | array   | false     | Allows to add own Envs to the container |
| [resources](#resources)        | map     | false     | CPU & Memory (Limit & Request) definition for the container  |
| [ports](ports)                 | array   | false     | Define open ports for the container  |

{{< back >}}

#### standby

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| standby_host                   | string  | true      | Endpoint of the primary cluster  |
| standby_port                   | string  | true      | PostgreSQL port of the primary cluster  |

{{< back >}}

---

#### streams

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| applicationId                  | string  | true      | The application name to which the database and CDC belongs to.  |
| database                       | string  | true      | Name of the database from where events will be published via Postgres' logical decoding feature.  |
| tables                         | map     | true      | Defines a map of table names and their properties (eventType, idColumn and payloadColumn).  |
| batchSize                      | int     | false     | Defines the size of batches in which events are consumed. Defaults to 1  |
| enableRecovery                 | boolean | false     | Flag to enable a dead letter queue recovery for all streams tables.  |
| filter                         | string  | false     | Streamed events can be filtered by a jsonpath expression for each table.  |
| standby_port                   | string  | false     | PostgreSQL port of the primary cluster  |

{{< back >}}

---

#### tde

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| enable                         | boolean | true      | enable TDE during initDB  |

{{< back >}}

---

#### tolerations

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| key                            | string  | false     | Key for the taint attribute of the node  |
| operator                       | string  | false     | Comparison operator (Equal or Exists).  |
| value                          | string  | false     | Value of the taint (only relevant for ‘Equal’).  |
| effect                         | string  | false     | Specifies how the node handles the pod (NoExecute, NoSchedule, PreferNoSchedule)  |
| tolerationSeconds              | int     | false     | Specifies how long the pod tolerates the taint (only for NoExecute).  |


{{< back >}}

---

#### volume

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| size                           | string  | true      | the size of the target volume. Usual Kubernetes size modifiers, i.e. Gi or Mi, apply  |
| storageClass                   | string  | false     | the name of the Kubernetes storage class to draw the persistent volume from. If empty K8s will choose the default StorageClass  |
| subPath                        | string  | false     | Subpath to use when mounting volume into PostgreSQL container.  |
| iops                           | int     | false     | When running the operator on AWS the latest generation of EBS volumes (gp3) allows for configuring the number of IOPS. Maximum is 16000  |
| throughput                     | int     | false     | When running the operator on AWS the latest generation of EBS volumes (gp3) allows for configuring the throughput in MB/s. Maximum is 1000  |
| selector                       | map     | false     | A label query over PVs to consider for binding. See the [Kubernetes documentation](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) for details on using matchLabels and matchExpressions  |

{{< back >}}

---

#### volumeSource

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| emptyDir                       | string  | false      | emptyDir: {} |
| [PersistentVolumeClaim](#volumeSource-PersistentVolumeClaim) | map   | false      | PersistentVolumeClaim-Objekt |
| [configMap](#volumeSource-configMap)             | map   | false      | configMap-Objekt |

{{< back >}}

---

#### volumeSource-PersistentVolumeClaim

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| claimName                      | string  | true      | Name of the PersistentVolumeClaim  |
| readyOnly                      | boolean | false     | PersistentVolumeClaim-Objekt |

{{< back >}}

---

#### volumeSource-configMap

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| name                           | string  | true      | Name of the Configmap  |

{{< back >}}

---

#### multisite

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| enable                         | boolean | true      | Enable multisite-feature |
| [etcd](#etcd)                  | map     | true      | Enables the definition of a pgbackrest-setup for the cluster |
| retry_timeout                  | int     | true      | Patroni `retry_timeout` parameter value for the global etcd, optional. The default is set by the PostgreSQL image.  |
| site                           | string  | true      | Name for the site of this cluster |
| ttl                            | int     | true      | Patroni `ttl` parameter value for the global etcd, optional. The default is set by the PostgreSQL image.  |

{{< back >}}

---

#### slots

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| type                           | string  | true      | Slot-Type (`physical` or `logical`) |
| database                       | string  | false     | Databasename - for logical replication only  |
| plugin                         | string  | false     | Plugin - for logical replication only  |

{{< back >}}
---

#### schemas

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| defaultRoles                   | boolean | false     | Creates schema exclusiv roles with `NOLOGIN` permission and `_user`suffix Default: true |
| defaultUsers                   | boolean | false     | Creates schema exclusiv roles with `LOGIN` permission and `_user`suffix Default: false |

#### etcd

| Name           |  Type  | required |                                                                                               Description |
|----------------|:------:|---------:|----------------------------------------------------------------------------------------------------------:|
| hosts          | string |     true | list of etcd hosts, including etcd-client-port (default: `2379`), comma separated like in the etcd config |
| password       | string |    false |                                                                              Password for the global etcd |
| protocol       | string |     true |                                                              Protocol for the global etcd (http or https) |
| user           | string |    false |                                                                              Username for the global etcd |
| certSecretName | string |    false |                   Secret for client certificates (tls.crt/key) and server certificate validation (ca.crt) |

{{< back >}}

---

#### requests

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| cpu                            | string  | true      | cpu definitons Example: 1000m|
| memory                         | string  | true      | memory definitons Example: 1000Mi|

{{< back >}}

---

#### limits

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| cpu                            | string  | true      | cpu definitons Example: 1000m|
| memory                         | string  | true      | memory definitons Example: 1000Mi| 

{{< back >}}

---

#### pgbackrest

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| [configuration](#configuration)| object  | false     | Enables the definition of a pgbackrest-setup for the cluster |
| global                         | object  | false     |  |
| image                          | string  | true      |  |
| [repos](#repos)                | array   | true      |  |
| [resources](#resources)        | object  | false     | CPU & Memory (Limit & Request) definition for the pgBackRest container|

{{< back >}}

---

#### configuration

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| secret                         | object  | false     | Secretname with the contained S3 credentials (AccessKey & SecretAccessKey) (Note: must be placed in the same namespace as the cluster) |
| [protection](#protection)      | object  | false     | Enable Protection-Options |

{{< back >}}

---



#### protection

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| restore                        | boolean | false     | A restore is ignored as long as this option is set to true. |

{{< back >}}

---

#### repos

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| name                           | string  | true      | Name of the repository Required:Repo[1-4] |
| storage                        | string  | true      | Defines the used backup-storage (Choose from List: pvc,s3,blob,gcs) |
| resource                       | string  | true      | Bucket-/Instance-/Storage- or PVC-Name |
| endpoint                       | string  | false     | The Endpoint for the choosen Storage (Not required for local storage) |
| region                         | string  | false     | Region for the choosen Storage (S3 only) |
| [schedule](#schedule)          | string  | false     | Object for defining automatic backups |

{{< back >}}

---

#### schedule

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| full                           | string  | false     | (Cron-Syntax) 	Define full backup |
| incr                           | string  | false     | (Cron-Syntax) 	Define incremental backup |
| diff                           | string  | false     | (Cron-Syntax) 	Define differential backup |

{{< back >}}

---

#### status

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| PostgresClusterStatus          | string  | false     | Shows the cluster status. Filled by the Operator |

{{< back >}}
