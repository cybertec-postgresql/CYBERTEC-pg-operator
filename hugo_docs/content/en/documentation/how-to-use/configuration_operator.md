---
title: "Operator-Configuration"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 32020
---

Users who are already used to working with PostgreSQL from Baremetal or VMs are already familiar with the need for various files to configure PostgreSQL. These include
- postgresql.conf 
- pg_hba.conf
- ...

Although these files are available in the container, direct modification is not planned. As part of the declarative mode of operation of the operator, these files are defined via the operator. The modifying intervention within the container also represents a contradiction to the immutability of the container.

For these reasons, the operator provides a way to make adjustments to the various files, from PostgreSQL to Patroni. 

We differentiate between two main objects in the cluster manifest:
- [`postgresql`](documentation/how-to-use/configuration/#postgresql) with the child objects `version` and `parameters`
- [`patroni`](documentation/how-to-use/configuration/#patroni) with objects for the `pg_hab`, `slots` and much more

## postgresql

The `postgresql `object consists of the following elements:
- `version` - allows you to select the major version of PostgreSQL used. 
- `parameters`- enables the postgresql.conf to be changed

```
spec:
  postgresql:
    parameters:
      shared_preload_libraries: 'pg_stat_statements,pgnodemx, timescaledb'
      shared_buffers: '512MB'
    version: '16'
```

Any known PostgreSQL parameter from postgresql.conf can be entered here and will be delivered by the operator to all nodes of the cluster accordingly. 

You can find more information about the parameters in the [PostgreSQL documentation](https://www.postgresql.org/docs/)

## patroni

The patroni object contains numerous options for customising the patroni-setu, and the pg_hba.conf is also configured here. A complete list of all available elements can be found here. 

The most important elements include 
- `pg_hba` - pg_hba.conf
- `slots` 
- `synchronous_mode` - enables synchronous mode in the cluster. The default is set to `false`
- `maximum_lag_on_failover` - Specifies the maximum lag so that the pod is still considered healthy in the event of a failover.
- `failsafe_mode` Allows you to cancel the downgrading of the leader if all cluster members can be reached via the Patroni Rest Api. 
You can find more information on this in the [Patroni documentation](https://patroni-readthedocs-io.translate.goog/en/master/dcs_failsafe_mode.html?_x_tr_sl=auto&_x_tr_tl=de&_x_tr_hl=de&_x_tr_pto=wapp)

### pg_hba

The pg_hba.conf contains all defined authentication rules for PostgreSQL. 

When customising this configuration, it is important that the entire version of pg_hba is written to the manifest. 
The current configuration can be read out in the database using table pg_hba_file_rules ;. 

Further information can be found in the [PostgreSQL documentation](https://www.postgresql.org/docs/current/auth-pg-hba-conf.html)


### slots

When using user-defined slots, for example for the use of CDC using Debezium, there are problems when interacting with Patroni, as the slot and its current status are not automatically synchronised to the replicas. 

In the event of a failover, the client cannot start replication as both the entire slot and the information about the data that has already been synchronised are missing. 

To resolve this problem, slots must be defined in the cluster manifest rather than in PostgreSQL. 

```
spec:
  patroni:
    slots:
      cdc-example:
        database: app_db
        plugin: pgoutput
        type: logical
```
This example creates a logical replication slot with the name `cdc-example` within the `app_db` database and uses the `pgoutput` plugin for the slot.


> **_ATTENTION:_**  Slots are only synchronised from the leader/standby leader to the replicas. This means that using the slots read-only on the replicas will cause a problem in the event of a failover.


