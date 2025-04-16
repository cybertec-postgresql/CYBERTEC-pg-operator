
# CYBERTEC PG Operator (CPO)

**The CYBERTEC PG Operator (CPO)** is a powerful Kubernetes operator that dramatically simplifies the creation and management of highly available **PostgreSQL clusters**. 
Fully integrated with GitOps/CI/CD workflows and infrastructure-as-code principles, CPO enables consistent, secure and automated database provisioning - **without direct access to the Kubernetes API**.

<img src="docs/diagrams/cpo_logo.svg" width="350">

---

## Highlights

- **Completely declarative configuration** via custom resources
- Highly available PostgreSQL clusters** with [Patroni](https://github.com/zalando/patroni)
- Seamless integration into CI/CD pipelines** (e.g. ArgoCD, Flux)
- Compatible with OpenShift**
- **Support for cloud & on-prem environments**

---

### Operator features

- Rolling updates for cluster changes & minor version upgrades
- Live Volume Resize (without pod restarts if supported by storage)
- Database Connection Pooling via **pgBouncer**
- In-place major upgrades of all clusters (fast & secure)
- Backup & restore to PVC, AWS, GCS and Azure
- Client-side backup encryption
- Support for standby clusters & multi-site topologies
- User & credential management at K8s level
- Support for own TLS certificates
- **TDE** integration with **[CYBERTEC PGEE](https://www.cybertec-postgresql.com/en/products/cybertec-postgresql-enterprise-edition/)**
- Migration from AWS EBS `gp2` to `gp3` with IOPS and throughput config

---

### Cloud native architecture

The CYBERTEC PG Operator is designed from the ground up with a **cloud-native approach**:

- **Declarative configuration** via Kubernetes CRDs - completely in the spirit of *Infrastructure as Code*.
- **Self-healing and automation** through Kubernetes and [Patroni](https://github.com/zalando/patroni), including automatic failover, leader election and rolling updates.
- CI/CD-friendly**: No direct access to the Kubernetes API required - ideal for GitOps workflows and automated deployments.
- Platform-independent**: Runs on any Kubernetes-compatible infrastructure - whether public cloud, on-prem or hybrid.
- API-driven control**: Patroni provides a REST API to query the cluster state and trigger failover - essential for dynamic, service-oriented architectures.

This architecture forms the basis for a modern, highly available and scalable PostgreSQL platform in the cloud era.

---

## PostgreSQL features

- PostgreSQL 13 to 17
- Streaming replication via **Patroni**
- Fully integrated backup & PITR with `pgBackRest` or `pg_basebackup`
- Extensions like:
- [PostGIS](https://postgis.net/)
- [pgAudit](https://github.com/pgaudit/pgaudit)
- [TimescaleDB](https://github.com/timescale/timescaledb)
- [pg_cron](https://github.com/citusdata/pg_cron)
- [credcheck](https://github.com/MigOpsRepos/credcheck)
- [set_user](https://github.com/pgaudit/set_user)
- Minimal downtime during upgrades thanks to rolling updates and failover mechanisms
- Self-healing, redundancy and pod anti-affinity for maximum availability

---

## Compatibility

| Release | PostgreSQL | pgBackRest | Patroni | Kubernetes | Go      |
|---------|------------|------------|---------|------------|---------|
| 0.8.0   | 13 - 17    | 2.53       | 4.0.2   | 1.21+      | 1.21.7  |
| 0.8.3   | 13 - 17    | 2.54-2     | 4.0.5   | 1.21+      | 1.22.12 |

--- 

## Getting Started

Want to get started quickly? This way:

- [Quickstart-Guide (documentation)](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/quickstart/)
- [Tutorials & examples](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials)

---

## Documentation

You can find the complete and searchable documentation here:

- [Official Documentation](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/)

---

## Licence

This project is licensed under the **Apache 2.0 licence**. Further information can be found in the [LICENSE](./LICENSE) file.

---

## participate

Pull requests and feedback are always welcome. Please read our [Contribution Guidelines](CONTRIBUTING.md) in advance if you would like to participate.

---