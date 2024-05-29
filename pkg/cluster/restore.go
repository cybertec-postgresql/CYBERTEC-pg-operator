package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/constants"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

const (
	restoreSpecAnnotation = "InProgressRestoreSpec"
)

/*
Execute a restore when specified to do so and keep track of the progress to be able to retry.

The important events in the process of a restore are the following, with each following step adding to the
conditions of previous steps.

  - restore configmap.Data.restore_enabled set - All pods have been shut down. Restore state has been recorded into
    restore configmap, currently active restore ID is in configmap.Data.restore_id.
  - primary pod ready - Restore of primary has succeeded, there is a primary running on a new timeline with restored
    database contents.
  - replica count matches spec and all replicas are ready - Replicas have also been restored to desired state and are
    running.
  - Postgresql.Status.RestoreID == Restore.ID - Restore is successfully completed.
*/
func (c *Cluster) processRestore(spec *cpov1.Postgresql) error {
	/*
		Step 1: Prepare for restore by recording current spec state and marking the restore as in progress.

		To avoid restoring to multiple different backups or any other such issues a restore specification
		can not be modified after we have started the restore. This is enforced by having storing the restore
		in an annotation in the restore configmap and using this while the ID hasn't changed. To change restore
		settings a different ID needs to be used.

		This state has been completed when these conditions are true:
			Postgresql.Status.InProgressRestoreID == Spec.Backup.Pgbackrest.Restore.ID
			restoreConfigmap.Data["restore_enabled"] == "true"

	*/
	currentRestoreId := spec.Spec.GetBackup().GetRestoreID()
	if currentRestoreId == "" && c.RestoreCM != nil {
		// If there was a restore in progress, continue that
		currentRestoreId = c.RestoreCM.Data["restore_id"]
	}
	if currentRestoreId == "" {
		return fmt.Errorf("BUG: restore id should not be empty")
	}

	c.logger.Infof("Running restore job %s", currentRestoreId)

	var restore *cpov1.Restore
	if c.RestoreCM == nil || c.RestoreCM.Data["restore_id"] != currentRestoreId {
		// Scale main statefulset down to 0
		if err := c.stopCluster(); err != nil {
			return err
		}

		// Record in progress restore
		restore = &spec.Spec.GetBackup().Pgbackrest.Restore
		if err := c.syncRestoreConfigMap(restore); err != nil {
			return err
		}

		c.logger.Infof("Recorded restore specification: %s", restore)
		// After this state the contents of the main PVC could be broken, at least some restore must run to completion
	} else {
		// When this state has been completed used the recorded restore specification from previous time
		if err := json.Unmarshal([]byte(c.RestoreCM.Annotations[restoreSpecAnnotation]), &restore); err != nil {
			return fmt.Errorf("invalid json stored in restore specification: %s", err)
		}
		c.logger.Infof("Restore initialization already completed, continuing restore using restore specification: %s", restore)
	}

	c.KubeClient.SetPostgresCRDStatus(c.clusterName(), cpov1.ClusterStatusRestoring)

	/*
		Step 2: Run restore on pod-0 using init container. This is considered to be successful when the pod
		becomes ready.
	*/
	pod0, err := c.getPod0()
	if err != nil {
		return fmt.Errorf("error getting primary pod during restore: %s", err)
	}

	if !postgresPodIsReady(pod0) {
		c.logger.Debugf("Current replicas value: %v", *c.GetStatefulSet().Spec.Replicas)
		if !PointerEqualsValue(c.GetStatefulSet().Spec.Replicas, 1) {
			c.logger.Infof("Scaling number of pods to 1")
			if err := c.setStatefulSetReplicaCount(1); err != nil {
				return fmt.Errorf("unable to scale up stateful set for restore: %s", err)
			}
		}
		c.logger.Infof("waiting for first pod to become ready")
		if err := c.waitStatefulsetPodsReady(); err != nil {
			/* TODO: Right now this will error out after OpConfig.ResourceCheckTimeout. A retry running cluster sync
			should end up back here, effectively creating a retry loop. Don't want to run with no timeout to avoid
			endlessly tying up a sync worker for a potential denial of service problem when restoring too many large
			clusters at once. Ideally should fire off a background go routine that polls the state and fires off an
			event for the main sync loop when something interesting happens.
			*/
			return fmt.Errorf("error when waiting for pod 0 to become ready, retry will wait again: %s", err)
		}
	}
	c.logger.Infof("first pod restore has succeeded")

	/*
		Step 3: Run restore on replica pods, making sure that they don't try to create a new timeline.
		Restore is assumed to have succeeded when the pod becomes ready.
	*/

	if !PointerEqualsValue(c.GetStatefulSet().Spec.Replicas, spec.Spec.NumberOfInstances) {
		c.logger.Infof("scaling cluster back up to full size")
		c.setStatefulSetReplicaCount(int(spec.Spec.NumberOfInstances))
	}

	c.waitForAllPodsLabelReady()
	c.logger.Infof("All replicas have been restored")

	/*
		Step 4: Restore completed, cleanup.
	*/
	spec, err = c.KubeClient.SetCRDRestoreStatus(c.clusterName(), currentRestoreId)
	if err != nil {
		return fmt.Errorf("error recording restore crd status: %s", err)
	}

	err = c.deletePgbackrestRestoreConfig()
	if err != nil {
		return fmt.Errorf("error removing restore configuration: %s", err)
	}
	c.RestoreCM = nil

	c.Postgresql = *spec
	c.logger.Infof("Restore successfully completed")

	return nil
}

