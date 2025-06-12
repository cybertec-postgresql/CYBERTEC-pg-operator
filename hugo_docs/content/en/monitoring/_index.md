---
title: "Monitoring"
date: 2023-12-28T14:26:51+01:00
draft: false
weight: 2000
---
The CPO-Project has prepared severall Tools which allows to setup a Monitoring-Stack including Alerting and Metric-Viewer.
These Stack is based on: 
- Prometheus
- Alertmanager
- Grafana
- exporter-container

CPO has prepared an own Exporter for the PostgreSQl-Pod which can used as a sidecar.

#### Setting up the Monitoring Stack
To setup the Monitoring-Stack we suggest that you create an own namespace and use the prepared kustomization file inside the Operator-Tutorials.
```
$ kubectl create namespace cpo-monitoring
namespace/cpo-monitoring created
$ kubectl get pods -n cpo-monitoring
No resources found in cpo-monitoring namespace.

git clone https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorial
cd CYBERTEC-operator-tutorial/setup/monitoring

# Hint: Please check if youn want to use a specific storage-class the file pvcs.yaml and add your storageclass on the commented part. Please ensure that you removed the comment-char.

$ kubectl apply -n cpo-monitoring -k .
serviceaccount/cpo-monitoring created
serviceaccount/cpo-monitoring-tools created
clusterrole.rbac.authorization.k8s.io/cpo-monitoring unchanged
clusterrolebinding.rbac.authorization.k8s.io/cpo-monitoring unchanged
configmap/alertmanager-config created
configmap/alertmanager-rules-config created
configmap/cpo-prometheus-cm created
configmap/grafana-dashboards created
configmap/grafana-datasources created
secret/grafana-secret created
service/cpo-monitoring-alertmanager created
service/cpo-monitoring-grafana created
service/cpo-monitoring-prometheus created
persistentvolumeclaim/alertmanager-pvc created
persistentvolumeclaim/grafana-pvc created
persistentvolumeclaim/prometheus-pvc created
deployment.apps/cpo-monitoring-alertmanager created
deployment.apps/cpo-monitoring-grafana created
deployment.apps/cpo-monitoring-prometheus created

Hint: If you're not running Openshift you will get a error like this: 
error: resource mapping not found for name: "grafana" namespace: "" from ".":
no matches for kind "Route" in version "route.openshift.io/v1" ensure CRDs are installed first

You can ignore this, because it depends on an object with the type route which is part of Openshift. 
It is not needed replaced by ingress-rules or an loadbalancer-service.
```

After installing the Monitoring-Stack we're able to check the created pods inside the namespace
```
$ kubectl get pods -n cpo-monitoring
----------------------------------------------------------------------------------------
NAME                                          | READY  | STATUS      | RESTARTS  | AGE
cpo-monitoring-alertmanager-5bb8bc79f7-8pdv4  | 1/1    | Running     | 0         | 3m35s
cpo-monitoring-grafana-7c7c4f787b-jbj2f       | 1/1    | Running     | 0         | 3m35s
cpo-monitoring-prometheus-67969b757f-k26jd    | 1/1    | Running     | 0         | 3m35s

```
The configuration of this monitoring-stack is based on severall configmaps which can be modified.

#### Prometheus-Configuration


#### Alertmanager-Configuration


#### Grafana-Configuration


#### Configure a PostgreSQL-Cluster to allow Prometheus to gather metrics

To allow Prometheus to gather metrics from your cluster you need to do some small modfications on the Cluster-Manifest.
We need to create the monitor-object for this:
```
kubectl edit postgresqls.cpo.opensource.cybertec.at cluster-1

...
spec:
  ...
  monitor:
    image: docker.io/cybertecpostgresql/cybertec-pg-container:exporter-16.2-1
```

The Operator will add automatically the monitoring sidecar to your pods, create a new postgres-user and add some structure inside the postgres-database to enable everthing needed for the Monitoring. Also every Ressource of your Cluster will get a new label: cpo_monitoring_stack=true. This is needed for Prometheus to identify all clusters which should be added to the monitoring.
Removing this label will stop Prometheus to gather data from this cluster.

After changing your Cluster-Manifest the Pods needs to be recreated which is done by a rolling update. 
After this you can see that the pod has now more than just one container.

```
kubectl get pods
-----------------------------------------------------------------------------
NAME                             | READY  | STATUS           | RESTARTS | AGE
cluster-1-0                      | 2/2    | Running          | 0        | 54s
cluster-1-1                      | 2/2    | Running          | 0        | 31s

```
You can check the logs to see that the exporter is working and with curl you can see the output of the exporter.

```
kubectl logs cluster-1-0 -c postgres-exporter
kubectl exec --stdin --tty cluster-1-0 -c postgres-exporter -- /bin/bash
[exporter@cluster-1-0 /]# curl http://127.0.0.1:9187/metrics

```
You can now setup a LoadBalancer-Service or create an Ingress-Rule to allow access von outside to the grafana. Alternativ you can use a port-forward.

##### LoadBalancer or Nodeport

##### Ingress-Rule

##### Port-Forwarding
```
$ kubectl get pods -n cpo-monitoring
----------------------------------------------------------------------------------------
NAME                                          | READY  | STATUS      | RESTARTS  | AGE
cpo-monitoring-alertmanager-5bb8bc79f7-8pdv4  | 1/1    | Running     | 0         | 6m42s
cpo-monitoring-grafana-7c7c4f787b-jbj2f       | 1/1    | Running     | 0         | 6m42s
cpo-monitoring-prometheus-67969b757f-k26jd    | 1/1    | Running     | 0         | 6m42s

$ kubectl port-forward cpo-monitoring-grafana-7c7c4f787b-jbj2f -n cpo-monitoring 9000:9000
Forwarding from 127.0.0.1:9000 -> 9000
Forwarding from [::1]:9000 -> 9000

```
Call http://localhost:9000 in the [Browser](http://localhost:9000)

##### Use a Route (Openshift only)

```
kubectl get route -n cpo-monitoring

```
Use the Route-Adress to access Grafana