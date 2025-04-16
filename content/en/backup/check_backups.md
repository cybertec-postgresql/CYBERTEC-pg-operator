---
title: "Check/Monitor Backups"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 7
---
There are several ways to gain an insight into the current status of pgBackRest. 
One of these is to use pgBackRest within the container. This can be done both via the repo host and the Postgres pod.

### pgbackrest via terminal (Repo-Host-Pod)
```
kubectl exec cluster-5-pgbackrest-repo-host-0 --stdin --tty -- pgbackrest info 
stanza: db
    status: ok
    cipher: none

    db (current)
        wal archive min/max (16): 00000006000000000000005C/000000070000000000000092

        full backup: 20240517-125730F
            timestamp start/stop: 2024-05-17 12:57:30+00 / 2024-05-17 12:57:41+00
            wal start/stop: 00000007000000000000005E / 00000007000000000000005E
            database size: 22.9MB, database backup size: 22.9MB
            repo1: backup set size: 3MB, backup size: 3MB

        incr backup: 20240517-125730F_20240517-130003I
            timestamp start/stop: 2024-05-17 13:00:03+00 / 2024-05-17 13:00:05+00
            wal start/stop: 000000070000000000000060 / 000000070000000000000060
            database size: 22.9MB, database backup size: 904.3KB
            repo1: backup set size: 3MB, backup size: 149.4KB
            backup reference list: 20240517-125730F

        incr backup: 20240517-125730F_20240517-131503I
            timestamp start/stop: 2024-05-17 13:15:03+00 / 2024-05-17 13:15:04+00
            wal start/stop: 000000070000000000000062 / 000000070000000000000062
            database size: 22.9MB, database backup size: 24.3KB
            repo1: backup set size: 3MB, backup size: 2.9KB
            backup reference list: 20240517-125730F, 20240517-125730F_20240517-130003I
```
### pgbackrest via terminal (Postgres-Pod)
```
kubectl exec cluster-5-0 --stdin --tty -- pgbackrest info 
Defaulted container "postgres" out of: postgres, postgres-exporter, pgbackrest-restore (init)
stanza: db
    status: ok
    cipher: none

    db (current)
        wal archive min/max (16): 00000006000000000000005C/000000070000000000000092

        full backup: 20240517-125730F
            timestamp start/stop: 2024-05-17 12:57:30+00 / 2024-05-17 12:57:41+00
            wal start/stop: 00000007000000000000005E / 00000007000000000000005E
            database size: 22.9MB, database backup size: 22.9MB
            repo1: backup set size: 3MB, backup size: 3MB

        incr backup: 20240517-125730F_20240517-130003I
            timestamp start/stop: 2024-05-17 13:00:03+00 / 2024-05-17 13:00:05+00
            wal start/stop: 000000070000000000000060 / 000000070000000000000060
            database size: 22.9MB, database backup size: 904.3KB
            repo1: backup set size: 3MB, backup size: 149.4KB
            backup reference list: 20240517-125730F

        incr backup: 20240517-125730F_20240517-131503I
            timestamp start/stop: 2024-05-17 13:15:03+00 / 2024-05-17 13:15:04+00
            wal start/stop: 000000070000000000000062 / 000000070000000000000062
            database size: 22.9MB, database backup size: 24.3KB
            repo1: backup set size: 3MB, backup size: 2.9KB
            backup reference list: 20240517-125730F, 20240517-125730F_20240517-130003I
```
There is the "normal" output, as well as the output format Json, which can be processed directly in the terminal. 

```
kubectl exec cluster-5-0 --stdin --tty -- pgbackrest info  --output=json
```

### Check pgBackrest via Monitoring

In addition to reading the status via the containers, pgBackRest can also be analysed and monitored via the monitoring stack. You can find information on setting up the monitoring stack and further information [here](documentation/how-to-use/monitoring).