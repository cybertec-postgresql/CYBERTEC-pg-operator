---
title: "Modify Cluster"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 32070
---
Starting with the Single-Node-Cluster from the previous section, we want to modify the Instance a bit to see. 
## CPU and Memory
```
spec:
  resources:
    limits:
      cpu: 1000m
      memory: 500Mi
    requests:
      cpu: 500m
      memory: 500mi
```
Based on the ressources-Definiton we're able to modify the reserved Hardware (requests) and the limits, which allows use to consume more than the reserved definitons if the k8s-worker has this hardware available. There are some Restrictions when modifiying the limits-section. Because of the behaviour of Databases we should never define a diff between requests.memory and limits.memory. A Database is after some time using all available Memory, for Cache and other things. Limits are optional and the worker node can force them back. forcing back memory will create big problems inside a database like creating corruption, forcing OutOfMemory-Killer and so on.
CPU on the other side is a ressource we can use inside the limits definiton to allow our database using more cpu if needed and available.

## Sidecars
Sidecars are further Containers running on the same Pod as the Database. We can use them for serveral different Jobs.
The Operator allows us to define them directly inside the Cluster-Manifest.
```
spec:
  sidecars:
   - name: "telegraf-sidecar"
     image: "telegraf:latest"
     ports:
     - name: metrics
       containerPort: 8094
       protocol: TCP
     resources:
       limits:
         cpu: 500m
         memory: 500Mi
       requests:
         cpu: 100m
         memory: 100Mi
     env:
       - name: "USEFUL_VAR"
         value: "perhaps-true"
```
This Example will add a second Container to our Pods. This will trigger a restart, which creates Downtime if you're not running a HA-Cluster.

## Init-Containers
We can exactly the same as for sidecars also for Init-Containers. 
The difference is, that a sidecar is running normally on a pod. 
An Init-Container will just run as first container when the pod is created and it will ends after his job is done. 
The "normal" Containers has to wait till all init-Containers finished their jobs and ended with a exit-status.
```
spec:
  initContainers:
  - name: date
    image: busybox
    command: [ "/bin/date" ]
```

## TLS-Certificates
One Startup the Containers will create a custom TLS-Certificate which allows creating tls-secured-connections to the Database.
But this Certificates cannot verified, because the application has no information about the CA. Because of this the certificates are no protection against MITM-Attacks. 
You're able to configure your own Certificates and CA to ensure, that you can use secured and verified connections between your application and your database. 
```
spec:
  tls:
    secretName: ""  # should correspond to a Kubernetes Secret resource to load
    certificateFile: "tls.crt"
    privateKeyFile: "tls.key"
    caFile: ""  # optionally configure Postgres with a CA certificate
    caSecretName: "" # optionally the ca.crt can come from this secret instead.
```
You need to store the needed values from tls.crt, tls.key and ca.crt in a secret and define the secrtetname inside the tls-object. 
if you want you can create a separate sercet just for the ca and use this secret for every cluster inside the Namespace. 
To get Information about creating Certificates and the secrets check the Tutorial in the additonal-Section or click [here](additonal/tutorials/tls)

## Node-Affinity
Node-Affinity will ensure that the Cluster-pods only deployed on Kubernetes-Nodes which has the defined Labelkey and -Value
```
spec:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: cpo
              operator: In
              values:
                - enabled
```
This allowes you to use specific database-nodes in a mixed cluster for example. 
In the Example above the Cluster-Pods are just deployed on Nodes with the Key: cpo and the value: enabled
So you're able to seperate your Workload. 

## PostgreSQL-Configuration
Every Cluster will start with the default PostgreSQL-Configuration. Every Parameter can be overriden based in definitions inside the Cluster-Manifest. 
Therefore we just need a add the section parameters to the postgresql-Object
```
spec:
  postgresql:
    version: 16
    parameters:
      max_connections: "53"
      log_statement: "all"
      track_io_timing: "true"
```
These Definitions will change the PostgreSQL-Configuration. Based on the needs of Parameter changes the Pods may needs a restart, which creates a Downtime if its not a HA-Cluster.
You can check Parameters and allowed Values on this Sources to ensure a correct Value.
- PostgreSQL Documentation
- [PostgreSQL.org](https://postgresql.org)
- [PostgreSQLco.nf](https://postgresqlco.nf/)
