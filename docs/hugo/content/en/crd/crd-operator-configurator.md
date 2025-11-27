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
| enable_team_id_clustername_prefix                 | boolean       | true      |                    |
| etcd_host                                         | string        |           |                    |
| kubernetes_use_configmaps                         | boolean       | false     |                    |
| max_instances                                     | int           | -1        |                    |
| min_instances                                     | int           | -1        |                    |
| postgres_pod_resources                            | string        | true      |                    |
| repair_period                                     | string        | 5m        |                    |
| resync_period                                     | string        | 30m       |                    |
| set_memory_request_to_limit                       | boolean       | false     |                    |
| workers                                           | int           | 8         |                    |

{{< back >}}

---

#### kubernetes

| Name                                          | Type          | default     | Description        |
| --------------------------------------------- |:-------------:| -----------:| ------------------:|
| cluster_labels                                | map           |             | a map of key-value pairs adding labels |
| cluster_domain                                | string        | `cluster.local`      |                    |
| cluster_name_label                            | string        | `cluster.cpo.opensource.cybertec.at/name` |  |
| container_readonly_root_filesystem            | boolean       | `false`     |                    |
| enable_cross_namespace_secret                 | boolean       | `false`     |                    |
| enable_init_containers                        | boolean       | `true`      |                    |
| enable_pod_antiaffinity                       | boolean       | `true`      |                    |
| enable_pod_disruption_budget                  | boolean       | `true`      |                    |
| enable_readiness_probe                        | boolean       | `true`      |                    |
| enable_liveness _probe                        | boolean       | `false`     |                    |
| enable_sidecars                               | boolean       | `true`      |                    |
| inherited_labels                              | list          |             |                    |
| master_pod_move_timeout                       | string        | `20m`       |                    |
| oauth_token_secret_name                       | string        | `postgresql-operator` |                    |
| pdb_name_format                               | string        | `postgres-{cluster}-pdb` |                    |
| pod_management_policy                         | string        | `true`      |                    |
| pod_antiaffinity_topology_key                 | string        | `kubernetes.io/hostname` |                    |
| pod_antiaffinity_preferred_during_scheduling  | boolean       | `false`     |                    |
| pod_role_label                                | string        | `member.cpo.opensource.cybertec.at/role` |                    |
| pod_service_account_definition                | string        | `''`        |                    |
| pod_service_account_name                      | string        | `cpo-pod`   |                    |
| pod_service_account_role_binding_definition   | string        | `''`        |                    |
| pod_terminate_grace_period                    | string        | `5m`        |                    |
| secret_name_template                          | string        | `{username}.{cluster}.credentials.{tprkind}.{tprgroup}` |                    |
| share_pgsocket_with_sidecars                  | boolean       | `false`     |                    |
| spilo_allow_privilege_escalation              | boolean       | `false`     |                    |
| spilo_privileged                              | boolean       | `false`     |                    |
| storage_resize_mode                           | string        | `pvc`       |                    |
| watched_namespace                             | string        | `*`         |                    |


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
| db_hosted_zone                                | string        | `db.example.com` |                    |
| enable_master_load_balancer                   | boolean       | `false`          |                    |
| enable_master_pooler_load_balancer            | boolean       | `false`          |                    | 
| enable_replica_load_balancer                  | boolean       | `false`          |                    | 
| enable_replica_pooler_load_balancer           | boolean       | `false`          |                    | 
| external_traffic_policy                       | string        | `Cluster`        |                    | 
| master_dns_name_format                        | string        | `{cluster}.{namespace}.{hostedzone}` |                    | 
| master_legacy_dns_name_format                 | string        | `{cluster}.{team}.{hostedzone}`      |                    | 
| replica_legacy_dns_name_format                | string        | `{cluster}-repl.{team}.{hostedzone}` |                    | 
| replica_dns_name_format                       | string        | `{cluster}-repl.{namespace}.{hostedzone}` |                    | 

{{< back >}}

---

#### major_version_upgrade

| Name                                          | Type          | default   | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| major_version_upgrade_mode                    | string        | `manual`  |                    |
| minimal_major_version                         | string        | `13`      |                    | 
| target_major_version                          | string        | `18`      |                    | 

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
| debug_logging                                 | boolean        | `true`    |                    |
| enable_database_access                        | boolean        | `true`    |                    | 

{{< back >}}

---

#### logical_backup (deprecated)

| Name                                          | Type          | default       | Description        |
| --------------------------------------------- |:-------------:| -------------:| ------------------:|
| logical_backup_docker_image                   | string        |               |                    |
| logical_backup_job_prefix                     | string        | `logical-backup-` |                    |
| logical_backup_provider                       | string        | `s3`          |                    | 
| logical_backup_s3_bucket                      | string        | `my-bucket-url` |                    | 
| logical_backup_s3_sse                         | string        | `AES256`      |                    | 
| logical_backup_schedule                       | string        | `30 00 * * *` |                    | 

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