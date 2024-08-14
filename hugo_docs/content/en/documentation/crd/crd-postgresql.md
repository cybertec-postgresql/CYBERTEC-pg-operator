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
| [backup](#backup)              | object  | false     | Enables the definition of a customised backup solution for the cluster |
| teamId                         | string  | true      | name of the team the cluster belongs to |
| numberOfInstances              | Int     | true      | Number of nodes of the cluster |
| dockerImages                   | string  | false     | Define a custom image to override the default |
| schedulerName                  | string  | false     | Define a custom Name to override the default |
| spiloRunAsUser                 | string  | false     | Define an User id which should be used for the pods |
| spiloRunAsGroup                | string  | false     | Define an Group id which should be used for the pods |
| spiloFSGroup                   | string  | false     | Persistent Volumes for the pods in the StatefulSet will be owned and writable by the group ID specified. |
| enableMasterLoadBalancer       | boolean | false     | Define whether to enable the load balancer pointing to the Postgres primary |
| enableMasterPoolerLoadBalancer | boolean | false     | Define whether to enable the load balancer pointing to the primary ConnectionPooler |
| enableReplicaLoadBalancer      | boolean | false     | Define whether to enable the load balancer pointing to the Postgres replicas |
| enableReplicaPoolerLoadBalancer| boolean | false     | Define whether to enable the load balancer pointing to the Replica-ConnectionPooler |
| allowedSourceRange             | string  | false     | Defines the range of IP networks (in CIDR-notation). The corresponding load balancer is accessible only to the networks defined by this parameter. |
| users                          | map     | false     | a map of usernames to user flags for the users that should be created in the cluster by the operator |
| usersWithSecretRotation        | list    | false     | list of users to enable credential rotation in K8s secrets. The rotation interval can only be configured globally. |
| usersWithInPlaceSecretRotation | list    | false     | list of users to enable in-place password rotation in K8s secrets. The rotation interval can only be configured globally. |
| databases                      | map     | false     | a map of databases that should be created in the cluster by the operator |
| tolerations                    | list    | false     | a list of tolerations that apply to the cluster pods. Each element of that list is a dictionary with the following fields: 
key, operator, value, effect and tolerationSeconds |
| podPriorityClassName           | string  | false     | a name of the priority class that should be assigned to the cluster pods. If not set then the default priority class is taken. The priority class itself must be defined in advance |
| podAnnotations                 | map     | false     | A map of key value pairs that gets attached as annotations to each pod created for the database. |
| ServiceAnnotations             | map     | false     | A map of key value pairs that gets attached as annotations to each Service created for the database. |
| enableShmVolume                | boolean | false     | Start a database pod without limitations on shm memory. By default Docker limit /dev/shm to 64M (see e.g. the docker issue, which could be not enough if PostgreSQL uses parallel workers heavily. If this option is present and value is true, to the target database pod will be mounted a new tmpfs volume to remove this limitation. |
| enableConnectionPooler         | boolean | false     | creates a ConnectionPooler for the primary Database |
| enableReplicaConnectionPooler  | boolean | false     | creates a ConnectionPooler for the replica Databases |
| enableLogicalBackup            | boolean | false     | Enable logical Backups for this Cluster (Stored on S3) - s3-configuration for Operator is needed (Not for pgBackRest) |
| logicalBackupSchedule          | string  | false     | Schedule for the logical backup K8s cron job.  (Not for pgBackRest) |
| additionalVolumes              | list    | false     | List of additional volumes to mount in each container of the statefulset pod. Each item must contain a name, mountPath, and volumeSource which is a kubernetes volumeSource. It allows you to mount existing PersistentVolumeClaims, ConfigMaps and Secrets inside the StatefulSet. |

[back to parent](#postgresql)

---

#### backup

| Name                           | Type    | required  | Description        |
| ------------------------------ |:-------:| ---------:| ------------------:|
| [pgbackrest](#pgbackrest)      | object  | false     | Enables the definition of a pgbackrest-setup for the cluster |

[back to parent](#spec)

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
