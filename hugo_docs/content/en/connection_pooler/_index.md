---
title: "connection pooler"
date: 2024-05-31T14:26:51+01:00
draft: false
weight: 1700
---

A connection pooler is a tool that acts as a proxy between the application and the database and enables the performance of the application to be improved and the load on the database to be reduced. The reason for this lies in the connection handling of PostgreSQL. 

## How PostgreSQL handles connection
PostgreSQL use a new Process for every database-connection created by the postmaster. This process is handling the connection. On the positive side, this enables a stable connection and isolation, but it is not particularly efficient for short-lived connections due to the effort required to create them.

## How Connection Pooling solves this problem

With connection pooling, the application connects to the pooler, which in turn maintains a number of connections to the PostgreSQL database. 
This makes it possible to use the connections from the pooler to the database for a long time instead of short-lived connections and to recycle them accordingly.

In addition to utilising long-term connections, a ConnectionPooler also makes it possible to reduce the number of connections required to the database. For example, if you have 3 application nodes, each of which maintains 100 connections to the database at the same time, that would be 300 connections in total. The application usually does not even begin to utilise this number of connections. With the pgBouncer, this can be optimised so that the applications open the 300 connections to the pgBouncer, but the pgBouncer only generates 100 connections to PostgreSQL, for example, thus reducing the load by 2/3. 

{{< hint type=Info >}}It is important to correctly configure the bouncer and thus the connections to be created between pgBouncer and PostgreSQL so that enough connections are available for the workload. {{< /hint >}}

## How does this work with CPO
CPO relies on pgBouncer, a popular and above all lightweight open source tool. pgBouncer manages individual user-database connections for each user used, which can be used immediately for incoming client connections. 

## How do I create a pooler for a cluster?

- connection_pooler.number_of_instances - How many instances of connection pooler to create. Default is 2 which is also the required minimum.
- connection_pooler.schema - Database schema to create for credentials lookup function to be used by the connection pooler. Is is created in every database of the Postgres cluster. You can also choose an existing schema. Default schema is pooler.
- connection_pooler.user - User to create for connection pooler to be able to connect to a database. You can also choose an existing role, but make sure it has the LOGIN privilege. Default role is pooler.
- connection_pooler.image - Docker image to use for connection pooler deployment. Default: “registry.opensource.zalan.do/acid/pgbouncer”
- connection_poole.max_db_connections - How many connections the pooler can max hold. This value is divided among the pooler pods. Default is 60 which will make up 30 connections per pod for the default setup with two instances.
- connection_pooler.mode - Defines pooler mode. Available Value:  `session`,  `transaction` or `statement`. Default is `transaction`.
- connection_pooler.resources - Hardware definition for the pooler pods

- enableConnectionPooler - Defines whether poolers for read/write access should be created based on the spec.connectionPooler definition. 
- enableReplicaConnectionPooler- Defines whether poolers for read-only access should be created based on the spec.connectionPooler definition. 

```
spec:
  connectionPooler:
    mode: transaction
    numberOfInstances: 2
    resources:
      limits:
        cpu: '1'
        memory: 100Mi
      requests:
        cpu: 500m
        memory: 100Mi
    schema: pooler
    user: pooler
  enableConnectionPooler: true
  enableReplicaConnectionPooler: true
```


