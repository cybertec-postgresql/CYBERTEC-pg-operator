
# CYBERTEC PG Operator

CPO (CYBERTEC PG Operator) allows you to create and run PostgreSQL clusters on Kubernetes.

The operator reduces your efforts and simplifies the administration of your PostgreSQL clusters so that you can concentrate on other things.
<img src="docs/diagrams/cpo_logo.svg" width="350">

The Postgres Operator delivers an easy to run highly-available [PostgreSQL](https://www.postgresql.org/)
clusters on Kubernetes (K8s) powered by [Patroni](https://github.com/zalando/patroni).
It is configured only through Postgres manifests (CRDs) to ease integration into automated CI/CD
pipelines with no access to Kubernetes API directly, promoting infrastructure as code vs manual operations.

### Operator features

* Rolling updates on Postgres cluster changes, incl. quick minor version updates
* Live volume resize without pod restarts if supported by the storage-system (PVC)
* Database connection pooling with pgBouncer
* Support fast in place major version upgrade. Supports global upgrade of all clusters.
* Restore and cloning Postgres clusters on PVC, AWS, GCS and Azure
* Client-side backup encryption
* Standby cluster
* Patroni-Multisite-Cluster with automatic failover
* Configurable for non-cloud environments
* Basic credential and user management on K8s, eases application deployments
* Support for custom TLS certificates
* Support for AWS EBS gp2 to gp3 migration, supporting iops and throughput configuration
* Compatible with OpenShift
* TDE-Integration with [CYBERTEC PGEE](https://www.cybertec-postgresql.com/en/products/cybertec-postgresql-enterprise-edition/)

### PostgreSQL features

* Supports PostgreSQL 17, starting from 13+
* Streaming replication cluster via Patroni
* Integrated backup solution, automatic backups and very easy restore (Backup & PITR)
* Rolling update procedure for adjustments to the pods and minor updates
* Major upgrade with minimum interruption time
* Reduction of downtime thanks to redundancy, pod anti-affinity, auto-failover and self-healing
* Supports PostgreSQL 17, starting from 13+
* Several extensions such as Postgis, TimescaleDB, pgAudit and many more are included
* Streaming replication cluster via Patroni
* Point-In-Time-Recovery with
[pg_basebackup](https://www.postgresql.org/docs/16/app-pgbasebackup.html) /
[pgBackRest](https://pgbackrest.org/) via [CYBERTEC-pg-container](https://github.com/cybertec-postgresql/CYBERTEC-pg-container)
[pg_stat_statements](https://www.postgresql.org/docs/16/pgstatstatements.html),
* Incl. popular Postgres extensions such as
[pgaudit](https://github.com/pgaudit/pgaudit),
[pgauditlogtofile](https://github.com/fmbiete/pgauditlogtofile),
<!-- [pg_partman](https://github.com/pgpartman/pg_partman), -->
[postgis](https://postgis.net/),
[set_user](https://github.com/pgaudit/set_user)
[pg_cron](https://github.com/citusdata/pg_cron),
[timescaledb](https://github.com/timescale/timescaledb)
[credcheck](https://github.com/MigOpsRepos/credcheck)

The Operator project is being driven forward by CYBERTEC and is currently in production at various locations.

## Supported Postgres & K8s versions

| Release   | Postgres versions | pgBackRest versions   | Patroni versions | K8s versions      | Golang  |
| :-------- | :---------------: | :-------------------: | :--------------: | :----------------:| :-----: |
| 0.8.0     | 13 &rarr; 17      | 2.53                  | 4.0.2            | 1.21+             | 1.21.7  |
| 0.8.3     | 13 &rarr; 17      | 2.54-2                | 4.0.5            | 1.21+             | 1.22.12  |


## Getting started

[Getting started - Documentation](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/quickstart/) 

[Tutorials](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials).


## Documentation

There is a browser-friendly version of this documentation at
[CPO-Documentation](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/)


