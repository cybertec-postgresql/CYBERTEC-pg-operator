---
title: "Operator-Configuration"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 332
---

#### CRD for kind OperatorConfiguration

| Name        | Type                | required  | Description        |
| ----------- |:-------------------:| ---------:| ------------------:|
| apiVersion  | string              | true      | cpo.opensource.cybertec.at/v1 |
| kind        | string              | true      | OperatorConfiguration                  |
| metadata    | object              | true      |                    |
| [configuration](#configuration) | object         | true      |                    |

---

#### configuration

| Name                                              | Type          | default   | Description        |
| ------------------------------------------------- |:-------------:| ---------:| ------------------:|
| [kubernetes](#kubernetes)                         | object        |           |                    |
| [users](#users)                                   | object        |           |                    |
| [connection_pooler](#connection_pooler)           | object        |           |                    |
| [logging_rest_api](#logging_rest_api)             | object        |           |                    |
| [load_balancer](#load_balancer)                   | object        |           |                    |
| [major_version_upgrade](#major_version_upgrade)   | object        |           |                    |
| [teams_api](#teams_api)                           | object        |           |                    |
| [timeouts](#timeouts)                             | object        |           |                    |
| [debug](#debug)                                   | object        |           |                    |
| [logical_backup](#logical_backup)                 | object        |           |                    |
| [aws_or_gcp](#aws_or_gcp)                         | object        |           |                    |
| [sidecars](#sidecars)                             | list          |           | Each item is of type [Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#container-v1-core) |
| docker_image                                      | string        |           |                    |
| enable_crd_registration                           | boolean       | `true`    | True, Operator updates the crd itself |
| enable_crd_validation                             | boolean       | `true `   | deprecated         |
| enable_lazy_spilo_upgrade                         | boolean       | `false `  | If true, update statefulset with new images without rolling update. |
| enable_pgversion_env_var                          | boolean       | `true `   | Set PGVersion via ENV-Label. Changes can create issues |
| enable_shm_volume                                 | boolean       | `true`    | True adds tmpfs-Volume to remove shm memory-limitations |
| enable_spilo_wal_path_compat                      | boolean       | `false`   |                    |
| enable_team_id_clustername_prefix                 | boolean       | false     |                    |
| etcd_host                                         | string        |           | Only required if the Kubernetes-native approach is not used. |
| kubernetes_use_configmaps                         | boolean       | true      | Recommended! Uses configmaps for Patroni instead of entrypoints. |
| max_instances                                     | int           | -1        | Maximum number of Postgres pods per cluster. |
| min_instances                                     | int           | -1        | Minimal number of Postgres pods per cluster. |
| postgres_pod_resources                            | string        | true      |                    |
| repair_period                                     | string        | 5m        | Period between subsequent repair requests |
| resync_period                                     | string        | 30m       | Period between subsequent resync requests |
| set_memory_request_to_limit                       | boolean       | false     |                    |
| workers                                           | int           | 8         | Number of workers in the operator that simultaneously process tasks such as create/update/delete clusters |

{{< back >}}

---

#### kubernetes

| Name                                          | Type          | default     | Description        |
| --------------------------------------------- |:-------------:| -----------:| ------------------:|
| cluster_labels                                | map           |             | a map of key-value pairs adding labels |
| cluster_domain                                | string        | `cluster.local` | DNS domain used inside the K8s-Cluster. Used by the operator to communicate with clusters |
| cluster_name_label                            | string        | `cluster.cpo.opensource.cybertec.at/name` | Label to identify all resources of a cluster |
| container_readonly_root_filesystem            | boolean       | `false`     | Enables ReadOnlyRootFilesystem in the SecurityContext of the pods |
| enable_cross_namespace_secret                 | boolean       | `false`     | Enables the storage of secrets in another namespace, provided that it is activated. The namespace is defined in the cluster manifest. |
| enable_init_containers                        | boolean       | `true`      | Allows the definition of init containers in the cluster manifest |
| enable_pod_antiaffinity                       | boolean       | `true`      | The pod anti-affinity rules are applied when activated. |
| enable_pod_disruption_budget                  | boolean       | `true`      | Pod Disruption Budgets (PDB) are generated for clusters when activated. |
| enable_readiness_probe                        | boolean       | `true`      | Operator adds readiness probe for resources when enabled |
| enable_liveness _probe                        | boolean       | `false`     | Operator adds liveness probe for resources when enabled |
| enable_sidecars                               | boolean       | `true`      | Allows the definition of sidecars in the cluster manifest |
| inherited_labels                              | list          |             | Labels added to each resource |
| master_pod_move_timeout                       | string        | `20m`       | Timeout for waiting for a primary pod to switch to another Kubernetes node. |
| oauth_token_secret_name                       | string        | `postgresql-operator` |                    |
| pdb_name_format                               | string        | `postgres-{cluster}-pdb` | Naming scheme for generated pod disruption budgets (PDB) |
| pod_management_policy                         | string        | `ordered_ready` | Pod-Management-Strategy for the [statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) |
| pod_antiaffinity_topology_key                 | string        | `kubernetes.io/hostname` | Defines the anti-affinity [topology Key](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) |
| pod_antiaffinity_preferred_during_scheduling  | boolean       | `false`     |                    |
| pod_role_label                                | string        | `member.cpo.opensource.cybertec.at/role` | Defines the label for the pod-role|
| pod_service_account_definition                | string        | `''`        |                    |
| pod_service_account_name                      | string        | `cpo-pod`   | ServiceAccount used for all cluster-pods |
| pod_service_account_role_binding_definition   | string        | `''`        |                    |
| pod_terminate_grace_period                    | string        | `5m`        |                    |
| secret_name_template                          | string        | `{username}.{cluster}.credentials.{tprkind}.{tprgroup}` |                    |
| share_pgsocket_with_sidecars                  | boolean       | `false`     |                    |
| spilo_allow_privilege_escalation              | boolean       | `false`     | Defines privilege-escalation attribut in SecurityContext |
| spilo_privileged                              | boolean       | `false`     | Defines privileged attribut in SecurityContext |
| storage_resize_mode                           | string        | `pvc`       |                    |
| watched_namespace                             | string        | `*`         | Operator watches for Objects in the defined Namespace. `*` means all, `` means only operator-namespace, `NAMESPACE_NAME` means specific namespace |


{{< back >}}

---

#### users

| Name                                          | Type          | default   | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| enable_password_rotation                      | boolean       | `false`   | password rotation by the Operator for all Login Roles excluding DB_Owner                   |
| password_rotation_interval                    | int           | `90`      | Interval in days   |
| password_rotation_user_retention              | int           | `180`     | To avoid a constantly growing number of new users due to password rotation, the operator deletes the created users after a certain number of days. The number can be configured with this parameter. However, the operator checks whether the retention policy is at least twice as long as the rotation interval and updates it to this minimum if this is not the case.                  |
| replication_username                          | string        | `cpo_replication` | Name for the replication-user| 
| super_username                                | string        | `postgres` | Name for the Superuser. Changes can create issues | 

{{< back >}}

---

#### connection_pooler

| Name                                          | Type          | default       | Description            |
| --------------------------------------------- |:-------------:| -------------:| ----------------------:|
| connection_pooler_default_cpu_request         | int           | `500m`        | CPU-Request for Pod    |
| connection_pooler_default_cpu_limit           | string        | `1`           | CPU-Limit for Pod      |
| connection_pooler_default_memory_request      | string        | `100Mi`       | Memory-Request for Pod |
| connection_pooler_default_memory_limit        | string        | `100Mi`       | Memory-Limit for Pod   |
| connection_pooler_image                       | string        |               | Container-Image        | 
| connection_pooler_max_db_connections          | int           | `60`          | Max Connections between DB and Pooler. Divided by the `connection_pooler_number_of_instances` |
| connection_pooler_mode                        | string        | `transaction` | Pooler mode | 
| connection_pooler_number_of_instances         | int           | `2`           | Number of Instances    | 
| connection_pooler_schema                      | string        | `pooler`      | Schema to create needed Objects like lookup function | 
| connection_pooler_user                        | int           | `pooler`      | Database-User for pooler |

{{< back >}}

---

#### logging_rest_api

| Name                                          | Type          | default   | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| api_port                                      | int           | `8080`    | REST-API port      |
| cluster_history_entries                       | int           | `1000`    | Number of lines used to store cluster logs.                   | 
| ring_log_lines                                | int           | `100`     | number of entries  | 

{{< back >}}

---

#### load_balancer

| Name                                          | Type          | default          | Description        |
| --------------------------------------------- |:-------------:| ----------------:| ------------------:|
| db_hosted_zone                                | string        | `db.example.com` | DNS-Definition for the Cluster DNS |
| enable_master_load_balancer                   | boolean       | `false`          | Creates loadbalancer service for the primary pod, if enabled |
| enable_master_pooler_load_balancer            | boolean       | `false`          | Creates loadbalancer service for the primary pooler, if enabled | 
| enable_replica_load_balancer                  | boolean       | `false`          | Creates loadbalancer service for the replica pods, if enabled | 
| enable_replica_pooler_load_balancer           | boolean       | `false`          | Creates loadbalancer service for the replica pooler, if enabled | 
| external_traffic_policy                       | string        | `Cluster`        | Defines traffic policy for loadbalancers. Possible Values: `Cluster`, `local`| 
| master_dns_name_format                        | string        | `{cluster}.{namespace}.{hostedzone}` | DNS-Format for the primary loadbalancer | 
| replica_dns_name_format                       | string        | `{cluster}-repl.{namespace}.{hostedzone}` | DNS-Format for the replica loadbalancer | 
| master_legacy_dns_name_format                 | string        | `{cluster}.{team}.{hostedzone}`      | deprecated | 
| replica_legacy_dns_name_format                | string        | `{cluster}-repl.{team}.{hostedzone}` | deprecated | 

{{< back >}}

---

#### major_version_upgrade

| Name                                          | Type          | default   | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| major_version_upgrade_mode                    | string        | `manual`  | Mode for Major-Upgrades. `manual` Upgrade triggert bei `PGVERSION`-defintion in Cluster-Manifest, `full` Upgrade triggert by the operator based on `target_major_version`, `off` The operator never triggers an upgrade. |
| minimal_major_version                         | string        | `13`      | The minimum Postgres major version that will not be automatically `updated when major_version_upgrade_mode = full` | 
| target_major_version                          | string        | `18`      | Target Postgres Major if the upgrade is triggered automatically via
 `updated when major_version_upgrade_mode = full` | 

{{< back >}}

---

#### teams_api

| Name                                          | Type          | default    | Description         |
| --------------------------------------------- |:-------------:| ----------:| ------------------:|
| enable_team_superuser                         | boolean       | `false`    |                    |
| teams_api_url                                 | string        | `https://teams.example.com/api/` |                    |
| team_admin_role                               | string        | `admin`    |                    | 
| enable_postgres_team_crd_superusers           | boolean       | `false`    |                    | 
| protected_role_names                          | list          |            |                    | 
| pam_role_name                                 | string        | `cpo_pam`  |                    | 
| pam_configuration                             | string        | `https://info.example.com/oauth2/tokeninfo?access_token= uid realm=/employees` |                    | 
| team_api_role_configuration                   | map           |            | a map of key-value pairs adding labels |
| enable_teams_api                              | boolean       | `false`    |                    | 
| enable_team_member_deprecation                | boolean       | `false`    |                    | 
| enable_admin_role_for_users                   | boolean       | `false`    |                    | 
| role_deletion_suffix                          | string        | `_deleted` |                    | 
| enable_postgres_team_crd                      | boolean       | `false`    |                    | 

{{< back >}}

---

#### timeouts

| Name                                          | Type          | default   | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| patroni_api_check_interval                    | string        | `1s`      |                    |
| patroni_api_check_timeout                     | string        | `5s`      |                    |
| pod_deletion_wait_timeout                     | string        | `10m`     |                    | 
| pod_label_wait_timeout                        | string        | `10m`     |                    | 
| ready_wait_interval                           | string        | `4s`      |                    | 
| ready_wait_timeout                            | string        | `30s`     |                    | 
| resource_check_interval                       | string        | `3s`      |                    | 
| resource_check_timeout                        | string        | `10m`     |                    | 

{{< back >}}

---

#### debug

| Name                                          | Type           | default   | Description        |
| --------------------------------------------- |:--------------:| ---------:| ------------------:|
| debug_logging                                 | boolean        | `true`    | Enable Debug-Logs  |
| enable_database_access                        | boolean        | `true`    | Allows the Operator to connect to the database (to create users and for other actions) | 

{{< back >}}

---

#### logical_backup (deprecated)

| Name                                          | Type          | default       | Description        |
| --------------------------------------------- |:-------------:| -------------:| ------------------:|
| logical_backup_docker_image                   | string        |               | deprecated         |
| logical_backup_job_prefix                     | string        | `logical-backup-` | deprecated     |
| logical_backup_provider                       | string        | `s3`          | deprecated         | 
| logical_backup_s3_bucket                      | string        | `my-bucket-url` | deprecated       | 
| logical_backup_s3_sse                         | string        | `AES256`      | deprecated         | 
| logical_backup_schedule                       | string        | `30 00 * * *` | deprecated         | 

{{< back >}}

---

#### aws_or_gcp

| Name                                          | Type          | default        | Description        |
| --------------------------------------------- |:-------------:| --------------:| ------------------:|
| additional_secret_mount_path                  | string        | `/meta/credentials` |                    |
| aws_region                                    | string        | `eu-central-1` |                    |
| enable_ebs_gp3_migration                      | boolean       | `false`        |                    | 
| enable_ebs_gp3_migration_max_size             | int           | `1000`         |                    | 

{{< back >}}

---