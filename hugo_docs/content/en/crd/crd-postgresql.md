---
title: "PostgreSQL"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 331
---
#### postgresql

| Name        | Type           | required  | Description        |
| ----------- |:--------------:| ---------:| ------------------:|
| apiVersion  | string         | true      | acid.zalando.do/v1 |
| kind        | string         | true      |                    |
| metadata    | object         | true      |                    |
| [spec](#spec)        | object         | true      |                    |
| [status](#status)      | object         | false     |                    |

---

#### spec

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| [additionalVolumes](additionalVolumes) | array   | false     | List of additional volumes to mount in each container of the statefulset pod |
| allowedSourceRanges            | array   | false     | The corresponding load balancer is accessible only to the networks defined by this parameter |
| [backup](#backup)              | object  | false     | Enables the definition of a customised backup solution for the cluster |
| [clone](#clone)                | object  | false     | Defines the clone-target for the Cluster |
| [connectionPooler](#connectionPooler) | object  | false     | Defines the configuration and settings for every type of a connectionPoolers (Primary and Replica). |
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
| [env](#env)                    | array   | false     | Allows to add own Envs (added as map) |
| [initContainers](#initContainers) | array   | false    | Enables the definition of init-containers |
| logicalBackupSchedule          | string  | false     | Enables the scheduling of logical backups based on cron-syntax. Example: `30 00 * * *` |
| maintenanceWindows             | array   | false     | Enables the definition of maintenance windows for the cluster. Example: `Sat:00:00-04:00` |
| masterServiceAnnotations       | map     | false     | Enables the definition of annotations for the Primary Service |
| [monitor](#monitor)            | map     | false     | Enables monitoring on the basis of the defined image |
| nodeAffinity                   | map     | false     | Enables overwriting of the nodeAffinity |
| numberOfInstances              | int     | true      | Number of nodes of the cluster |
| [patroni](#patroni             | map     | false     | Enables the customisation of patroni settings |
| podPriorityClassName           | string  | false     | a name of the priority class that should be assigned to the cluster pods. If not set then the default priority class is taken. The priority class itself must be defined in advance |
| podAnnotations                 | map     | false     | A map of key value pairs that gets attached as annotations to each pod created for the database. |
| [postgresql](#postgresql)      | map     | false     | Enables the customisation of PostgreSQL settings and parameters |
| [preparedDatabases](#preparedDatabases) | map     | false     | Allows you to define databases including owner, schemas and extension and have the operator generate them. See [tutorial](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/tree/main/cluster-tutorials/prepared_databases) |
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


[back to parent](#postgresql)

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

[back to parent](#spec)

---

#### backup

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| [pgbackrest](#pgbackrest)      | object  | false     | Enables the definition of a pgbackrest-setup for the cluster |

[back to parent](#spec)

---

#### clone

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| cluster                        | string  | true      | Name of the cluster to be cloned. Random value if the cluster does not exist locally.  |
| [pgbackrest](#pgbackrest)      | object  | false     | Enables the definition of a pgbackrest-setup for the cluster |

[back to parent](#spec)

---

#### volumeSource

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| emptyDir                       | string  | false      | emptyDir: {} |
| [PersistentVolumeClaim](#volumeSource-PersistentVolumeClaim) | map   | false      | PersistentVolumeClaim-Objekt |
| [configMap](#volumeSource-configMap)             | map   | false      | configMap-Objekt |

[back to parent](#additionalVolumes)

---

#### volumeSource-PersistentVolumeClaim

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| claimName                      | string  | true      | Name of the PersistentVolumeClaim  |
| readyOnly                      | boolean | false     | PersistentVolumeClaim-Objekt |

[back to parent](#volumeSource)

---

#### volumeSource-configMap

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| name                           | string  | true      | Name of the Configmap  |

[back to parent](#volumeSource)

---

#### pgbackrest

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| [configuration](#configuration)| object  | false     | Enables the definition of a pgbackrest-setup for the cluster |
| global                         | object  | false     |  |
| image                          | string  | true      |  |
| [repos](#repos)                | array   | true      |  |
| resources:                     | object  | false     | Resource definition (limits.cpu, limits.memory & requests.cpu & requests.memory)|

[back to parent](#backup)

---

#### configuration

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| secret                         | object  | false     | Secretname with the contained S3 credentials (AccessKey & SecretAccessKey) (Note: must be placed in the same namespace as the cluster) |
| [protection](#protection)      | object  | false     | Enable Protection-Options |

[back to parent](#pgbackrest)

---

#### protection

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| restore                        | boolean | false     | A restore is ignored as long as this option is set to true. |

[back to parent](#configuration)

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

[back to parent](#pgbackrest)

---

#### schedule

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| full                           | string  | false     | (Cron-Syntax) 	Define full backup |
| incr                           | string  | false     | (Cron-Syntax) 	Define incremental backup |
| diff                           | string  | false     | (Cron-Syntax) 	Define differential backup |

[back to parent](#pgbackrest)

---

#### status

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| PostgresClusterStatus          | string  | false     | Shows the cluster status. Filled by the Operator |

[back to parent](#postgresql)
