---
title: "Custom Labels"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 2
---

To manage and organise pods flexibly within a cluster, the operator allows labels to be defined at various levels. This enables both global labelling and specific metadata for individual components. Unlike environment variables, labels always refer to the pod as a whole, not to individual containers.

| object | Scope | Description |
| :--- | :--- | :--- |
| `spec.labels` | **Global** | These labels are adopted by **all** pods within the cluster (**PostgreSQL**, **Backup**, **Pooler**, etc.). |
| `spec.postgresql.labels` | **PostgreSQL** | These labels apply exclusively to the PostgreSQL pods. **PostgreSQL pods**. |
| `spec.backup.pgbackrest.labels` | **pgBackRest** | These labels apply exclusively to the backup pods **pgBackRest pods**. |
| `spec.connectionPooler.labels` | **ConnectionPooler** | These labels apply exclusively to the  **ConnectionPooler pods**. |

{{< hint type=Warning >}}Updating the labels triggers a rolling update to the respective pods.{{< /hint >}}


### Configuration Logic 

The definition of labels follows the standard Kubernetes schema for key-value pairs. 

```yaml 
labels: 
  custom_label: ‘value’
```