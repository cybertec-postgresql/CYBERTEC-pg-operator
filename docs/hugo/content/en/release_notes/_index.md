---
title: "Release-Notes"
date: 2024-03-11T14:26:51+01:00
draft: false
weight: 2500
---
### 0.9.0

#### Features
- Adding PG18
- Liveness check added (can be activated via Operator Configuration)
- ReadOnlyRootFilesystem added as SecurityContext (can be activated via Operator Configuration)

#### Changes
- Add OwnerReference for Statefulsets
- Optimisations for major upgrade
- Statefulsert receives OwnerReference to the CR

#### Fixes
- cert-Handling for Multisite
- pgBackRest Restore with TDE
- Fix for Monitoring pgBackRest
- Dependency updates and several small changes

#### Notification of upcoming deprecation
- PG13 has reached its EoL 

#### Supported Versions

- PG: 13 - 18
- Patroni: 4.1.0
- pgBackRest: 2.57.0
- Kubernetes: 1.21 - 1.32
- Openshift: 4.8 - 4.18

### 0.8.3

#### Fixes
- Majorupgrade updated for Patroni 4.x.x
- Fixes for PGEE
- Fix for Monitoring-User
- Dependency updates and several small changes

#### Supported Versions

- PG: 13 - 17
- Patroni: 4.0.5
- pgBackRest: 2.54.2
- Kubernetes: 1.21 - 1.32
- Openshift: 4.8 - 4.18

### 0.8.2

#### Features
- Added Clone-Functionality with pgBackRest

#### Supported Versions

- PG: 13 - 17
- Patroni: 3.3.2
- pgBackRest: 2.54.0
- Kubernetes: 1.21 - 1.32
- Openshift: 4.8 - 4.18

### 0.8.1

#### Features
- Added pgbackrest to Monitoring

#### Fixes
- Fixed role creation for monitoring

#### Supported Versions

- PG: 13 - 17
- Patroni: 3.3.2
- pgBackRest: 2.53
- Kubernetes: 1.21 - 1.32
- Openshift: 4.8 - 4.18

### 0.8.0

#### Features
- Multisite - Support
- use icu as default for pg > 14

#### Fixes
- Fixed role creation for monitoring.
- Fix for the use of gcs with pgBackRest

#### Supported Versions

- PG: 13 - 16 & 17Beta2
- Patroni: 3.3.2
- pgBackRest: 2.53
- Kubernetes: 1.21 - 1.32
- Openshift: 4.8 - 4.18

### 0.7.1

#### Fixes
- Fixed role creation for monitoring.
- Fix for the use of gcs with pgBackRest

#### Supported Versions

- PG: 13 - 16 & 17Beta2
- Patroni: 3.3.2
- pgBackRest: 2.53
- Kubernetes: 1.21 - 1.28
- Openshift: 4.8 - 4.13

### 0.7.0

#### Features
- Monitoring-Sidecar integrated via CRD [Start with Monitoring](documentation/cluster/monitoring)
- Password-Hash per default set to scram-sha-256
- pgBackRest with blockstorage using RepoHost
- Internal Certification-Management for RepoHost-Certificates
- Compatible with PG17Beta2

#### Changes
- API Change acid.zalan.do is replaced by cpo.opensource.cybertec.at - If you're updating your Operator from previous Versions, please check this  [HowTo Migrate to new API](documentation/operator/migrateToNewApi/)
- Patroni-Compatibility has increased to Version 3.3.2
- pgBackRest-Compatbility has increased to Version 2.52.1
- Revision of the restore process
- Revision of the backup jobs
- Operator now using Rocky9 as Baseimage
- Updates Go-Package to 1.22.5 

#### Fixes
- PDB Bug fixed - Single-Node Clusters are not creating PDBs anymore which can break Kubernetes-Update
- Wrong Templates inside Cronjobs fixed

#### Supported Versions

- PG: 13 - 16 & 17Beta2
- Patroni: 3.3.2
- pgBackRest: 2.52.1
- Kubernetes: 1.21 - 1.28
- Openshift: 4.8 - 4.13

### 0.6.1

Release with fixes

#### Fixes
- Backup-Pod now runs with "best-effort" resource definition
- Der Init-Container für die Wiederherstellung verwendet nun die gleiche Ressource-Definition wie der Datenbank-Container, wenn es keine spezifische Definition im Cluster-Manifest gibt (spec.backup.pgbackrest.resources)

#### Software-Versions

- PostgreSQL: 15.3 14.8, 13.11, 12.15
- Patroni: 3.0.4
- pgBackRest: 2.47
- OS: Rocky-Linux 9.1 (4.18)
</br></br>
___
</br></br>
### 0.6.0

Release with some improvements and stabilisation measuresm

#### Features
- Added [Pod Topology Spread Constraints](https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/)
- Added support for TDE based on the CYBERTEC PostgreSQL Enterprise Images (Licensed Container Suite)

#### Software-Versions

- PostgreSQL: 15.3 14.8, 13.11, 12.15
- Patroni: 3.0.4
- pgBackRest: 2.47
- OS: Rocky-Linux 9.1 (4.18)
</br></br>
___
</br></br>
### 0.5.0

Release with new Software-Updates and some internal Improvements
### Features
- Updated to Zalando Operator 1.9

#### Fixes
- internal Problems with Cronjobs
- updates for some API-Definitions

#### Software-Versions

- PostgreSQL: 15.2 14.7, 13.10, 12.14
- Patroni: 3.0.2
- pgBackRest: 2.45
- OS: Rocky-Linux 9.1 (4.18)
</br></br>
___
</br></br>
### 0.3.0

Release with some improvements and stabilisation measuresm

#### Fixes
- missing pgbackrest_restore configmap fixed

#### Software-Versions

- PostgreSQL: 15.1 14.7, 13.9, 12.13, 11.18 and 10.23
- Patroni: 3.0.1
- pgBackRest: 2.44
- OS: Rocky-Linux 9.1 (4.18)
</br></br>
___
</br></br>
### 0.1.0 
    
Initial Release as a Fork of the Zalando-Operator

#### Features

- Added Support for pgBackRest (PoC-State)
    - Stanza-create and Initial-Backup are executed automatically
    - Schedule automatic updates (Full/Incremental/Differential-Backup)
    - Securely store backups on AWS S3 and S3-compatible storage

#### Software-Versions

- PostgreSQL: 14.6, 13.9, 12.13, 11.18 and 10.23
- Patroni: 2.4.1
- pgBackRest: 2.42
- OS: Rocky-Linux 9.0 (4.18)
