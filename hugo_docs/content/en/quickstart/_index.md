---
title: "Quickstart"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 400
---

We can tell and document so much about our project but it seems you just want to get started. Let us show you the fastest way to use CPO.

## Preconditions

- git
- helm (optional)
- kubectl or oc

## Let's start

### Step 1 - Preparations
To get started, you can fork our tutorial repository on Github and then download it.
[CYBERTEC-operator-tutorials](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/fork)

```
git clone https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials.git
cd CYBERTEC-operator-tutorials
```

### Step 2 - Install the Operator
Two options are available for the installation: 
- Installation via Helm-Chart
- Installation via apply

#### Installation via Helm-Chart
```
kubectl apply -k setup/namespace/.
helm install cpo setup/helm/operator/ -n cpo
```

#### Installation via apply
```
kubectl apply -k setup/namespace/.
kubectl apply -k setup/helm/operator/. -n cpo
```

You can check if the operator pod is in operation.
```
kubectl get pods -n cpo --selector=cpo.cybertec.at/pod/type=postgres-operator
```
The result should look like this:
```
NAME                                 READY   STATUS    RESTARTS   AGE
postgres-operator-599688d948-fw8pw   1/1     Running   0          41s
```

The operator is ready and the setup is complete. The next step is the creation of a Postgres cluster

### Step 3 - Create a Cluster
To create a simple cluster, the following command is sufficient
```
kubectl apply -f cluster-tutorials/single-cluster
```

```
watch kubectl get pods --selector cluster-name=cluster-1
```
The result should look like this:
```
Alle 2.0s: kubectl get pods --selector cluster-name=cluster-1                                                                                                            

NAME          READY   STATUS            RESTARTS   AGE
cluster-1-0   2/2     Running           0          28s
cluster-1-1   0/2     PodInitializing   0          9s
```

### Step 4 - Connect to the Database
Get your login information from the secret.
```
kubectl get secret postgres.cluster-1.credentials.postgresql.cpo.opensource.cybertec.at -o jsonpath='{.data}' | jq '.|map_values(@base64d)'
```
The result should look like this:
```
{
  "password": "2rZG1Kx9asdHscswQGzff4Ru0xW6uasacy3GQ0sjdCH3wWr0kguUXUZek6dkemsf",
  "username": "postgres"
}
```
#### Connection via port-forward

```
kubectl port-forward cluster-1-0 5432:5432
```

```
# using psql
PGPASSWORD=2rZG1Kx9asdHscswQGzffjdCH3wWr0kguUXUZek6dkemsf psql -h 127.0.0.1 -p 5432 -U postgres

# using usql
PGPASSWORD=2rZG1Kx9asdHscswQGzffjdCH3wWr0kguUXUZek6dkemsf usql postgresql://postgres@127.0.0.1/postgres
```

## Next Steps
Congratulations, your first cluster is ready and you were able to connect to it. On the following pages we have put together an introduction with lots of information and details to show you the different possibilities and components of CPO. 