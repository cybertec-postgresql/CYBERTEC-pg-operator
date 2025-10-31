---
title: "Operator-Configuration"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 332
---

#### CRD for kind OperatorConfiguration

| Name        | Type           | required  | Description        |
| ----------- |:--------------:| ---------:| ------------------:|
| apiVersion  | string         | true      | cpo.opensource.cybertec.at/v1 |
| kind        | string         | true      | OperatorConfiguration                  |
| metadata    | object         | true      |                    |
| [configuration](#configuration)        | object         | true      |                    |

---

#### configuration

| Name                                              | Type          | required  | Description        |
| ------------------------------------------------- |:-------------:| ---------:| ------------------:|
| [kubernetes](#kubernetes)                         | object        | true      |                    |
| [users](#users)                                   | object        | true      |                    |
| [connection_pooler](#connection_pooler)           | object        | true      |                    |
| [logging_rest_api](#logging_rest_api)             | object        | true      |                    |
| [load_balancer](#load_balancer)                   | object        | true      |                    |
| [major_version_upgrade](#major_version_upgrade)   | object        | true      |                    |
| [teams_api](#teams_api)                           | object        | true      |                    |
| [timeouts](#timeouts)                             | object        | true      |                    |
| [debug](#debug)                                   | object        | true      |                    |
| [logical_backup](#logical_backup)                 | object        | true      |                    |
| [aws_or_gcp](#aws_or_gcp)                         | object        | true      |                    |
| docker_image                                      | string        | true      |                    |
| enable_crd_registration                           | boolean       | true      |                    |
| enable_crd_validation                             | boolean       | true      |                    |
| enable_lazy_spilo_upgrade                         | boolean       | true      |                    |
| enable_pgversion_env_var                          | string        | true      |                    |
| enable_shm_volume                                 | boolean       | true      |                    |
| enable_spilo_wal_path_compat                      | string        | true      |                    |
| enable_team_id_clustername_prefix                 | boolean       | true      |                    |
| etcd_host                                         | string        | true      |                    |
| kubernetes_use_configmaps                         | boolean       | true      |                    |
| max_instances                                     | int           | true      |                    |
| min_instances                                     | int           | true      |                    |
| postgres_pod_resources                            | string        | true      |                    |
| repair_period                                     | string        | true      |                    |
| resync_period                                     | string        | true      |                    |
| set_memory_request_to_limit                       | boolean       | true      |                    |
| workers                                           | int           | true      |                    |

{{< back >}}

---

#### kubernetes

| Name                                          | Type          | required  | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| cluster_labels                                | map           | true      | a map of key-value pairs adding labels |
| cluster_domain                                | string        | true      |                    |
| cluster_name_label                            | string        | true      | default: cluster.cpo.opensource.cybertec.at/name |
| container_readonly_root_filesystem            | boolean       | true      |  
| enable_cross_namespace_secret                 | boolean       | true      |                    |
| enable_init_containers                        | boolean       | true      |                    |
| enable_pod_antiaffinity                       | boolean       | true      |                    |
| enable_pod_disruption_budget                  | boolean       | true      |                    |
| enable_readiness_probe                        | boolean       | true      |                    |
| enable_sidecars                               | boolean       | true      |                    |
| inherited_labels                              | list          | true      |                    |
| master_pod_move_timeout                       | string        | true      |                    |
| oauth_token_secret_name                       | string        | true      |                    |
| pdb_name_format                               | string        | true      |                    |
| pod_management_policy                         | string        | true      |                    |
| pod_antiaffinity_topology_key                 | string        | true      |                    |
| pod_antiaffinity_preferred_during_scheduling  | boolean       | true      |                    |
| pod_role_label                                | string        | true      |                    |
| pod_service_account_definition                | string        | true      |                    |
| pod_service_account_name                      | string        | true      |                    |
| pod_service_account_role_binding_definition   | string        | true      |                    |
| pod_terminate_grace_period                    | string        | true      |                    |
| secret_name_template                          | string        | true      |                    |
| share_pgsocket_with_sidecars                  | boolean       | true      |                    |
| spilo_allow_privilege_escalation              | boolean       | true      |                    |
| spilo_privileged                              | boolean       | true      |                    |
| storage_resize_mode                           | string        | true      |                    |
| watched_namespace                             | string        | true      |                    |


{{< back >}}

---

#### users

| Name                                          | Type          | required  | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| enable_password_rotation                      | boolean       | true      |                    |
| password_rotation_interval                    | int           | true      |                    |
| password_rotation_user_retention              | int           | true      |                    |
| replication_username                          | string        | true      |                    | 
| super_username                                | string        | true      |                    | 

{{< back >}}

---

#### connection_pooler

| Name                                          | Type          | required  | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| connection_pooler_default_cpu_request         | int           | true      |                    |
| connection_pooler_default_cpu_limit           | string        | true      |                    | 
| connection_pooler_default_memory_request      | string        | true      |                    | 
| connection_pooler_default_memory_limit        | string        | true      |                    | 
| connection_pooler_image                       | string        | true      |                    | 
| connection_pooler_max_db_connections          | int           | true      |                    |
| connection_pooler_mode                        | string        | true      |                    | 
| connection_pooler_number_of_instances         | int           | true      |                    | 
| connection_pooler_schema                      | string        | true      |                    | 
| connection_pooler_user                        | int           | true      |                    |

{{< back >}}

---

#### logging_rest_api

| Name                                          | Type          | required  | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| api_port                                      | int           | true      |                    |
| cluster_history_entries                       | int           | true      |                    | 
| ring_log_lines                                | int           | true      |                    | 

{{< back >}}

---

#### load_balancer

{{< back >}}

---

#### major_version_upgrade

| Name                                          | Type          | required  | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| major_version_upgrade_mode                    | string        | true      |                    |
| minimal_major_version                         | string        | true      |                    | 
| target_major_version                          | string        | true      |                    | 

{{< back >}}

---

#### teams_api

{{< back >}}

---

#### timeouts

{{< back >}}

---

#### debug

| Name                                          | Type          | required  | Description        |
| --------------------------------------------- |:-------------:| ---------:| ------------------:|
| debug_logging                                 | boolean        | true      |                    |
| enable_database_access                        | boolean        | true      |                    | 

{{< back >}}

---

#### logical_backup

{{< back >}}

---

#### aws_or_gcp

{{< back >}}

---

---

| Name                              | Type    | default  | Description        |
| --------------------------------- |:-------:| --------:| ------------------:|
| enable_crd_registration           | boolean | true     |  |
| crd_categories                    | string  | all      |  |
| enable_lazy_spilo_upgrade         | boolean | false    |  |
| enable_pgversion_env_var          | boolean | true     |  |
| enable_spilo_wal_path_combat      | boolean | false    |  |
| etcd_host                         | string  |          |  |
| kubernetes_use_configmaps         | boolean | false    |  |
| docker_image                      | string  |          |  |
| sidecars                          | list    |          |  |
| enable_shm_volume                 | boolean | true     |  |
| workers                           | int     | 8        |  |
| max_instances                     | int     | -1       |  |
| min_instances                     | int     | -1       |  |
| resync_period                     | string  | 30m      |  |
| repair_period                     | string  |  5m      |  |
| set_memory_request_to_limit       | boolean | false    |  |
| debug_logging                     | boolean | true     |  |
| enable_db_access                  | boolean | true     |  |
| spilo_privileged                  | boolean | false    |  |
| spilo_allow_privilege_escalation  | boolean | true     |  |
| container_readonly_root_filesystem | boolean | false  |  |
| enable_readiness_probe            | boolean | true     |  |
| enable_liveness_probe             | boolean | false    |  |
| watched_namespace                 | string  | *        |  |

#### major-upgrade-specific

| Name                                  | Type    | default  | Description        |
| ------------------------------------- |:-------:| --------:| ------------------:|
| major_version_upgrade_mode            | string  | off      |  |
| major_version_upgrade_team_allow_list | string  |          |  |
| minimal_major_version                 | string  | 13       |  |
| target_major_version                  | string  | 17       |  |

#### aws-specific

| Name                                  | Type    | default  | Description        |
| ------------------------------------- |:-------:| --------:| ------------------:|
| wal_s3_bucket                         | string  |          |  |
| log_s3_bucket                         | string  |          |  |
| kube_iam_role                         | string  |          |  |
| aws_region                            | string  |          |  |
| additional_secret_mount               | string  |          |  |
| additional_secret_mount_path          | string  |          |  |
| enable_ebs_gp3_migration              | boolean |          |  |
| enable_ebs_gp3_migration_max_size     | int     |          |  |

#### logical-backup-specific

| Name                                  | Type    | default  | Description        |
| ------------------------------------- |:-------:| --------:| ------------------:|
| logical_backup_docker_image           | string  |          |  |
| logical_backup_google_application_credentials | string  |          |  |
| logical_backup_job_prefix             | string  |          |  |
| logical_backup_provider               | string  |          |  |
| logical_backup_s3_access_key_id       | string  |          |  |
| logical_backup_s3_bucket              | string  |          |  |
| logical_backup_s3_endpoint            | string  |          |  |
| logical_backup_s3_region              | string  |          |  |
| logical_backup_s3_secret_access_key   | string  |          |  |
| logical_backup_s3_sse                 | string  |          |  |
| logical_backup_s3_retention_time      | string  |          |  |
| logical_backup_schedule               | string  |          | (Cron-Syntax) |

#### team-api-specific

| Name                                  | Type    | default  | Description        |
| ------------------------------------- |:-------:| --------:| ------------------:|
| enable_teams_api                      | string  |          |  |
| teams_api_url                         | string  |          |  |
| teams_api_role_configuration          | string  |          |  |
| enable_team_superuser                 | boolean |          |  |
| team_admin_role                       | boolean |          |  |
| enable_admin_role_for_users           | boolean |          |  |
| pam_role_name                         | string  |          |  |
| pam_configuration                     | string  |          |  |
| protected_role_names                  | list    |          |  |
| postgres_superuser_teams              | string  |          |  |
| role_deletion_suffix                  | string  |          |  |
| enable_team_member_deprecation        | boolean |          |  |
| enable_postgres_team_crd              | boolean |          |  |
| enable_postgres_team_crd_superusers   | boolean |          |  |
| enable_team_id_clustername_prefix     | boolean |          |  |