---
title: "Container Images"
date: 2024-03-11T14:26:51+01:00
draft: false
weight: 202
---

For each version of the operator and the required PostgreSQL and other required containers, the corresponding image is provided on Dockerhub.

#### Operator container images
The operator images are the central components that control the operation and administration of the PostgreSQL databases. These images are available in the following repository on DockerHub:

[Operator Images](https://hub.docker.com/repository/docker/cybertecpostgresql/cybertec-pg-operator)

The repository contains all the necessary images for running the Cybertec PG Operator in a Kubernetes environment. These images are updated regularly to ensure the latest features and security updates.

#### Additional container images
In addition to the operator images, various container images are required to support a complete PostgreSQL environment. These images are available in the following repository:
[CYBERTEC-PG-Container Images](https://hub.docker.com/repository/docker/cybertecpostgresql/cybertec-pg-container/general)

This repository contains images for the following components:

- PostgreSQL: The main database image, which contains all supported major versions of PostgreSQL. The name of the tag always reflects the latest release, e.g. currently `16.4` for PostgreSQL `16.4`. For the other major versions, the corresponding minor versions released by the PostgreSQL community at the same time are included.
- Postgres-GIS: A specialised image that combines PostgreSQL with the PostGIS extension to support spatial data processing functions. You can find more information about Postgis [here](../../postgis).  
The tag for Postgis also includes the included version of Postgis. Example: postgres-16.4-34-1 Postgis: `3.4.x`
- PGBackRest: A backup and restore tool developed specifically for PostgreSQL and available as a separate container image.
- Exporter: Images for monitoring PostgreSQL databases that collect metrics and make them available for monitoring tools such as Prometheus.
- PgBouncer: A lightweight connection pooler for PostgreSQL that manages and optimises the number of concurrent connections.


#### Extensions
You can view the versions included in the [Extensions](../../extensions/pg16/) section.