---
title: "Software-Components"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 2
---

Various software components are used to operate CPO. This chapter lists the most important components and their respective purposes. 

Basically, the CPO project focusses on the main tasks of each individual component. This means that each component does what it does best and only that. 
In addition to reliable operation, this should also ensure efficient development and project management that utilises existing approaches rather than fighting against them.

### 1. CYBERTEC-pg-operator
The CYBERTEC-pg-operator is a Kubernetes operator that automates the operation and management of PostgreSQL databases on Kubernetes clusters. It facilitates the provisioning, scaling, backup and recovery of PostgreSQL clusters and integrates tools such as Patroni and pgBackRest for high availability and backup management. 

The main focus of the operator is the creation of the necessary templates and objects for Kubernetes, the regular check whether the declarative description of the cluster is still up to date and for the implementation of various tasks in the cluster, which were commissioned by the user. 

### 2. Kubernetes

Kubernetes is an open source platform for automating the deployment, scaling and management of containerised applications. It enables the management of container clusters in different environments and offers functions such as automatic load balancing, self-healing and rollouts. Kubernetes ensures that applications are always available and scalable and provides a framework for managing infrastructure in a cloud-native environment.

The focus of Kubernetes in the context of CPO is the use of the operator's templates to create the required objects. 
For example, the statefulset controller creates the desired pods based on the template. Kubernetes or the respective controllers monitor the generated objects independently and react if they are missing or do not correspond to the template. 
This means, for example, that pods that have been removed are automatically regenerated even if the operator is not currently running. This avoids the operator as a single point of failure.

### 3. Patroni
Patroni is an open source tool for managing PostgreSQL high availability clusters. It uses a distributed consensus mechanism, often based on Etcd, Consul or Zookeeper, to manage the role of the PostgreSQL primary node and perform automatic failovers. Patroni ensures that only one primary database server is active at a time, enabling consistency and availability of PostgreSQL databases in a cluster.

The focus of Patroni is to build, configure and monitor the PostgreSQL cluster based on the configuration created by the operator. Patroni therefore takes over all tasks such as leader selection, cluster monitoring, auto-failover and much more independently. 
Patroni is included in every PostgreSQL container and therefore pod and focussed on the individual cluster. 
This means that cluster management is guaranteed even without a currently running operator and therefore runs independently of the operator. This avoids the operator as a single point of failure.

### 4. PostgreSQL
PostgreSQL is a powerful, open source object-relational database management system (ORDBMS). It is known for its reliability, robustness and compliance with SQL standards. PostgreSQL supports advanced data types, functions and offers extensive customisation options. It is suitable for applications of any size and offers strong support for ACID transactions and Multi-Version Concurrency Control (MVCC).

The main role of PostgreSQL in the context of CPO is quite clear. Controlled by Patroni, PostgreSQL takes care of its task as a DBMS. 

### 5. pgBackRest
pgBackRest is a reliable backup and restore tool for PostgreSQL databases. It offers features such as incremental backups, parallel backup and restore, compression and encryption. pgBackRest is designed for use in large databases and offers both local and remote backup options. It integrates well into Kubernetes environments and enables automated and efficient backup strategies.

pgBackRest is configured based on the cluster manifest and therefore via the operator. Automatic backups, on the other hand, are based on Kubernetes cron jobs and are therefore independent of the operator, apart from the template generation by the operator. 

### 6. pgBouncer
PgBouncer is a lightweight connection pooler for PostgreSQL. It reduces the load on the database server by consolidating and efficiently managing incoming client connections. PgBouncer improves the performance and scalability of PostgreSQL-based applications by reducing the number of active connections while enabling fast switching times between different connections.