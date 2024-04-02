# CYBERTEC PG Operator

CPO (CYBERTEC PG Operator) allows you to create and run PostgreSQL clusters on Kubernetes.

The operator reduces your efforts and simplifies the administration of your PostgreSQL clusters so that you can concentrate on other things.
# CYBERTEC PG Operator

CPO (CYBERTEC PG Operator) allows you to create and run PostgreSQL clusters on Kubernetes.

The operator reduces your efforts and simplifies the administration of your PostgreSQL clusters so that you can concentrate on other things.
<img src="docs/diagrams/logo.png" width="200">

The Postgres Operator delivers an easy to run highly-available [PostgreSQL](https://www.postgresql.org/)
clusters on Kubernetes (K8s) powered by [Patroni](https://github.com/zalando/patroni).
It is configured only through Postgres manifests (CRDs) to ease integration into automated CI/CD
pipelines with no access to Kubernetes API directly, promoting infrastructure as code vs manual operations.

### Operator features

* Rolling updates on Postgres cluster changes, incl. quick minor version updates
* Live volume resize without pod restarts if supported by the storage-system (PVC)
* Database connection pooling with PGBouncer
* Support fast in place major version upgrade. Supports global upgrade of all clusters.
* Restore and cloning Postgres clusters on PVC, AWS, GCS and Azure
* Additionally logical backups to S3 or GCS bucket can be configured
* Standby cluster
* Configurable for non-cloud environments
* Basic credential and user management on K8s, eases application deployments
* Support for custom TLS certificates
* UI to create and edit Postgres cluster manifests
* Support for AWS EBS gp2 to gp3 migration, supporting iops and throughput configuration
* Compatible with OpenShift

### PostgreSQL features

* Supports PostgreSQL 15, starting from 10+
* Streaming replication cluster via Patroni
* Point-In-Time-Recovery with
[pg_basebackup](https://www.postgresql.org/docs/16/app-pgbasebackup.html) /
[pgBackRest](https://pgbackrest.org/) via [CYBERTEC-pg-container](https://github.com/cybertec-postgresql/CYBERTEC-pg-container)
[pg_stat_statements](https://www.postgresql.org/docs/15/pgstatstatements.html),
* Incl. popular Postgres extensions such as
[pg_cron](https://github.com/citusdata/pg_cron),
[pg_partman](https://github.com/pgpartman/pg_partman),
[postgis](https://postgis.net/),
[set_user](https://github.com/pgaudit/set_user) and
[timescaledb](https://github.com/timescale/timescaledb)
[credcheck](https://github.com/MigOpsRepos/credcheck)

The Postgres Operator has been developed at Zalando and is being used in
production for over five years.

## Supported Postgres & K8s versions

| Release   | Postgres versions | pgBackRest versions   | Patroni versions | K8s versions      | Golang  |
| :-------- | :---------------: | :-------------------: | :--------------: | :----------------:| :-----: |
| latest    | 13 &rarr; 16      | 2.51                  | 3.2.2            | 1.21+             | 1.19.8  |
| next rc   | 13 &rarr; 16      | 2.51                  | 3.2.2            | 1.21+             | 1.22.1  |

* Integrated backup solution, automatic backups and very easy restore (snapshot & PITR)
* Rolling update procedure for adjustments to the pods and minor updates
* Major upgrade with minimum interruption time
* Reduction of downtime thanks to redundancy, pod anti-affinity, auto-failover and self-healing
* Supports PostgreSQL 16, starting from 13+
* Streaming replication cluster via Patroni
* Point-In-Time-Recovery with
[pg_basebackup](https://www.postgresql.org/docs/11/app-pgbasebackup.html) /
[pgBackRest](https://pgbackrest.org/) via [CYBERTEC-pg-container](https://github.com/cybertec-postgresql/CYBERTEC-pg-container)
[pg_stat_statements](https://www.postgresql.org/docs/15/pgstatstatements.html),
* Incl. popular Postgres extensions such as
[pg_cron](https://github.com/citusdata/pg_cron),
[pg_partman](https://github.com/pgpartman/pg_partman),
[postgis](https://postgis.net/),
[set_user](https://github.com/pgaudit/set_user) and
[timescaledb](https://github.com/timescale/timescaledb)
[credcheck](https://github.com/MigOpsRepos/credcheck)

## Getting started

[Getting started - Documentation](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/documentation/how-to-use/installation/) 

[Tutorials](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials).


## Documentation

There is a browser-friendly version of this documentation at
[CPO-Documentation](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/)

## Community

Coming soon 