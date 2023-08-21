# CYBERTEC PG Operator

CPO (CYBERTEC PG Operator) allows you to create and run PostgreSQL clusters on Kubernetes.

The operator reduces your efforts and simplifies the administration of your PostgreSQL clusters so that you can concentrate on other things.

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

* Integrated backup solution, automatic backups and very easy restore (snapshot & PITR)
* Rolling update procedure for adjustments to the pods and minor updates
* Major upgrade with minimum interruption time
* Reduction of downtime thanks to redundancy, pod anti-affinity, auto-failover and self-healing

## Getting starteds

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