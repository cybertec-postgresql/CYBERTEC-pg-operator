# CYBERTEC PostgreSQL Operator for Kubernetes: Complete Guide

## Part 1: Choosing a PostgreSQL Operator for Kubernetes

There are several good open-source PostgreSQL operators. This is an honest evaluation framework, the questions to ask of any of them, and where the **CYBERTEC PG Operator (CPO)** focuses rather than a scorecard. Operator features move fast; verify each project's current capabilities at evaluation time.

### Main Open-Source PostgreSQL Operators
* **CYBERTEC PG Operator (CPO)**
* **CloudNativePG**
* **Zalando postgres-operator**
* **Crunchy PGO**
* **StackGres**

All are credible; they differ in HA mechanism, cross-cluster story, ecosystem, and support model.

---

### The Dimensions That Matter

* **High-availability mechanism:** How is the leader elected and failover performed? CPO uses Patroni, the widely-deployed, battle-tested HA layer for PostgreSQL. *Ask:* How mature is the failover logic, and how is split-brain prevented?
* **Cross-cluster / multi-site DR:** Can it fail across clusters (regions/data centres) automatically, not just across nodes? This is where CPO invests heavily: multi-site with an external quorum and automatic cross-site failover and failback. Many operators stop at in-cluster HA or one-way standby.
* **Day-2 operations:** Rolling resize/reconfigure, scaling, minor and major version upgrades, backups & PITR, connection pooling, monitoring—how much is automated vs. DIY?
* **Backups & recovery:** Which backup tool, what object stores, and how simple is point-in-time recovery and restore-to-new-cluster?
* **Configuration model & CRDs:** Is everything declarative? How readable is the CRD, and how much PostgreSQL can you actually tune through it?
* **Footprint & dependencies:** External etcd/consul vs. Kubernetes-native DCS; sidecars; resource overhead.
* **License & governance:** Open-source license, who governs the project, and how active it is.
* **Commercial support:** Can you buy 24×7 support with real SLAs from a team with deep PostgreSQL expertise, and will they fix issues at the source?

> **Key Strength:** CPO is backed by CYBERTEC, a dedicated PostgreSQL company with experts who actively contribute to PostgreSQL and provide executive-level support.

---

### Where CPO Fits

* **Patroni-based HA:** Proven reliability with a strong multi-site story (validated: automatic cross-site failover in ≈ 90s, clean automatic failback in ≈ 20s).
* **Open Source:** No lock-in, deployable on any Kubernetes or OpenShift cluster.
* **Expert Backing:** 24×7 support, consulting, and training directly from contributors to PostgreSQL and Patroni.

It is a particularly good fit when **cross-site / regional disaster recovery is a hard requirement** and when you want a commercial support partner behind your open-source database.

---

### How to Evaluate Fairly

1. **Run a spike:** Execute the same workload across two or three operators.
2. **Test real failover:** Kill the leader or simulate a site outage to measure real RTO/RPO. *(CPO's tutorials and multi-site scaffold simplify this setup).*
3. **Exercise recovery:** Test point-in-time restores, not just backup creation.
4. **Evaluate actual support costs:** Price the support tier you would realistically purchase, rather than relying solely on free tier capabilities.

---

### Where to Go Next
* **Whitepaper & Architecture:** Check the multi-site operator whitepaper in the resources section.
* **Source Code:** [CYBERTEC PG Operator GitHub Repository](https://github.com/cybertec-postgresql/CYBERTEC-pg-operator)
* **Documentation:** [CYBERTEC PG Operator Official Docs](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/)
* **Commercial Support:** Visit CYBERTEC 24×7 support services to resolve enterprise queries.

<br>

---

## Part 2: How the CYBERTEC PG Operator Works

A concept guide for engineers who want to understand what's happening under the YAML, the reconcile loop, where high availability actually lives, and which component is responsible for what.

---

### The Mental Model: Desired State, Reconciled Continuously

The operator is a Kubernetes controller. You don't tell it how to build a cluster; you declare what you want as a `postgresql` custom resource, and the operator continuously drives reality toward that declaration.

#### The Loop, Concretely

1. **Apply Manifest:** You `kubectl apply` a `postgresql` resource (desired state), or edit an existing one.
2. **Reconcile:** The operator's informer wakes on the change and reconciles: it diffs desired vs. actual and creates/updates the StatefulSet, Services, Secrets, ConfigMaps, and (if configured) the pgBackRest repo host and backup CronJobs.
3. **Continuous Watching:** It keeps watching. Drift, a deleted Service, a changed replica count, or a new parameter is corrected on the next sync. This is why *"change the manifest, let the operator reconcile"* is the answer to almost every day-2 task.

---

### Two Layers of High Availability and Who Owns Each

This is the single most important thing to internalise: **Kubernetes and Patroni protect different things.**

| Concern | Owner | What it does |
| :--- | :--- | :--- |
| **Pod/process liveness** | Kubernetes (StatefulSet) | Recreates a crashed or evicted database pod, even if the operator is down. |
| **Which pod is primary, replication, failover** | Patroni (inside every DB pod) | Holds a leader lease in a DCS, streams WAL, promotes a replica on leader loss. |
| **Cluster shape, config, lifecycle** | The operator | Turns your manifest into the right pods/Services/Secrets and reconciles changes. |

Because Patroni runs inside the database pods and uses the Kubernetes API as its distributed configuration store (DCS), **failover does not depend on the operator being up**. The operator builds and evolves the cluster; Patroni keeps PostgreSQL correct moment-to-moment. That separation is why an operator restart never risks your data path.

---

### What the Operator Creates per Cluster

* **StatefulSet:** Database pods (`<cluster>-0`, `-1`, …), each running PostgreSQL + Patroni.
* **Services (3x):**
  * `<cluster>` (always points at the current leader, for writes)
  * `<cluster>-repl` (replicas, for read-only)
  * `<cluster>-clusterpods` (headless)
* **Secrets:** Generated credentials (superuser, replication, application users).
* **Labels:** Tracks state across failovers:
  * `cluster.cpo.opensource.cybertec.at/name=<cluster>` on pods
  * `member.cpo.opensource.cybertec.at/role=master|replica` kept in sync with Patroni's view (ensures leader/replica Services route correctly).
* **Backups (if configured):** A pgBackRest repo-host pod and backup CronJobs.

---

### Day-to-Day Operations, Mechanically

Every operation follows the same pattern: an edit to the `postgresql` resource that the operator turns into a safe, rolling change:

* **Resize / Reconfigure:** Update `spec.resources` / `spec.postgresql.parameters` → rolling update, replica-first, with a leader switchover so there's no full outage.
* **Scale:** Change `spec.numberOfInstances` → a new pod is cloned from the leader and starts streaming; removing one drains it (failing over first if it's the leader).
* **Version Update:** Change `spec.dockerImage` → the pod restarts onto the new image on the same data directory (rolling on multi-instance clusters).
* **Backups / PITR:** Edit `spec.backup.pgbackrest` and a restore block → pgBackRest, driven by the operator.

---

### Beyond One Cluster: Multi-Site

A single Kubernetes cluster is still one failure domain. Multi-site layers a second consensus (an external etcd holding a per-cluster leader lease) on top of each site's own Patroni HA. Consequently, the loss of an entire site triggers an automatic cross-site failover under the same declarative model with one extra coordination layer.