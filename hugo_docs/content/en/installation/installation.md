---
title: "Install CPO"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 502
---

## Prerequisites

For the installation you either need our CPO tutorial repository or you install CPO directly from our registry.<br>
Exception: Installation via Operatorhub (Openshift only)

### CPO-Tutorial-Repository

To get started, you can fork our tutorial repository on Github and then download it.
[CYBERTEC-operator-tutorials](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials/fork)

```
GITHUB_USER='[YOUR_USERNAME]'
git clone https://github.com/$GITHUB_USER/CYBERTEC-operator-tutorials.git
cd CYBERTEC-operator-tutorials
```

### CPO-Registry


### Create Namespace

```
# kubectl
kubectl create namespace cpo

# oc
oc create namespace cpo
```

## Install CPO

There are several ways to install CPO:
- [Use Helm](#helm)
- [Use apply](#apply)
- [Use Operatorhub (On Openshift only)](#operatorhub)

### Helm

You can check and change the value.yaml of the helm diagram under the path helm/operator/values.yaml
By default, the operator is defined so that it is configured via crd-configuration. If you wish, you can change this to configmap. There are also some other default settings.

```
helm install -n cpo cpo helm/operator/.
```

The installation uses a standard configuration. On the following page you will find more information on how to [configure cpo](/documentation/how-to-use/configuration/) and thus adapt it to your requirements.

### Apply

The installation uses a standard configuration. On the following page you will find more information on how to [configure cpo](/documentation/how-to-use/configuration/) and thus adapt it to your requirements.

### Operatorhub

The installation uses a standard configuration. On the following page you will find more information on how to [configure cpo](/documentation/how-to-use/configuration/) and thus adapt it to your requirements.