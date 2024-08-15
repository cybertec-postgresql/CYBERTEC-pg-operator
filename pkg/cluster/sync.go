package cluster

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/spec"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/constants"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/k8sutil"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var requirePrimaryRestartWhenDecreased = []string{
	"max_connections",
	"max_prepared_transactions",
	"max_locks_per_transaction",
	"max_worker_processes",
	"max_wal_senders",
}

const (
	certAuthoritySecretKey        = "pgbackrest.ca-roots"      // #nosec G101 this is a name, not a credential
	certClientPrivateKeySecretKey = "pgbackrest-client.key"    // #nosec G101 this is a name, not a credential
	certClientSecretKey           = "pgbackrest-client.crt"    // #nosec G101 this is a name, not a credential
	certRepoSecretKey             = "pgbackrest-repo-host.crt" // #nosec G101 this is a name, not a credential
	certRepoPrivateKeySecretKey   = "pgbackrest-repo-host.key" // #nosec G101 this is a name, not a credential

	// pemLabelCertificate is the textual encoding label for an X.509 certificate
	// according to RFC 7468. See https://tools.ietf.org/html/rfc7468.
	pemLabelCertificate = "CERTIFICATE"

	// pemLabelECDSAKey is the textual encoding label for an elliptic curve private key
	// according to RFC 5915. See https://tools.ietf.org/html/rfc5915.
	pemLabelECDSAKey = "EC PRIVATE KEY"
)

type RootCertificateAuthority struct {
	Certificate Certificate
	PrivateKey  PrivateKey
}

// Certificate represents an X.509 certificate that conforms to the Internet
// PKI Profile, RFC 5280.
type Certificate struct{ x509 *x509.Certificate }

// PrivateKey represents the private key of a Certificate.
type PrivateKey struct{ ecdsa *ecdsa.PrivateKey }

var (
	_ encoding.TextMarshaler   = Certificate{}
	_ encoding.TextMarshaler   = (*Certificate)(nil)
	_ encoding.TextUnmarshaler = (*Certificate)(nil)

	_ encoding.TextMarshaler   = PrivateKey{}
	_ encoding.TextMarshaler   = (*PrivateKey)(nil)
	_ encoding.TextUnmarshaler = (*PrivateKey)(nil)
)

// certificateSignatureAlgorithm is ECDSA with SHA-384, the recommended
// signature algorithm with the P-256 curve.
const certificateSignatureAlgorithm = x509.ECDSAWithSHA384

// currentTime returns the current local time. It is a variable so it can be
// replaced during testing.
var currentTime = time.Now

// MarshalText returns a PEM encoding of c that OpenSSL understands.
func (c Certificate) MarshalText() ([]byte, error) {
	if c.x509 == nil || len(c.x509.Raw) == 0 {
		_, err := x509.ParseCertificate(nil)
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  pemLabelCertificate,
		Bytes: c.x509.Raw,
	}), nil
}

// UnmarshalText populates c from its PEM encoding.
func (c *Certificate) UnmarshalText(data []byte) error {
	block, _ := pem.Decode(data)

	if block == nil || block.Type != pemLabelCertificate {
		return fmt.Errorf("not a PEM-encoded certificate")
	}

	parsed, err := x509.ParseCertificate(block.Bytes)
	if err == nil {
		c.x509 = parsed
	}
	return err
}

// MarshalText returns a PEM encoding of k that OpenSSL understands.
func (k PrivateKey) MarshalText() ([]byte, error) {
	if k.ecdsa == nil {
		k.ecdsa = new(ecdsa.PrivateKey)
	}

	der, err := x509.MarshalECPrivateKey(k.ecdsa)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  pemLabelECDSAKey,
		Bytes: der,
	}), nil
}

