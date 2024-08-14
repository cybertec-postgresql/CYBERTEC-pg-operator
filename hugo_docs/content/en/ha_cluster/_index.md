---
title: "High Availability"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 1100
---
To ensure continiues productive usage you can create a HA-Cluster or modify a Single-Node-Cluster to a HA-Cluster.
The needed changes are less complicated
```
spec: 
  numberOfInstances: 2 # or more
```
The example above will create a HA-Cluster based on two Nodes.
```
kubectl get pods
-----------------------------------------------------------------------------
NAME                             | READY  | STATUS           | RESTARTS | AGE
cluster-1-0                      | 1/1    | Running          | 0        | 54s
cluster-1-1                      | 1/1    | Running          | 0        | 31s

```
