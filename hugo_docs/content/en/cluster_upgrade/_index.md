---
title: "Major-Upgrade"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 2100
---

CPO enables the use of the in-place upgrade, which makes it possible to upgrade a cluster to a new PG major. For this purpose, pg_upgrade is used in the background.

> **_PAY ATTENTION:_**  Note that an in-place upgrade generates both a pod restore in the form of a rolling update and an operational interruption of the cluster during the actual execution of the restore.

## How does the upgrade work?

### Preconditions:
1. Pod restart - Use the rolling update strategy to replace all pods based on the new ENV `PGVERSION` with the version you want to update to.
2. Check - Check that the new `PGVERSION` is larger than the previously used one.
3. Check whether the new `PGVERSION` is larger than the previously used one and the maintenance mode of the cluster must be deactivated. In addition, the replicas should not have a high lag.

### Preliminary checks

1. use initdb to prepare a new data_dir (`data_new`) based on the new `PGVERSION`.
2. check the upgrade possibility with `pg_upgrade --check`

> **_HINT:_**  If one of the steps is aborted, a cleanup is performed

### Prepare the Upgrade
1. remove dependencies that can cause problems. For example, the extensions `pg_stat_statements` and `pgaudit`.
2. activate the maintenance mode of the cluster
3. terminate PostgreSQL in an orderly manner
4. check pg_controldata for the checkpoint position and wait until all replicas apply the latest checkpoint location
5. use port `5432` for rsyncd and start it 

### Start the Upgrade

1. Call pg_upgrade -k to start the Upgrade
> **_ATTENTION_** if the process failed, we need to rollback, if it was sucessful we're reaching the point of no return
2. Rename the directories. `data -> data_old` and `data_new -> data`
3. Update the Patroni.config (`postgres.yml`)
4. Call Checkpoint on every replica and trigger rsync on the Replicas
5. Wait for Replicas to complete rsxnc. `Timeout: 300` 
6. Stop rsyncd on Primary and remove ininitialize key from DCS, because its based on the old sysid
7. Start Patroni on the Primary and start the postgres locally
8. Reset custom staticstics, warmup the Memory and start Analyze in stages in separate threads
9. Wait for every Replica to become ready
10. Disable the maintenance mode for the Cluster
11. Restore custom statistics, analyze these tables and restore dropped objetcs from `Prepare the upgrade`

### Completion of the upgrade
1. Drop directory `data_old`
2. Trigger new Backup

### How a rollback is working?
1. Stop rsynd if its running
2. Disable the maintenance mode for the Cluster
3. Drop directory `data_new`


## How to trigger a In-Place-Upgrade with cpo?

```
spec:
  postgresql:
    version: "16"
```
To trigger an In-Place-Upgrade you have just to increase the parameter `spec.postgresql.version`. If you choose a valid number the Operator will start with the prozedure, described above. 
If you choosse a not allowed value, you will give an error and if you decrease the value, the operator will just ignore it with the following log-Entry.
```
Operator-Log
````