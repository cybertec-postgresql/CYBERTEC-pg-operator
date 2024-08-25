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

### Network-Traffic


#### PG-Cluster-intern Traffic
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