// UnmarshalText populates k from its PEM encoding.
func (k *PrivateKey) UnmarshalText(data []byte) error {
	block, _ := pem.Decode(data)

	if block == nil || block.Type != pemLabelECDSAKey {
		return fmt.Errorf("not a PEM-encoded private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err == nil {
		k.ecdsa = key
	}
	return err
}

// LeafCertificate is a certificate and private key pair that can be validated
// by RootCertificateAuthority.
type LeafCertificate struct {
	Certificate Certificate
	PrivateKey  PrivateKey
}

// Equal reports whether c and other have the same value.
func (c Certificate) Equal(other Certificate) bool {
	return c.x509.Equal(other.x509)
}

// generateKey returns a random ECDSA key using a P-256 curve. This curve is
// roughly equivalent to an RSA 3072-bit key but requires less bits to achieve
// the equivalent cryptographic strength. Additionally, ECDSA is FIPS 140-2
// compliant.
func generateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// generateSerialNumber returns a random 128-bit integer.
func generateSerialNumber() (*big.Int, error) {
	return rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
}

// Sync syncs the cluster, making sure the actual Kubernetes objects correspond to what is defined in the manifest.
// Unlike the update, sync does not error out if some objects do not exist and takes care of creating them.
func (c *Cluster) Sync(newSpec *cpov1.Postgresql) error {
	var err error
	c.mu.Lock()
	defer c.mu.Unlock()

	oldSpec := c.Postgresql
	c.setSpec(newSpec)

	defer func() {
		if err != nil {
			c.logger.Warningf("error while syncing cluster state: %v", err)
			c.KubeClient.SetPostgresCRDStatus(c.clusterName(), cpov1.ClusterStatusSyncFailed)
		} else if !c.Status.Running() {
			c.KubeClient.SetPostgresCRDStatus(c.clusterName(), cpov1.ClusterStatusRunning)
		}
	}()

	// Make sure we know about any in progress restores before touching other stuff
	if err = c.refreshRestoreConfigMap(); err != nil {
		return fmt.Errorf("error refreshing restore configmap: %v", err)
	}

	if err = c.initUsers(); err != nil {
		err = fmt.Errorf("could not init users: %v", err)
		return err
	}

	//TODO: mind the secrets of the deleted/new users
	if err = c.syncSecrets(); err != nil {
		err = fmt.Errorf("could not sync secrets: %v", err)
		return err
	}

	if err = c.syncServices(); err != nil {
		err = fmt.Errorf("could not sync services: %v", err)
		return err
	}

	if err = c.syncPgbackrestConfig(); err != nil {
		err = fmt.Errorf("could not sync pgbackrest repo-host config: %v", err)
		return err
	}

	if err = c.syncPgbackrestRepoHostConfig(&c.Spec); err != nil {
		err = fmt.Errorf("could not sync pgbackrest config: %v", err)
		return err
	}

	//sync volume may already transition volumes to gp3, if iops/throughput or type is specified
	if err = c.syncVolumes(); err != nil {
		return err
	}

	if c.OpConfig.EnableEBSGp3Migration && len(c.EBSVolumes) > 0 {
		err = c.executeEBSMigration()
		if nil != err {
			return err
		}
	}

	c.logger.Debug("syncing statefulsets")
	if err = c.syncStatefulSet(); err != nil {
		if !k8sutil.ResourceAlreadyExists(err) {
			err = fmt.Errorf("could not sync statefulsets: %v", err)
			return err
		}
	}

	c.logger.Debug("syncing pod disruption budgets")
	if err = c.syncPodDisruptionBudget(false); err != nil {
		err = fmt.Errorf("could not sync pod disruption budget: %v", err)
		return err
	}

	// create a logical backup job unless we are running without pods or disable that feature explicitly
	if c.Spec.EnableLogicalBackup && c.getNumberOfInstances(&c.Spec) > 0 {

		c.logger.Debug("syncing logical backup job")
		if err = c.syncLogicalBackupJob(); err != nil {
			err = fmt.Errorf("could not sync the logical backup job: %v", err)
			return err
		}
	}

	c.logger.Debug("syncing pgbackrest jobs")
	deleteBackupJobs := c.Spec.GetBackup().Pgbackrest == nil
	if err = c.syncPgbackrestJob(deleteBackupJobs); err != nil {
		err = fmt.Errorf("could not sync the pgbackrest jobs: %v", err)
		return err
	}

	// create database objects unless we are running without pods or disabled that feature explicitly
	if !(c.databaseAccessDisabled() || c.getNumberOfInstances(&newSpec.Spec) <= 0 || c.Spec.StandbyCluster != nil || c.restoreInProgress()) {
		c.logger.Debug("syncing roles")
		if err = c.syncRoles(); err != nil {
			c.logger.Errorf("could not sync roles: %v", err)
		}
		c.logger.Debug("syncing databases")
		if err = c.syncDatabases(); err != nil {
			c.logger.Errorf("could not sync databases: %v", err)
		}
		c.logger.Debug("syncing prepared databases with schemas")
		if err = c.syncPreparedDatabases(); err != nil {
			c.logger.Errorf("could not sync prepared database: %v", err)
		}
	}

	// sync connection pooler
	if _, err = c.syncConnectionPooler(&oldSpec, newSpec, c.installLookupFunction); err != nil {
		return fmt.Errorf("could not sync connection pooler: %v", err)
	}

	// sync monitoring
	if err = c.syncMonitoringSecret(&oldSpec, newSpec); err != nil {
		return fmt.Errorf("could not sync monitoring: %v", err)
	}

	if err = c.syncWalPvc(&oldSpec, newSpec); err != nil {
		return fmt.Errorf("could not sync WAL-PVC: %v", err)
	}

	if len(c.Spec.Streams) > 0 {
		c.logger.Debug("syncing streams")
		if err = c.syncStreams(); err != nil {
			err = fmt.Errorf("could not sync streams: %v", err)
			return err
		}
	}

	// If we are requested to replace database contents with a restore only do so after we have everything
	// properly set up, but before we try to run the upgrade.
	restoreId := newSpec.Spec.GetBackup().GetRestoreID()
	if c.restoreInProgress() || c.Status.RestoreID != restoreId {
		if err := c.processRestore(newSpec); err != nil {
			return fmt.Errorf("restoring backup failed: %v", err)
		}
	}

	// Major version upgrade must only run after success of all earlier operations, must remain last item in sync
	if err := c.majorVersionUpgrade(); err != nil {
		c.logger.Errorf("major version upgrade failed: %v", err)
	}

	return err
}

func (c *Cluster) deletePgbackrestRepoHostObjects() error {
	c.setProcessName("Deleting pgbackrest repo-host")
	c.logger.Info("Deleting pgbackrest repo-host")

	var err error
	if err = c.KubeClient.StatefulSets(c.Namespace).Delete(context.TODO(), c.getPgbackrestRepoHostName(), metav1.DeleteOptions{}); err != nil {
		c.logger.Errorf("Could not delete Pgbackrest repo-host statefulset %v", err)
	} else {
		c.logger.Info("Repo-host statefulset is now deleted")
	}
	if err = c.KubeClient.Pods(c.Namespace).Delete(context.TODO(), c.getPgbackrestRepoHostName()+"-0", metav1.DeleteOptions{}); err != nil {
		c.logger.Errorf("Could not delete Pgbackrest repo-host pods %v", err)
	} else {
		c.logger.Info("Repo-host pods are now deleted")
	}
	if err = c.deleteRepoHostPersistentVolumeClaims(); err != nil {
		c.logger.Errorf("Could not delete Pgbackrest repo-host pvc %v", err)
	} else {
		c.logger.Info("Repo-host pvcs are now deleted")
	}
	if err = c.KubeClient.Secrets(c.Namespace).Delete(context.TODO(), c.getPgbackrestCertSecretName(), metav1.DeleteOptions{}); err != nil {
		c.logger.Errorf("Could not delete Pgbackrest repo-host secrets %v", err)
	} else {
		c.logger.Info("Repo-host secret is now deleted")
	}
	if err = c.KubeClient.ConfigMaps(c.Namespace).Delete(context.TODO(), c.getPgbackrestRepoHostConfigmapName(), metav1.DeleteOptions{}); err != nil {
		c.logger.Errorf("Could not delete Pgbackrest repo-host configmap %v", err)
	} else {
		c.logger.Info("Repo-host configmap is now deleted")
	}

	return nil
}

func (c *Cluster) syncServices() error {
	for _, role := range []PostgresRole{Master, Replica, ClusterPods} {
		c.logger.Debugf("syncing %s service", role)

		if !c.patroniKubernetesUseConfigMaps() {
			if err := c.syncEndpoint(role); err != nil {
				return fmt.Errorf("could not sync %s endpoint: %v", role, err)
			}
		}
		if err := c.syncService(role); err != nil {
			return fmt.Errorf("could not sync %s service: %v", role, err)
		}
	}

	return nil
}

func (c *Cluster) syncService(role PostgresRole) error {
	var (
		svc *v1.Service
		err error
	)
	c.setProcessName("syncing %s service", role)

	if svc, err = c.KubeClient.Services(c.Namespace).Get(context.TODO(), c.serviceName(role), metav1.GetOptions{}); err == nil {
		c.Services[role] = svc
		desiredSvc := c.generateService(role, &c.Spec)
		if match, reason := c.compareServices(svc, desiredSvc); !match {
			c.logServiceChanges(role, svc, desiredSvc, false, reason)
			updatedSvc, err := c.updateService(role, svc, desiredSvc)
			if err != nil {
				return fmt.Errorf("could not update %s service to match desired state: %v", role, err)
			}
			c.Services[role] = updatedSvc
			c.logger.Infof("%s service %q is in the desired state now", role, util.NameFromMeta(desiredSvc.ObjectMeta))
		}
		return nil
	}
	if !k8sutil.ResourceNotFound(err) {
		return fmt.Errorf("could not get %s service: %v", role, err)
	}
	// no existing service, create new one
	c.Services[role] = nil
	c.logger.Infof("could not find the cluster's %s service", role)

	if svc, err = c.createService(role); err == nil {
		c.logger.Infof("created missing %s service %q", role, util.NameFromMeta(svc.ObjectMeta))
	} else {
		if !k8sutil.ResourceAlreadyExists(err) {
			return fmt.Errorf("could not create missing %s service: %v", role, err)
		}
		c.logger.Infof("%s service %q already exists", role, util.NameFromMeta(svc.ObjectMeta))
		if svc, err = c.KubeClient.Services(c.Namespace).Get(context.TODO(), c.serviceName(role), metav1.GetOptions{}); err != nil {
			return fmt.Errorf("could not fetch existing %s service: %v", role, err)
		}
	}
	c.Services[role] = svc
	return nil
}

func (c *Cluster) syncEndpoint(role PostgresRole) error {
	var (
		ep  *v1.Endpoints
		err error
	)
	c.setProcessName("syncing %s endpoint", role)

	if ep, err = c.KubeClient.Endpoints(c.Namespace).Get(context.TODO(), c.endpointName(role), metav1.GetOptions{}); err == nil {
		// TODO: No syncing of endpoints here, is this covered completely by updateService?
		c.Endpoints[role] = ep
		return nil
	}
	if !k8sutil.ResourceNotFound(err) {
		return fmt.Errorf("could not get %s endpoint: %v", role, err)
	}
	// no existing endpoint, create new one
	c.Endpoints[role] = nil
	c.logger.Infof("could not find the cluster's %s endpoint", role)

	if ep, err = c.createEndpoint(role); err == nil {
		c.logger.Infof("created missing %s endpoint %q", role, util.NameFromMeta(ep.ObjectMeta))
	} else {
		if !k8sutil.ResourceAlreadyExists(err) {
			return fmt.Errorf("could not create missing %s endpoint: %v", role, err)
		}
		c.logger.Infof("%s endpoint %q already exists", role, util.NameFromMeta(ep.ObjectMeta))
		if ep, err = c.KubeClient.Endpoints(c.Namespace).Get(context.TODO(), c.endpointName(role), metav1.GetOptions{}); err != nil {
			return fmt.Errorf("could not fetch existing %s endpoint: %v", role, err)
		}
	}
	c.Endpoints[role] = ep
	return nil
}

func (c *Cluster) syncPodDisruptionBudget(isUpdate bool) error {
	var (
		pdb *policyv1.PodDisruptionBudget
		err error
	)
	if pdb, err = c.KubeClient.PodDisruptionBudgets(c.Namespace).Get(context.TODO(), c.podDisruptionBudgetName(), metav1.GetOptions{}); err == nil {
		c.PodDisruptionBudget = pdb
		newPDB := c.generatePodDisruptionBudget()
		if match, reason := k8sutil.SamePDB(pdb, newPDB); !match {
			c.logPDBChanges(pdb, newPDB, isUpdate, reason)
			if err = c.updatePodDisruptionBudget(newPDB); err != nil {
				return err
			}
		} else {
			c.PodDisruptionBudget = pdb
		}
		return nil

	}
	if !k8sutil.ResourceNotFound(err) {
		return fmt.Errorf("could not get pod disruption budget: %v", err)
	}
	// no existing pod disruption budget, create new one
	c.PodDisruptionBudget = nil
	c.logger.Infof("could not find the cluster's pod disruption budget")

	// When number of instances is 1, we don't need to create a pod disruption budget.
	if c.Spec.NumberOfInstances <= 1 {
		if c.PodDisruptionBudget != nil {
			c.logger.Warning("deleting pod disruption budget creation, number of instances is less than 1")
			if err := c.deletePodDisruptionBudget(); err != nil {
				c.logger.Warningf("could not delete pod disruption budget: %v", err)
			}
		}
		c.logger.Warning("skipping pod disruption budget creation, number of instances is less than 1")
	} else {
		if pdb, err = c.createPodDisruptionBudget(); err != nil {
			if !k8sutil.ResourceAlreadyExists(err) {
				return fmt.Errorf("could not create pod disruption budget: %v", err)
			}
			c.logger.Infof("pod disruption budget %q already exists", util.NameFromMeta(pdb.ObjectMeta))
			if pdb, err = c.KubeClient.PodDisruptionBudgets(c.Namespace).Get(context.TODO(), c.podDisruptionBudgetName(), metav1.GetOptions{}); err != nil {
				return fmt.Errorf("could not fetch existing %q pod disruption budget", util.NameFromMeta(pdb.ObjectMeta))
			}
		}
		c.logger.Infof("created missing pod disruption budget %q", util.NameFromMeta(pdb.ObjectMeta))
		c.PodDisruptionBudget = pdb

	}

	return nil
}

func (c *Cluster) syncStatefulSet() error {
	var (
		restartWait         uint32
		configPatched       bool
		restartPrimaryFirst bool
	)
	podsToRecreate := make([]v1.Pod, 0)
	isSafeToRecreatePods := true
	switchoverCandidates := make([]spec.NamespacedName, 0)

	pods, err := c.listPodsOfType(TYPE_POSTGRESQL)
	if err != nil {
		c.logger.Warnf("could not list pods of the statefulset: %v", err)
	}
	if c.Spec.Monitoring != nil { // XXX: Why are we generating a sidecar in the sync code?
		monitor := c.Spec.Monitoring
		sidecar := &cpov1.Sidecar{
			Name:        "postgres-exporter",
			DockerImage: monitor.Image,
			Ports: []v1.ContainerPort{
				{
					ContainerPort: monitorPort,
					Protocol:      v1.ProtocolTCP,
				},
			},
			Env: c.generateMonitoringEnvVars(),
		}
		c.Spec.Sidecars = append(c.Spec.Sidecars, *sidecar) //populate the sidecar spec so that the sidecar is automatically created
	}
	// NB: Be careful to consider the codepath that acts on podsRollingUpdateRequired before returning early.
	sset, err := c.KubeClient.StatefulSets(c.Namespace).Get(context.TODO(), c.statefulSetName(), metav1.GetOptions{})
	if err != nil {
		if !k8sutil.ResourceNotFound(err) {
			return fmt.Errorf("error during reading of statefulset: %v", err)
		}
		// statefulset does not exist, try to re-create it
		c.Statefulset = nil
		c.logger.Infof("cluster's statefulset does not exist")

		sset, err = c.createStatefulSet()
		if err != nil {
			return fmt.Errorf("could not create missing statefulset: %v", err)
		}

		if err = c.waitStatefulsetPodsReady(); err != nil {
			return fmt.Errorf("cluster is not ready: %v", err)
		}

		if len(pods) > 0 {
			for _, pod := range pods {
				if err = c.markRollingUpdateFlagForPod(&pod, "pod from previous statefulset"); err != nil {
					c.logger.Warnf("marking old pod for rolling update failed: %v", err)
				}
				podsToRecreate = append(podsToRecreate, pod)
			}
		}
		c.logger.Infof("created missing statefulset %q", util.NameFromMeta(sset.ObjectMeta))

	} else {
		// check if there are still pods with a rolling update flag
		for _, pod := range pods {
			if c.getRollingUpdateFlagFromPod(&pod) {
				podsToRecreate = append(podsToRecreate, pod)
			} else {
				role := PostgresRole(pod.Labels[c.OpConfig.PodRoleLabel])
				if role == Master {
					continue
				}
				switchoverCandidates = append(switchoverCandidates, util.NameFromMeta(pod.ObjectMeta))
			}
		}

		if len(podsToRecreate) > 0 {
			c.logger.Debugf("%d / %d pod(s) still need to be rotated", len(podsToRecreate), len(pods))
		}

		// statefulset is already there, make sure we use its definition in order to compare with the spec.
		c.Statefulset = sset

		desiredSts, err := c.generateStatefulSet(&c.Spec)
		if err != nil {
			return fmt.Errorf("could not generate statefulset: %v", err)
		}

		if c.restoreInProgress() {
			c.applyRestoreStatefulSetSyncOverrides(desiredSts, c.Statefulset)
		}

		cmp := c.compareStatefulSetWith(c.Statefulset, desiredSts)
		if !cmp.match {
			if cmp.rollingUpdate {
				podsToRecreate = make([]v1.Pod, 0)
				switchoverCandidates = make([]spec.NamespacedName, 0)
				for _, pod := range pods {
					if err = c.markRollingUpdateFlagForPod(&pod, "pod changes"); err != nil {
						return fmt.Errorf("updating rolling update flag for pod failed: %v", err)
					}
					podsToRecreate = append(podsToRecreate, pod)
				}
			}

			c.logStatefulSetChanges(c.Statefulset, desiredSts, false, cmp.reasons)

			if !cmp.replace {
				if err := c.updateStatefulSet(desiredSts); err != nil {
					return fmt.Errorf("could not update statefulset: %v", err)
				}
			} else {
				if err := c.replaceStatefulSet(&c.Statefulset, desiredSts); err != nil {
					return fmt.Errorf("could not replace statefulset: %v", err)
				}
			}
		}

		if len(podsToRecreate) == 0 && !c.OpConfig.EnableLazySpiloUpgrade {
			// even if the desired and the running statefulsets match
			// there still may be not up-to-date pods on condition
			//  (a) the lazy update was just disabled
			// and
			//  (b) some of the pods were not restarted when the lazy update was still in place
			for _, pod := range pods {
				effectivePodImage := getPostgresContainer(&pod.Spec).Image
				stsImage := getPostgresContainer(&desiredSts.Spec.Template.Spec).Image

				if stsImage != effectivePodImage {
					if err = c.markRollingUpdateFlagForPod(&pod, "pod not yet restarted due to lazy update"); err != nil {
						c.logger.Warnf("updating rolling update flag failed for pod %q: %v", pod.Name, err)
					}
					podsToRecreate = append(podsToRecreate, pod)
				} else {
					role := PostgresRole(pod.Labels[c.OpConfig.PodRoleLabel])
					if role == Master {
						continue
					}
					switchoverCandidates = append(switchoverCandidates, util.NameFromMeta(pod.ObjectMeta))
				}
			}
		}
	}

	// apply PostgreSQL parameters that can only be set via the Patroni API.
	// it is important to do it after the statefulset pods are there, but before the rolling update
	// since those parameters require PostgreSQL restart.
	pods, err = c.listPodsOfType(TYPE_POSTGRESQL)
	if err != nil {
		c.logger.Warnf("could not get list of pods to apply PostgreSQL parameters only to be set via Patroni API: %v", err)
	}

	requiredPgParameters := make(map[string]string)
	for k, v := range c.Spec.Parameters {
		requiredPgParameters[k] = v
	}
	// if streams are defined wal_level must be switched to logical
	if len(c.Spec.Streams) > 0 {
		requiredPgParameters["wal_level"] = "logical"
	}

	// sync Patroni config
	c.logger.Debug("syncing Patroni config")
	if configPatched, restartPrimaryFirst, restartWait, err = c.syncPatroniConfig(pods, c.Spec.Patroni, requiredPgParameters); err != nil {
		c.logger.Warningf("Patroni config updated? %v - errors during config sync: %v", configPatched, err)
		isSafeToRecreatePods = false
	}

	// restart Postgres where it is still pending
	if err = c.restartInstances(pods, restartWait, restartPrimaryFirst); err != nil {
		c.logger.Errorf("errors while restarting Postgres in pods via Patroni API: %v", err)
		isSafeToRecreatePods = false
	}

	// if we get here we also need to re-create the pods (either leftovers from the old
	// statefulset or those that got their configuration from the outdated statefulset)
	if len(podsToRecreate) > 0 {
		if isSafeToRecreatePods {
			c.logger.Debugln("performing rolling update")
			c.eventRecorder.Event(c.GetReference(), v1.EventTypeNormal, "Update", "Performing rolling update")
			if err := c.recreatePods(podsToRecreate, switchoverCandidates); err != nil {
				return fmt.Errorf("could not recreate pods: %v", err)
			}
			c.eventRecorder.Event(c.GetReference(), v1.EventTypeNormal, "Update", "Rolling update done - pods have been recreated")
		} else {
			c.logger.Warningf("postpone pod recreation until next sync because of errors during config sync")
		}
	}

	return nil
}

func (c *Cluster) syncPatroniConfig(pods []v1.Pod, requiredPatroniConfig cpov1.Patroni, requiredPgParameters map[string]string) (bool, bool, uint32, error) {
	var (
		effectivePatroniConfig cpov1.Patroni
		effectivePgParameters  map[string]string
		loopWait               uint32
		configPatched          bool
		restartPrimaryFirst    bool
		err                    error
	)

	errors := make([]string, 0)

	// get Postgres config, compare with manifest and update via Patroni PATCH endpoint if it differs
	for i, pod := range pods {
		podName := util.NameFromMeta(pods[i].ObjectMeta)
		effectivePatroniConfig, effectivePgParameters, err = c.getPatroniConfig(&pod)
		if err != nil {
			errors = append(errors, fmt.Sprintf("could not get Postgres config from pod %s: %v", podName, err))
			continue
		}
		loopWait = effectivePatroniConfig.LoopWait

		// empty config probably means cluster is not fully initialized yet, e.g. restoring from backup
		if reflect.DeepEqual(effectivePatroniConfig, cpov1.Patroni{}) || len(effectivePgParameters) == 0 {
			errors = append(errors, fmt.Sprintf("empty Patroni config on pod %s - skipping config patch", podName))
		} else {
			configPatched, restartPrimaryFirst, err = c.checkAndSetGlobalPostgreSQLConfiguration(&pod, effectivePatroniConfig, requiredPatroniConfig, effectivePgParameters, requiredPgParameters)
			if err != nil {
				errors = append(errors, fmt.Sprintf("could not set PostgreSQL configuration options for pod %s: %v", podName, err))
				continue
			}

			// it could take up to LoopWait to apply the config
			if configPatched {
				time.Sleep(time.Duration(loopWait)*time.Second + time.Second*2)
				// Patroni's config endpoint is just a "proxy" to DCS.
				// It is enough to patch it only once and it doesn't matter which pod is used
				break
			}
		}
	}

	if len(errors) > 0 {
		err = fmt.Errorf("%v", strings.Join(errors, `', '`))
	}

	return configPatched, restartPrimaryFirst, loopWait, err
}

func (c *Cluster) restartInstances(pods []v1.Pod, restartWait uint32, restartPrimaryFirst bool) (err error) {
	errors := make([]string, 0)
	remainingPods := make([]*v1.Pod, 0)

	skipRole := Master
	if restartPrimaryFirst {
		skipRole = Replica
	}

	for i, pod := range pods {
		role := PostgresRole(pod.Labels[c.OpConfig.PodRoleLabel])
		if role == skipRole {
			remainingPods = append(remainingPods, &pods[i])
			continue
		}
		if err = c.restartInstance(&pod, restartWait); err != nil {
			errors = append(errors, fmt.Sprintf("%v", err))
		}
	}

	// in most cases only the master should be left to restart
	if len(remainingPods) > 0 {
		for _, remainingPod := range remainingPods {
			if err = c.restartInstance(remainingPod, restartWait); err != nil {
				errors = append(errors, fmt.Sprintf("%v", err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%v", strings.Join(errors, `', '`))
	}

	return nil
}

func (c *Cluster) restartInstance(pod *v1.Pod, restartWait uint32) error {
	// if the config update requires a restart, call Patroni restart
	podName := util.NameFromMeta(pod.ObjectMeta)
	role := PostgresRole(pod.Labels[c.OpConfig.PodRoleLabel])
	memberData, err := c.getPatroniMemberData(pod)
	if err != nil {
		return fmt.Errorf("could not restart Postgres in %s pod %s: %v", role, podName, err)
	}

	// do restart only when it is pending
	if memberData.PendingRestart {
		c.eventRecorder.Event(c.GetReference(), v1.EventTypeNormal, "Update", fmt.Sprintf("restarting Postgres server within %s pod %s", role, podName))
		if err := c.patroni.Restart(pod); err != nil {
			return err
		}
		time.Sleep(time.Duration(restartWait) * time.Second)
		c.eventRecorder.Event(c.GetReference(), v1.EventTypeNormal, "Update", fmt.Sprintf("Postgres server restart done for %s pod %s", role, podName))
	}

	return nil
}

// AnnotationsToPropagate get the annotations to update if required
// based on the annotations in postgres CRD
func (c *Cluster) AnnotationsToPropagate(annotations map[string]string) map[string]string {

	if annotations == nil {
		annotations = make(map[string]string)
	}

	pgCRDAnnotations := c.ObjectMeta.Annotations

	if pgCRDAnnotations != nil {
		for _, anno := range c.OpConfig.DownscalerAnnotations {
			for k, v := range pgCRDAnnotations {
				matched, err := regexp.MatchString(anno, k)
				if err != nil {
					c.logger.Errorf("annotations matching issue: %v", err)
					return nil
				}
				if matched {
					annotations[k] = v
				}
			}
		}
	}

	if len(annotations) > 0 {
		return annotations
	}

	return nil
}

// checkAndSetGlobalPostgreSQLConfiguration checks whether cluster-wide API parameters
// (like max_connections) have changed and if necessary sets it via the Patroni API
func (c *Cluster) checkAndSetGlobalPostgreSQLConfiguration(pod *v1.Pod, effectivePatroniConfig, desiredPatroniConfig cpov1.Patroni, effectivePgParameters, desiredPgParameters map[string]string) (bool, bool, error) {
	configToSet := make(map[string]interface{})
	parametersToSet := make(map[string]string)
	restartPrimary := make([]bool, 0)
	configPatched := false
	requiresMasterRestart := false

	// compare effective and desired Patroni config options
	if desiredPatroniConfig.LoopWait > 0 && desiredPatroniConfig.LoopWait != effectivePatroniConfig.LoopWait {
		configToSet["loop_wait"] = desiredPatroniConfig.LoopWait
	}
	if desiredPatroniConfig.MaximumLagOnFailover > 0 && desiredPatroniConfig.MaximumLagOnFailover != effectivePatroniConfig.MaximumLagOnFailover {
		configToSet["maximum_lag_on_failover"] = desiredPatroniConfig.MaximumLagOnFailover
	}
	if desiredPatroniConfig.PgHba != nil && !reflect.DeepEqual(desiredPatroniConfig.PgHba, effectivePatroniConfig.PgHba) {
		configToSet["pg_hba"] = desiredPatroniConfig.PgHba
	}
	if desiredPatroniConfig.RetryTimeout > 0 && desiredPatroniConfig.RetryTimeout != effectivePatroniConfig.RetryTimeout {
		configToSet["retry_timeout"] = desiredPatroniConfig.RetryTimeout
	}
	if desiredPatroniConfig.SynchronousMode != effectivePatroniConfig.SynchronousMode {
		configToSet["synchronous_mode"] = desiredPatroniConfig.SynchronousMode
	}
	if desiredPatroniConfig.SynchronousModeStrict != effectivePatroniConfig.SynchronousModeStrict {
		configToSet["synchronous_mode_strict"] = desiredPatroniConfig.SynchronousModeStrict
	}
	if desiredPatroniConfig.TTL > 0 && desiredPatroniConfig.TTL != effectivePatroniConfig.TTL {
		configToSet["ttl"] = desiredPatroniConfig.TTL
	}

	var desiredFailsafe *bool
	if desiredPatroniConfig.FailsafeMode != nil {
		desiredFailsafe = desiredPatroniConfig.FailsafeMode
	} else if c.OpConfig.EnablePatroniFailsafeMode != nil {
		desiredFailsafe = c.OpConfig.EnablePatroniFailsafeMode
	}

	effectiveFailsafe := effectivePatroniConfig.FailsafeMode

	if desiredFailsafe != nil {
		if effectiveFailsafe == nil || *desiredFailsafe != *effectiveFailsafe {
			configToSet["failsafe_mode"] = *desiredFailsafe
		}
	}

	slotsToSet := make(map[string]interface{})
	// check if there is any slot deletion
	for slotName, effectiveSlot := range c.replicationSlots {
		if desiredSlot, exists := desiredPatroniConfig.Slots[slotName]; exists {
			if reflect.DeepEqual(effectiveSlot, desiredSlot) {
				continue
			}
		}
		slotsToSet[slotName] = nil
		delete(c.replicationSlots, slotName)
	}
	// check if specified slots exist in config and if they differ
	for slotName, desiredSlot := range desiredPatroniConfig.Slots {
		// only add slots specified in manifest to c.replicationSlots
		for manifestSlotName, _ := range c.Spec.Patroni.Slots {
			if manifestSlotName == slotName {
				c.replicationSlots[slotName] = desiredSlot
			}
		}
		if effectiveSlot, exists := effectivePatroniConfig.Slots[slotName]; exists {
			if reflect.DeepEqual(desiredSlot, effectiveSlot) {
				continue
			}
		}
		slotsToSet[slotName] = desiredSlot
	}
	if len(slotsToSet) > 0 {
		configToSet["slots"] = slotsToSet
	}

	// compare effective and desired parameters under postgresql section in Patroni config
	for desiredOption, desiredValue := range desiredPgParameters {
		effectiveValue := effectivePgParameters[desiredOption]
		if isBootstrapOnlyParameter(desiredOption) && (effectiveValue != desiredValue) {
			parametersToSet[desiredOption] = desiredValue
			if util.SliceContains(requirePrimaryRestartWhenDecreased, desiredOption) {
				effectiveValueNum, errConv := strconv.Atoi(effectiveValue)
				desiredValueNum, errConv2 := strconv.Atoi(desiredValue)
				if errConv != nil || errConv2 != nil {
					continue
				}
				if effectiveValueNum > desiredValueNum {
					restartPrimary = append(restartPrimary, true)
					continue
				}
			}
			restartPrimary = append(restartPrimary, false)
		}
	}

	// check if there exist only config updates that require a restart of the primary
	if len(restartPrimary) > 0 && !util.SliceContains(restartPrimary, false) && len(configToSet) == 0 {
		requiresMasterRestart = true
	}

	if len(parametersToSet) > 0 {
		configToSet["postgresql"] = map[string]interface{}{constants.PatroniPGParametersParameterName: parametersToSet}
	}

	if len(configToSet) == 0 {
		return configPatched, requiresMasterRestart, nil
	}

	configToSetJson, err := json.Marshal(configToSet)
	if err != nil {
		c.logger.Debugf("could not convert config patch to JSON: %v", err)
	}

	// try all pods until the first one that is successful, as it doesn't matter which pod
	// carries the request to change configuration through
	podName := util.NameFromMeta(pod.ObjectMeta)
	c.logger.Debugf("patching Postgres config via Patroni API on pod %s with following options: %s",
		podName, configToSetJson)
	if err = c.patroni.SetConfig(pod, configToSet); err != nil {
		return configPatched, requiresMasterRestart, fmt.Errorf("could not patch postgres parameters within pod %s: %v", podName, err)
	}
	configPatched = true

	return configPatched, requiresMasterRestart, nil
}

func (c *Cluster) syncSecrets() error {
	c.logger.Info("syncing secrets")
	c.setProcessName("syncing secrets")
	generatedSecrets := c.generateUserSecrets()
	retentionUsers := make([]string, 0)
	currentTime := time.Now()

	for secretUsername, generatedSecret := range generatedSecrets {
		secret, err := c.KubeClient.Secrets(generatedSecret.Namespace).Create(context.TODO(), generatedSecret, metav1.CreateOptions{})
		if err == nil {
			c.Secrets[secret.UID] = secret
			c.logger.Debugf("created new secret %s, namespace: %s, uid: %s", util.NameFromMeta(secret.ObjectMeta), generatedSecret.Namespace, secret.UID)
			continue
		}
		if k8sutil.ResourceAlreadyExists(err) {
			if err = c.updateSecret(secretUsername, generatedSecret, &retentionUsers, currentTime); err != nil {
				c.logger.Warningf("syncing secret %s failed: %v", util.NameFromMeta(secret.ObjectMeta), err)
			}
		} else {
			return fmt.Errorf("could not create secret for user %s: in namespace %s: %v", secretUsername, generatedSecret.Namespace, err)
		}
	}

	// remove rotation users that exceed the retention interval
	if len(retentionUsers) > 0 {
		err := c.initDbConn()
		if err != nil {
			return fmt.Errorf("could not init db connection: %v", err)
		}
		if err = c.cleanupRotatedUsers(retentionUsers, c.pgDb); err != nil {
			return fmt.Errorf("error removing users exceeding configured retention interval: %v", err)
		}
		if err := c.closeDbConn(); err != nil {
			c.logger.Errorf("could not close database connection after removing users exceeding configured retention interval: %v", err)
		}
	}

	return nil
}

func (c *Cluster) getNextRotationDate(currentDate time.Time) (time.Time, string) {
	nextRotationDate := currentDate.AddDate(0, 0, int(c.OpConfig.PasswordRotationInterval))
	return nextRotationDate, nextRotationDate.Format(time.RFC3339)
}

func (c *Cluster) updateSecret(
	secretUsername string,
	generatedSecret *v1.Secret,
	retentionUsers *[]string,
	currentTime time.Time) error {
	var (
		secret          *v1.Secret
		err             error
		updateSecret    bool
		updateSecretMsg string
	)

	// get the secret first
	if secret, err = c.KubeClient.Secrets(generatedSecret.Namespace).Get(context.TODO(), generatedSecret.Name, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("could not get current secret: %v", err)
	}
	c.Secrets[secret.UID] = secret

	// fetch user map to update later
	var userMap map[string]spec.PgUser
	var userKey string
	if secretUsername == c.systemUsers[constants.SuperuserKeyName].Name {
		userKey = constants.SuperuserKeyName
		userMap = c.systemUsers
	} else if secretUsername == c.systemUsers[constants.ReplicationUserKeyName].Name {
		userKey = constants.ReplicationUserKeyName
		userMap = c.systemUsers
	} else {
		userKey = secretUsername
		userMap = c.pgUsers
	}

	// use system user when pooler is enabled and pooler user is specfied in manifest
	if _, exists := c.systemUsers[constants.ConnectionPoolerUserKeyName]; exists {
		if secretUsername == c.systemUsers[constants.ConnectionPoolerUserKeyName].Name {
			userKey = constants.ConnectionPoolerUserKeyName
			userMap = c.systemUsers
		}
	}
	// use system user when streams are defined and fes_user is specfied in manifest
	if _, exists := c.systemUsers[constants.EventStreamUserKeyName]; exists {
		if secretUsername == c.systemUsers[constants.EventStreamUserKeyName].Name {
			userKey = constants.EventStreamUserKeyName
			userMap = c.systemUsers
		}
	}

	pwdUser := userMap[userKey]
	secretName := util.NameFromMeta(secret.ObjectMeta)

	// if password rotation is enabled update password and username if rotation interval has been passed
	// rotation can be enabled globally or via the manifest (excluding the Postgres superuser)
	rotationEnabledInManifest := secretUsername != constants.SuperuserKeyName &&
		(util.SliceContains(c.Spec.UsersWithSecretRotation, secretUsername) ||
			util.SliceContains(c.Spec.UsersWithInPlaceSecretRotation, secretUsername))

	// globally enabled rotation is only allowed for manifest and bootstrapped roles
	allowedRoleTypes := []spec.RoleOrigin{spec.RoleOriginManifest, spec.RoleOriginBootstrap}
	rotationAllowed := !pwdUser.IsDbOwner && util.SliceContains(allowedRoleTypes, pwdUser.Origin) && c.Spec.StandbyCluster == nil

	if (c.OpConfig.EnablePasswordRotation && rotationAllowed) || rotationEnabledInManifest {
		updateSecretMsg, err = c.rotatePasswordInSecret(secret, secretUsername, pwdUser.Origin, currentTime, retentionUsers)
		if err != nil {
			c.logger.Warnf("password rotation failed for user %s: %v", secretUsername, err)
		}
		if updateSecretMsg != "" {
			updateSecret = true
		}
	} else {
		// username might not match if password rotation has been disabled again
		if secretUsername != string(secret.Data["username"]) {
			*retentionUsers = append(*retentionUsers, secretUsername)
			secret.Data["username"] = []byte(secretUsername)
			secret.Data["password"] = []byte(util.RandomPassword(constants.PasswordLength))
			secret.Data["nextRotation"] = []byte{}
			updateSecret = true
			updateSecretMsg = fmt.Sprintf("secret %s does not contain the role %s - updating username and resetting password", secretName, secretUsername)
		}
	}

	// if this secret belongs to the infrastructure role and the password has changed - replace it in the secret
	if pwdUser.Password != string(secret.Data["password"]) && pwdUser.Origin == spec.RoleOriginInfrastructure {
		secret = generatedSecret
		updateSecret = true
		updateSecretMsg = fmt.Sprintf("updating the secret %s from the infrastructure roles", secretName)
	} else {
		// for non-infrastructure role - update the role with username and password from secret
		pwdUser.Name = string(secret.Data["username"])
		pwdUser.Password = string(secret.Data["password"])
		// update membership if we deal with a rotation user
		if secretUsername != pwdUser.Name {
			pwdUser.Rotated = true
			pwdUser.MemberOf = []string{secretUsername}
		}
		userMap[userKey] = pwdUser
	}

	if updateSecret {
		c.logger.Debugln(updateSecretMsg)
		if _, err = c.KubeClient.Secrets(secret.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("could not update secret %s: %v", secretName, err)
		}
		c.Secrets[secret.UID] = secret
	}

	return nil
}

func (c *Cluster) rotatePasswordInSecret(
	secret *v1.Secret,
	secretUsername string,
	roleOrigin spec.RoleOrigin,
	currentTime time.Time,
	retentionUsers *[]string) (string, error) {
	var (
		err                 error
		nextRotationDate    time.Time
		nextRotationDateStr string
		updateSecretMsg     string
	)

	secretName := util.NameFromMeta(secret.ObjectMeta)

	// initialize password rotation setting first rotation date
	nextRotationDateStr = string(secret.Data["nextRotation"])
	if nextRotationDate, err = time.ParseInLocation(time.RFC3339, nextRotationDateStr, currentTime.UTC().Location()); err != nil {
		nextRotationDate, nextRotationDateStr = c.getNextRotationDate(currentTime)
		secret.Data["nextRotation"] = []byte(nextRotationDateStr)
		updateSecretMsg = fmt.Sprintf("rotation date not found in secret %s. Setting it to %s", secretName, nextRotationDateStr)
	}

	// check if next rotation can happen sooner
	// if rotation interval has been decreased
	currentRotationDate, nextRotationDateStr := c.getNextRotationDate(currentTime)
	if nextRotationDate.After(currentRotationDate) {
		nextRotationDate = currentRotationDate
	}

	// update password and next rotation date if configured interval has passed
	if currentTime.After(nextRotationDate) {
		// create rotation user if role is not listed for in-place password update
		if !util.SliceContains(c.Spec.UsersWithInPlaceSecretRotation, secretUsername) {
			rotationUsername := fmt.Sprintf("%s%s", secretUsername, currentTime.Format(constants.RotationUserDateFormat))
			secret.Data["username"] = []byte(rotationUsername)
			c.logger.Infof("updating username in secret %s and creating rotation user %s in the database", secretName, rotationUsername)
			// whenever there is a rotation, check if old rotation users can be deleted
			*retentionUsers = append(*retentionUsers, secretUsername)
		} else {
			// when passwords of system users are rotated in place, pods have to be replaced
			if roleOrigin == spec.RoleOriginSystem {
				pods, err := c.listPodsOfType(TYPE_POSTGRESQL)
				if err != nil {
					return "", fmt.Errorf("could not list pods of the statefulset: %v", err)
				}
				for _, pod := range pods {
					if err = c.markRollingUpdateFlagForPod(&pod,
						fmt.Sprintf("replace pod due to password rotation of system user %s", secretUsername)); err != nil {
						c.logger.Warnf("marking pod for rolling update due to password rotation failed: %v", err)
					}
				}
			}

			// when password of connection pooler is rotated in place, pooler pods have to be replaced
			if roleOrigin == spec.RoleOriginConnectionPooler {
				listOptions := metav1.ListOptions{
					LabelSelector: c.poolerLabelsSet(true).String(),
				}
				poolerPods, err := c.listPoolerPods(listOptions)
				if err != nil {
					return "", fmt.Errorf("could not list pods of the pooler deployment: %v", err)
				}
				for _, poolerPod := range poolerPods {
					if err = c.markRollingUpdateFlagForPod(&poolerPod,
						fmt.Sprintf("replace pooler pod due to password rotation of pooler user %s", secretUsername)); err != nil {
						c.logger.Warnf("marking pooler pod for rolling update due to password rotation failed: %v", err)
					}
				}
			}

			// when password of stream user is rotated in place, it should trigger rolling update in FES deployment
			if roleOrigin == spec.RoleOriginStream {
				c.logger.Warnf("password in secret of stream user %s changed", constants.EventStreamSourceSlotPrefix+constants.UserRoleNameSuffix)
			}
		}
		secret.Data["password"] = []byte(util.RandomPassword(constants.PasswordLength))
		secret.Data["nextRotation"] = []byte(nextRotationDateStr)
		updateSecretMsg = fmt.Sprintf("updating secret %s due to password rotation - next rotation date: %s", secretName, nextRotationDateStr)
	}

	return updateSecretMsg, nil
}

func (c *Cluster) syncRoles() (err error) {
	c.setProcessName("syncing roles")

	var (
		dbUsers   spec.PgUserMap
		newUsers  spec.PgUserMap
		userNames []string
	)

	err = c.initDbConn()
	if err != nil {
		return fmt.Errorf("could not init db connection: %v", err)
	}

	defer func() {
		if err2 := c.closeDbConn(); err2 != nil {
			if err == nil {
				err = fmt.Errorf("could not close database connection: %v", err2)
			} else {
				err = fmt.Errorf("could not close database connection: %v (prior error: %v)", err2, err)
			}
		}
	}()

	//Check if monitoring user is added in manifest
	if _, ok := c.Spec.Users["cpo-exporter"]; ok {
		c.logger.Error("creating user of name cpo-exporter is not allowed as it is reserved for monitoring")
	}

	// mapping between original role name and with deletion suffix
	deletedUsers := map[string]string{}
	newUsers = make(map[string]spec.PgUser)

	// create list of database roles to query
	for _, u := range c.pgUsers {
		pgRole := u.Name
		userNames = append(userNames, pgRole)

		// when a rotation happened add group role to query its rolconfig
		if u.Rotated {
			userNames = append(userNames, u.MemberOf[0])
		}

		// add team member role name with rename suffix in case we need to rename it back
		if u.Origin == spec.RoleOriginTeamsAPI && c.OpConfig.EnableTeamMemberDeprecation {
			deletedUsers[pgRole+c.OpConfig.RoleDeletionSuffix] = pgRole
			userNames = append(userNames, pgRole+c.OpConfig.RoleDeletionSuffix)
		}
	}

	// add team members that exist only in cache
	// to trigger a rename of the role in ProduceSyncRequests
	for _, cachedUser := range c.pgUsersCache {
		if _, exists := c.pgUsers[cachedUser.Name]; !exists {
			userNames = append(userNames, cachedUser.Name)
		}
	}

	// search also for system users
	for _, systemUser := range c.systemUsers {
		userNames = append(userNames, systemUser.Name)
		newUsers[systemUser.Name] = systemUser
	}

	dbUsers, err = c.readPgUsersFromDatabase(userNames)
	if err != nil {
		return fmt.Errorf("error getting users from the database: %v", err)
	}

DBUSERS:
	for _, dbUser := range dbUsers {
		// copy rolconfig to rotation users
		for pgUserName, pgUser := range c.pgUsers {
			if pgUser.Rotated && pgUser.MemberOf[0] == dbUser.Name {
				pgUser.Parameters = dbUser.Parameters
				c.pgUsers[pgUserName] = pgUser
				// remove group role from dbUsers to not count as deleted role
				delete(dbUsers, dbUser.Name)
				continue DBUSERS
			}
		}

		// update pgUsers where a deleted role was found
		// so that they are skipped in ProduceSyncRequests
		originalUsername, foundDeletedUser := deletedUsers[dbUser.Name]
		// check if original user does not exist in dbUsers
		_, originalUserAlreadyExists := dbUsers[originalUsername]
		if foundDeletedUser && !originalUserAlreadyExists {
			recreatedUser := c.pgUsers[originalUsername]
			recreatedUser.Deleted = true
			c.pgUsers[originalUsername] = recreatedUser
		}
	}

	// last but not least copy pgUsers to newUsers to send to ProduceSyncRequests
	for _, pgUser := range c.pgUsers {
		newUsers[pgUser.Name] = pgUser
	}

	pgSyncRequests := c.userSyncStrategy.ProduceSyncRequests(dbUsers, newUsers)
	if err = c.userSyncStrategy.ExecuteSyncRequests(pgSyncRequests, c.pgDb); err != nil {
		return fmt.Errorf("error executing sync statements: %v", err)
	}

	return nil
}

func (c *Cluster) syncDatabases() error {
	c.setProcessName("syncing databases")
	errors := make([]string, 0)
	createDatabases := make(map[string]string)
	alterOwnerDatabases := make(map[string]string)
	preparedDatabases := make([]string, 0)

	if err := c.initDbConn(); err != nil {
		return fmt.Errorf("could not init database connection")
	}
	defer func() {
		if err := c.closeDbConn(); err != nil {
			c.logger.Errorf("could not close database connection: %v", err)
		}
	}()

	currentDatabases, err := c.getDatabases()
	if err != nil {
		return fmt.Errorf("could not get current databases: %v", err)
	}

	// if no prepared databases are specified create a database named like the cluster
	if c.Spec.PreparedDatabases != nil && len(c.Spec.PreparedDatabases) == 0 { // TODO: add option to disable creating such a default DB
		c.Spec.PreparedDatabases = map[string]cpov1.PreparedDatabase{strings.Replace(c.Name, "-", "_", -1): {}}
	}
	for preparedDatabaseName := range c.Spec.PreparedDatabases {
		_, exists := currentDatabases[preparedDatabaseName]
		if !exists {
			createDatabases[preparedDatabaseName] = fmt.Sprintf("%s%s", preparedDatabaseName, constants.OwnerRoleNameSuffix)
			preparedDatabases = append(preparedDatabases, preparedDatabaseName)
		}
	}

	for databaseName, newOwner := range c.Spec.Databases {
		currentOwner, exists := currentDatabases[databaseName]
		if !exists {
			createDatabases[databaseName] = newOwner
		} else if currentOwner != newOwner {
			alterOwnerDatabases[databaseName] = newOwner
		}
	}

	if len(createDatabases)+len(alterOwnerDatabases) == 0 {
		return nil
	}

	for databaseName, owner := range createDatabases {
		if err = c.executeCreateDatabase(databaseName, owner); err != nil {
			errors = append(errors, err.Error())
		}
	}
	for databaseName, owner := range alterOwnerDatabases {
		if err = c.executeAlterDatabaseOwner(databaseName, owner); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(createDatabases) > 0 {
		// trigger creation of pooler objects in new database in syncConnectionPooler
		if c.ConnectionPooler != nil {
			for _, role := range [3]PostgresRole{Master, Replica, ClusterPods} {
				c.ConnectionPooler[role].LookupFunction = false
			}
		}
	}

	// set default privileges for prepared database
	for _, preparedDatabase := range preparedDatabases {
		if err := c.initDbConnWithName(preparedDatabase); err != nil {
			errors = append(errors, fmt.Sprintf("could not init database connection to %s", preparedDatabase))
			continue
		}

		for _, owner := range c.getOwnerRoles(preparedDatabase, c.Spec.PreparedDatabases[preparedDatabase].DefaultUsers) {
			if err = c.execAlterGlobalDefaultPrivileges(owner, preparedDatabase); err != nil {
				errors = append(errors, err.Error())
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("error(s) while syncing databases: %v", strings.Join(errors, `', '`))
	}

	return nil
}

func (c *Cluster) syncPreparedDatabases() error {
	c.setProcessName("syncing prepared databases")
	errors := make([]string, 0)

	for preparedDbName, preparedDB := range c.Spec.PreparedDatabases {
		if err := c.initDbConnWithName(preparedDbName); err != nil {
			errors = append(errors, fmt.Sprintf("could not init connection to database %s: %v", preparedDbName, err))
			continue
		}

		c.logger.Debugf("syncing prepared database %q", preparedDbName)
		// now, prepare defined schemas
		preparedSchemas := preparedDB.PreparedSchemas
		if len(preparedDB.PreparedSchemas) == 0 {
			preparedSchemas = map[string]cpov1.PreparedSchema{"data": {DefaultRoles: util.True()}}
		}
		if err := c.syncPreparedSchemas(preparedDbName, preparedSchemas); err != nil {
			errors = append(errors, err.Error())
			continue
		}

		// install extensions
		if err := c.syncExtensions(preparedDB.Extensions); err != nil {
			errors = append(errors, err.Error())
		}

		if err := c.closeDbConn(); err != nil {
			c.logger.Errorf("could not close database connection: %v", err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("error(s) while syncing prepared databases: %v", strings.Join(errors, `', '`))
	}

	return nil
}

func (c *Cluster) syncPreparedSchemas(databaseName string, preparedSchemas map[string]cpov1.PreparedSchema) error {
	c.setProcessName("syncing prepared schemas")
	errors := make([]string, 0)

	currentSchemas, err := c.getSchemas()
	if err != nil {
		return fmt.Errorf("could not get current schemas: %v", err)
	}

	var schemas []string

	for schema := range preparedSchemas {
		schemas = append(schemas, schema)
	}

	if createPreparedSchemas, equal := util.SubstractStringSlices(schemas, currentSchemas); !equal {
		for _, schemaName := range createPreparedSchemas {
			owner := constants.OwnerRoleNameSuffix
			dbOwner := fmt.Sprintf("%s%s", databaseName, owner)
			if preparedSchemas[schemaName].DefaultRoles == nil || *preparedSchemas[schemaName].DefaultRoles {
				owner = fmt.Sprintf("%s_%s%s", databaseName, schemaName, owner)
			} else {
				owner = dbOwner
			}
			if err = c.executeCreateDatabaseSchema(databaseName, schemaName, dbOwner, owner); err != nil {
				errors = append(errors, err.Error())
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("error(s) while syncing schemas of prepared databases: %v", strings.Join(errors, `', '`))
	}

	return nil
}

func (c *Cluster) syncExtensions(extensions map[string]string) error {
	c.setProcessName("syncing database extensions")
	errors := make([]string, 0)
	createExtensions := make(map[string]string)
	alterExtensions := make(map[string]string)

	currentExtensions, err := c.getExtensions()
	if err != nil {
		return fmt.Errorf("could not get current database extensions: %v", err)
	}

	for extName, newSchema := range extensions {
		currentSchema, exists := currentExtensions[extName]
		if !exists {
			createExtensions[extName] = newSchema
		} else if currentSchema != newSchema {
			alterExtensions[extName] = newSchema
		}
	}

	for extName, schema := range createExtensions {
		if err = c.executeCreateExtension(extName, schema); err != nil {
			errors = append(errors, err.Error())
		}
	}
	for extName, schema := range alterExtensions {
		if err = c.executeAlterExtension(extName, schema); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("error(s) while syncing database extensions: %v", strings.Join(errors, `', '`))
	}

	return nil
}

func (c *Cluster) syncLogicalBackupJob() error {
	var (
		job        *batchv1.CronJob
		desiredJob *batchv1.CronJob
		err        error
	)
	c.setProcessName("syncing the logical backup job")

	// sync the job if it exists

	jobName := c.getLogicalBackupJobName()
	if job, err = c.KubeClient.CronJobsGetter.CronJobs(c.Namespace).Get(context.TODO(), jobName, metav1.GetOptions{}); err == nil {

		desiredJob, err = c.generateLogicalBackupJob()
		if err != nil {
			return fmt.Errorf("could not generate the desired logical backup job state: %v", err)
		}
		if match, reason := k8sutil.SameLogicalBackupJob(job, desiredJob); !match {
			c.logger.Infof("logical job %s is not in the desired state and needs to be updated",
				c.getLogicalBackupJobName(),
			)
			if reason != "" {
				c.logger.Infof("reason: %s", reason)
			}
			if err = c.patchLogicalBackupJob(desiredJob); err != nil {
				return fmt.Errorf("could not update logical backup job to match desired state: %v", err)
			}
			c.logger.Info("the logical backup job is synced")
		}
		return nil
	}
	if !k8sutil.ResourceNotFound(err) {
		return fmt.Errorf("could not get logical backp job: %v", err)
	}

	// no existing logical backup job, create new one
	c.logger.Info("could not find the cluster's logical backup job")

	if err = c.createLogicalBackupJob(); err == nil {
		c.logger.Infof("created missing logical backup job %s", jobName)
	} else {
		if !k8sutil.ResourceAlreadyExists(err) {
			return fmt.Errorf("could not create missing logical backup job: %v", err)
		}
		c.logger.Infof("logical backup job %s already exists", jobName)
		if _, err = c.KubeClient.CronJobsGetter.CronJobs(c.Namespace).Get(context.TODO(), jobName, metav1.GetOptions{}); err != nil {
			return fmt.Errorf("could not fetch existing logical backup job: %v", err)
		}
	}
	return nil
}

func (c *Cluster) syncPgbackrestConfig() error {
	if cm, err := c.KubeClient.ConfigMaps(c.Namespace).Get(context.TODO(), c.getPgbackrestConfigmapName(), metav1.GetOptions{}); err == nil {
		if err := c.updatePgbackrestConfig(cm); err != nil {
			return fmt.Errorf("could not update a pgbackrest config: %v", err)
		}
		c.logger.Info("a pgbackrest config has been successfully updated")
	} else {
		if err := c.createPgbackrestConfig(); err != nil {
			return fmt.Errorf("could not create a pgbackrest config: %v", err)
		}
		c.logger.Info("a pgbackrest config has been successfully created")
	}
	return nil
}

func (c *Cluster) syncPgbackrestRepoHostConfig(spec *cpov1.PostgresSpec) error {
	c.logger.Info("check if a sync for repo host configmap is needed ")

	repoNeeded := specHasPgbackrestPVCRepo(&c.Postgresql.Spec)

	c.setProcessName("Syncing pgbackrest repo-host")
	c.logger.Info("Syncing pgbackrest repo-host")

	curSts, err := c.KubeClient.StatefulSets(c.Namespace).Get(context.TODO(), c.getPgbackrestRepoHostName(), metav1.GetOptions{})
	if err != nil {
		if !k8sutil.ResourceNotFound(err) {
			return fmt.Errorf("Error during reading of repo-host statefulset: %v", err)
		}
		if repoNeeded {
			c.logger.Infof("Pgbackrest repo-host statefulset doesn't exist, creating")
			return c.createPgbackrestRepoHostObjects(spec)
		}
		// TODO: Should check and cleanup for other orphaned repo related objects?
		return nil
	} else {
		c.logger.Infof("Found existing pgbackrest repository")
		if !repoNeeded {
			c.logger.Infof("No pgbackrest repository host configured, deleting")
			return c.deletePgbackrestRepoHostObjects()
		}
	}

	desiredSts, err := c.generateRepoHostStatefulSet(spec)
	if err != nil {
		return fmt.Errorf("could not generate pgbackrest repo-host statefulset: %v", err)
	}

	cmp := c.compareStatefulSetWith(curSts, desiredSts)
	if !cmp.match {
		c.logStatefulSetChanges(curSts, desiredSts, false, cmp.reasons)

		// Replica count is only one that results in !cmp.replace and this is const 1 for repo host
		if err := c.replaceStatefulSet(&curSts, desiredSts); err != nil {
			return fmt.Errorf("could not replace pgbackrest repo-host statefulset: %v", err)
		}
	}

	if err = c.updatePgbackrestRepoHostConfig(); err != nil {
		return fmt.Errorf("could not update pgbackrest repo-host config: %v", err)
	}

	return nil
}

func (c *Cluster) syncPgbackrestJob(forceRemove bool) error {
	repos := []string{"repo1", "repo2", "repo3", "repo4"}
	schedules := []string{"full", "incr", "diff"}
	for _, rep := range repos {
		for _, schedul := range schedules {
			remove := true
			if !forceRemove && len(c.Postgresql.Spec.Backup.Pgbackrest.Repos) >= 1 {
				for _, repo := range c.Postgresql.Spec.Backup.Pgbackrest.Repos {
					for name, schedule := range repo.Schedule {
						if rep == repo.Name && name == schedul {
							job, err := c.generatePgbackrestJob(c.Postgresql.Spec.Backup.Pgbackrest, &repo, name, schedule)
							if err != nil {
								return fmt.Errorf("could not generate pgbackrest job: %v", err)
							}
							remove = false
							if _, err := c.KubeClient.CronJobsGetter.CronJobs(c.Namespace).Get(context.TODO(), c.getPgbackrestJobName(repo.Name, name), metav1.GetOptions{}); err == nil {
								if err := c.patchPgbackrestJob(job); err != nil {
									return fmt.Errorf("could not update a pgbackrest cronjob: %v", err)
								}
								c.logger.Infof("pgbackrest cronjob for %v %v has been successfully updated", rep, schedul)
							} else {
								if err := c.createPgbackrestJob(job); err != nil {
									return fmt.Errorf("could not create a pgbackrest cronjob: %v", err)
								}
								c.logger.Infof("pgbackrest cronjob for %v %v has been successfully created", rep, schedul)
							}
						}
					}
				}
			}
			if remove {
				deleted, err := c.deletePgbackrestJob(rep, schedul)
				if err != nil {
					c.logger.Warningf("failed to delete pgbackrest cronjob: %v", err)
				}
				if deleted {
					c.logger.Infof("pgbackrest cronjob for %v %v has been successfully deleted", rep, schedul)
				}
			}
		}
	}
	return nil
}

func (c *Cluster) createTDESecret() error {
	c.logger.Info("creating TDE secret")
	c.setProcessName("creating TDE secret")
	generatedKey := make([]byte, 16)
	rand.Read(generatedKey)

	generatedSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.getTDESecretName(),
			Namespace: c.Namespace,
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"key": []byte(fmt.Sprintf("%x", generatedKey)),
		},
	}
	secret, err := c.KubeClient.Secrets(generatedSecret.Namespace).Create(context.TODO(), &generatedSecret, metav1.CreateOptions{})
	if err == nil {
		c.Secrets[secret.UID] = secret
		c.logger.Debugf("created new secret %s, namespace: %s, uid: %s", util.NameFromMeta(secret.ObjectMeta), generatedSecret.Namespace, secret.UID)
	} else {
		return fmt.Errorf("could not create secret for TDE %s: in namespace %s: %v", util.NameFromMeta(secret.ObjectMeta), generatedSecret.Namespace, err)
	}

	return nil
}

func (c *Cluster) createMonitoringSecret() error {
	c.logger.Info("creating Monitoring secret")
	c.setProcessName("creating Monitoring secret")
	generatedKey := make([]byte, 16)
	rand.Read(generatedKey)

	generatedSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.getMonitoringSecretName(),
			Namespace: c.Namespace,
			Labels:    c.labelsSet(true),
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte(c.getMonitoringSecretName()),
			"password": []byte(fmt.Sprintf("%x", generatedKey)),
		},
	}
	secret, err := c.KubeClient.Secrets(generatedSecret.Namespace).Create(context.TODO(), &generatedSecret, metav1.CreateOptions{})
	if err == nil {
		c.Secrets[secret.UID] = secret
		c.logger.Debugf("created new secret %s, namespace: %s, uid: %s", util.NameFromMeta(secret.ObjectMeta), generatedSecret.Namespace, secret.UID)
	} else {
		if !k8sutil.ResourceAlreadyExists(err) {
			return fmt.Errorf("could not create secret for Monitoring %s: in namespace %s: %v", util.NameFromMeta(secret.ObjectMeta), generatedSecret.Namespace, err)
		}
	}

	return nil
}

// delete monitoring secret
func (c *Cluster) deleteMonitoringSecret() (err error) {
	// Repeat the same for the secret object
	secretName := c.getMonitoringSecretName()

	secret, err := c.KubeClient.
		Secrets(c.Namespace).
		Get(context.TODO(), secretName, metav1.GetOptions{})

	if err != nil {
		c.logger.Debugf("could not get monitoring secret %s: %v", secretName, err)
	} else {
		if err = c.deleteSecret(secret.UID, *secret); err != nil {
			return fmt.Errorf("could not delete monitoring secret: %v", err)
		}
	}
	return nil
}

// Sync monitoring
// In case of monitoring is added/deleted, we need to
// 1. Update sts to in/exclude the exporter contianer
// 2. Add/Delete the respective user
// 3. Add/Delete the respective secret
// Point 1 and 2 are taken care in Update func, so we only need to take care
// Point 3 here.
func (c *Cluster) syncMonitoringSecret(oldSpec, newSpec *cpov1.Postgresql) error {
	c.logger.Info("syncing Monitoring secret")
	c.setProcessName("syncing Monitoring secret")

	if newSpec.Spec.Monitoring != nil && oldSpec.Spec.Monitoring == nil {
		// Create monitoring secret
		if err := c.createMonitoringSecret(); err != nil {
			return fmt.Errorf("could not create the monitoring secret: %v", err)
		}
		c.logger.Info("monitoring secret was successfully created")
	} else if newSpec.Spec.Monitoring == nil && oldSpec.Spec.Monitoring != nil {
		// Delete the monitoring secret
		if err := c.deleteMonitoringSecret(); err != nil {
			return fmt.Errorf("could not delete the monitoring secret: %v", err)
		}
		c.logger.Info("monitoring secret was successfully deleted")
	}
	return nil
}

func (c *Cluster) syncWalPvc(oldSpec, newSpec *cpov1.Postgresql) error {
	c.logger.Info("syncing PVC for WAL")
	c.setProcessName("syncing PVC for WAL")

	if newSpec.Spec.WalPvc == nil && oldSpec.Spec.WalPvc != nil {
		pvcs, err := c.listPersistentVolumeClaims()
		if err != nil {
			return fmt.Errorf("Could not list PVCs")
		} else {
			for _, pvc := range pvcs {
				if strings.Contains(pvc.Name, getWALPVCName(c.Spec.ClusterName)) {
					c.logger.Infof("deleting WAL-PVC %q", util.NameFromMeta(pvc.ObjectMeta))
					if err := c.KubeClient.PersistentVolumeClaims(pvc.Namespace).Delete(context.TODO(), pvc.Name, c.deleteOptions); err != nil {
						return fmt.Errorf("could not delete WAL PVC: %v", err)
					}
				}
			}
		}
	}
	return nil
}

func generateRootCertificate(
	privateKey *ecdsa.PrivateKey, serialNumber *big.Int,
) (*x509.Certificate, error) {
	const rootCommonName = "postgres-operator-ca"
	const rootExpiration = time.Hour * 24 * 365 * 10
	const rootStartValid = time.Hour * -1

	now := currentTime()
	template := &x509.Certificate{
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		MaxPathLenZero:        true, // there are no intermediate certificates
		NotBefore:             now.Add(rootStartValid),
		NotAfter:              now.Add(rootExpiration),
		SerialNumber:          serialNumber,
		SignatureAlgorithm:    certificateSignatureAlgorithm,
		Subject: pkix.Name{
			CommonName: rootCommonName,
		},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	// A root certificate is self-signed, so pass in the template twice.
	bytes, err := x509.CreateCertificate(rand.Reader, template, template,
		privateKey.Public(), privateKey)

	parsed, _ := x509.ParseCertificate(bytes)
	return parsed, err
}

func generateLeafCertificate(
	signer *x509.Certificate, signerPrivate *ecdsa.PrivateKey,
	signeePublic *ecdsa.PublicKey, serialNumber *big.Int,
	commonName string, dnsNames []string, server bool,
) (*x509.Certificate, error) {
	const leafExpiration = time.Hour * 24 * 365
	const leafStartValid = time.Hour * -1
	extKey := []x509.ExtKeyUsage{
		x509.ExtKeyUsageClientAuth,
		x509.ExtKeyUsageServerAuth,
	}

	now := currentTime()
	template := &x509.Certificate{
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           extKey,
		NotBefore:             now.Add(leafStartValid),
		NotAfter:              now.Add(leafExpiration),
		SerialNumber:          serialNumber,
		SignatureAlgorithm:    certificateSignatureAlgorithm,
		Subject: pkix.Name{
			CommonName: commonName,
		},
	}

	bytes, err := x509.CreateCertificate(rand.Reader, template, signer,
		signeePublic, signerPrivate)

	parsed, _ := x509.ParseCertificate(bytes)
	return parsed, err
}

// GenerateLeafCertificate generates a new key and certificate signed by root.
func (root *RootCertificateAuthority) GenerateLeafCertificate(
	commonName string, dnsNames []string, server bool,
) (*LeafCertificate, error) {
	var leaf LeafCertificate
	var serial *big.Int

	key, err := generateKey()
	if err == nil {
		serial, err = generateSerialNumber()
	}
	if err == nil {
		leaf.PrivateKey.ecdsa = key
		leaf.Certificate.x509, err = generateLeafCertificate(
			root.Certificate.x509, root.PrivateKey.ecdsa, &key.PublicKey, serial,
			commonName, dnsNames, server)
	}

	return &leaf, err
}

// NewRootCertificateAuthority generates a new key and self-signed certificate
// for issuing other certificates.
func NewRootCertificateAuthority() (*RootCertificateAuthority, error) {
	var root RootCertificateAuthority
	var serial *big.Int

	key, err := generateKey()
	if err == nil {
		serial, err = generateSerialNumber()
	}
	if err == nil {
		root.PrivateKey.ecdsa = key
		root.Certificate.x509, err = generateRootCertificate(key, serial)
	}

	return &root, err
}

// certFile concatenates the results of multiple PEM-encoding marshalers.
func certFile(texts ...encoding.TextMarshaler) ([]byte, error) {
	var out []byte

	for i := range texts {
		if b, err := texts[i].MarshalText(); err == nil {
			out = append(out, b...)
		} else {
			return nil, err
		}
	}

	return out, nil
}

// clientCommonName returns a client certificate common name (CN) for cluster.
func (c *Cluster) clientCommonName() string {
	// The common name (ASN.1 OID 2.5.4.3) of a certificate must be
	// 64 characters or less. ObjectMeta.UID is a UUID in its 36-character
	// string representation.
	// - https://tools.ietf.org/html/rfc5280#appendix-A
	// - https://docs.k8s.io/concepts/overview/working-with-objects/names/#uids
	// - https://releases.k8s.io/v1.22.0/staging/src/k8s.io/apiserver/pkg/registry/rest/create.go#L111
	// - https://releases.k8s.io/v1.22.0/staging/src/k8s.io/apiserver/pkg/registry/rest/meta.go#L30
	return c.clusterName().Name
}

// ByteMap initializes m when it points to nil.
func ByteMap(m *map[string][]byte) {
	if m != nil && *m == nil {
		*m = make(map[string][]byte)
	}
}

func (c *Cluster) createPgbackrestCertSecret(secretname string) error {
	c.logger.Info("creating PVC secret")
	c.setProcessName("creating PVC secret")
	generatedKey := make([]byte, 16)
	rand.Read(generatedKey)

	// Save the CA and generate a TLS client certificate for the entire cluster.

	// The server verifies its "tls-server-auth" option contains the common
	// name (CN) of the certificate presented by a client. The entire
	// cluster uses a single client certificate so the "tls-server-auth"
	// option can stay the same when PostgreSQL instances and repository
	// hosts are added or removed.
	leaf := &LeafCertificate{}
	commonName := c.clientCommonName()
	mainServiceName := "*." + c.clusterName().Name + "." + c.Namespace + ".svc." + c.OpConfig.ClusterDomain
	auxServiceName := "*." + c.serviceName(ClusterPods) + "." + c.Namespace + ".svc." + c.OpConfig.ClusterDomain
	dnsNames := []string{commonName, mainServiceName, auxServiceName}

	inRoot := &RootCertificateAuthority{}
	inRoot, err := NewRootCertificateAuthority()
	if err != nil {
		c.logger.Errorf("Error in certificate creation %v", err)
	}

	leaf, err = inRoot.GenerateLeafCertificate(commonName, dnsNames, false)

	if err != nil {
		c.logger.Errorf("could not generate root certificate %s", err)
	}
	var tsl_certAuthoritySecretKey, tsl_certClientPrivateKeySecretKey, tsl_certClientSecretKey, tsl_certRepoPrivateKeySecretKey, tsl_certRepoSecretKey []byte

	if err == nil {
		tsl_certAuthoritySecretKey, err = certFile(inRoot.Certificate)
	}
	if err == nil {
		tsl_certClientPrivateKeySecretKey, err = certFile(leaf.PrivateKey)
	}
	if err == nil {
		tsl_certClientSecretKey, _ = certFile(leaf.Certificate)
	}

	repo_leaf := &LeafCertificate{}
	repo_leaf, err = inRoot.GenerateLeafCertificate(commonName, dnsNames, true)

	if err == nil {
		tsl_certRepoPrivateKeySecretKey, err = certFile(repo_leaf.PrivateKey)
	}
	if err == nil {
		tsl_certRepoSecretKey, err = certFile(repo_leaf.Certificate)
	}
	if err != nil {
		c.logger.Errorf("Error in certificate creation %v", err)
	}
	generatedSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretname,
			Namespace: c.Namespace,
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"key":                         []byte(fmt.Sprintf("%x", generatedKey)),
			certAuthoritySecretKey:        tsl_certAuthoritySecretKey,
			certClientPrivateKeySecretKey: tsl_certClientPrivateKeySecretKey,
			certClientSecretKey:           tsl_certClientSecretKey,
			certRepoPrivateKeySecretKey:   tsl_certRepoPrivateKeySecretKey,
			certRepoSecretKey:             tsl_certRepoSecretKey,
		},
	}
	secret, err := c.KubeClient.Secrets(generatedSecret.Namespace).Create(context.TODO(), &generatedSecret, metav1.CreateOptions{})
	if err == nil {
		c.Secrets[secret.UID] = secret
		c.logger.Debugf("created new secret %s, namespace: %s, uid: %s", util.NameFromMeta(secret.ObjectMeta), generatedSecret.Namespace, secret.UID)
	} else {
		if !k8sutil.ResourceAlreadyExists(err) {
			return fmt.Errorf("could not create secret for PVC %s: in namespace %s: %v", util.NameFromMeta(secret.ObjectMeta), generatedSecret.Namespace, err)
		}
	}

	return nil
}
