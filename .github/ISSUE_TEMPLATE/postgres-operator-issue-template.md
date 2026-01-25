---
name: CYBERTEC-PG-Operator issue template
about: How are you using the operator?
title: '[BUG] Brief description of the problem'
labels: ''
assignees: ''

---

To quickly narrow down the problem, we need details about your infrastructure.

- **Operator version / image:** [e.g. docker.io/cybertecpostgresql/cybertec-pg-operator:v0.9.0-1]
- **Postgres Docker Image:** [e.g., containers.cybertec.at/cybertec-pg-container/postgres:rocky9-18.1-1]
- **Kubernetes Platform:** [Please check or add]
    - [ ] Vanilla Kubernetes (Bare Metal / VM)
    - [ ] OpenShift
    - [ ] Rancher / RKE / RKE2
    - [ ] AWS EKS
    - [ ] Google GKE
    - [ ] Azure AKS
    - [ ] Minikube / Kind / K3s
    - [ ] Other: ________
- **Kubernetes version:** [Output of `kubectl version`, e.g. v1.28.3]
- **Storage Class:** [e.g., standard, gp2, longhorn, ceph-rbd]
- **Is the operator running in production?** [Yes / No]

---

### Problem Description

**What happened?**
Describe the unexpected behavior.

**What did you expect?**
Describe what should have happened.

---

### Steps to reproduce

How can we reproduce the problem?
1. Applied Postgres cluster manifest (see below)
2. Executed command: `...`
3. Error occurs when ...

---

### Relevant logs and manifests

**IMPORTANT:** Please use code blocks (```) to format logs and YAML files in a readable way.

**1. Operator logs:**
(Please check the logs of the operator pod for errors)
```text
Insert logs here...
```

**2. Postgres / Patroni /pgBackRest logs: (If the pod starts but the DB is not running):**
(Please check the logs of the postgres or pgbackrest pod for errors)
```text
Insert logs here...
```

**3. Postgres Cluster Manifest (YAML): (Please remove passwords or sensitive data!)**
```yaml
Insert Manifest here...
```

**Additional information**
Are there any special circumstances in your setup? (e.g., service mesh such as Istio/Linkerd, special network policies, air-gapped environment?)