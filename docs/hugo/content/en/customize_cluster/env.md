---
title: "Environment variables"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 1
---

To flexibly manage containers within a cluster, the operator allows environment variables to be defined at various levels. This enables both global settings and specific configurations for individual components.
Hierarchy and Scope
The variables are defined within the Custom Resource (CR). The following logic applies for inheritance and assignment:

| object | Scope | Description |
| :--- | :--- | :--- |
| `spec.env` | **Global** | These ENVs are inherited by **all** containers within the cluster (PostgreSQL, Backup, Monitoring, etc.). |
| `spec.postgresql.env` | **PostgreSQL** | These ENVs apply exclusively to the **PostgreSQL containers**. |
| `spec.backup.pgbackrest.env` | **pgBackRest** | These ENVs apply exclusively to the **Backup containers**. |
| `spec.monitor.env` | **Exporter-Sidecar** | These ENVs apply exclusively to the **ConnectionPooler containers**. |
| `spec.connectionPooler.env` | **ConnectionPooler** | These ENVs apply exclusively to the **Monitoring sidecars**. |

{{< hint type=Warning >}}Updating the ENVs triggers a rolling update to the respective containers.{{< /hint >}}


### Configuration Logic 

The definition of variables follows the standard Kubernetes schema for key-value pairs. 

```yaml 
env: 
  - name: ENV_NAME 
    value: ‘value’
```