---
title: "Storage"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 32050
---
Storage is crucial for the performance of a database and is therefore a central element. As with systems based on bare metal or virtual machines, the same requirements apply to Kubernetes workloads, such as constant availability, good performance, consistency and durability. 

A basic distinction is made between local storage, which is directly connected to the worker node, and network storage, which is mounted on the worker node and thus made available to the pod. 

In probably the vast majority of Kubernetes systems, network storage is used, for example from systems from hyperscalers or other cloud providers or own systems such as CEPH. 

With network storage in particular, attention must be paid to performance in terms of throughput (speed and guaranteed IOPS) and, above all, latency. It is also important to ensure that the different volumes do not compete with each other in terms of load.

> **_PAY ATTENTION:_**  Before using a CPO cluster, make sure that the storage is suitable for the intended use and provides the necessary performance. In addition, check the storage with benchmarks before use. We recommend the use of [pgbench](https://www.postgresql.org/docs/current/pgbench.html) for this purpose.

## Define Storage-Volume

The storage is defined via the volume object and enables the size and storage class for the storage to be defined, among other things. 
```
spec:
  volume:
    size: 5Gi
    storageClass: default-provisioner
  ...
```

The volume is currently used for both PG and WAL data. In future, there will be an optional option to create a separate WAL volume.
Please check our [roadmap](roadmap)

> **_PAY ATTENTION:_**  Please ensure, that the storageClass exists and is usable. If a Volume cannot provide the Volume will stand in the pending-State as like the Database-Pod.

The volume is currently used for both PG and WAL data. In future, there will be an optional option to create a separate WAL volume.

## Expanding Volume

> **_HINT:_**  Kubernetes is able to forward requests to expand the storage to the storage system and enable the expand without the need to restart the container. However, this also requires the associated storage system and the driver used to support this. This information can be found in the storage class under the object: allowVolumeExpansion. A distinction must also be made between online and offline expand. The latter requires a restart of the pod. To do this, the pod must be deleted manually.

To Expand the Volume, the value of the object volume.size must be increased
```
spec:
  volume:
    size: 10Gi
    storageClass: default-provisioner
  ...
```
This will trigger the expand of your Cluster-Volumes. It will need some time and you can check the current state inside the pvc.
```
kubectl get pvc pgdata-cluster-1-0 -o yaml
-------------------------------------------------------
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: crc-csi-hostpath-provisioner
  volumeMode: Filesystem
  volumeName: pvc-800d7ecc-2d5f-4ef4-af83-1cd94c766d37
status:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 5Gi
  phase: Bound
```

## Creating additonal Volumes
The Operator allows you to modify your cluster with additonal Volumes.
```
spec:
  ...
  additionalVolumes:
    - name: empty
      mountPath: /opt/empty
      targetContainers:
        - all
      volumeSource:
        emptyDir: {}
```
This example will create an emptyDir and mount it to all Containers inside the Database-Pod.


## Specific Settings for aws gp3 Storage
For the gp3 Storage aws you can define more informations 
```
  volume:
    size: 1Gi
    storageClass: gp3
    iops: 1000  # for EBS gp3
    throughput: 250  # in MB/s for EBS gp3

```
The defined IOPS and Throughput will include in the PersistentVolumeClaim and send to the storage-Provisioner.
Please keep in Mind, that on aws there is a CoolDown-Time as a limitation defined. For new Changes you need to wait 6 hours. 
Please also ensure to check the default and allowed values for IOPS and Throughput [AWS docs](https://aws.amazon.com/ebs/general-purpose/).

To ensure that the settings are updates properly please define the Operator-Configuration 'storage_resize_mode' from default to 'mixed'
