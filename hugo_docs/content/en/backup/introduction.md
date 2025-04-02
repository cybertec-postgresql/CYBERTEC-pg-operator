---
title: "Introduction"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 1
---
Backups are essential for databases. From broken storage to deployments gone wrong, backups often save the day. Starting with pg_dump, which was released in the late 1990s, to the archiving of WAL files (PostgreSQL 8.0 / 2005) and pg_basebackup (PostgreSQL 9.0 / 2010), PostgreSQL already offers built-in options for backups and restores based on logical and physical backups. 

### Backups with pgBackRest

CPO relies on [pgBackRest](www.pgbackrest.org) as its backup solution, a tried-and-tested tool with extensive backup and restore options.
The backup is based on two elements: 
- Snapshots in the form of physical backups
- WAL archive: Continuous archiving of the WAL files

### Backups

Backups represent a snapshot of the database in the form of pyhsical files. This contains all relevant information that PostgreSQL holds in its data folder.
With pgBackRest it is possible to create different types of Backups: 
- full Snapshot: This captures and saves all files at the time of the backup
- Differential backup: Only captures all files that have been changed since the last full Backup
- Incremental backup: Only records the files that have been changed since the last backup (of any kind). 

When restoring using differential or incremental Backup, it is necessary to also use the previous Backup that provide the basis for the selected Backup. 

{{< hint type=Info >}}The choice of Backup types depends on factors such as the size of the database, the time available for backups and the restore.{{< /hint >}}

### WAL-Archive

The WAL (Write-Ahead-Log) refers to log files which record all changes to the database data before they are written to the actual files. The basic idea here is to guarantee the consistency and recoverability of the comitted data even in the event of failures. 

PostgreSQL normally cleans up or recycles the WAL files that are no longer required. By using WAL archiving, the WAL files are saved to a different location before this process so that they can be used for various activities in the future. 
These activities include
- Providing the WAL files for replicas to keep them up to date
- Restoring instances that have lost parts of the WAL files in the event of a failure and cannot return to a consistent state without them without losing data
- Point-In-Time-Recovery (PITR): In contrast to Backups, which map a fixed point in time, WAL files make it possible to jump dynamically to a desired point in time and restore the database to the closest available consistent data point

{{< hint type=Info >}}WAL archiving is an indispensable tool for data availability, recoverability and the continuous availability of PostgreSQL.{{< /hint >}}

### Backup your Cluster

With pgBackRest, backups can be stored on different types of storage: 
- Block storage (PVC)
- S3 / S3-compatible storage
- Azure blob storage
- GCS

### How a Backup works

The operator creates a cronjob object on Kubernetes based on the defined times for automatic backups. This means that the Kubernetes core ([CronJob Controller](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)) will take care of processing the automatic backups and create a job and thus a pod at the appropriate time. 
The pod will send the backup command to the primary or, if block storage is used, to the repo host and monitor it. As soon as the backup is successfully completed, the pod stops with Completed and thus completes the job. 

```
kubectl get cronjobs
---------------------------------------------------------------------------------------
NAME                          | SCHEDULE     | SUSPEND | ACTIVE | LAST SCHEDULE | AGE
pgbackrest-cluster-repo1-full | 30 2 * * *   | False   | 0      | 4h46m         | 14h
pgbackrest-cluster-repo1-incr | */30 * * * * | False   | 1      | 81s           | 106m

kubectl get jobs
-----------------------------------------------------------------------
NAME                                   | COMPLETIONS | DURATION | AGE
pgbackrest-cluster-repo1-full-28597110 | 1/1         | 52s      | 140m
pgbackrest-cluster-repo1-incr-28597365 | 1/1         | 2m37s    | 32m
pgbackrest-cluster-repo1-incr-28597380 | 1/1         | 2m38s    | 17m
pgbackrest-cluster-repo1-incr-28597395 | 0/1         | 2m3s     | 2m3s

```

If there are problems such as a timeout, the pod will stop with exit code 1 and thus indicate an error. In this case, a new pod will be created which will attempt to complete the backup. The maximum number of attempts is 6, so if the backup fails six times, the job is deemed to have failed and will not be attempted again until the next cronjob execution. The job pod log provides information about the problems.

```
kubectl get pods
-----------------------------------------------------------------------------------
NAME                                         | READY | STATUS    | RESTARTS | AGE
cluster-0                                    | 2/2   | Running   | 2        | 14h
cluster-pgbackrest-repo-host-0               | 1/1   | Running   | 0        | 107m
pgbackrest-cluster-repo1-full-28597110-x8zpw | 0/1   | Completed | 0        | 143m
pgbackrest-cluster-repo1-incr-28597365-7bb5l | 0/1   | Completed | 0        | 34m
pgbackrest-cluster-repo1-incr-28597380-j76rr | 0/1   | Completed | 0        | 19m
pgbackrest-cluster-repo1-incr-28597395-rh86t | 0/1   | Completed | 0        | 4m27s
postgres-operator-66bbff5c54-5sjmk           | 1/1   | Running   | 0        | 47m
```
