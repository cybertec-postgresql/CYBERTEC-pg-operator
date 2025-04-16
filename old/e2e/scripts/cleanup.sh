#!/bin/bash
kubectl delete postgresql acid-minimal-cluster
kubectl delete deployments -l application=db-connection-pooler,cluster.cpo.opensource.cybertec.at/name=acid-minimal-cluster
kubectl delete statefulsets -l application=cpo,cluster.cpo.opensource.cybertec.at/name=acid-minimal-cluster
kubectl delete services -l application=cpo,cluster.cpo.opensource.cybertec.at/name=acid-minimal-cluster
kubectl delete configmap postgres-operator
kubectl delete deployment postgres-operator