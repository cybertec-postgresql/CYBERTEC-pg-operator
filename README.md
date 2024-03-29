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
* Live volume resize without pod restarts (AWS EBS, PVC)
* Database connection pooling with PGBouncer
* Support fast in place major version upgrade. Supports global upgrade of all clusters.
* Restore and cloning Postgres clusters on AWS, GCS and Azure
* Additionally logical backups to S3 or GCS bucket can be configured
* Standby cluster from S3 or GCS WAL archive
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
[pg_basebackup](https://www.postgresql.org/docs/11/app-pgbasebackup.html) /
[WAL-E](https://github.com/wal-e/wal-e) via [Spilo](https://github.com/zalando/spilo)
* Preload libraries: [bg_mon](https://github.com/CyberDem0n/bg_mon),
[pg_stat_statements](https://www.postgresql.org/docs/15/pgstatstatements.html),
[pgextwlist](https://github.com/dimitri/pgextwlist),
[pg_auth_mon](https://github.com/RafiaSabih/pg_auth_mon)
* Incl. popular Postgres extensions such as
[decoderbufs](https://github.com/debezium/postgres-decoderbufs),
[hypopg](https://github.com/HypoPG/hypopg),
[pg_cron](https://github.com/citusdata/pg_cron),
[pg_partman](https://github.com/pgpartman/pg_partman),
[pg_stat_kcache](https://github.com/powa-team/pg_stat_kcache),
[pgq](https://github.com/pgq/pgq),
[plpgsql_check](https://github.com/okbob/plpgsql_check),
[postgis](https://postgis.net/),
[set_user](https://github.com/pgaudit/set_user) and
[timescaledb](https://github.com/timescale/timescaledb)

The Postgres Operator has been developed at Zalando and is being used in
production for over five years.

## Supported Postgres & K8s versions

| Release   | Postgres versions | K8s versions      | Golang  |
| :-------- | :---------------: | :---------------: | :-----: |
| v1.10.*   | 10 &rarr; 15      | 1.21+             | 1.19.8  |
| v1.9.0    | 10 &rarr; 15      | 1.21+             | 1.18.9  |
| v1.8.*    | 9.5 &rarr; 14     | 1.20 &rarr; 1.24  | 1.17.4  |
| v1.7.1    | 9.5 &rarr; 14     | 1.20 &rarr; 1.24  | 1.16.9  |

* Integrated backup solution, automatic backups and very easy restore (snapshot & PITR)
* Rolling update procedure for adjustments to the pods and minor updates
* Major upgrade with minimum interruption time
* Reduction of downtime thanks to redundancy, pod anti-affinity, auto-failover and self-healing
* Supports PostgreSQL 15, starting from 10+
* Streaming replication cluster via Patroni
* Point-In-Time-Recovery with
[pg_basebackup](https://www.postgresql.org/docs/11/app-pgbasebackup.html) /
[WAL-E](https://github.com/wal-e/wal-e) via [Spilo](https://github.com/zalando/spilo)
* Preload libraries: [bg_mon](https://github.com/CyberDem0n/bg_mon),
[pg_stat_statements](https://www.postgresql.org/docs/15/pgstatstatements.html),
[pgextwlist](https://github.com/dimitri/pgextwlist),
[pg_auth_mon](https://github.com/RafiaSabih/pg_auth_mon)
* Incl. popular Postgres extensions such as
[decoderbufs](https://github.com/debezium/postgres-decoderbufs),
[hypopg](https://github.com/HypoPG/hypopg),
[pg_cron](https://github.com/citusdata/pg_cron),
[pg_partman](https://github.com/pgpartman/pg_partman),
[pg_stat_kcache](https://github.com/powa-team/pg_stat_kcache),
[pgq](https://github.com/pgq/pgq),
[plpgsql_check](https://github.com/okbob/plpgsql_check),
[postgis](https://postgis.net/),
[set_user](https://github.com/pgaudit/set_user) and
[timescaledb](https://github.com/timescale/timescaledb)

The Postgres Operator has been developed at Zalando and is being used in
production for over five years.

## Supported Postgres & K8s versions

| Release   | Postgres versions | K8s versions      | Golang  |
| :-------- | :---------------: | :---------------: | :-----: |
| v1.10.*   | 10 &rarr; 15      | 1.25+             | 1.19.8  |
| v1.9.0    | 10 &rarr; 15      | 1.25+             | 1.18.9  |
| v1.8.*    | 9.5 &rarr; 14     | 1.20 &rarr; 1.24  | 1.17.4  |
| v1.7.1    | 9.5 &rarr; 14     | 1.20 &rarr; 1.24  | 1.16.9  |

* Integrated backup solution, automatic backups and very easy restore (snapshot & PITR)
* Rolling update procedure for adjustments to the pods and minor updates
* Major upgrade with minimum interruption time
* Reduction of downtime thanks to redundancy, pod anti-affinity, auto-failover and self-healing

## Getting started

[Getting started - Documentation](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/documentation/how-to-use/installation/) 

[Tutorials](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials).


## Documentation

Coming soon 

Until then, please use the following:

There is a browser-friendly version of this documentation at
[postgres-operator.readthedocs.io](https://postgres-operator.readthedocs.io)

* [How it works](docs/index.md)
* [Installation](docs/quickstart.md#deployment-options)
* [The Postgres experience on K8s](docs/user.md)
* [The Postgres Operator UI](docs/operator-ui.md)
* [DBA options - from RBAC to backup](docs/administrator.md)
* [Build, debug and extend the operator](docs/developer.md)
* [Configuration options](docs/reference/operator_parameters.md)
* [Postgres manifest reference](docs/reference/cluster_manifest.md)
* [Command-line options and environment variables](docs/reference/command_line_and_environment.md)

## Community

Coming soon 