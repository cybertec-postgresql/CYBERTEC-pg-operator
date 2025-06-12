---
title: "The Project"
date: 2024-03-11T14:26:51+01:00
draft: false
weight: 201
---
The CYBERTEC PostgreSQL Operator (CPO) enables the simple provision and management of PostgreSQL clusters on Kubernetes. It reduces the administration effort and facilitates the management of single-node and HA clusters.
## Main components
- [CYBERTEC-pg-operator](https://github.com/cybertec-postgresql/CYBERTEC-pg-operator): Kubernetes operator for the automation of PostgreSQL clusters.
- [CYBERTEC-pg-container](https://github.com/cybertec-postgresql/CYBERTEC-pg-container): Docker container suite for PostgreSQL, Patroni and etcd for the provision of HA clusters.
- [CYBERTEC-operator-tutorials](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials): Tutorials and instructions for installing and using the operator.
## Features
- Cluster management:
    - Single-node and HA (High Availability) clusters via [Patroni](https://patroni.readthedocs.io/en/latest/)
    - Reduction of downtime thanks to redundancy, pod anti-affinity, auto-failover and self-healing
    - Automated failover
    - Live volume resize without pod restarts
    - Basic credential and user management on K8s, eases application deployments
    - Compatible with OpenShift and Rancher
- PostgreSQL compatibility:
    - Supports PostgreSQL versions 13 to 17
    - Inplace upgrades for smooth version changes and minimal downtime
    - Extensive extension support, including pgAudit, TimescaleDB and PostGIS
    - Standby-Cluster
- Backup & Restore:
    - Integrated pgBackRest support
    - Automatic backups
    - Point-in-Time- and Snapshot-based Restores / Disaster Recovery
- Connection management:
    - pgBouncer for connection pooling
- Monitoring & alerting stack
    - Integrated metrics exporter
    - Prometheus, alert manager for metrics collection and alerting
    - Grafana for visual monitoring of the clusters
- Operator UI:
    - Web interface for managing clusters

## Installation
Detailed instructions on installation and configuration can be found in the CYBERTEC operator tutorials and in the following chapters
Example of installation via Helm:
```
helm repo add cybertec https://cybertec-postgresql.github.io/helm-charts/
helm install pg-operator cybertec/cybertec-pg-operator
```

More information: [Installation]({{< relref "installation/install_operator" >}})

## Contribution
This project is open source, and contributions to its further development are expressly encouraged.
Possible forms of contribution:
- Bug reports and feature requests
- Code contributions (pull requests welcome)
- Improvement of the documentation
Further details on contributions can be found in the respective GitHub repositories.
## Licence
The CYBERTEC PostgreSQL Operator is licensed under the Apache 2.0 licence.