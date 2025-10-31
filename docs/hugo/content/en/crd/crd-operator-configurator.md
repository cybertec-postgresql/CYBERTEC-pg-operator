---
title: "Operator-Configuration"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 332
---

| Name                             | Type    | default  | Description        |
| -------------------------------- |:-------:| --------:| ------------------:|
| enable_crd_registration          | boolean | true     |  |
| crd_categories                   | string  | all      |  |
| enable_lazy_spilo_upgrade        | boolean | false    |  |
| enable_pgversion_env_var         | boolean | true     |  |
| enable_spilo_wal_path_combat     | boolean | false    |  |
| etcd_host                        | string  |          |  |
| kubernetes_use_configmaps        | boolean | false    |  |
| docker_image                     | string  |          |  |
| sidecars                         | list    |          |  |
| enable_shm_volume                | boolean | true     |  |
| workers                          | int     | 8        |  |
| max_instances                    | int     | -1       |  |
| min_instances                    | int     | -1       |  |
| resync_period                    | string  | 30m      |  |
| repair_period                    | string  |  5m      |  |
| set_memory_request_to_limit      | boolean | false    |  |
| debug_logging                    | boolean | true     |  |
| enable_db_access                 | boolean | true     |  |
| spilo_privileged                 | boolean | false    |  |
| spilo_allow_privilege_escalation | boolean | true     |  |
| watched_namespace                | string  | *        |  |

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