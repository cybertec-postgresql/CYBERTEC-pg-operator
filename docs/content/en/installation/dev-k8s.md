---
title: "Setup local Kubernetes"
date: 2023-03-07T14:26:51+01:00
draft: false
weight: 501
---

There are various options for setting up a local Kubernetes environment. This chapter deals with the following two variants:
- minikube
- crc (CodeReadyContainers from RedHat)

### Minikube
Minikube is a tool that makes it possible to run Kubernetes locally on a single computer. It sets up a minimal but functional Kubernetes environment suitable for development and testing purposes. Minikube supports most Kubernetes features and provides an easy way to launch and manage Kubernetes clusters on local machines without the need for a complex cloud infrastructure.

#### Install Kubectl & Minikube
To use Minikube, it is essential to install the Kubectl client.

[Here](https://kubernetes.io/docs/tasks/tools/) you will find all the information you need to install kubectl on your Linux, Mac or Windows device.

You can Install Minikube on your Linux-, Mac- or Windows-Devide using [this Documentation](https://minikube.sigs.k8s.io/docs/start/?arch=%2Flinux%2Fx86-64%2Fstable%2Fbinary+download).

#### Use Minikube

Before starting minikube, it is advisable to define a path for the kubeconfig.
```bash
export KUBECONFIG=/home/USERNAME/kubeconfig_minikube.conf
```
You can then start minikube and all the necessary data is written directly to the conf. The definition of a user-defined path ensures that other configs are not inadvertently overwritten. 
The path must be defined again via ENV in each new user session. Alternatively, this can also be permanently defined via .bashrc. 
If the default path is not used for any other purpose, the ENV does not need to be set. 
```bash
# Start minikube
minikube start

# get pods from default namespace
kubectl get pods

# change default namespace to cpo
kubectl config set-context --namespace=cpo
```

### CRC
CRC (CodeReady Containers) is a tool from Red Hat that provides a local OpenShift environment. It is specifically designed to run a compact version of OpenShift on a local machine to provide developers and testers with an easy way to develop and test applications optimised for use in OpenShift. CRC includes all the necessary OpenShift components and makes it possible to use Red Hat's container platform locally without building a full cloud infrastructure.

#### Install oc-client & CRC
To use CRC, it is essential to install the oc-client or the kubectl-client.

[Here](https://docs.openshift.com/container-platform/latest/cli_reference/openshift_cli/getting-started-cli.html) you will find all the information you need to install kubectl on your Linux, Mac or Windows device.

You can Download and install CRC on your Linux-, Mac- or Windows-Devide using [this informations](https://developers.redhat.com/products/openshift-local/overview).

#### Use CRC

Before installing crc, it is advisable to define a path for the kubeconfig.
```bash
export KUBECONFIG=/home/USERNAME/kubeconfig_crc.conf
```
You can then install and start crc and all the necessary data is written directly to the conf. The definition of a user-defined path ensures that other configs are not inadvertently overwritten. 
The path must be defined again via ENV in each new user session. Alternatively, this can also be permanently defined via .bashrc. 
If the default path is not used for any other purpose, the ENV does not need to be set. 
```bash
# Install crc
crc setup

# Start crc
crc start

# get pods from default namespace
oc get pods

# change default namespace to cpo
oc project cpo
```