func PointerEqualsValue[T comparable](ptr *T, v T) bool {
	if ptr == nil {
		return false
	}
	return *ptr == v
}

func (c *Cluster) syncRestoreConfigMap(restore *cpov1.Restore) error {
	c.logger.Infof("Recording restore specification into restore configmap")

	pgbackrestRestoreConfigmapSpec, err := c.generatePgbackrestRestoreConfigmap(restore)
	if err != nil {
		return fmt.Errorf("could not generate pgbackrest restore configmap spec: %v", err)
	}
	c.logger.Debugf("Generated pgbackrest restore configmapSpec: %v", pgbackrestRestoreConfigmapSpec)

	if c.RestoreCM == nil {
		cm, err := c.KubeClient.ConfigMaps(c.Namespace).Create(context.TODO(), pgbackrestRestoreConfigmapSpec, metav1.CreateOptions{})
		if err == nil {
			c.RestoreCM = cm
			return nil
		}
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("could not create pgbackrest restore configmap: %v", err)
		}
	}
	c.logger.Infof("Overwriting existing restore configmap")

	cm, err := c.KubeClient.ConfigMaps(c.Namespace).Update(context.TODO(), pgbackrestRestoreConfigmapSpec, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("could not update pgbackrest restore configmap: %v", err)
	}
	c.RestoreCM = cm
	return nil
}

func (c *Cluster) configMapLabels() map[string]string {
	cmLabels := labels.Set(map[string]string{
		"member.cpo.opensource.cybertec.at/type": string(TYPE_POSTGRESQL),
		c.OpConfig.ClusterNameLabel:              c.Name,
	})
	return labels.Merge(c.OpConfig.ClusterLabels, cmLabels)
}

func (c *Cluster) generatePgbackrestRestoreConfigmap(restore *cpov1.Restore) (*v1.ConfigMap, error) {
	data := make(map[string]string)
	data["restore_enable"] = "true"
	data["restore_id"] = restore.ID

	optionsArray := make([]string, 0)
	for key, value := range restore.Options {
		optionsArray = append(optionsArray, fmt.Sprintf("--%s=%s", key, value))
	}
	options := strings.Join(optionsArray, " ")
	data["restore_command"] = fmt.Sprintf(" --repo=%v %s",
		repoNumberFromName(restore.Repo),
		options)

	restoreBytes, err := json.Marshal(restore)
	if err != nil {
		return nil, fmt.Errorf("error marshaling restore specification: %s", err)
	}

	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      c.getPgbackrestRestoreConfigmapName(),
			Annotations: map[string]string{
				restoreSpecAnnotation: string(restoreBytes),
			},
			Labels: c.configMapLabels(),
		},
		Data: data,
	}
	return configmap, nil
}

