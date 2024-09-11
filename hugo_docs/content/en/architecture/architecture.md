---
title: "Architecture"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 1
---
This chapter covers all important aspects relating to the architecture of CPO and the associated components. In addition to the underlying Kubertnetes, the various components and their interaction for the operation of a PostgreSQL cluster are analysed.

### Brief overview of the components
<div style="text-align:center">
    <img src="images/architecture_overview.png" alt="drawing" align="center" width="60%" />
</div>

Once the operator is employed and a Postgres cluster manifest is submitted, following objects are created:

#### Services
There are atmost two services created by the operator. One is master and the other is replica. Master is the read and write serivce. 
On the other hand, replica is read only service. It is important to connect your application to replica for the read only operations. 
Master service should be used only for the read and write purposes. This helps in load balancing and better availability of the services. 
Each of these services can be reached by their respective endpoints that are available as master and -repl respectively.
    
#### Pod Disruption Budget
The minimum number of instances required is set to 1 by default. It is recommended to use atleast two instances, it helps in efficient pod restart, in case of upgrades, etc.

#### Pods
Another important entity in this setup are the postgres pods. Based on the number of instances field in the cluster manifest, the number of pods are created. There is always one master pod reacheable by the master service. Rest of all the pods are read-only Replica pods which can be reached via -repl service. Note that, if there are more than one replica pods, then which pod will be reached by the service is undefined. Any of the replica pod can be acheived, and they are all identical in all the aspects.

There is always ONLY ONE master pod for every cluster and multiple replicas pods based on the number of instances in the cluster manifest.

E.g. if the number of instances is given 1 in the cluster manifest, then only one master pod will be created. If the number of instances is 2, then there will be 1 master and one replica pod. Similarly, if it is 3, then there will be 1 master and 2 replica pods, and so on.

### Network-Traffic

#### PG-Cluster-internal Traffic
With internal PG cluster-internal traffic, we are talking about all traffic that is necessary for the operation of the cluster itself. This includes 
- Communication for the sync of the replicas:
    - pg_basebackup & streaming replication
- Communication with pgBackRest (if configured)
    - Backups
    - WAL archiving
    - replica-create for new replicas

The figure below shows the internal traffic flows with pgBackRest based on block storage (left) or cloud storage (right)

 <div style="text-align:center">
    <img src="images/architecture_cluster_backup_pvc.png" alt="drawing" width="45%" />
    <img src="images/architecture_cluster_backup_cloud_storage.png" alt="drawing" width="45%" />

</div>   

#### External Traffic

External traffic, i.e. the connection to the database for the user or the application, takes place via defined Kubernetes services. A distinction must be made here between read/write and read only traffic.

##### read/write

##### read-only