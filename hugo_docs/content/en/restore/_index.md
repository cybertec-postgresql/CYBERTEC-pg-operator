---
title: "Restore"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 1400
---

Restore or recovery is the process of starting a PostgreSQL instance or a cluster based on a defined and existing backup. This can be just a Backup or a combination of a Backup and additional WAL files. The difference is that a Backup represents a fixed point in time, whereas the combination with WAL enables a point-in-time recovery(PITR). 

You can find more information about backups [here](../backup/introduction/)

### Rescue my cluster

CPO enables the restore based on an existing backup using pgBackRest. To do this, it needs the relevant information about the point in time or snapBackupshot to which it should restore and where the data for this comes from. 
As we have already provided the operator with all the information relating to the storage of backups in the previous chapter, it only needs the following information: 
- `id`: Control variable, must be increased for each restore process 
- `type`: What type of restore is required
- `repo`: Which repo the data should come from
- `set`: Specific Backup to restore - Check [backup](../backup/check_backups/) to see how to get the identifier

{{< hint type=Info >}}To ensure that the operator does not repeat an already done restore, the defined object `id` in the restore section is saved by the operator, so the value of this `id` must be changed for a new restore.{{< /hint >}}

#### Details for a Backup restore
With this information, we define a fixed Backup from `repo1` and that pgBackRest should stop at the end of the Backup
```
restore:
  id: '1'
  options:
    type: 'immediate'
    set: '20240515-164100F'
  repo: 'repo1'
```

{{< hint type=info >}} Without the specification `--type=immediate`, pgBackRest would then consume the entire WAL that is available and thus restore the last available consistent data point. {{< /hint >}}

#### Details for a point-in-time recoery (PITR)
We use this information to define a point-in-time recovery (PITR) and define the end point using a timestamp and the start point using a Backup specification. The latter is optional. Without this information, pgBackRest would automatically start at the last previous full Backup. 
```
restore:
  id: '1'
  options:
    type: 'time'
    set: '20240515-164100F'
    target: '2024-05-16 07:46:05.506817+00'

  repo: '1'
```
{{< hint type=info >}}`--type=time` indicates that it is a point-in-time recovery (PITR).  {{< /hint >}}

## Example in a cluster manifest

```
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-5
  namespace: cpo
spec:
  backup:
    pgbackrest:
      configuration:
        secret: cluster-1-pvc-credentials
      global:
        repo1-retention-full: '7'
        repo1-retention-full-type: count
      image: 'docker.io/cybertecpostgresql/cybertec-pg-container:pgbackrest-16.4-1'
      repos:
        - name: repo1
          schedule:
            full: 30 2 * * *
          storage: pvc
          volume:
            size: 1Gi
      restore:
        id: '1'
        options:
          type: 'time'
          set: '20240515-164100F'
          target: '2024-05-16 07:46:05.506817+00'
```
An example of this can also be found in our tutorials. For a point-in-time recovery (PITR) you can find it [here](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/tree/main/cluster-tutorials/restore_pitr).

{{< hint type=warning >}} Incorrect information for the Backup or the timestamp can result in pgBackRest not being able to complete the restore successfully. In the event of an error, the information must be corrected and another restore must be started. {{< /hint >}}