func (c *Cluster) patchRestoreConfigMap(addAnnotation []map[string]interface{}) error {
	patch, err := json.Marshal(addAnnotation)
	if err != nil {
		return err
	}

	_, err = c.KubeClient.ConfigMaps(c.Namespace).Patch(
		context.TODO(), c.getPgbackrestRestoreConfigmapName(), types.JSONPatchType, patch, metav1.PatchOptions{})
	return err
}

func (c *Cluster) setStatefulSetReplicaCount(n int) error {
	c.logger.Infof("Scaling database replica count to %v", n)
	patch, err := json.Marshal([]map[string]interface{}{
		{
			"op":    "add",
			"path":  "/spec/replicas",
			"value": n,
		},
	})
	if err != nil {
		return err
	}

	newSts, err := c.KubeClient.StatefulSets(c.Namespace).Patch(
		context.TODO(), c.Name, types.JSONPatchType, patch, metav1.PatchOptions{})
	if err == nil {
		c.Statefulset = newSts
	}
	return err
}

func (c *Cluster) stopCluster() error {
	c.logger.Infof("Making sure all database pods have been stopped")
	if err := c.setStatefulSetReplicaCount(0); err != nil {
		return err
	}
	if err := c.deletePods(); err != nil {
		return err
	}
	return nil
}

func (c *Cluster) getPod0() (*v1.Pod, error) {
	pods, err := c.listPods()
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		if pod.Name == c.Name+"-0" {
			return &pod, nil
		}
	}
	return nil, nil
}

func postgresPodIsReady(pod *v1.Pod) bool {
	if pod == nil {
		return false
	}

	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == constants.PostgresContainerName {
			return container.Ready
		}
	}
	return false
}

func (c *Cluster) applyRestoreStatefulSetSyncOverrides(newSts, oldSts *appsv1.StatefulSet) {
	c.logger.Debugf("Restore active: ignoring Replicas value %v, keeping value %v", newSts.Spec.Replicas, oldSts.Spec.Replicas)
	*newSts.Spec.Replicas = *oldSts.Spec.Replicas

	//TODO: right now changing selectors or labels mid-flight will probably make things break. Should we forbid those
	// from being changed?

	/* TODO: these are semi-optional. We will still operate with wrong config but suboptimally.
	// Want parallel pod management to bring up replicas all at once. Possibly should be configurable.
	newSts.Spec.PodManagementPolicy = appsv1.ParallelPodManagement
	// Want PVCs to be retained
	pvcRetentionPolicy := appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
		WhenDeleted: newSts.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted,
		WhenScaled:appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
	}
	newSts.Spec.PersistentVolumeClaimRetentionPolicy = &pvcRetentionPolicy
	*/
}

func (c *Cluster) restoreInProgress() bool {
	return c.RestoreCM != nil
}

func (c *Cluster) generateRestoreContainerEnvVars() ([]v1.EnvVar, error) {
	envVars := []v1.EnvVar{
		{
			Name:  "SCOPE",
			Value: c.Name,
		},
		{
			Name:  "PGROOT",
			Value: constants.PostgresDataPath,
		},
		{
			Name:  "PGUSER_SUPERUSER",
			Value: c.OpConfig.SuperUsername,
		},
		{
			Name: "PGPASSWORD_SUPERUSER",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c.credentialSecretName(c.OpConfig.SuperUsername),
					},
					Key: "password",
				},
			},
		},
		{
			Name:  "USE_PGBACKREST",
			Value: "true",
		},
	}

	return envVars, nil
}

func (c *Cluster) refreshRestoreConfigMap() error {
	cm, err := c.KubeClient.ConfigMaps(c.Namespace).Get(context.TODO(), c.getPgbackrestRestoreConfigmapName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			if c.RestoreCM != nil {
				c.logger.Warningf("Forgetting about in progress restore: %s", c.RestoreCM)
			}
			c.RestoreCM = nil
			return nil
		}
		return err
	}
	c.RestoreCM = cm
	return nil
}
