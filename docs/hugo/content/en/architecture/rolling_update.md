---
title: "Rolling-Updates"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 3
---

Whether updating the minor version, changing the hardware definitions of the cluster or other adjustments that require a pod restart, CPO ensures that the update is as uninterrupted as possible. 

This means that adjustments are carried out on the various pods of a particular cluster one after the other and in a sensible sequence. This happens as soon as a cluster consists of more than 1 PostgreSQL node. 

In the event of a necessary restart, the operator independently stops the pods and does not leave this to Kubernetes. The idea behind this is that all replica pods are restarted one after the other first. The operator recognises these by the label cpo.opensource.cybertec.at/role=replica set by Patroni

As soon as all replicas are ready again, the operator checks whether one of the replicas is able to take over cluster operation and performs a switchover. Only then is the former leader pod stopped and restarted. 

This ensures that the only effect on the application is the switchover.
{{< hint type=info >}} A completely uninterrupted handover of operation is not possible due to the architecture and connection handling of PostgreSQL. {{< /hint >}}