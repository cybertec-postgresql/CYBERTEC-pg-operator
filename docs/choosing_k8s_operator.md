# Choosing a PostgreSQL Operator for Kubernetes

There are several good open-source PostgreSQL operators. This is an honest evaluation framework, the questions to ask of any of them, and where the **CYBERTEC PG Operator (CPO)** focuses rather than a scorecard. Operator features move fast; verify each project's current capabilities at evaluation time.

### Main Open-Source PostgreSQL Operators
* **CYBERTEC PG Operator (CPO)**
* **CloudNativePG**
* **Zalando postgres-operator**
* **Crunchy PGO**
* **StackGres**

All are credible; they differ in HA mechanism, cross-cluster story, ecosystem, and support model.

---

## The Dimensions That Matter

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

## Where CPO Fits

* **Patroni-based HA:** Proven reliability with a strong multi-site story (validated: automatic cross-site failover in ≈ 90s, clean automatic failback in ≈ 20s).
* **Open Source:** No lock-in, deployable on any Kubernetes or OpenShift cluster.
* **Expert Backing:** 24×7 support, consulting, and training directly from contributors to PostgreSQL and Patroni.

It is a particularly good fit when **cross-site / regional disaster recovery is a hard requirement** and when you want a commercial support partner behind your open-source database.

---

## How to Evaluate Fairly

1. **Run a spike:** Execute the same workload across two or three operators.
2. **Test real failover:** Kill the leader or simulate a site outage to measure real RTO/RPO. *(CPO's tutorials and multi-site scaffold simplify this setup).*
3. **Exercise recovery:** Test point-in-time restores, not just backup creation.
4. **Evaluate actual support costs:** Price the support tier you would realistically purchase, rather than relying solely on free tier capabilities.

---

## Where to Go Next

<!-- * **Whitepaper & Architecture:** Check the multi-site operator whitepaper in the resources section. -->
* **Source Code:** [CYBERTEC PG Operator GitHub Repository](https://github.com/cybertec-postgresql/CYBERTEC-pg-operator)
* **Documentation:** [CYBERTEC PG Operator Official Docs](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/)
* **Commercial Support:** Visit CYBERTEC 24×7 [support services](https://www.cybertec-postgresql.com/en/services/postgresql-support/) to resolve enterprise queries.
