---
title: "Apply Ressources"
date: 2024-04-28T14:26:51+01:00
draft: false
weight: 700
---

Kubernetes workloads are often deployed without a direct resource definition. This means that, apart from the limits specified by the administrators, the workloads can use the required resources of the worker node very dynamically. 

The cluster manifest is used to define the Postgres pod resources via the typical resources objects.

There are basically two different definitions:
- `requests`: Basic requirement and guaranteed by the worker node
- `limits`: maximum availability, allocation is increased dynamically if the worker node can provide the resources.

For the planning of the cluster, a proper definition should be carried out in terms of the required hardware, which is then defined as `requests`. These resources are thus guaranteed to the cluster and are taken into account when deploying the pod. Accordingly, a pod can only be deployed on a worker if it can provide these resources. Any limits beyond this are not taken into account when deploying.

The unit of measurement should be taken into account when planning the necessary CPUs: 
cpu specifications are based on millicores
- `1 cpu` corresponds to `1 core`
- `1 core `corresponds to `1000 millicores (m)`
- `1/2 core` corresponds to `500 m`

```
  resources:
    limits:
      cpu: 500m
      memory: 1Gi
    requests:
      cpu: 1000m
      memory: 1Gi
```

This example corresponds to a guaranteed availability of half a core and 1 Gibibyte. However, if necessary and available, the container can use up to one core. The allocation takes place dynamically and for the required time.

Pods can be categorised into three Quality of Services (QoS) based on the defined information on the resources. 
    
- `Best-Effort`: The containers of a pod have no resource information
- `Burstable`: A container of the pod has a memory or CPU `requests` or `limits`.  
- `Guaranteed`: Each container of a pod has both cpu and memory `requests` and `limits`. In addition, the details of the respective `limits` correspond to the `requests` details 

If you would like more information and explanations, you can take a look at the [Kubernetes documentation on QoS](https://kubernetes.io/docs/tasks/configure-pod-container/quality-service-pod/#qos-classes).

We generally recommend using the Guaranteed Status for PostgreSQL workloads. However, many users very successfully use the deviation of the CPU limit by factors such as 2. 
For example: 
```
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 2000m
      memory: 1Gi
```
This is intended to create the possibility of additional CPU resources for sudden load peaks. 

> **_HINT:_**  The use of burstable definitions does not release you from a correct resource calculation, as `limits` resources are not guaranteed and therefore an undersupply can occur if the requests are not properly defined.