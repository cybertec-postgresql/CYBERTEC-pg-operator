---
title: "Operator-Configuration"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 503
---

# Configuring the PostgreSQL Operator

The PostgreSQL Operator is configured based on the custom resource type **`OperatorConfiguration`**.
This resource allows you to control the behaviour of the operator in detail and adapt it to individual requirements.

The supplied **Helm chart** already contains a **default configuration** that is suitable for most use cases.
These default values cover typical operating requirements and enable a quick start without additional adjustments.

The assignment to OperatorConfiguration is based on the ENV section in the operator deployment:
```yaml
  containers:
    - name: postgres-operator
      env:
        - name: POSTGRES_OPERATOR_CONFIGURATION_OBJECT
          value: postgresql-operator-configuration
```

In addition, the `OperatorConfiguration` offers a wide range of options for specifically influencing the behaviour of the operator.
Among other things, the following aspects can be configured:

## Behaviour for major upgrades
The operator allows you to configure the behaviour during major upgrades using the following fields:
```yaml
    major_version_upgrade_mode: manual
    minimal_major_version: '13'
    target_major_version: '18'
```
### Explanation of parameters

- **major_version_upgrade_mode**
Controls how major upgrades are performed:

  - `manual`: The upgrade is triggered **manually** via the cluster manifest. 
  - `off`: Upgrades via Operator are disabled.
  - `full`: The operator compares the version in the manifest with the configured `minimal_major_version`. If the version is lower, the operator starts an **automatic upgrade** to the configured `target_major_version`.

- **minimal_major_version** 
Specifies The minimum Postgres major version that will not be automatically upgraded (only relevant in `full` mode).

- **target_major_version** 
The version to which the automatic upgrade should update (only relevant in `full` mode).
 

## Readiness and liveness probes
The operator allows health checks to be configured using the following fields:

```yaml
enable_readiness_probe: true
enable_liveness_probe: false
```
### Explanation of parameters

- **enable_readiness_probe**
Specifies whether the readiness probe definition should be added to the container.

- **enable_liveness_probe**
Specifies whether the liveness probe definition should be added to the container.

## SecurityContext settings
The operator allows the configuration of the **SecurityContext** for the PostgreSQL containers via the following fields:

```yaml
  spilo_privileged: false
  spilo_allow_privilege_escalation: false
  container_readonly_root_filesystem: true
```
### Explanation of parameters

- **spilo_privileged**
Specifies whether the container should run in **privileged mode**.
  - `true`: Privileged mode enabled
  - `false`: Privileged mode disabled (recommended for production environments)

- **container_readonly_root_filesystem**
Enables a **read-only root filesystem** to increase security.
  - `true`: Root filesystem is read-only
  - `false`: Write access to root filesystem allowed

- **spilo_allow_privilege_escalation** 
Specifies whether the container is allowed to **escalate privileges**.
  - `true`: Privilege escalation allowed
  - `false`: Privilege escalation disabled (security-friendly)

## Connection pooler configuration
The operator enables detailed configuration of the **connection pooler** via the following fields:

```yaml
  connection_pooler:
    connection_pooler_user: pooler
    connection_pooler_default_memory_request: 100Mi
    connection_pooler_max_db_connections: 60
    connection_pooler_default_cpu_request: 500m
    connection_pooler_image: 'docker.io/cybertecpostgresql/cybertec-pg-container:pgbouncer-1.24.1-4'
    connection_pooler_default_memory_limit: 100Mi
    connection_pooler_default_cpu_limit: '1'
    connection_pooler_schema: pooler
    connection_pooler_number_of_instances: 2
    connection_pooler_mode: transaction
```
### Explanation of parameters

- **connection_pooler_user**
Username for the pooler role in the database.

- **connection_pooler_default_memory_request / connection_pooler_default_cpu_request**
Resource request for the pooler container (memory and CPU).

- **connection_pooler_default_memory_limit / connection_pooler_default_cpu_limit**
Resource limits for the pooler container (memory and CPU).

- **connection_pooler_max_db_connections**
Maximum number of simultaneous connections that the pooler creates to PostgreSQL.

- **connection_pooler_image**
Container image for the pooler.

- **connection_pooler_schema**
Database schema used by the pooler.

- **connection_pooler_number_of_instances**
Number of pooler pods.

- **connection_pooler_mode**
Operating mode of the pooler:
- `transaction`: Pooler manages connections per transaction
- `session`: Pooler manages connections per session$
- `statement`: Pooler manages connections per statement


## Debug options
*(Here you can explain which debug features are available and how they are activated)*
```yaml
  debug:
    debug_logging: true
    enable_database_access: true
```
### Explanation of parameters

- **debug_logging**
Enables or disables debug output in the operator log

- **enable_database_access**
Defines whether the operator is permitted to access the database, for example to create users or databases

The complete structure and description of all available parameters is documented in the [OperatorConfiguration](crd/crd-operator-configurator).
