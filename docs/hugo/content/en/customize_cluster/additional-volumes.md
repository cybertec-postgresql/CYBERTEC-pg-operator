---
title: "Additional Volumes"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 3
---

```
  additionalVolumes:
    - name: empty
      mountPath: /opt/empty
      targetContainers:
        - all
      volumeSource:
        emptyDir: {}
#    - name: data
#      mountPath: /home/postgres/pgdata/partitions
#      targetContainers:
#        - postgres
#      volumeSource:
#        PersistentVolumeClaim:
#          claimName: pvc-postgresql-data-partitions
#          readyOnly: false
#    - name: conf
#      mountPath: /etc/telegraf
#      subPath: telegraf.conf
#      targetContainers:
#        - telegraf-sidecar
#      volumeSource:
#        configMap:
#          name: my-config-map
```