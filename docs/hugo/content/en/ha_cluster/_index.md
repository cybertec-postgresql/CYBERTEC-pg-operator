---
title: "High Availability"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 1100
---

High availability (HA) is a critical aspect of running database systems, especially in mission-critical applications where downtime is unacceptable. This section explains why high availability is important for PostgreSQL and how Patroni acts as a solution to ensure HA.
Why High Availability (HA) for PostgreSQL?
1. To minimise downtime: In modern, data-driven applications, downtime can cause significant financial and reputational losses. High availability ensures that the database remains available even in the event of hardware failures or network problems.
2. Data integrity and security: A database failure can lead to data loss or data inconsistencies. High-availability solutions protect against such scenarios through continuous data replication and automatic failover.
3. Scalability and load balancing: HA setups make it possible to distribute the load across multiple nodes, resulting in better performance and faster response times. This is particularly important in environments with high data traffic.
4. Ease of maintenance: By setting up high availability, database maintenance can be performed without interrupting services. Nodes can be maintained incrementally while the database remains available.

#### Patroni - the cluster manager
In our PostgreSQL environment, we use [Patroni](../../patroni) in the PG containers by default. This has the advantage that even single-node instances basically function as Patroni clusters. This configuration offers several important advantages:
- Easy scalability: by using Patroni in all PG containers, scaling pods up and down is possible at any time. You can easily add additional pods as needed to improve performance or increase capacity, or remove pods to free up resources. This flexibility is particularly useful in dynamic environments where requirements can change quickly.
- Automated cluster management: Patroni automatically takes over the management of the cluster. When a new pod is added to an existing cluster, Patroni takes care of setting up the new node itself, including initialising and starting replication. This means you don't have to perform any manual steps to configure or manage new nodes - Patroni does it all for you automatically.
- Seamless integration: As Patroni is active in every PG container by default, you don't have to worry about compatibility or manual configuration. This makes deployment and maintenance much easier, as all the necessary components are already preconfigured.
- Optimisation of resources: Even with a minimal setup (single-node instance), you benefit from the advantages of a Patroni cluster, including the possibility of easy expansion and automatic failover in the event of a failure. This ensures optimal resource utilisation and minimises downtime.

#### Upgrade the cluster to high availability 

The necessary changes to a high-availability cluster are very limited. 
Only the number of desired instances needs to be increased. 

```
apiVersion: cpo.opensource.cybertec.at/v1
kind: postgresql
metadata:
  name: cluster-1
spec:
  dockerImage: "docker.io/cybertecpostgresql/cybertec-pg-container:postgres-18.0-1"
  numberOfInstances: 2
  postgresql:
    version: "18"
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

The example above will create a HA-Cluster based on two Nodes.
```
kubectl get pods
-----------------------------------------------------------------------------
NAME                             | READY  | STATUS           | RESTARTS | AGE
cluster-1-0                      | 1/1    | Running          | 0        | 3d
cluster-1-1                      | 1/1    | Running          | 0        | 31s

```
