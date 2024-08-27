---
title: "Create a Cluster"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 600
---

To set up a cluster, the implementation is based on a description, as with the other Kubernetes deplyoments. To do this, the operator uses a document of type `postgresql`.

You can also find the basic minimum specifications for a single-node cluster in our [tutorial project on Github](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/blob/main/cluster-tutorials/single-cluster/postgres.yaml)

## minimal Single-Node Cluster
```
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-1
spec:
  dockerImage: "docker.io/cybertecpostgresql/cybertec-pg-container:postgres-16.1-6-dev"
  numberOfInstances: 1
  postgresql:
    version: "16"
  resources:
    limits:
      cpu: 500m
      memory: 500Mi
    requests:
      cpu: 500m
      memory: 500Mi
  volume:
    size: 5Gi 
```
Based on this Manifest the Operator will deploy a single-Node-Cluster based on the defined dockerImage and start the included Postgres-16-Server. 
Also created is a volume based on your default-storage Class. The Ressource-Definiton means, that we reserve a half cpu and a half GB Memory for this Cluster with the same Definition as limit.

After some seconds we should see, that the operator creates our cluster based on the declared definitions.
```
kubectl get pods
-----------------------------------------------------------------------------
NAME                             | READY  | STATUS           | RESTARTS | AGE
cluster-1-0                      | 1/1    | Running          | 0        | 50s

```

> **_HINT:_**  [Here](documentation/crd/crd-postgresql/) you will find a complete overview of the available options within the cluster manifest.

## minimal High-Availability Cluster

No more effort is required to create a High-Availablity cluster than for a Single-Node Cluster. Only the `Cluster-Manifest` needs to be modified slightly. 
The difference lies in the object numberOfInstances, which must be set > 1.

You can also find the basic minimum specifications for a High-Availability-Cluster cluster in our [tutorial project on Github](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/blob/main/cluster-tutorials/high-availability-cluster/ha-postgres.yaml)


```
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-1
spec:
  dockerImage: "docker.io/cybertecpostgresql/cybertec-pg-container:postgres-16.1-6-dev"
  numberOfInstances: 2
  postgresql:
    version: "16"
  resources:
    limits:
      cpu: 500m
      memory: 500Mi
    requests:
      cpu: 500m
      memory: 500Mi
  volume:
    size: 5Gi 
```

You can either create a new cluster with the document or update an existing cluster with it. 
This makes it possible to scale the cluster up and down during operation.

```
kubectl get pods
-----------------------------------------------------------------------------
NAME                             | READY  | STATUS           | RESTARTS | AGE
cluster-1-0                      | 1/1    | Running          | 0        | 2m12s
cluster-1-1                      | 1/1    | Running          | 0        | 50s

```


<!-- 
We can now starting to modify our cluster with some more Definitons. 
### Use a specific Storageclass
```
spec:
  ...
  volume:
    size: 5Gi
    storageClass: default-provisioner
  ...
```
Using the storageClass-Definiton allows us to define a specific storageClass for this Cluster. Please ensure, that the storageClass exists and is usable. If a Volume cannot provide the Volume will stand in the pending-State as like the Database-Pod.

### Expanding Volume
The Operator allows to you expand your volume if the storage-System is able to do this. 
```
spec:
  ...
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

### Creating additonal Volumes
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


### Specific Settings for aws gp3 Storage
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

To ensure that the settings are updates properly please define the Operator-Configuration 'storage_resize_mode' from default to 'mixed' -->