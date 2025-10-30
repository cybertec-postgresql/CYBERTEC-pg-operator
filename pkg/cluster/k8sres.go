package cluster

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	cpo "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at"
	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/spec"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/config"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/constants"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/k8sutil"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/patroni"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/retryutil"
	"golang.org/x/exp/maps"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	pgBinariesLocationTemplate     = "/usr/lib/postgresql/%v/bin"
	patroniPGBinariesParameterName = "bin_dir"
	patroniPGHBAConfParameterName  = "pg_hba"
	localHost                      = "127.0.0.1/32"
	scalyrSidecarName              = "scalyr-sidecar"
	logicalBackupContainerName     = "logical-backup"
	connectionPoolerContainer      = "connection-pooler"
	pgPort                         = 5432
	operatorPort                   = 8080
	monitorPort                    = 9187
	monitorUsername                = "cpo_exporter"
)

type pgUser struct {
	Password string   `json:"password"`
	Options  []string `json:"options"`
}

type patroniDCS struct {
	TTL                      uint32                       `json:"ttl,omitempty"`
	LoopWait                 uint32                       `json:"loop_wait,omitempty"`
	RetryTimeout             uint32                       `json:"retry_timeout,omitempty"`
	MaximumLagOnFailover     float32                      `json:"maximum_lag_on_failover,omitempty"`
	SynchronousMode          bool                         `json:"synchronous_mode,omitempty"`
	SynchronousModeStrict    bool                         `json:"synchronous_mode_strict,omitempty"`
	SynchronousNodeCount     uint32                       `json:"synchronous_node_count,omitempty"`
	PGBootstrapConfiguration map[string]interface{}       `json:"postgresql,omitempty"`
	Slots                    map[string]map[string]string `json:"slots,omitempty"`
	FailsafeMode             *bool                        `json:"failsafe_mode,omitempty"`
}

type pgBootstrap struct {
	Initdb []interface{}     `json:"initdb"`
	Users  map[string]pgUser `json:"users"`
	DCS    patroniDCS        `json:"dcs,omitempty"`
}

type spiloConfiguration struct {
	PgLocalConfiguration map[string]interface{} `json:"postgresql"`
	Bootstrap            pgBootstrap            `json:"bootstrap"`
}

func (c *Cluster) statefulSetName() string {
	return c.Name
}

func (c *Cluster) endpointName(role PostgresRole) string {
	name := c.Name
	if role == Replica {
		name = fmt.Sprintf("%s-%s", name, "repl")
	}

	return name
}

func (c *Cluster) serviceName(role PostgresRole) string {
	name := c.Name
	if role == Replica {
		name = fmt.Sprintf("%s-%s", name, "repl")
	}
	if role == ClusterPods {
		name = fmt.Sprintf("%s-%s", name, "clusterpods")
	}

	return name
}

func (c *Cluster) serviceAddress(role PostgresRole) string {
	service, exist := c.Services[role]

	if exist {
		return service.ObjectMeta.Name
	}

	defaultAddress := c.serviceName(role)
	c.logger.Warningf("No service for role %s - defaulting to %s", role, defaultAddress)
	return defaultAddress
}

func (c *Cluster) servicePort(role PostgresRole) int32 {
	service, exist := c.Services[role]

	if exist {
		return service.Spec.Ports[0].Port
	}

	c.logger.Warningf("No service for role %s - defaulting to port %d", role, pgPort)
	return pgPort
}

func (c *Cluster) podDisruptionBudgetName() string {
	return c.OpConfig.PDBNameFormat.Format("cluster", c.Name)
}

func makeDefaultResources(config *config.Config) cpov1.Resources {

	defaultRequests := cpov1.ResourceDescription{
		CPU:    config.Resources.DefaultCPURequest,
		Memory: config.Resources.DefaultMemoryRequest,
	}
	defaultLimits := cpov1.ResourceDescription{
		CPU:    config.Resources.DefaultCPULimit,
		Memory: config.Resources.DefaultMemoryLimit,
	}

	return cpov1.Resources{
		ResourceRequests: defaultRequests,
		ResourceLimits:   defaultLimits,
	}
}

func makeLogicalBackupResources(config *config.Config) cpov1.Resources {

	logicalBackupResourceRequests := cpov1.ResourceDescription{
		CPU:    config.LogicalBackup.LogicalBackupCPURequest,
		Memory: config.LogicalBackup.LogicalBackupMemoryRequest,
	}
	logicalBackupResourceLimits := cpov1.ResourceDescription{
		CPU:    config.LogicalBackup.LogicalBackupCPULimit,
		Memory: config.LogicalBackup.LogicalBackupMemoryLimit,
	}

	return cpov1.Resources{
		ResourceRequests: logicalBackupResourceRequests,
		ResourceLimits:   logicalBackupResourceLimits,
	}
}

func (c *Cluster) enforceMinResourceLimits(resources *v1.ResourceRequirements) error {
	var (
		isSmaller bool
		err       error
		msg       string
	)

	// setting limits too low can cause unnecessary evictions / OOM kills
	cpuLimit := resources.Limits[v1.ResourceCPU]
	minCPULimit := c.OpConfig.MinCPULimit
	if minCPULimit != "" {
		isSmaller, err = util.IsSmallerQuantity(cpuLimit.String(), minCPULimit)
		if err != nil {
			return fmt.Errorf("could not compare defined CPU limit %s for %q container with configured minimum value %s: %v",
				cpuLimit.String(), constants.PostgresContainerName, minCPULimit, err)
		}
		if isSmaller {
			msg = fmt.Sprintf("defined CPU limit %s for %q container is below required minimum %s and will be increased",
				cpuLimit.String(), constants.PostgresContainerName, minCPULimit)
			c.logger.Warningf("%s", msg)
			c.eventRecorder.Eventf(c.GetReference(), v1.EventTypeWarning, "ResourceLimits", msg)
			resources.Limits[v1.ResourceCPU], _ = resource.ParseQuantity(minCPULimit)
		}
	}

	memoryLimit := resources.Limits[v1.ResourceMemory]
	minMemoryLimit := c.OpConfig.MinMemoryLimit
	if minMemoryLimit != "" {
		isSmaller, err = util.IsSmallerQuantity(memoryLimit.String(), minMemoryLimit)
		if err != nil {
			return fmt.Errorf("could not compare defined memory limit %s for %q container with configured minimum value %s: %v",
				memoryLimit.String(), constants.PostgresContainerName, minMemoryLimit, err)
		}
		if isSmaller {
			msg = fmt.Sprintf("defined memory limit %s for %q container is below required minimum %s and will be increased",
				memoryLimit.String(), constants.PostgresContainerName, minMemoryLimit)
			c.logger.Warningf("%s", msg)
			c.eventRecorder.Eventf(c.GetReference(), v1.EventTypeWarning, "ResourceLimits", msg)
			resources.Limits[v1.ResourceMemory], _ = resource.ParseQuantity(minMemoryLimit)
		}
	}

	return nil
}

func (c *Cluster) enforceMaxResourceRequests(resources *v1.ResourceRequirements) error {
	var (
		err error
	)

	cpuRequest := resources.Requests[v1.ResourceCPU]
	maxCPURequest := c.OpConfig.MaxCPURequest
	maxCPU, err := util.MinResource(maxCPURequest, cpuRequest.String())
	if err != nil {
		return fmt.Errorf("could not compare defined CPU request %s for %q container with configured maximum value %s: %v",
			cpuRequest.String(), constants.PostgresContainerName, maxCPURequest, err)
	}
	resources.Requests[v1.ResourceCPU] = maxCPU

	memoryRequest := resources.Requests[v1.ResourceMemory]
	maxMemoryRequest := c.OpConfig.MaxMemoryRequest
	maxMemory, err := util.MinResource(maxMemoryRequest, memoryRequest.String())
	if err != nil {
		return fmt.Errorf("could not compare defined memory request %s for %q container with configured maximum value %s: %v",
			memoryRequest.String(), constants.PostgresContainerName, maxMemoryRequest, err)
	}
	resources.Requests[v1.ResourceMemory] = maxMemory

	return nil
}

func setMemoryRequestToLimit(resources *v1.ResourceRequirements, containerName string, logger *logrus.Entry) {

	requests := resources.Requests[v1.ResourceMemory]
	limits := resources.Limits[v1.ResourceMemory]
	isSmaller := requests.Cmp(limits) == -1
	if isSmaller {
		logger.Warningf("memory request of %s for %q container is increased to match memory limit of %s",
			requests.String(), containerName, limits.String())
		resources.Requests[v1.ResourceMemory] = limits
	}
}

func fillResourceList(spec cpov1.ResourceDescription, defaults cpov1.ResourceDescription) (v1.ResourceList, error) {
	var err error
	requests := v1.ResourceList{}

	if spec.CPU != "" {
		requests[v1.ResourceCPU], err = resource.ParseQuantity(spec.CPU)
		if err != nil {
			return nil, fmt.Errorf("could not parse CPU quantity: %v", err)
		}
	} else {
		requests[v1.ResourceCPU], err = resource.ParseQuantity(defaults.CPU)
		if err != nil {
			return nil, fmt.Errorf("could not parse default CPU quantity: %v", err)
		}
	}
	if spec.Memory != "" {
		requests[v1.ResourceMemory], err = resource.ParseQuantity(spec.Memory)
		if err != nil {
			return nil, fmt.Errorf("could not parse memory quantity: %v", err)
		}
	} else {
		requests[v1.ResourceMemory], err = resource.ParseQuantity(defaults.Memory)
		if err != nil {
			return nil, fmt.Errorf("could not parse default memory quantity: %v", err)
		}
	}

	return requests, nil
}

func (c *Cluster) generateResourceRequirements(
	resources *cpov1.Resources,
	defaultResources cpov1.Resources,
	containerName string) (*v1.ResourceRequirements, error) {
	var err error
	specRequests := cpov1.ResourceDescription{}
	specLimits := cpov1.ResourceDescription{}
	result := v1.ResourceRequirements{}

	if resources != nil {
		specRequests = resources.ResourceRequests
		specLimits = resources.ResourceLimits
	}

	result.Requests, err = fillResourceList(specRequests, defaultResources.ResourceRequests)
	if err != nil {
		return nil, fmt.Errorf("could not fill resource requests: %v", err)
	}

	result.Limits, err = fillResourceList(specLimits, defaultResources.ResourceLimits)
	if err != nil {
		return nil, fmt.Errorf("could not fill resource limits: %v", err)
	}

	// enforce minimum cpu and memory limits for Postgres containers only
	if containerName == constants.PostgresContainerName {
		if err = c.enforceMinResourceLimits(&result); err != nil {
			return nil, fmt.Errorf("could not enforce minimum resource limits: %v", err)
		}
	}

	if c.OpConfig.SetMemoryRequestToLimit {
		setMemoryRequestToLimit(&result, containerName, c.logger)
	}

	// enforce maximum cpu and memory requests for Postgres containers only
	if containerName == constants.PostgresContainerName {
		if err = c.enforceMaxResourceRequests(&result); err != nil {
			return nil, fmt.Errorf("could not enforce maximum resource requests: %v", err)
		}
	}

	return &result, nil
}

func generateSpiloJSONConfiguration(pg *cpov1.PostgresqlParam, patroni *cpov1.Patroni, opConfig *config.Config, enableTDE bool, logger *logrus.Entry) (string, error) {
	config := spiloConfiguration{}

	config.Bootstrap = pgBootstrap{}

	pgVersion, err := strconv.Atoi(pg.PgVersion)
	if err != nil {
		fmt.Println("Problem to get PGVersion:", err)
		pgVersion = 16
	}
	if pgVersion > 14 {
		config.Bootstrap.Initdb = []interface{}{map[string]string{"auth-host": "scram-sha-256"},
			map[string]string{"auth-local": "trust"},
			map[string]string{"encoding": "UTF8"},
			map[string]string{"locale": "en_US.UTF-8"},
			map[string]string{"locale-provider": "icu"},
			map[string]string{"icu-locale": "en_US"}}
	} else {
		config.Bootstrap.Initdb = []interface{}{map[string]string{"auth-host": "scram-sha-256"},
			map[string]string{"auth-local": "trust"}}
	}

	if enableTDE {
		config.Bootstrap.Initdb = append(config.Bootstrap.Initdb, map[string]string{"encryption-key-command": "/tmp/tde.sh"})
	}

	initdbOptionNames := []string{}

	for k := range patroni.InitDB {
		initdbOptionNames = append(initdbOptionNames, k)
	}
	/* We need to sort the user-defined options to more easily compare the resulting specs */
	sort.Strings(initdbOptionNames)

	// Initdb parameters in the manifest take priority over the default ones
	// The whole type switch dance is caused by the ability to specify both
	// maps and normal string items in the array of initdb options. We need
	// both to convert the initial key-value to strings when necessary, and
	// to de-duplicate the options supplied.
PatroniInitDBParams:
	for _, k := range initdbOptionNames {
		v := patroni.InitDB[k]
		for i, defaultParam := range config.Bootstrap.Initdb {
			switch t := defaultParam.(type) {
			case map[string]string:
				{
					for k1 := range t {
						if k1 == k {
							(config.Bootstrap.Initdb[i]).(map[string]string)[k] = v
							continue PatroniInitDBParams
						}
					}
				}
			case string:
				{
					/* if the option already occurs in the list */
					if t == v {
						continue PatroniInitDBParams
					}
				}
			default:
				logger.Warningf("unsupported type for initdb configuration item %s: %T", defaultParam, defaultParam)
				continue PatroniInitDBParams
			}
		}
		// The following options are known to have no parameters
		if v == "true" {
			switch k {
			case "data-checksums", "debug", "no-locale", "noclean", "nosync", "sync-only":
				config.Bootstrap.Initdb = append(config.Bootstrap.Initdb, k)
				continue
			}
		}
		config.Bootstrap.Initdb = append(config.Bootstrap.Initdb, map[string]string{k: v})
	}

	if patroni.MaximumLagOnFailover >= 0 {
		config.Bootstrap.DCS.MaximumLagOnFailover = patroni.MaximumLagOnFailover
	}
	if patroni.LoopWait != 0 {
		config.Bootstrap.DCS.LoopWait = patroni.LoopWait
	}
	if patroni.RetryTimeout != 0 {
		config.Bootstrap.DCS.RetryTimeout = patroni.RetryTimeout
	}
	if patroni.TTL != 0 {
		config.Bootstrap.DCS.TTL = patroni.TTL
	}
	if patroni.Slots != nil {
		config.Bootstrap.DCS.Slots = patroni.Slots
	}
	if patroni.SynchronousMode {
		config.Bootstrap.DCS.SynchronousMode = patroni.SynchronousMode
	}
	if patroni.SynchronousModeStrict {
		config.Bootstrap.DCS.SynchronousModeStrict = patroni.SynchronousModeStrict
	}
	if patroni.SynchronousNodeCount >= 1 {
		config.Bootstrap.DCS.SynchronousNodeCount = patroni.SynchronousNodeCount
	}
	if patroni.FailsafeMode != nil {
		config.Bootstrap.DCS.FailsafeMode = patroni.FailsafeMode
	} else if opConfig.EnablePatroniFailsafeMode != nil {
		config.Bootstrap.DCS.FailsafeMode = opConfig.EnablePatroniFailsafeMode
	}

	config.PgLocalConfiguration = make(map[string]interface{})

	// the newer and preferred way to specify the PG version is to use the `PGVERSION` env variable
	// setting postgresq.bin_dir in the SPILO_CONFIGURATION still works and takes precedence over PGVERSION
	// so we add postgresq.bin_dir only if PGVERSION is unused
	// see PR 222 in Spilo
	if !opConfig.EnablePgVersionEnvVar {
		config.PgLocalConfiguration[patroniPGBinariesParameterName] = fmt.Sprintf(pgBinariesLocationTemplate, pg.PgVersion)
	}
	if len(pg.Parameters) > 0 {
		local, bootstrap := getLocalAndBoostrapPostgreSQLParameters(pg.Parameters)

		if len(local) > 0 {
			config.PgLocalConfiguration[constants.PatroniPGParametersParameterName] = local
		}
		if len(bootstrap) > 0 {
			config.Bootstrap.DCS.PGBootstrapConfiguration = make(map[string]interface{})
			config.Bootstrap.DCS.PGBootstrapConfiguration[constants.PatroniPGParametersParameterName] = bootstrap
		}
	}
	// Patroni gives us a choice of writing pg_hba.conf to either the bootstrap section or to the local postgresql one.
	// We choose the local one, because we need Patroni to change pg_hba.conf in PostgreSQL after the user changes the
	// relevant section in the manifest.
	if len(patroni.PgHba) > 0 {
		config.PgLocalConfiguration[patroniPGHBAConfParameterName] = patroni.PgHba
	}

	res, err := json.Marshal(config)
	return string(res), err
}

func getLocalAndBoostrapPostgreSQLParameters(parameters map[string]string) (local, bootstrap map[string]string) {
	local = make(map[string]string)
	bootstrap = make(map[string]string)
	for param, val := range parameters {
		if isBootstrapOnlyParameter(param) {
			bootstrap[param] = val
		} else {
			local[param] = val
		}
	}
	return
}

func generateCapabilities(capabilities []string) *v1.Capabilities {
	additionalCapabilities := make([]v1.Capability, 0, len(capabilities))
	for _, capability := range capabilities {
		additionalCapabilities = append(additionalCapabilities, v1.Capability(strings.ToUpper(capability)))
	}
	if len(additionalCapabilities) > 0 {
		return &v1.Capabilities{
			Add: additionalCapabilities,
		}
	}
	return nil
}

func (c *Cluster) nodeAffinity(nodeReadinessLabel map[string]string, nodeAffinity *v1.NodeAffinity) *v1.Affinity {
	if len(nodeReadinessLabel) == 0 && nodeAffinity == nil {
		return nil
	}
	nodeAffinityCopy := v1.NodeAffinity{}
	if nodeAffinity != nil {
		nodeAffinityCopy = *nodeAffinity.DeepCopy()
	}
	if len(nodeReadinessLabel) > 0 {
		matchExpressions := make([]v1.NodeSelectorRequirement, 0)
		for k, v := range nodeReadinessLabel {
			matchExpressions = append(matchExpressions, v1.NodeSelectorRequirement{
				Key:      k,
				Operator: v1.NodeSelectorOpIn,
				Values:   []string{v},
			})
		}
		nodeReadinessSelectorTerm := v1.NodeSelectorTerm{MatchExpressions: matchExpressions}
		if nodeAffinityCopy.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			nodeAffinityCopy.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					nodeReadinessSelectorTerm,
				},
			}
		} else {
			if c.OpConfig.NodeReadinessLabelMerge == "OR" {
				manifestTerms := nodeAffinityCopy.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
				manifestTerms = append(manifestTerms, nodeReadinessSelectorTerm)
				nodeAffinityCopy.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{
					NodeSelectorTerms: manifestTerms,
				}
			} else if c.OpConfig.NodeReadinessLabelMerge == "AND" {
				for i, nodeSelectorTerm := range nodeAffinityCopy.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
					manifestExpressions := nodeSelectorTerm.MatchExpressions
					manifestExpressions = append(manifestExpressions, matchExpressions...)
					nodeAffinityCopy.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i] = v1.NodeSelectorTerm{MatchExpressions: manifestExpressions}
				}
			}
		}
	}

	return &v1.Affinity{
		NodeAffinity: &nodeAffinityCopy,
	}
}

func podAffinity(
	labels labels.Set,
	topologyKey string,
	nodeAffinity *v1.Affinity,
	preferredDuringScheduling bool,
	anti bool) *v1.Affinity {

	var podAffinity v1.Affinity

	podAffinityTerm := v1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		TopologyKey: topologyKey,
	}

	if anti {
		podAffinity.PodAntiAffinity = generatePodAntiAffinity(podAffinityTerm, preferredDuringScheduling)
	} else {
		podAffinity.PodAffinity = generatePodAffinity(podAffinityTerm, preferredDuringScheduling)
	}

	if nodeAffinity != nil && nodeAffinity.NodeAffinity != nil {
		podAffinity.NodeAffinity = nodeAffinity.NodeAffinity
	}

	return &podAffinity
}

func generatePodAffinity(podAffinityTerm v1.PodAffinityTerm, preferredDuringScheduling bool) *v1.PodAffinity {
	podAffinity := &v1.PodAffinity{}

	if preferredDuringScheduling {
		podAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []v1.WeightedPodAffinityTerm{{
			Weight:          1,
			PodAffinityTerm: podAffinityTerm,
		}}
	} else {
		podAffinity.RequiredDuringSchedulingIgnoredDuringExecution = []v1.PodAffinityTerm{podAffinityTerm}
	}

	return podAffinity
}

func generatePodAntiAffinity(podAffinityTerm v1.PodAffinityTerm, preferredDuringScheduling bool) *v1.PodAntiAffinity {
	podAntiAffinity := &v1.PodAntiAffinity{}

	if preferredDuringScheduling {
		podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []v1.WeightedPodAffinityTerm{{
			Weight:          1,
			PodAffinityTerm: podAffinityTerm,
		}}
	} else {
		podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = []v1.PodAffinityTerm{podAffinityTerm}
	}

	return podAntiAffinity
}

func tolerations(tolerationsSpec *[]v1.Toleration, podToleration map[string]string) []v1.Toleration {
	// allow to override tolerations by postgresql manifest
	if len(*tolerationsSpec) > 0 {
		return *tolerationsSpec
	}

	if len(podToleration["key"]) > 0 ||
		len(podToleration["operator"]) > 0 ||
		len(podToleration["value"]) > 0 ||
		len(podToleration["effect"]) > 0 {

		return []v1.Toleration{
			{
				Key:      podToleration["key"],
				Operator: v1.TolerationOperator(podToleration["operator"]),
				Value:    podToleration["value"],
				Effect:   v1.TaintEffect(podToleration["effect"]),
			},
		}
	}

	return []v1.Toleration{}
}

func topologySpreadConstraints(topologySpreadConstraintsSpec *[]v1.TopologySpreadConstraint) []v1.TopologySpreadConstraint {
	// allow to override tolerations by postgresql manifest
	if len(*topologySpreadConstraintsSpec) > 0 {
		return *topologySpreadConstraintsSpec
	}

	return []v1.TopologySpreadConstraint{}
}

// isBootstrapOnlyParameter checks against special Patroni bootstrap parameters.
// Those parameters must go to the bootstrap/dcs/postgresql/parameters section.
// See http://patroni.readthedocs.io/en/latest/dynamic_configuration.html.
func isBootstrapOnlyParameter(param string) bool {
	params := map[string]bool{
		"archive_command":                  false,
		"shared_buffers":                   false,
		"logging_collector":                false,
		"log_destination":                  false,
		"log_directory":                    false,
		"log_filename":                     false,
		"log_file_mode":                    false,
		"log_rotation_age":                 false,
		"log_truncate_on_rotation":         false,
		"ssl":                              false,
		"ssl_ca_file":                      false,
		"ssl_crl_file":                     false,
		"ssl_cert_file":                    false,
		"ssl_key_file":                     false,
		"shared_preload_libraries":         false,
		"bg_mon.listen_address":            false,
		"bg_mon.history_buckets":           false,
		"pg_stat_statements.track_utility": false,
		"extwlist.extensions":              false,
		"extwlist.custom_path":             false,
	}
	result, ok := params[param]
	if !ok {
		result = true
	}
	return result
}

func generateVolumeMounts(volume cpov1.Volume) []v1.VolumeMount {
	return []v1.VolumeMount{
		{
			Name:      constants.DataVolumeName,
			MountPath: constants.PostgresDataMount, //TODO: fetch from manifest
			SubPath:   volume.SubPath,
		},
	}
}

func generateContainer(
	name string,
	dockerImage *string,
	resourceRequirements *v1.ResourceRequirements,
	envVars []v1.EnvVar,
	volumeMounts []v1.VolumeMount,
	privilegedMode bool,
	privilegeEscalationMode *bool,
	readOnlyRootFilesystem *bool,
	additionalPodCapabilities *v1.Capabilities,
) *v1.Container {
	return &v1.Container{
		Name:            name,
		Image:           *dockerImage,
		ImagePullPolicy: v1.PullIfNotPresent,
		Resources:       *resourceRequirements,
		Ports: []v1.ContainerPort{
			{
				ContainerPort: patroni.ApiPort,
				Protocol:      v1.ProtocolTCP,
			},
			{
				ContainerPort: pgPort,
				Protocol:      v1.ProtocolTCP,
			},
			{
				ContainerPort: operatorPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
		VolumeMounts: volumeMounts,
		Env:          envVars,
		SecurityContext: &v1.SecurityContext{
			AllowPrivilegeEscalation: privilegeEscalationMode,
			Privileged:               &privilegedMode,
			ReadOnlyRootFilesystem:   readOnlyRootFilesystem,
			Capabilities:             additionalPodCapabilities,
		},
	}
}

func (c *Cluster) generateSidecarContainers(sidecars []cpov1.Sidecar,
	defaultResources cpov1.Resources, startIndex int) ([]v1.Container, error) {

	if len(sidecars) > 0 {
		result := make([]v1.Container, 0)
		for index, sidecar := range sidecars {
			var resourcesSpec cpov1.Resources
			if sidecar.Resources == nil {
				resourcesSpec = cpov1.Resources{}
			} else {
				sidecar.Resources.DeepCopyInto(&resourcesSpec)
			}

			resources, err := c.generateResourceRequirements(&resourcesSpec, defaultResources, sidecar.Name)
			if err != nil {
				return nil, err
			}

			sc := getSidecarContainer(sidecar, startIndex+index, resources)
			result = append(result, *sc)
		}
		return result, nil
	}
	return nil, nil
}

// adds common fields to sidecars
func patchSidecarContainers(in []v1.Container, volumeMounts []v1.VolumeMount, superUserName string, credentialsSecretName string, logger *logrus.Entry, privilegedMode bool, privilegeEscalationMode *bool, additionalPodCapabilities *v1.Capabilities) []v1.Container {
	result := []v1.Container{}

	for _, container := range in {
		container.VolumeMounts = append(container.VolumeMounts, volumeMounts...)
		env := []v1.EnvVar{
			{
				Name: "POD_NAME",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.name",
					},
				},
			},
			{
				Name: "POD_NAMESPACE",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.namespace",
					},
				},
			},
			{
				Name:  "POSTGRES_USER",
				Value: superUserName,
			},
			{
				Name: "POSTGRES_PASSWORD",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: credentialsSecretName,
						},
						Key: "password",
					},
				},
			},
		}
		container.Env = appendEnvVars(env, container.Env...)

		result = append(result, container)
	}

	return result
}

// Check whether or not we're requested to mount an shm volume,
// taking into account that PostgreSQL manifest has precedence.
func mountShmVolumeNeeded(opConfig config.Config, spec *cpov1.PostgresSpec) *bool {
	if spec.ShmVolume != nil && *spec.ShmVolume {
		return spec.ShmVolume
	}

	return opConfig.ShmVolume
}

func (c *Cluster) generatePodTemplate(
	namespace string,
	labels labels.Set,
	annotations map[string]string,
	spiloContainer *v1.Container,
	initContainers []v1.Container,
	sidecarContainers []v1.Container,
	sharePgSocketWithSidecars *bool,
	tolerationsSpec *[]v1.Toleration,
	topologySpreadConstraintsSpec *[]v1.TopologySpreadConstraint,
	spiloRunAsUser *int64,
	spiloRunAsGroup *int64,
	spiloFSGroup *int64,
	nodeAffinity *v1.Affinity,
	schedulerName *string,
	terminateGracePeriod int64,
	podServiceAccountName string,
	kubeIAMRole string,
	priorityClassName string,
	shmVolume *bool,
	podAntiAffinity bool,
	podAntiAffinityTopologyKey string,
	podAntiAffinityPreferredDuringScheduling bool,
	additionalSecretMount string,
	additionalSecretMountPath string,
	additionalVolumes []cpov1.AdditionalVolume,
	isRepoHost bool,
) (*v1.PodTemplateSpec, error) {

	terminateGracePeriodSeconds := terminateGracePeriod
	containers := []v1.Container{*spiloContainer}
	containers = append(containers, sidecarContainers...)
	securityContext := v1.PodSecurityContext{}

	if spiloRunAsUser != nil {
		securityContext.RunAsUser = spiloRunAsUser
	}

	if spiloRunAsGroup != nil {
		securityContext.RunAsGroup = spiloRunAsGroup
	}

	if spiloFSGroup != nil {
		securityContext.FSGroup = spiloFSGroup
	}

	podSpec := v1.PodSpec{
		ServiceAccountName:            podServiceAccountName,
		TerminationGracePeriodSeconds: &terminateGracePeriodSeconds,
		Containers:                    containers,
		InitContainers:                initContainers,
		Tolerations:                   *tolerationsSpec,
		TopologySpreadConstraints:     *topologySpreadConstraintsSpec,
		SecurityContext:               &securityContext,
	}

	if schedulerName != nil {
		podSpec.SchedulerName = *schedulerName
	}

	if shmVolume != nil && *shmVolume {
		addShmVolume(&podSpec)
	}

	if podAntiAffinity {
		podSpec.Affinity = podAffinity(
			labels,
			podAntiAffinityTopologyKey,
			nodeAffinity,
			podAntiAffinityPreferredDuringScheduling,
			true,
		)
	} else if nodeAffinity != nil {
		podSpec.Affinity = nodeAffinity
	}

	if priorityClassName != "" {
		podSpec.PriorityClassName = priorityClassName
	}

	if c.Postgresql.Spec.Monitoring != nil {
		addEmptyDirVolume(&podSpec, "exporter-tmp", "postgres-exporter", "/tmp")
	}

	if c.OpConfig.ReadOnlyRootFilesystem != nil && *c.OpConfig.ReadOnlyRootFilesystem && !isRepoHost {
		addRunVolume(&podSpec, "postgres-run", "postgres", "/run")
		addEmptyDirVolume(&podSpec, "postgres-tmp", "postgres", "/tmp")
	}

	if c.OpConfig.ReadOnlyRootFilesystem != nil && *c.OpConfig.ReadOnlyRootFilesystem && isRepoHost {
		addEmptyDirVolume(&podSpec, "pgbackrest-tmp", "pgbackrest", "/tmp")
	}

	if sharePgSocketWithSidecars != nil && *sharePgSocketWithSidecars {
		addVarRunVolume(&podSpec)
	}

	if additionalSecretMount != "" {
		addSecretVolume(&podSpec, additionalSecretMount, additionalSecretMountPath)
	}

	if additionalVolumes != nil {
		c.addAdditionalVolumes(&podSpec, additionalVolumes)
	}

	if c.Postgresql.Spec.Monitoring != nil {
		labels["cpo_monitoring_stack"] = "true"
	}

	template := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: podSpec,
	}
	if kubeIAMRole != "" {
		if template.Annotations == nil {
			template.Annotations = make(map[string]string)
		}
		template.Annotations[constants.KubeIAmAnnotation] = kubeIAMRole
	}

	return &template, nil
}

// generatePodEnvVars generates environment variables for the Spilo Pod
func (c *Cluster) generateSpiloPodEnvVars(
	spec *cpov1.PostgresSpec,
	uid types.UID,
	spiloConfiguration string) ([]v1.EnvVar, error) {

	// hard-coded set of environment variables we need
	// to guarantee core functionality of the operator
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
			Name: "POD_IP",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name:  "PGUSER_SUPERUSER",
			Value: c.OpConfig.SuperUsername,
		},
		{
			Name:  "KUBERNETES_SCOPE_LABEL",
			Value: c.OpConfig.ClusterNameLabel,
		},
		{
			Name:  "KUBERNETES_ROLE_LABEL",
			Value: c.OpConfig.PodRoleLabel,
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
			Name:  "PGUSER_STANDBY",
			Value: c.OpConfig.ReplicationUsername,
		},
		{
			Name: "PGPASSWORD_STANDBY",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c.credentialSecretName(c.OpConfig.ReplicationUsername),
					},
					Key: "password",
				},
			},
		},
		{
			Name:  "PAM_OAUTH2",
			Value: c.OpConfig.PamConfiguration,
		},
		{
			Name:  "HUMAN_ROLE",
			Value: c.OpConfig.PamRoleName,
		},
		// NSS WRAPPER
		{
			Name:  "LD_PRELOAD",
			Value: "/usr/lib64/libnss_wrapper.so",
		},
		{
			Name:  "NSS_WRAPPER_PASSWD",
			Value: "/tmp/nss_wrapper/passwd",
		},
		{
			Name:  "NSS_WRAPPER_GROUP",
			Value: "/tmp/nss_wrapper/group",
		},
	}

	if c.OpConfig.EnableSpiloWalPathCompat {
		envVars = append(envVars, v1.EnvVar{Name: "ENABLE_WAL_PATH_COMPAT", Value: "true"})
	}

	if spec.Backup != nil && spec.Backup.Pgbackrest != nil {
		envVars = append(envVars, v1.EnvVar{Name: "USE_PGBACKREST", Value: "true"})
	}

	if c.OpConfig.ReadOnlyRootFilesystem != nil && *c.OpConfig.ReadOnlyRootFilesystem {
		envVars = append(envVars, v1.EnvVar{Name: "HOME", Value: "/home/postgres"})
	}

	if spec.TDE != nil && spec.TDE.Enable {
		envVars = append(envVars, v1.EnvVar{Name: "TDE", Value: "true"})
		// envVars = append(envVars, v1.EnvVar{Name: "PGENCRKEYCMD", Value: "/tmp/tde.sh"})
		envVars = append(envVars, v1.EnvVar{Name: "TDE_KEY", ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: c.getTDESecretName(),
				},
				Key: "key",
			},
		},
		})
	}
	if spec.Monitoring != nil {
		envVars = append(envVars, v1.EnvVar{Name: "cpo_monitoring_stack", Value: "true"})
	}

	if c.OpConfig.EnablePgVersionEnvVar {
		envVars = append(envVars, v1.EnvVar{Name: "PGVERSION", Value: c.GetDesiredMajorVersion()})
	}
	// Spilo expects cluster labels as JSON
	typeLabel := labels.Set(map[string]string{"member.cpo.opensource.cybertec.at/type": string(TYPE_POSTGRESQL)})
	databaseClusterLabels := labels.Merge(labels.Set(c.OpConfig.ClusterLabels), typeLabel)
	if clusterLabels, err := json.Marshal(databaseClusterLabels); err != nil {
		envVars = append(envVars, v1.EnvVar{Name: "KUBERNETES_LABELS", Value: databaseClusterLabels.String()})
	} else {
		envVars = append(envVars, v1.EnvVar{Name: "KUBERNETES_LABELS", Value: string(clusterLabels)})
	}
	if spiloConfiguration != "" {
		envVars = append(envVars, v1.EnvVar{Name: "SPILO_CONFIGURATION", Value: spiloConfiguration})
	}

	if c.patroniUsesKubernetes() {
		envVars = append(envVars, v1.EnvVar{Name: "DCS_ENABLE_KUBERNETES_API", Value: "true"})
	} else {
		envVars = append(envVars, v1.EnvVar{Name: "ETCD_HOST", Value: c.OpConfig.EtcdHost})
	}

	if c.patroniKubernetesUseConfigMaps() {
		envVars = append(envVars, v1.EnvVar{Name: "KUBERNETES_USE_CONFIGMAPS", Value: "true"})
	}

	// fetch cluster-specific variables that will override all subsequent global variables
	if len(spec.Env) > 0 {
		envVars = appendEnvVars(envVars, spec.Env...)
	}

	if spec.Clone != nil && (spec.Clone.ClusterName != "" || spec.Clone.Pgbackrest != nil) {
		envVars = append(envVars, c.generateCloneEnvironment(spec.Clone)...)
	}

	if spec.StandbyCluster != nil {
		envVars = append(envVars, c.generateStandbyEnvironment(spec.StandbyCluster)...)
	}

	// fetch variables from custom environment Secret
	// that will override all subsequent global variables
	secretEnvVarsList, err := c.getPodEnvironmentSecretVariables()
	if err != nil {
		return nil, err
	}
	envVars = appendEnvVars(envVars, secretEnvVarsList...)

	// fetch variables from custom environment ConfigMap
	// that will override all subsequent global variables
	configMapEnvVarsList, err := c.getPodEnvironmentConfigMapVariables()
	if err != nil {
		return nil, err
	}
	envVars = appendEnvVars(envVars, configMapEnvVarsList...)

	// global variables derived from operator configuration
	opConfigEnvVars := make([]v1.EnvVar, 0)
	if c.OpConfig.WALES3Bucket != "" {
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "WAL_S3_BUCKET", Value: c.OpConfig.WALES3Bucket})
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "WAL_BUCKET_SCOPE_SUFFIX", Value: getBucketScopeSuffix(string(uid))})
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "WAL_BUCKET_SCOPE_PREFIX", Value: ""})
	}

	if c.OpConfig.WALGSBucket != "" {
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "WAL_GS_BUCKET", Value: c.OpConfig.WALGSBucket})
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "WAL_BUCKET_SCOPE_SUFFIX", Value: getBucketScopeSuffix(string(uid))})
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "WAL_BUCKET_SCOPE_PREFIX", Value: ""})
	}

	if c.OpConfig.WALAZStorageAccount != "" {
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "AZURE_STORAGE_ACCOUNT", Value: c.OpConfig.WALAZStorageAccount})
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "WAL_BUCKET_SCOPE_SUFFIX", Value: getBucketScopeSuffix(string(uid))})
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "WAL_BUCKET_SCOPE_PREFIX", Value: ""})
	}

	if c.OpConfig.GCPCredentials != "" {
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: c.OpConfig.GCPCredentials})
	}

	if c.OpConfig.LogS3Bucket != "" {
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "LOG_S3_BUCKET", Value: c.OpConfig.LogS3Bucket})
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "LOG_BUCKET_SCOPE_SUFFIX", Value: getBucketScopeSuffix(string(uid))})
		opConfigEnvVars = append(opConfigEnvVars, v1.EnvVar{Name: "LOG_BUCKET_SCOPE_PREFIX", Value: ""})
	}
	if c.Postgresql.Spec.Backup != nil && c.Postgresql.Spec.Backup.Pgbackrest != nil {
		for _, repo := range c.Postgresql.Spec.Backup.Pgbackrest.Repos {
			if repo.Storage == "pvc" {
				envVars = append(envVars, v1.EnvVar{Name: "COMMAND", Value: "repo-host"})
			}
		}
	}

	envVars = appendEnvVars(envVars, opConfigEnvVars...)

	return envVars, nil
}

// generatePodEnvVars generates environment variables for the Spilo Pod
func (c *Cluster) generatepgBackRestPodEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name:  "USE_PGBACKREST",
			Value: "true",
		},
		{
			Name:  "MODE",
			Value: "repo",
		},
	}
}

func copyEnvVars(envs []v1.EnvVar) []v1.EnvVar {
	return append([]v1.EnvVar{}, envs...)
}

func appendEnvVars(envs []v1.EnvVar, appEnv ...v1.EnvVar) []v1.EnvVar {
	collectedEnvs := envs
	for _, env := range appEnv {
		if !isEnvVarPresent(collectedEnvs, env.Name) {
			collectedEnvs = append(collectedEnvs, env)
		}
	}
	return collectedEnvs
}

func isEnvVarPresent(envs []v1.EnvVar, key string) bool {
	for _, env := range envs {
		if strings.EqualFold(env.Name, key) {
			return true
		}
	}
	return false
}

// Return list of variables the pod received from the configured ConfigMap
func (c *Cluster) getPodEnvironmentConfigMapVariables() ([]v1.EnvVar, error) {
	configMapPodEnvVarsList := make([]v1.EnvVar, 0)

	if c.OpConfig.PodEnvironmentConfigMap.Name == "" {
		return configMapPodEnvVarsList, nil
	}

	cm, err := c.KubeClient.ConfigMaps(c.OpConfig.PodEnvironmentConfigMap.Namespace).Get(
		context.TODO(),
		c.OpConfig.PodEnvironmentConfigMap.Name,
		metav1.GetOptions{})
	if err != nil {
		// if not found, try again using the cluster's namespace if it's different (old behavior)
		if k8sutil.ResourceNotFound(err) && c.Namespace != c.OpConfig.PodEnvironmentConfigMap.Namespace {
			cm, err = c.KubeClient.ConfigMaps(c.Namespace).Get(
				context.TODO(),
				c.OpConfig.PodEnvironmentConfigMap.Name,
				metav1.GetOptions{})
		}
		if err != nil {
			return nil, fmt.Errorf("could not read PodEnvironmentConfigMap: %v", err)
		}
	}

	for k, v := range cm.Data {
		configMapPodEnvVarsList = append(configMapPodEnvVarsList, v1.EnvVar{Name: k, Value: v})
	}
	sort.Slice(configMapPodEnvVarsList, func(i, j int) bool { return configMapPodEnvVarsList[i].Name < configMapPodEnvVarsList[j].Name })
	return configMapPodEnvVarsList, nil
}

// Return list of variables the pod received from the configured Secret
func (c *Cluster) getPodEnvironmentSecretVariables() ([]v1.EnvVar, error) {
	secretPodEnvVarsList := make([]v1.EnvVar, 0)

	if c.OpConfig.PodEnvironmentSecret == "" {
		return secretPodEnvVarsList, nil
	}

	secret := &v1.Secret{}
	var notFoundErr error
	err := retryutil.Retry(c.OpConfig.ResourceCheckInterval, c.OpConfig.ResourceCheckTimeout,
		func() (bool, error) {
			var err error
			secret, err = c.KubeClient.Secrets(c.Namespace).Get(
				context.TODO(),
				c.OpConfig.PodEnvironmentSecret,
				metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					notFoundErr = err
					return false, nil
				}
				return false, err
			}
			return true, nil
		},
	)
	if notFoundErr != nil && err != nil {
		err = errors.Wrap(notFoundErr, err.Error())
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not read Secret PodEnvironmentSecretName")
	}

	for k := range secret.Data {
		secretPodEnvVarsList = append(secretPodEnvVarsList,
			v1.EnvVar{Name: k, ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c.OpConfig.PodEnvironmentSecret,
					},
					Key: k,
				},
			}})
	}

	sort.Slice(secretPodEnvVarsList, func(i, j int) bool { return secretPodEnvVarsList[i].Name < secretPodEnvVarsList[j].Name })
	return secretPodEnvVarsList, nil
}

func getSidecarContainer(sidecar cpov1.Sidecar, index int, resources *v1.ResourceRequirements) *v1.Container {
	name := sidecar.Name
	if name == "" {
		name = fmt.Sprintf("sidecar-%d", index)
	}

	return &v1.Container{
		Name:            name,
		Image:           sidecar.DockerImage,
		ImagePullPolicy: v1.PullIfNotPresent,
		Resources:       *resources,
		Env:             sidecar.Env,
		Ports:           sidecar.Ports,
		SecurityContext: sidecar.SecurityContext,
		VolumeMounts:    sidecar.VolumeMounts,
	}
}

func getBucketScopeSuffix(uid string) string {
	if uid != "" {
		return fmt.Sprintf("/%s", uid)
	}
	return ""
}

func makeResources(cpuRequest, memoryRequest, cpuLimit, memoryLimit string) cpov1.Resources {
	return cpov1.Resources{
		ResourceRequests: cpov1.ResourceDescription{
			CPU:    cpuRequest,
			Memory: memoryRequest,
		},
		ResourceLimits: cpov1.ResourceDescription{
			CPU:    cpuLimit,
			Memory: memoryLimit,
		},
	}
}

func extractPgVersionFromBinPath(binPath string, template string) (string, error) {
	var pgVersion float32
	_, err := fmt.Sscanf(binPath, template, &pgVersion)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", pgVersion), nil
}

func generateSpiloReadinessProbe() *v1.Probe {
	return &v1.Probe{
		FailureThreshold: 3,
		ProbeHandler: v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   "/readiness",
				Port:   intstr.IntOrString{IntVal: patroni.ApiPort},
				Scheme: v1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 6,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		TimeoutSeconds:      5,
	}
}

func generatePatroniLivenessProbe() *v1.Probe {
	return &v1.Probe{
		FailureThreshold: 6,
		ProbeHandler: v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   "/liveness",
				Port:   intstr.IntOrString{IntVal: patroni.ApiPort},
				Scheme: v1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 30,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
	}
}

func (c *Cluster) generateStatefulSet(spec *cpov1.PostgresSpec) (*appsv1.StatefulSet, error) {

	var (
		err                 error
		initContainers      []v1.Container
		sidecarContainers   []v1.Container
		podTemplate         *v1.PodTemplateSpec
		volumeClaimTemplate *v1.PersistentVolumeClaim
		additionalVolumes   = spec.AdditionalVolumes
	)

	defaultResources := makeDefaultResources(&c.OpConfig)
	resourceRequirements, err := c.generateResourceRequirements(
		spec.Resources, defaultResources, constants.PostgresContainerName)
	if err != nil {
		return nil, fmt.Errorf("could not generate resource requirements: %v", err)
	}

	if spec.InitContainers != nil && len(spec.InitContainers) > 0 {
		if c.OpConfig.EnableInitContainers != nil && !(*c.OpConfig.EnableInitContainers) {
			c.logger.Warningf("initContainers specified but disabled in configuration - next statefulset creation would fail")
		}
		initContainers = spec.InitContainers
	}

	// backward compatible check for InitContainers
	if spec.InitContainersOld != nil {
		msg := "manifest parameter init_containers is deprecated."
		if spec.InitContainers == nil {
			c.logger.Warningf("%s Consider using initContainers instead.", msg)
			spec.InitContainers = spec.InitContainersOld
		} else {
			c.logger.Warningf("%s Only value from initContainers is used", msg)
		}
	}

	// backward compatible check for PodPriorityClassName
	if spec.PodPriorityClassNameOld != "" {
		msg := "manifest parameter pod_priority_class_name is deprecated."
		if spec.PodPriorityClassName == "" {
			c.logger.Warningf("%s Consider using podPriorityClassName instead.", msg)
			spec.PodPriorityClassName = spec.PodPriorityClassNameOld
		} else {
			c.logger.Warningf("%s Only value from podPriorityClassName is used", msg)
		}
	}

	enableTDE := false
	if spec.TDE != nil && spec.TDE.Enable {
		enableTDE = true
	}
	spiloConfiguration, err := generateSpiloJSONConfiguration(&spec.PostgresqlParam, &spec.Patroni, &c.OpConfig, enableTDE, c.logger)
	if err != nil {
		return nil, fmt.Errorf("could not generate Spilo JSON configuration: %v", err)
	}

	// generate environment variables for the spilo container
	spiloEnvVars, err := c.generateSpiloPodEnvVars(spec, c.Postgresql.GetUID(), spiloConfiguration)
	if err != nil {
		return nil, fmt.Errorf("could not generate Spilo env vars: %v", err)
	}

	// pickup the docker image for the spilo container
	effectiveDockerImage := util.Coalesce(spec.DockerImage, c.OpConfig.DockerImage)

	// determine the User, Group and FSGroup for the spilo pod
	effectiveRunAsUser := c.OpConfig.Resources.SpiloRunAsUser
	if spec.SpiloRunAsUser != nil {
		effectiveRunAsUser = spec.SpiloRunAsUser
	}

	effectiveRunAsGroup := c.OpConfig.Resources.SpiloRunAsGroup
	if spec.SpiloRunAsGroup != nil {
		effectiveRunAsGroup = spec.SpiloRunAsGroup
	}

	effectiveFSGroup := c.OpConfig.Resources.SpiloFSGroup
	if spec.SpiloFSGroup != nil {
		effectiveFSGroup = spec.SpiloFSGroup
	}

	volumeMounts := generateVolumeMounts(spec.Volume)

	// configure TLS with a custom secret volume
	if spec.TLS != nil && spec.TLS.SecretName != "" {
		getSpiloTLSEnv := func(k string) string {
			keyName := ""
			switch k {
			case "tls.crt":
				keyName = "SSL_CERTIFICATE_FILE"
			case "tls.key":
				keyName = "SSL_PRIVATE_KEY_FILE"
			case "tls.ca":
				keyName = "SSL_CA_FILE"
			default:
				panic(fmt.Sprintf("TLS env key unknown %s", k))
			}
			// this is combined with the FSGroup in the section above
			// to give read access to the postgres user
			//defaultMode := int32(0644)
			defaultMode := int32(0600)
			mountPath := "/tls"
			additionalVolumes = append(additionalVolumes, cpov1.AdditionalVolume{
				Name:      spec.TLS.SecretName,
				MountPath: mountPath,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName:  spec.TLS.SecretName,
						DefaultMode: &defaultMode,
					},
				},
			})

			// use the same filenames as Secret resources by default
			certFile := ensurePath(spec.TLS.CertificateFile, mountPath, "tls.crt")
			privateKeyFile := ensurePath(spec.TLS.PrivateKeyFile, mountPath, "tls.key")
			spiloEnvVars = appendEnvVars(
				spiloEnvVars,
				v1.EnvVar{Name: "SSL_CERTIFICATE_FILE", Value: certFile},
				v1.EnvVar{Name: "SSL_PRIVATE_KEY_FILE", Value: privateKeyFile},
			)
			return keyName
		}

		tlsEnv, tlsVolumes := c.generateTlsMounts(spec, getSpiloTLSEnv)
		for _, env := range tlsEnv {
			spiloEnvVars = appendEnvVars(spiloEnvVars, env)
		}
		additionalVolumes = append(additionalVolumes, tlsVolumes...)
	}

	repo_host_mode := false
	// Add this envVar so that it is not added to the pgbackrest initcontainer
	if specHasPgbackrestPVCRepo(spec) {
		repo_host_mode = true
		spiloEnvVars = append(spiloEnvVars, v1.EnvVar{
			Name:  "REPO_HOST",
			Value: "true",
		})
	}

	if c.multisiteEnabled() {
		multisiteEnvVars, multisiteVolumes := c.generateMultisiteEnvVars()
		spiloEnvVars = appendEnvVars(spiloEnvVars, multisiteEnvVars...)
		additionalVolumes = append(additionalVolumes, multisiteVolumes...)
	}

	// generate the spilo container
	spiloContainer := generateContainer(constants.PostgresContainerName,
		&effectiveDockerImage,
		resourceRequirements,
		spiloEnvVars,
		volumeMounts,
		c.OpConfig.Resources.SpiloPrivileged,
		c.OpConfig.Resources.SpiloAllowPrivilegeEscalation,
		c.OpConfig.Resources.ReadOnlyRootFilesystem,
		generateCapabilities(c.OpConfig.AdditionalPodCapabilities),
	)

	// Patroni responds 200 to probe only if it either owns the leader lock or postgres is running and DCS is accessible
	if c.OpConfig.EnableReadinessProbe {
		spiloContainer.ReadinessProbe = generateSpiloReadinessProbe()
	}
	//
	if c.OpConfig.EnableLivenessProbe {
		spiloContainer.LivenessProbe = generatePatroniLivenessProbe()
	}

	// generate container specs for sidecars specified in the cluster manifest
	clusterSpecificSidecars := []v1.Container{}
	if spec.Sidecars != nil && len(spec.Sidecars) > 0 {
		// warn if sidecars are defined, but globally disabled (does not apply to globally defined sidecars)
		if c.OpConfig.EnableSidecars != nil && !(*c.OpConfig.EnableSidecars) {
			c.logger.Warningf("sidecars specified but disabled in configuration - next statefulset creation would fail")
		}

		if clusterSpecificSidecars, err = c.generateSidecarContainers(spec.Sidecars, defaultResources, 0); err != nil {
			return nil, fmt.Errorf("could not generate sidecar containers: %v", err)
		}
	}

	// decrapted way of providing global sidecars
	var globalSidecarContainersByDockerImage []v1.Container
	var globalSidecarsByDockerImage []cpov1.Sidecar
	for name, dockerImage := range c.OpConfig.SidecarImages {
		globalSidecarsByDockerImage = append(globalSidecarsByDockerImage, cpov1.Sidecar{Name: name, DockerImage: dockerImage})
	}
	if globalSidecarContainersByDockerImage, err = c.generateSidecarContainers(globalSidecarsByDockerImage, defaultResources, len(clusterSpecificSidecars)); err != nil {
		return nil, fmt.Errorf("could not generate sidecar containers: %v", err)
	}
	// make the resulting list reproducible
	// c.OpConfig.SidecarImages is unsorted by Golang definition
	// .Name is unique
	sort.Slice(globalSidecarContainersByDockerImage, func(i, j int) bool {
		return globalSidecarContainersByDockerImage[i].Name < globalSidecarContainersByDockerImage[j].Name
	})

	// generate scalyr sidecar container
	var scalyrSidecars []v1.Container
	if scalyrSidecar, err :=
		c.generateScalyrSidecarSpec(c.Name,
			c.OpConfig.ScalyrAPIKey,
			c.OpConfig.ScalyrServerURL,
			c.OpConfig.ScalyrImage,
			c.OpConfig.ScalyrCPURequest,
			c.OpConfig.ScalyrMemoryRequest,
			c.OpConfig.ScalyrCPULimit,
			c.OpConfig.ScalyrMemoryLimit,
			defaultResources); err != nil {
		return nil, fmt.Errorf("could not generate Scalyr sidecar: %v", err)
	} else {
		if scalyrSidecar != nil {
			scalyrSidecars = append(scalyrSidecars, *scalyrSidecar)
		}
	}

	sidecarContainers, conflicts := mergeContainers(clusterSpecificSidecars, c.Config.OpConfig.SidecarContainers, globalSidecarContainersByDockerImage, scalyrSidecars)
	for containerName := range conflicts {
		c.logger.Warningf("a sidecar is specified twice. Ignoring sidecar %q in favor of %q with high a precedence",
			containerName, containerName)
	}

	sidecarContainers = patchSidecarContainers(sidecarContainers, volumeMounts, c.OpConfig.SuperUsername, c.credentialSecretName(c.OpConfig.SuperUsername), c.logger, c.OpConfig.Resources.SpiloPrivileged, c.OpConfig.Resources.SpiloAllowPrivilegeEscalation, generateCapabilities(c.OpConfig.AdditionalPodCapabilities))

	tolerationSpec := tolerations(&spec.Tolerations, c.OpConfig.PodToleration)
	topologySpreadConstraintsSpec := topologySpreadConstraints(&spec.TopologySpreadConstraints)
	effectivePodPriorityClassName := util.Coalesce(spec.PodPriorityClassName, c.OpConfig.PodPriorityClassName)

	podAnnotations := c.generatePodAnnotations(spec)

	if spec.GetBackup().Pgbackrest != nil {
		initContainers = append(initContainers, c.generatePgbackrestRestoreContainer(spec, repo_host_mode, volumeMounts, resourceRequirements, c.OpConfig.Resources.SpiloPrivileged, c.OpConfig.Resources.SpiloAllowPrivilegeEscalation, generateCapabilities(c.OpConfig.AdditionalPodCapabilities)))

		additionalVolumes = append(additionalVolumes, c.generatePgbackrestConfigVolume(spec.Backup.Pgbackrest, false))

		if specHasPgbackrestPVCRepo(spec) {
			additionalVolumes = append(additionalVolumes, c.generateCertSecretVolume())
		}
	}

	if specHasPgbackrestClone(spec) {
		additionalVolumes = append(additionalVolumes, c.generatePgbackrestCloneConfigVolumes(spec.Clone)...)
	}

	// generate pod template for the statefulset, based on the spilo container and sidecars
	podTemplate, err = c.generatePodTemplate(
		c.Namespace,
		c.labelsSetWithType(true, TYPE_POSTGRESQL),
		c.annotationsSet(podAnnotations),
		spiloContainer,
		initContainers,
		sidecarContainers,
		c.OpConfig.SharePgSocketWithSidecars,
		&tolerationSpec,
		&topologySpreadConstraintsSpec,
		effectiveRunAsUser,
		effectiveRunAsGroup,
		effectiveFSGroup,
		c.nodeAffinity(c.OpConfig.NodeReadinessLabel, spec.NodeAffinity),
		spec.SchedulerName,
		int64(c.OpConfig.PodTerminateGracePeriod.Seconds()),
		c.OpConfig.PodServiceAccountName,
		c.OpConfig.KubeIAMRole,
		effectivePodPriorityClassName,
		mountShmVolumeNeeded(c.OpConfig, spec),
		c.OpConfig.EnablePodAntiAffinity,
		c.OpConfig.PodAntiAffinityTopologyKey,
		c.OpConfig.PodAntiAffinityPreferredDuringScheduling,
		c.OpConfig.AdditionalSecretMount,
		c.OpConfig.AdditionalSecretMountPath,
		additionalVolumes,
		false)

	if err != nil {
		return nil, fmt.Errorf("could not generate pod template: %v", err)
	}

	if volumeClaimTemplate, err = c.generatePersistentVolumeClaimTemplate(spec.Volume.Size,
		spec.Volume.StorageClass, spec.Volume.Selector, constants.DataVolumeName); err != nil {
		return nil, fmt.Errorf("could not generate volume claim template: %v", err)
	}

	// global minInstances and maxInstances settings can overwrite manifest
	numberOfInstances := c.getNumberOfInstances(spec)

	// the operator has domain-specific logic on how to do rolling updates of PG clusters
	// so we do not use default rolling updates implemented by stateful sets
	// that leaves the legacy "OnDelete" update strategy as the only option
	updateStrategy := appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType}

	var podManagementPolicy appsv1.PodManagementPolicyType
	if c.OpConfig.PodManagementPolicy == "ordered_ready" {
		podManagementPolicy = appsv1.OrderedReadyPodManagement
	} else if c.OpConfig.PodManagementPolicy == "parallel" {
		podManagementPolicy = appsv1.ParallelPodManagement
	} else {
		return nil, fmt.Errorf("could not set the pod management policy to the unknown value: %v", c.OpConfig.PodManagementPolicy)
	}

	var persistentVolumeClaimRetentionPolicy appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy
	if c.OpConfig.PersistentVolumeClaimRetentionPolicy["when_deleted"] == "delete" {
		persistentVolumeClaimRetentionPolicy.WhenDeleted = appsv1.DeletePersistentVolumeClaimRetentionPolicyType
	} else {
		persistentVolumeClaimRetentionPolicy.WhenDeleted = appsv1.RetainPersistentVolumeClaimRetentionPolicyType
	}

	if c.OpConfig.PersistentVolumeClaimRetentionPolicy["when_scaled"] == "delete" {
		persistentVolumeClaimRetentionPolicy.WhenScaled = appsv1.DeletePersistentVolumeClaimRetentionPolicyType
	} else {
		persistentVolumeClaimRetentionPolicy.WhenScaled = appsv1.RetainPersistentVolumeClaimRetentionPolicyType
	}

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.statefulSetName(),
			Namespace:   c.Namespace,
			Labels:      c.labelsSetWithType(true, TYPE_POSTGRESQL),
			Annotations: c.AnnotationsToPropagate(c.annotationsSet(nil)),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:                             &numberOfInstances,
			Selector:                             c.labelsSelector(TYPE_POSTGRESQL),
			ServiceName:                          c.serviceName(Master),
			Template:                             *podTemplate,
			VolumeClaimTemplates:                 []v1.PersistentVolumeClaim{*volumeClaimTemplate},
			UpdateStrategy:                       updateStrategy,
			PodManagementPolicy:                  podManagementPolicy,
			PersistentVolumeClaimRetentionPolicy: &persistentVolumeClaimRetentionPolicy,
		},
	}

	return statefulSet, nil
}

func (c *Cluster) generatePgbackrestRestoreContainer(spec *cpov1.PostgresSpec, repo_host_mode bool, volumeMounts []v1.VolumeMount, resourceRequirements *v1.ResourceRequirements, privilegedMode bool, privilegeEscalationMode *bool, additionalPodCapabilities *v1.Capabilities) v1.Container {
	isOptional := true
	pgbackrestRestoreEnvVars := []v1.EnvVar{
		{
			Name:  "USE_PGBACKREST",
			Value: "true",
		},
		{
			Name:  "MODE",
			Value: "restore",
		},
		{
			Name:  "PGROOT",
			Value: constants.PostgresDataPath,
		},
		{
			Name:  "PGVERSION",
			Value: c.GetDesiredMajorVersion(),
		},
		{
			Name: "POD_INDEX",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: "RESTORE_COMMAND",
			ValueFrom: &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c.getPgbackrestRestoreConfigmapName(),
					},
					Key:      "restore_command",
					Optional: &isOptional,
				},
			},
		},
		{
			Name: "RESTORE_ENABLE",
			ValueFrom: &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c.getPgbackrestRestoreConfigmapName(),
					},
					Key:      "restore_enable",
					Optional: &isOptional,
				},
			},
		},
		{
			Name: "RESTORE_ID",
			ValueFrom: &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c.getPgbackrestRestoreConfigmapName(),
					},
					Key:      "restore_id",
					Optional: &isOptional,
				},
			},
		},
	}
	if repo_host_mode {
		pgbackrestRestoreEnvVars = appendEnvVars(
			pgbackrestRestoreEnvVars, v1.EnvVar{
				Name:  "REPO_HOST",
				Value: "true",
			},
		)
	}
	if spec.TDE != nil && spec.TDE.Enable {
		pgbackrestRestoreEnvVars = append(pgbackrestRestoreEnvVars, v1.EnvVar{Name: "TDE", Value: "true"})
		pgbackrestRestoreEnvVars = append(pgbackrestRestoreEnvVars, v1.EnvVar{Name: "TDE_KEY", ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: c.getTDESecretName(),
				},
				Key: "key",
			},
		},
		})
	}

	return v1.Container{
		Name:         constants.RestoreContainerName,
		Image:        spec.Backup.Pgbackrest.Image,
		Env:          pgbackrestRestoreEnvVars,
		VolumeMounts: volumeMounts,
		Resources:    *resourceRequirements,
		SecurityContext: &v1.SecurityContext{
			AllowPrivilegeEscalation: privilegeEscalationMode,
			Privileged:               &privilegedMode,
			ReadOnlyRootFilesystem:   util.True(),
			Capabilities:             additionalPodCapabilities,
		},
	}
}

func (c *Cluster) generateRepoHostStatefulSet(spec *cpov1.PostgresSpec) (*appsv1.StatefulSet, error) {
	var (
		err               error
		initContainers    []v1.Container
		sidecarContainers []v1.Container
		podTemplate       *v1.PodTemplateSpec
		additionalVolumes []cpov1.AdditionalVolume
	)
	volume := []v1.PersistentVolumeClaim{}

	defaultResources := makeDefaultResources(&c.OpConfig)
	resourceRequirements, err := c.generateResourceRequirements(
		spec.Backup.Pgbackrest.Resources,
		defaultResources, constants.RepoContainerName)
	if err != nil {
		return nil, fmt.Errorf("could not generate resource requirements: %v", err)
	}

	// generate environment variables for the spilo container
	repoEnvVars := c.generatepgBackRestPodEnvVars()

	// determine the User, Group and FSGroup for the spilo pod
	effectiveRunAsUser := c.OpConfig.Resources.SpiloRunAsUser
	effectiveRunAsGroup := c.OpConfig.Resources.SpiloRunAsGroup
	effectiveFSGroup := c.OpConfig.Resources.SpiloFSGroup
	effectiveDockerImage := c.Spec.Backup.Pgbackrest.Image

	repoHostMountPath := ""
	repoHostName := ""
	volumeMounts := []v1.VolumeMount{}
	if c.Spec.Backup.Pgbackrest.Repos != nil {
		for i, repo := range c.Spec.Backup.Pgbackrest.Repos {
			if repo.Storage == "pvc" {
				repoHostMountPath = "/data/pgbackrest/repo" + fmt.Sprintf("%d", i+1)
				repoHostName = "repo" + fmt.Sprintf("%d", i+1)
				volumeMounts = append(volumeMounts, v1.VolumeMount{
					Name:      repoHostName,
					MountPath: repoHostMountPath,
				})
			}

		}
	}

	additionalVolumes = append(additionalVolumes, c.generatePgbackrestConfigVolume(spec.Backup.Pgbackrest, true))
	additionalVolumes = append(additionalVolumes, c.generateCertSecretVolume())

	// generate the spilo container
	repoContainer := generateContainer(constants.RepoContainerName,
		&effectiveDockerImage,
		resourceRequirements,
		repoEnvVars,
		volumeMounts,
		c.OpConfig.Resources.SpiloPrivileged,
		c.OpConfig.Resources.SpiloAllowPrivilegeEscalation,
		c.OpConfig.Resources.ReadOnlyRootFilesystem,
		generateCapabilities(c.OpConfig.AdditionalPodCapabilities),
	)

	// TODO: validate that we want to use the same settings here as the main STS
	tolerationSpec := tolerations(&spec.Tolerations, c.OpConfig.PodToleration)
	topologySpreadConstraintsSpec := topologySpreadConstraints(&spec.TopologySpreadConstraints)
	effectivePodPriorityClassName := util.Coalesce(spec.PodPriorityClassName, c.OpConfig.PodPriorityClassName)

	podAnnotations := c.generatePodAnnotations(spec)
	repoHostLabels := c.labelsSetWithType(true, TYPE_REPOSITORY)

	// generate pod template for the statefulset, based on the spilo container and sidecars
	podTemplate, err = c.generatePodTemplate(
		c.Namespace,
		repoHostLabels,
		c.annotationsSet(podAnnotations),
		repoContainer,
		initContainers,
		sidecarContainers,
		c.OpConfig.SharePgSocketWithSidecars,
		&tolerationSpec,
		&topologySpreadConstraintsSpec,
		effectiveRunAsUser,
		effectiveRunAsGroup,
		effectiveFSGroup,
		c.nodeAffinity(c.OpConfig.NodeReadinessLabel, spec.NodeAffinity),
		spec.SchedulerName,
		int64(c.OpConfig.PodTerminateGracePeriod.Seconds()),
		c.OpConfig.PodServiceAccountName,
		c.OpConfig.KubeIAMRole,
		effectivePodPriorityClassName,
		nil,
		c.OpConfig.EnablePodAntiAffinity,
		c.OpConfig.PodAntiAffinityTopologyKey,
		c.OpConfig.PodAntiAffinityPreferredDuringScheduling,
		c.OpConfig.AdditionalSecretMount,
		c.OpConfig.AdditionalSecretMountPath,
		additionalVolumes,
		true)

	if err != nil {
		return nil, fmt.Errorf("could not generate pod template: %v", err)
	}

	// create pvc for each backrest repo with pvc storage
	if spec.Backup.Pgbackrest.Repos != nil {
		for _, repo := range c.Spec.Backup.Pgbackrest.Repos {
			if repo.Storage == "pvc" {
				if !regexp.MustCompile("^repo[1-4]$").MatchString(repo.Name) {
					return nil, fmt.Errorf("invalid repo name: %s", repo.Name)
				}
				v, err := c.generatePersistentVolumeClaimTemplate(repo.Volume.Size,
					repo.Volume.StorageClass, repo.Volume.Selector, repo.Name)
				if err != nil {
					return nil, fmt.Errorf("could not generate volume claim template: %v", err)
				} else {
					volume = append(volume, *v)
				}
			}
		}
	}

	// For RepoHost only 1 instance is fixed always
	numberOfInstances := int32(1)

	// We let StatefulSet controller handle repo pod updates
	updateStrategy := appsv1.StatefulSetUpdateStrategy{Type: appsv1.RollingUpdateStatefulSetStrategyType}
	podManagementPolicy := appsv1.OrderedReadyPodManagement

	var persistentVolumeClaimRetentionPolicy appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy
	if c.OpConfig.PersistentVolumeClaimRetentionPolicy["when_deleted"] == "delete" {
		persistentVolumeClaimRetentionPolicy.WhenDeleted = appsv1.DeletePersistentVolumeClaimRetentionPolicyType
	} else {
		persistentVolumeClaimRetentionPolicy.WhenDeleted = appsv1.RetainPersistentVolumeClaimRetentionPolicyType
	}

	if c.OpConfig.PersistentVolumeClaimRetentionPolicy["when_scaled"] == "delete" {
		persistentVolumeClaimRetentionPolicy.WhenScaled = appsv1.DeletePersistentVolumeClaimRetentionPolicyType
	} else {
		persistentVolumeClaimRetentionPolicy.WhenScaled = appsv1.RetainPersistentVolumeClaimRetentionPolicyType
	}

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.getPgbackrestRepoHostName(),
			Namespace:   c.Namespace,
			Labels:      repoHostLabels,
			Annotations: c.AnnotationsToPropagate(c.annotationsSet(nil)),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:                             &numberOfInstances,
			Selector:                             c.labelsSelector(TYPE_REPOSITORY),
			ServiceName:                          c.serviceName(ClusterPods),
			Template:                             *podTemplate,
			VolumeClaimTemplates:                 volume,
			UpdateStrategy:                       updateStrategy,
			PodManagementPolicy:                  podManagementPolicy,
			PersistentVolumeClaimRetentionPolicy: &persistentVolumeClaimRetentionPolicy,
		},
	}

	return statefulSet, nil
}

func (c *Cluster) generateTlsMounts(spec *cpov1.PostgresSpec, tlsEnv func(key string) string) ([]v1.EnvVar, []cpov1.AdditionalVolume) {
	// this is combined with the FSGroup in the section above
	// to give read access to the postgres user
	defaultMode := int32(0640)
	mountPath := "/tls"
	env := make([]v1.EnvVar, 0)
	volumes := make([]cpov1.AdditionalVolume, 0)

	volumes = append(volumes, cpov1.AdditionalVolume{
		Name:      spec.TLS.SecretName,
		MountPath: mountPath,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName:  spec.TLS.SecretName,
				DefaultMode: &defaultMode,
			},
		},
	})

	// use the same filenames as Secret resources by default
	certFile := ensurePath(spec.TLS.CertificateFile, mountPath, "tls.crt")
	privateKeyFile := ensurePath(spec.TLS.PrivateKeyFile, mountPath, "tls.key")
	env = append(env, v1.EnvVar{Name: tlsEnv("tls.crt"), Value: certFile})
	env = append(env, v1.EnvVar{Name: tlsEnv("tls.key"), Value: privateKeyFile})

	if spec.TLS.CAFile != "" {
		// support scenario when the ca.crt resides in a different secret, diff path
		mountPathCA := mountPath
		if spec.TLS.CASecretName != "" {
			mountPathCA = mountPath + "ca"
		}

		caFile := ensurePath(spec.TLS.CAFile, mountPathCA, "")
		env = append(env, v1.EnvVar{Name: tlsEnv("tls.ca"), Value: caFile})

		// the ca file from CASecretName secret takes priority
		if spec.TLS.CASecretName != "" {
			volumes = append(volumes, cpov1.AdditionalVolume{
				Name:      spec.TLS.CASecretName,
				MountPath: mountPathCA,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName:  spec.TLS.CASecretName,
						DefaultMode: &defaultMode,
					},
				},
			})
		}
	}

	return env, volumes
}

func (c *Cluster) generatePgbackrestConfigVolume(backrestSpec *cpov1.Pgbackrest, forRepoHost bool) cpov1.AdditionalVolume {
	defaultMode := int32(0640)

	var configMapName string
	if forRepoHost {
		configMapName = c.getPgbackrestRepoHostConfigmapName()
	} else {
		configMapName = c.getPgbackrestConfigmapName()
	}

	projections := []v1.VolumeProjection{
		{ConfigMap: &v1.ConfigMapProjection{
			LocalObjectReference: v1.LocalObjectReference{Name: configMapName},
			Optional:             util.True(),
		},
		},
	}

	if backrestSpec.Configuration.Secret != "" {
		projections = append(projections, v1.VolumeProjection{
			Secret: &v1.SecretProjection{
				LocalObjectReference: v1.LocalObjectReference{Name: backrestSpec.Configuration.Secret},
				Optional:             util.True(),
			},
		})
	}
	return cpov1.AdditionalVolume{
		Name:      "pgbackrest-config",
		MountPath: "/etc/pgbackrest/conf.d",
		TargetContainers: []string{
			constants.PostgresContainerName,
			constants.RepoContainerName,
			constants.RestoreContainerName,
			constants.BackupContainerName,
		},
		VolumeSource: v1.VolumeSource{
			Projected: &v1.ProjectedVolumeSource{
				DefaultMode: &defaultMode,
				Sources:     projections,
			},
		},
	}
}

func (c *Cluster) generatePgbackrestCloneConfigVolumes(description *cpov1.CloneDescription) []cpov1.AdditionalVolume {
	defaultMode := int32(0640)

	projections := []v1.VolumeProjection{{
		ConfigMap: &v1.ConfigMapProjection{
			LocalObjectReference: v1.LocalObjectReference{Name: c.getPgbackrestCloneConfigmapName()},
		},
	}}

	if description.Pgbackrest.Configuration.Secret != "" {
		projections = append(projections, v1.VolumeProjection{
			Secret: &v1.SecretProjection{
				LocalObjectReference: v1.LocalObjectReference{Name: description.Pgbackrest.Configuration.Secret},
			},
		})
	}

	volumes := []cpov1.AdditionalVolume{
		{
			Name:      "pgbackrest-clone",
			MountPath: "/etc/pgbackrest/clone-conf.d",
			VolumeSource: v1.VolumeSource{
				Projected: &v1.ProjectedVolumeSource{
					DefaultMode: &defaultMode,
					Sources:     projections,
				},
			},
		},
	}

	if description.Pgbackrest.Repo.Storage == "pvc" && description.ClusterName != "" {
		// Cloning from another cluster, mount that clusters certs
		volumes = append(volumes, cpov1.AdditionalVolume{
			Name:      "pgbackrest-clone-certs",
			MountPath: "/etc/pgbackrest/clone-certs",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName:  description.ClusterName + "-pgbackrest-cert", // TODO: refactor name generation
					DefaultMode: &defaultMode,
				},
			},
		})
	}

	return volumes
}

func (c *Cluster) generateCertSecretVolume() cpov1.AdditionalVolume {
	defaultMode := int32(0640)

	return cpov1.AdditionalVolume{
		Name:      "cert-secret",
		MountPath: "/etc/pgbackrest/certs",
		TargetContainers: []string{
			constants.PostgresContainerName,
			constants.RepoContainerName,
			constants.RestoreContainerName,
		},
		VolumeSource: v1.VolumeSource{
			Projected: &v1.ProjectedVolumeSource{
				DefaultMode: &defaultMode,
				Sources: []v1.VolumeProjection{
					{Secret: &v1.SecretProjection{
						LocalObjectReference: v1.LocalObjectReference{Name: c.getPgbackrestCertSecretName()},
						Optional:             util.True(),
					},
					},
				},
			},
		},
	}
}

func (c *Cluster) generatePodAnnotations(spec *cpov1.PostgresSpec) map[string]string {
	annotations := make(map[string]string)
	for k, v := range c.OpConfig.CustomPodAnnotations {
		annotations[k] = v
	}
	if spec != nil || spec.PodAnnotations != nil {
		for k, v := range spec.PodAnnotations {
			annotations[k] = v
		}
	}

	if len(annotations) == 0 {
		return nil
	}

	return annotations
}

func (c *Cluster) generateScalyrSidecarSpec(clusterName, APIKey, serverURL, dockerImage string,
	scalyrCPURequest string, scalyrMemoryRequest string, scalyrCPULimit string, scalyrMemoryLimit string,
	defaultResources cpov1.Resources) (*v1.Container, error) {
	if APIKey == "" || dockerImage == "" {
		if APIKey == "" && dockerImage != "" {
			c.logger.Warning("Not running Scalyr sidecar: SCALYR_API_KEY must be defined")
		}
		return nil, nil
	}
	resourcesScalyrSidecar := makeResources(
		scalyrCPURequest,
		scalyrMemoryRequest,
		scalyrCPULimit,
		scalyrMemoryLimit,
	)
	resourceRequirementsScalyrSidecar, err := c.generateResourceRequirements(
		&resourcesScalyrSidecar, defaultResources, scalyrSidecarName)
	if err != nil {
		return nil, fmt.Errorf("invalid resources for Scalyr sidecar: %v", err)
	}
	env := []v1.EnvVar{
		{
			Name:  "SCALYR_API_KEY",
			Value: APIKey,
		},
		{
			Name:  "SCALYR_SERVER_HOST",
			Value: clusterName,
		},
	}
	if serverURL != "" {
		env = append(env, v1.EnvVar{Name: "SCALYR_SERVER_URL", Value: serverURL})
	}
	return &v1.Container{
		Name:            scalyrSidecarName,
		Image:           dockerImage,
		Env:             env,
		ImagePullPolicy: v1.PullIfNotPresent,
		Resources:       *resourceRequirementsScalyrSidecar,
	}, nil
}

func (c *Cluster) getNumberOfInstances(spec *cpov1.PostgresSpec) int32 {
	min := c.OpConfig.MinInstances
	max := c.OpConfig.MaxInstances
	instanceLimitAnnotationKey := c.OpConfig.IgnoreInstanceLimitsAnnotationKey
	cur := spec.NumberOfInstances
	newcur := cur

	if instanceLimitAnnotationKey != "" {
		if value, exists := c.ObjectMeta.Annotations[instanceLimitAnnotationKey]; exists && value == "true" {
			return cur
		}
	}

	if spec.StandbyCluster != nil {
		if newcur == 1 {
			min = newcur
			max = newcur
		} else {
			c.logger.Warningf("operator only supports standby clusters with 1 pod")
		}
	}
	if max >= 0 && newcur > max {
		newcur = max
	}
	if min >= 0 && newcur < min {
		newcur = min
	}
	if newcur != cur {
		c.logger.Infof("adjusted number of instances from %d to %d (min: %d, max: %d)", cur, newcur, min, max)
	}

	return newcur
}

// To avoid issues with limited /dev/shm inside docker environment, when
// PostgreSQL can't allocate enough of dsa segments from it, we can
// mount an extra memory volume
//
// see https://docs.okd.io/latest/dev_guide/shared_memory.html
func addShmVolume(podSpec *v1.PodSpec) {

	postgresContainerIdx := 0

	volumes := append(podSpec.Volumes, v1.Volume{
		Name: constants.ShmVolumeName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				Medium: "Memory",
			},
		},
	})

	for i, container := range podSpec.Containers {
		if container.Name == constants.PostgresContainerName {
			postgresContainerIdx = i
		}
	}

	mounts := append(podSpec.Containers[postgresContainerIdx].VolumeMounts,
		v1.VolumeMount{
			Name:      constants.ShmVolumeName,
			MountPath: constants.ShmVolumePath,
		})

	podSpec.Containers[postgresContainerIdx].VolumeMounts = mounts

	podSpec.Volumes = volumes
}

func addEmptyDirVolume(podSpec *v1.PodSpec, volumeName string, containerName string, path string) {
	vol := v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
	podSpec.Volumes = append(podSpec.Volumes, vol)

	mount := v1.VolumeMount{
		Name:      vol.Name,
		MountPath: path,
	}

	for i := range podSpec.Containers {
		if podSpec.Containers[i].Name == containerName {
			podSpec.Containers[i].VolumeMounts = append(podSpec.Containers[i].VolumeMounts, mount)
		}
	}
	if vol.Name == "postgres-tmp" && len(podSpec.InitContainers) > 0 {
		for i := range podSpec.InitContainers {
			if podSpec.InitContainers[i].Name == "pgbackrest-restore" {
				podSpec.InitContainers[i].VolumeMounts = append(podSpec.InitContainers[i].VolumeMounts, mount)
			}
		}
	}
}

func addRunVolume(podSpec *v1.PodSpec, volumeName string, containerName string, path string) {
	vol := v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
	podSpec.Volumes = append(podSpec.Volumes, vol)

	mount := v1.VolumeMount{
		Name:      vol.Name,
		MountPath: path,
	}

	for i := range podSpec.Containers {
		if podSpec.Containers[i].Name == containerName {
			podSpec.Containers[i].VolumeMounts = append(podSpec.Containers[i].VolumeMounts, mount)
		}
	}
}

func addVarRunVolume(podSpec *v1.PodSpec) {
	volumes := append(podSpec.Volumes, v1.Volume{
		Name: "postgresql-run",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				Medium: "Memory",
			},
		},
	})

	for i := range podSpec.Containers {
		mounts := append(podSpec.Containers[i].VolumeMounts,
			v1.VolumeMount{
				Name:      constants.RunVolumeName,
				MountPath: constants.RunVolumePath,
			})
		podSpec.Containers[i].VolumeMounts = mounts
	}

	podSpec.Volumes = volumes
}

func addSecretVolume(podSpec *v1.PodSpec, additionalSecretMount string, additionalSecretMountPath string) {
	volumes := append(podSpec.Volumes, v1.Volume{
		Name: additionalSecretMount,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: additionalSecretMount,
			},
		},
	})

	for i := range podSpec.Containers {
		mounts := append(podSpec.Containers[i].VolumeMounts,
			v1.VolumeMount{
				Name:      additionalSecretMount,
				MountPath: additionalSecretMountPath,
			})
		podSpec.Containers[i].VolumeMounts = mounts
	}

	podSpec.Volumes = volumes
}

func (c *Cluster) addAdditionalVolumes(podSpec *v1.PodSpec,
	additionalVolumes []cpov1.AdditionalVolume) {

	volumes := podSpec.Volumes
	mountPaths := map[string]cpov1.AdditionalVolume{}
	for i, additionalVolume := range additionalVolumes {
		if previousVolume, exist := mountPaths[additionalVolume.MountPath]; exist {
			msg := "volume %+v cannot be mounted to the same path as %+v"
			c.logger.Warningf(msg, additionalVolume, previousVolume)
			continue
		}

		if additionalVolume.MountPath == constants.PostgresDataMount {
			msg := "cannot mount volume on postgresql data directory, %+v"
			c.logger.Warningf(msg, additionalVolume)
			continue
		}

		// if no target container is defined assign it to postgres container
		if len(additionalVolume.TargetContainers) == 0 {
			postgresContainer := getPostgresContainer(podSpec)
			additionalVolumes[i].TargetContainers = []string{postgresContainer.Name}
		}

		for _, target := range additionalVolume.TargetContainers {
			if target == "all" && len(additionalVolume.TargetContainers) != 1 {
				msg := `target containers could be either "all" or a list
						of containers, mixing those is not allowed, %+v`
				c.logger.Warningf(msg, additionalVolume)
				continue
			}
		}

		volumes = append(volumes,
			v1.Volume{
				Name:         additionalVolume.Name,
				VolumeSource: additionalVolume.VolumeSource,
			},
		)

		mountPaths[additionalVolume.MountPath] = additionalVolume
	}

	c.logger.Infof("Mount additional volumes: %+v", additionalVolumes)

	addMountsToMatchedContainers(podSpec.Containers, additionalVolumes)
	addMountsToMatchedContainers(podSpec.InitContainers, additionalVolumes)

	podSpec.Volumes = volumes
}

func addMountsToMatchedContainers(containers []v1.Container, additionalVolumes []cpov1.AdditionalVolume) {
	for i := range containers {
		mounts := containers[i].VolumeMounts
		for _, additionalVolume := range additionalVolumes {
			for _, target := range additionalVolume.TargetContainers {
				if containers[i].Name == target || target == "all" {
					mounts = append(mounts, v1.VolumeMount{
						Name:      additionalVolume.Name,
						MountPath: additionalVolume.MountPath,
						SubPath:   additionalVolume.SubPath,
					})
				}
			}
		}
		containers[i].VolumeMounts = mounts
	}
}

func (c *Cluster) generatePersistentVolumeClaimTemplate(volumeSize, volumeStorageClass string,
	volumeSelector *metav1.LabelSelector, name string) (*v1.PersistentVolumeClaim, error) {

	var storageClassName *string
	if volumeStorageClass != "" {
		storageClassName = &volumeStorageClass
	}

	quantity, err := resource.ParseQuantity(volumeSize)
	if err != nil {
		return nil, fmt.Errorf("could not parse volume size: %v", err)
	}

	volumeMode := v1.PersistentVolumeFilesystem
	volumeClaim := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: c.annotationsSet(nil),
			Labels:      c.labelsSet(true),
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: quantity,
				},
			},
			StorageClassName: storageClassName,
			VolumeMode:       &volumeMode,
			Selector:         volumeSelector,
		},
	}

	return volumeClaim, nil
}

func (c *Cluster) generateUserSecrets() map[string]*v1.Secret {
	secrets := make(map[string]*v1.Secret, len(c.pgUsers)+len(c.systemUsers))
	namespace := c.Namespace
	for username, pgUser := range c.pgUsers {
		//Skip users with no password i.e. human users (they'll be authenticated using pam)
		secret := c.generateSingleUserSecret(pgUser.Namespace, pgUser)
		if secret != nil {
			secrets[username] = secret
		}
		namespace = pgUser.Namespace
	}
	/* special case for the system user */
	for _, systemUser := range c.systemUsers {
		secret := c.generateSingleUserSecret(namespace, systemUser)
		if secret != nil {
			secrets[systemUser.Name] = secret
		}
	}

	return secrets
}

func (c *Cluster) generateSingleUserSecret(namespace string, pgUser spec.PgUser) *v1.Secret {
	//Skip users with no password i.e. human users (they'll be authenticated using pam)
	if pgUser.Password == "" {
		if pgUser.Origin != spec.RoleOriginTeamsAPI {
			c.logger.Warningf("could not generate secret for a non-teamsAPI role %q: role has no password",
				pgUser.Name)
		}
		return nil
	}

	//skip NOLOGIN users
	for _, flag := range pgUser.Flags {
		if flag == constants.RoleFlagNoLogin {
			return nil
		}
	}

	username := pgUser.Name
	lbls := c.labelsSet(true)

	if username == constants.ConnectionPoolerUserName {
		lbls = c.connectionPoolerLabels("", false).MatchLabels
	}

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.credentialSecretName(username),
			Namespace:   pgUser.Namespace,
			Labels:      lbls,
			Annotations: c.annotationsSet(nil),
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte(pgUser.Name),
			"password": []byte(pgUser.Password),
		},
	}

	return &secret
}

func (c *Cluster) shouldCreateLoadBalancerForService(role PostgresRole, spec *cpov1.PostgresSpec) bool {

	switch role {

	case Replica:

		// if the value is explicitly set in a Postgresql manifest, follow this setting
		if spec.EnableReplicaLoadBalancer != nil {
			return *spec.EnableReplicaLoadBalancer
		}

		// otherwise, follow the operator configuration
		return c.OpConfig.EnableReplicaLoadBalancer

	case Master:
		// If multisite is enabled at manifest or operator configuration we always
		// need a load balancer
		if c.multisiteEnabled() {
			return true
		}

		if spec.EnableMasterLoadBalancer != nil {
			return *spec.EnableMasterLoadBalancer
		}

		return c.OpConfig.EnableMasterLoadBalancer

	case ClusterPods:
		// Ignoring values, service is used internaly only
		return false

	default:
		panic(fmt.Sprintf("Unknown role %v", role))
	}

}

func (c *Cluster) generateService(role PostgresRole, spec *cpov1.PostgresSpec) *v1.Service {
	serviceSpec := v1.ServiceSpec{
		Ports: []v1.ServicePort{{Name: "postgresql", Port: pgPort, TargetPort: intstr.IntOrString{IntVal: pgPort}}},
		Type:  v1.ServiceTypeClusterIP,
	}

	if role == ClusterPods {
		serviceSpec = v1.ServiceSpec{
			ClusterIP:                "None",
			PublishNotReadyAddresses: true,
			Type:                     v1.ServiceTypeClusterIP,
		}
	}

	// no selector for master, see https://github.com/cybertec-postgresql/cybertec-pg-operator/issues/340
	// if kubernetes_use_configmaps is set master service needs a selector
	if role == Replica || c.patroniKubernetesUseConfigMaps() {
		// XXX: this seems broken when etcd_host is set. That makes use config maps false, but we should need a selector
		serviceSpec.Selector = c.roleLabelsSet(false, role)
	}

	if c.shouldCreateLoadBalancerForService(role, spec) {
		c.configureLoadBalanceService(&serviceSpec, spec.AllowedSourceRanges)
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.serviceName(role),
			Namespace:   c.Namespace,
			Labels:      c.roleLabelsSet(true, role),
			Annotations: c.annotationsSet(c.generateServiceAnnotations(role, spec)),
		},
		Spec: serviceSpec,
	}

	return service
}

func (c *Cluster) configureLoadBalanceService(serviceSpec *v1.ServiceSpec, sourceRanges []string) {
	// spec.AllowedSourceRanges evaluates to the empty slice of zero length
	// when omitted or set to 'null'/empty sequence in the PG manifest
	if len(sourceRanges) > 0 {
		serviceSpec.LoadBalancerSourceRanges = sourceRanges
	} else {
		// safe default value: lock a load balancer only to the local address unless overridden explicitly
		serviceSpec.LoadBalancerSourceRanges = []string{localHost}
	}

	c.logger.Debugf("final load balancer source ranges as seen in a service spec (not necessarily applied): %q", serviceSpec.LoadBalancerSourceRanges)
	serviceSpec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyType(c.OpConfig.ExternalTrafficPolicy)
	serviceSpec.Type = v1.ServiceTypeLoadBalancer
}

func (c *Cluster) generateServiceAnnotations(role PostgresRole, spec *cpov1.PostgresSpec) map[string]string {
	annotations := c.getCustomServiceAnnotations(role, spec)

	if c.shouldCreateLoadBalancerForService(role, spec) {
		dnsName := c.dnsName(role)

		// Just set ELB Timeout annotation with default value, if it does not
		// have a custom value
		if _, ok := annotations[constants.ElbTimeoutAnnotationName]; !ok {
			annotations[constants.ElbTimeoutAnnotationName] = constants.ElbTimeoutAnnotationValue
		}
		// External DNS name annotation is not customizable
		annotations[constants.ZalandoDNSNameAnnotation] = dnsName
	}

	if len(annotations) == 0 {
		return nil
	}

	return annotations
}

func (c *Cluster) getCustomServiceAnnotations(role PostgresRole, spec *cpov1.PostgresSpec) map[string]string {
	annotations := make(map[string]string)
	maps.Copy(annotations, c.OpConfig.CustomServiceAnnotations)

	if spec != nil {
		maps.Copy(annotations, spec.ServiceAnnotations)

		switch role {
		case Master:
			maps.Copy(annotations, spec.MasterServiceAnnotations)
		case Replica:
			maps.Copy(annotations, spec.ReplicaServiceAnnotations)
		case ClusterPods:
			maps.Copy(annotations, spec.ClusterPodsServiceAnnotations)
		}
	}

	return annotations
}

func (c *Cluster) generateEndpoint(role PostgresRole, subsets []v1.EndpointSubset) *v1.Endpoints {
	endpoints := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.endpointName(role),
			Namespace: c.Namespace,
			Labels:    c.roleLabelsSet(true, role),
		},
	}
	if len(subsets) > 0 {
		endpoints.Subsets = subsets
	}

	return endpoints
}

func (c *Cluster) generateCloneEnvironment(description *cpov1.CloneDescription) []v1.EnvVar {
	result := make([]v1.EnvVar, 0)

	if description.Pgbackrest != nil {
		result = append(result, v1.EnvVar{Name: "CLONE_METHOD", Value: "CLONE_WITH_PGBACKREST"})
		result = append(result, v1.EnvVar{Name: "CLONE_PGBACKREST_CONFIG", Value: "/etc/pgbackrest/clone-conf.d"})
		if description.EndTimestamp != "" {
			result = append(result, v1.EnvVar{Name: "CLONE_TARGET_TIME", Value: description.EndTimestamp})
		}

		return result
	}

	if description.ClusterName == "" {
		return result
	}

	cluster := description.ClusterName
	result = append(result, v1.EnvVar{Name: "CLONE_SCOPE", Value: cluster})
	if description.EndTimestamp == "" {
		c.logger.Infof("cloning with basebackup from %s", cluster)
		// cloning with basebackup, make a connection string to the cluster to clone from
		host, port := c.getClusterServiceConnectionParameters(cluster)
		// TODO: make some/all of those constants
		result = append(result, v1.EnvVar{Name: "CLONE_METHOD", Value: "CLONE_WITH_BASEBACKUP"})
		result = append(result, v1.EnvVar{Name: "CLONE_HOST", Value: host})
		result = append(result, v1.EnvVar{Name: "CLONE_PORT", Value: port})
		// TODO: assume replication user name is the same for all clusters, fetch it from secrets otherwise
		result = append(result, v1.EnvVar{Name: "CLONE_USER", Value: c.OpConfig.ReplicationUsername})
		result = append(result,
			v1.EnvVar{Name: "CLONE_PASSWORD",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: c.credentialSecretNameForCluster(c.OpConfig.ReplicationUsername,
								description.ClusterName),
						},
						Key: "password",
					},
				},
			})
	} else {
		c.logger.Info("cloning from WAL location")
		if description.S3WalPath == "" {
			c.logger.Info("no S3 WAL path defined - taking value from global config", description.S3WalPath)

			if c.OpConfig.WALES3Bucket != "" {
				c.logger.Debugf("found WALES3Bucket %s - will set CLONE_WAL_S3_BUCKET", c.OpConfig.WALES3Bucket)
				result = append(result, v1.EnvVar{Name: "CLONE_WAL_S3_BUCKET", Value: c.OpConfig.WALES3Bucket})
			} else if c.OpConfig.WALGSBucket != "" {
				c.logger.Debugf("found WALGSBucket %s - will set CLONE_WAL_GS_BUCKET", c.OpConfig.WALGSBucket)
				result = append(result, v1.EnvVar{Name: "CLONE_WAL_GS_BUCKET", Value: c.OpConfig.WALGSBucket})
				if c.OpConfig.GCPCredentials != "" {
					result = append(result, v1.EnvVar{Name: "CLONE_GOOGLE_APPLICATION_CREDENTIALS", Value: c.OpConfig.GCPCredentials})
				}
			} else if c.OpConfig.WALAZStorageAccount != "" {
				c.logger.Debugf("found WALAZStorageAccount %s - will set CLONE_AZURE_STORAGE_ACCOUNT", c.OpConfig.WALAZStorageAccount)
				result = append(result, v1.EnvVar{Name: "CLONE_AZURE_STORAGE_ACCOUNT", Value: c.OpConfig.WALAZStorageAccount})
			} else {
				c.logger.Error("cannot figure out S3 or GS bucket or AZ storage account. All options are empty in the config.")
			}

			// append suffix because WAL location name is not the whole path
			result = append(result, v1.EnvVar{Name: "CLONE_WAL_BUCKET_SCOPE_SUFFIX", Value: getBucketScopeSuffix(description.UID)})
		} else {
			c.logger.Debugf("use S3WalPath %s from the manifest", description.S3WalPath)

			result = append(result, v1.EnvVar{
				Name:  "CLONE_WALE_S3_PREFIX",
				Value: description.S3WalPath,
			})
		}

		result = append(result, v1.EnvVar{Name: "CLONE_METHOD", Value: "CLONE_WITH_WALE"})
		result = append(result, v1.EnvVar{Name: "CLONE_TARGET_TIME", Value: description.EndTimestamp})
		result = append(result, v1.EnvVar{Name: "CLONE_WAL_BUCKET_SCOPE_PREFIX", Value: ""})

		if description.S3Endpoint != "" {
			result = append(result, v1.EnvVar{Name: "CLONE_AWS_ENDPOINT", Value: description.S3Endpoint})
			result = append(result, v1.EnvVar{Name: "CLONE_WALE_S3_ENDPOINT", Value: description.S3Endpoint})
		}

		if description.S3AccessKeyId != "" {
			result = append(result, v1.EnvVar{Name: "CLONE_AWS_ACCESS_KEY_ID", Value: description.S3AccessKeyId})
		}

		if description.S3SecretAccessKey != "" {
			result = append(result, v1.EnvVar{Name: "CLONE_AWS_SECRET_ACCESS_KEY", Value: description.S3SecretAccessKey})
		}

		if description.S3ForcePathStyle != nil {
			s3ForcePathStyle := "0"

			if *description.S3ForcePathStyle {
				s3ForcePathStyle = "1"
			}

			result = append(result, v1.EnvVar{Name: "CLONE_AWS_S3_FORCE_PATH_STYLE", Value: s3ForcePathStyle})
		}
	}

	return result
}

func (c *Cluster) generateStandbyEnvironment(description *cpov1.StandbyDescription) []v1.EnvVar {
	result := make([]v1.EnvVar, 0)

	if description.StandbyHost != "" {
		c.logger.Info("standby cluster streaming from remote primary")
		result = append(result, v1.EnvVar{
			Name:  "STANDBY_HOST",
			Value: description.StandbyHost,
		})
		if description.StandbyPort != "" {
			result = append(result, v1.EnvVar{
				Name:  "STANDBY_PORT",
				Value: description.StandbyPort,
			})
		}
	} else {
		c.logger.Info("standby cluster streaming from WAL location")
		if description.S3WalPath != "" {
			result = append(result, v1.EnvVar{
				Name:  "STANDBY_WALE_S3_PREFIX",
				Value: description.S3WalPath,
			})
		} else if description.GSWalPath != "" {
			result = append(result, v1.EnvVar{
				Name:  "STANDBY_WALE_GS_PREFIX",
				Value: description.GSWalPath,
			})
		} else {
			c.logger.Error("no WAL path specified in standby section")
			return result
		}

		result = append(result, v1.EnvVar{Name: "STANDBY_METHOD", Value: "STANDBY_WITH_WALE"})
		result = append(result, v1.EnvVar{Name: "STANDBY_WAL_BUCKET_SCOPE_PREFIX", Value: ""})
	}

	return result
}

func (c *Cluster) generatePodDisruptionBudget() *policyv1.PodDisruptionBudget {
	minAvailable := intstr.FromInt(1)
	pdbEnabled := c.OpConfig.EnablePodDisruptionBudget

	// if PodDisruptionBudget is disabled or if there are no DB pods, set the budget to 0.
	if (pdbEnabled != nil && !(*pdbEnabled)) || c.Spec.NumberOfInstances <= 0 {
		minAvailable = intstr.FromInt(0)
	}

	return &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.podDisruptionBudgetName(),
			Namespace:   c.Namespace,
			Labels:      c.labelsSet(true),
			Annotations: c.annotationsSet(nil),
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: c.roleLabelsSet(false, Master),
			},
		},
	}
}

// getClusterServiceConnectionParameters fetches cluster host name and port
// TODO: perhaps we need to query the service (i.e. if non-standard port is used?)
// TODO: handle clusters in different namespaces
func (c *Cluster) getClusterServiceConnectionParameters(clusterName string) (host string, port string) {
	host = clusterName
	port = fmt.Sprintf("%d", pgPort)
	return
}

func (c *Cluster) generateLogicalBackupJob() (*batchv1.CronJob, error) {

	var (
		err                  error
		podTemplate          *v1.PodTemplateSpec
		resourceRequirements *v1.ResourceRequirements
	)

	// NB: a cron job creates standard batch jobs according to schedule; these batch jobs manage pods and clean-up

	c.logger.Debug("Generating logical backup pod template")

	// allocate configured resources for logical backup pod
	logicalBackupResources := makeLogicalBackupResources(&c.OpConfig)
	// if not defined only default resources from spilo pods are used
	resourceRequirements, err = c.generateResourceRequirements(
		&logicalBackupResources, makeDefaultResources(&c.OpConfig), logicalBackupContainerName)

	if err != nil {
		return nil, fmt.Errorf("could not generate resource requirements for logical backup pods: %v", err)
	}

	envVars := c.generateLogicalBackupPodEnvVars()
	logicalBackupContainer := generateContainer(
		logicalBackupContainerName,
		&c.OpConfig.LogicalBackup.LogicalBackupDockerImage,
		resourceRequirements,
		envVars,
		[]v1.VolumeMount{},
		c.OpConfig.SpiloPrivileged, // use same value as for normal DB pods
		c.OpConfig.SpiloAllowPrivilegeEscalation,
		util.False(),
		nil,
	)

	nodeAffinity := c.nodeAffinity(c.OpConfig.NodeReadinessLabel, nil)
	podAffinity := podAffinity(
		c.roleLabelsSet(false, Master),
		"kubernetes.io/hostname",
		nodeAffinity,
		true,
		false,
	)

	annotations := c.generatePodAnnotations(&c.Spec)

	// re-use the method that generates DB pod templates
	if podTemplate, err = c.generatePodTemplate(
		c.Namespace,
		c.labelsSetWithType(true, TYPE_LOGICAL_BACKUP),
		annotations,
		logicalBackupContainer,
		[]v1.Container{},
		[]v1.Container{},
		util.False(),
		&[]v1.Toleration{},
		&[]v1.TopologySpreadConstraint{},
		nil,
		nil,
		nil,
		c.nodeAffinity(c.OpConfig.NodeReadinessLabel, nil),
		nil,
		int64(c.OpConfig.PodTerminateGracePeriod.Seconds()),
		c.OpConfig.PodServiceAccountName,
		c.OpConfig.KubeIAMRole,
		"",
		util.False(),
		false,
		"",
		false,
		c.OpConfig.AdditionalSecretMount,
		c.OpConfig.AdditionalSecretMountPath,
		[]cpov1.AdditionalVolume{},
		false); err != nil {
		return nil, fmt.Errorf("could not generate pod template for logical backup pod: %v", err)
	}

	// overwrite specific params of logical backups pods
	podTemplate.Spec.Affinity = podAffinity
	podTemplate.Spec.RestartPolicy = "Never" // affects containers within a pod

	// configure a batch job

	jobSpec := batchv1.JobSpec{
		Template: *podTemplate,
	}

	// configure a cron job

	jobTemplateSpec := batchv1.JobTemplateSpec{
		Spec: jobSpec,
	}

	schedule := c.Postgresql.Spec.LogicalBackupSchedule
	if schedule == "" {
		schedule = c.OpConfig.LogicalBackupSchedule
	}

	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.getLogicalBackupJobName(),
			Namespace:   c.Namespace,
			Labels:      c.labelsSetWithType(true, TYPE_LOGICAL_BACKUP),
			Annotations: c.annotationsSet(nil),
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          schedule,
			JobTemplate:       jobTemplateSpec,
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
		},
	}

	return cronJob, nil
}

func (c *Cluster) generateLogicalBackupPodEnvVars() []v1.EnvVar {

	envVars := []v1.EnvVar{
		{
			Name:  "SCOPE",
			Value: c.Name,
		},
		{
			Name:  "CLUSTER_NAME_LABEL",
			Value: c.OpConfig.ClusterNameLabel,
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		// Bucket env vars
		{
			Name:  "LOGICAL_BACKUP_PROVIDER",
			Value: c.OpConfig.LogicalBackup.LogicalBackupProvider,
		},
		{
			Name:  "LOGICAL_BACKUP_S3_BUCKET",
			Value: c.OpConfig.LogicalBackup.LogicalBackupS3Bucket,
		},
		{
			Name:  "LOGICAL_BACKUP_S3_REGION",
			Value: c.OpConfig.LogicalBackup.LogicalBackupS3Region,
		},
		{
			Name:  "LOGICAL_BACKUP_S3_ENDPOINT",
			Value: c.OpConfig.LogicalBackup.LogicalBackupS3Endpoint,
		},
		{
			Name:  "LOGICAL_BACKUP_S3_SSE",
			Value: c.OpConfig.LogicalBackup.LogicalBackupS3SSE,
		},
		{
			Name:  "LOGICAL_BACKUP_S3_RETENTION_TIME",
			Value: c.OpConfig.LogicalBackup.LogicalBackupS3RetentionTime,
		},
		{
			Name:  "LOGICAL_BACKUP_S3_BUCKET_SCOPE_SUFFIX",
			Value: getBucketScopeSuffix(string(c.Postgresql.GetUID())),
		},
		{
			Name:  "LOGICAL_BACKUP_GOOGLE_APPLICATION_CREDENTIALS",
			Value: c.OpConfig.LogicalBackup.LogicalBackupGoogleApplicationCredentials,
		},
		{
			Name:  "LOGICAL_BACKUP_AZURE_STORAGE_ACCOUNT_NAME",
			Value: c.OpConfig.LogicalBackup.LogicalBackupAzureStorageAccountName,
		},
		{
			Name:  "LOGICAL_BACKUP_AZURE_STORAGE_CONTAINER",
			Value: c.OpConfig.LogicalBackup.LogicalBackupAzureStorageContainer,
		},
		{
			Name:  "LOGICAL_BACKUP_AZURE_STORAGE_ACCOUNT_KEY",
			Value: c.OpConfig.LogicalBackup.LogicalBackupAzureStorageAccountKey,
		},
		// Postgres env vars
		{
			Name:  "PG_VERSION",
			Value: c.Spec.PostgresqlParam.PgVersion,
		},
		{
			Name:  "PGPORT",
			Value: fmt.Sprintf("%d", pgPort),
		},
		{
			Name:  "PGUSER",
			Value: c.OpConfig.SuperUsername,
		},
		{
			Name:  "PGDATABASE",
			Value: c.OpConfig.SuperUsername,
		},
		{
			Name:  "PGSSLMODE",
			Value: "require",
		},
		{
			Name: "PGPASSWORD",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c.credentialSecretName(c.OpConfig.SuperUsername),
					},
					Key: "password",
				},
			},
		},
	}

	if c.OpConfig.LogicalBackup.LogicalBackupS3AccessKeyID != "" {
		envVars = append(envVars, v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: c.OpConfig.LogicalBackup.LogicalBackupS3AccessKeyID})
	}

	if c.OpConfig.LogicalBackup.LogicalBackupS3SecretAccessKey != "" {
		envVars = append(envVars, v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: c.OpConfig.LogicalBackup.LogicalBackupS3SecretAccessKey})
	}

	return envVars
}

// getLogicalBackupJobName returns the name; the job itself may not exists
func (c *Cluster) getLogicalBackupJobName() (jobName string) {
	return trimCronjobName(fmt.Sprintf("%s%s", c.OpConfig.LogicalBackupJobPrefix, c.clusterName().Name))
}

func (c *Cluster) getPgbackrestConfigmapName() (jobName string) {
	return fmt.Sprintf("%s-pgbackrest-config", c.Name)
}

func (c *Cluster) getPgbackrestRepoHostConfigmapName() (jobName string) {
	return fmt.Sprintf("%s-pgbackrest-repohost-config", c.Name)
}

func (c *Cluster) getPgbackrestCloneConfigmapName() (jobName string) {
	return fmt.Sprintf("%s-pgbackrest-clone-config", c.Name)
}

func (c *Cluster) getTDESecretName() string {
	return fmt.Sprintf("%s-tde", c.Name)
}

func (c *Cluster) getMonitoringSecretName() string {
	return c.OpConfig.SecretNameTemplate.Format(
		"username", "cpo-exporter",
		"cluster", c.clusterName().Name,
		"tprkind", cpov1.PostgresCRDResourceKind,
		"tprgroup", cpo.GroupName)
}

func (c *Cluster) generateMonitoringEnvVars() []v1.EnvVar {
	env := []v1.EnvVar{
		{
			Name:  "DATA_SOURCE_URI",
			Value: "localhost:5432/postgres?sslmode=disable",
		},
		{
			Name:  "DATA_SOURCE_USER",
			Value: monitorUsername,
		},
		{
			Name: "DATA_SOURCE_PASS",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c.getMonitoringSecretName(),
					},
					Key: "password",
				},
			},
		},
	}
	return env
}

func (c *Cluster) getPgbackrestRestoreConfigmapName() (jobName string) {
	return fmt.Sprintf("%s-pgbackrest-restore", c.Name)
}

// Return an array of ownerReferences to make an arbitraty object dependent on
// the StatefulSet. Dependency is made on StatefulSet instead of PostgreSQL CRD
// while the former is represent the actual state, and only it's deletion means
// we delete the cluster (e.g. if CRD was deleted, StatefulSet somehow
// survived, we can't delete an object because it will affect the functioning
// cluster).
func (c *Cluster) ownerReferences() []metav1.OwnerReference {
	controller := true

	if c.Statefulset == nil {
		c.logger.Warning("Cannot get owner reference, no statefulset")
		return []metav1.OwnerReference{}
	}

	return []metav1.OwnerReference{
		{
			UID:        c.Statefulset.ObjectMeta.UID,
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
			Name:       c.Statefulset.ObjectMeta.Name,
			Controller: &controller,
		},
	}
}

func ensurePath(file string, defaultDir string, defaultFile string) string {
	if file == "" {
		return path.Join(defaultDir, defaultFile)
	}
	if !path.IsAbs(file) {
		return path.Join(defaultDir, file)
	}
	return file
}

func (c *Cluster) generatePgbackrestConfigmap() (*v1.ConfigMap, error) {
	config := "[db]\npg1-path = /home/postgres/pgdata/pgroot/data\npg1-port = 5432\npg1-socket-path = /var/run/postgresql/\n"
	if c.Postgresql.Spec.TDE != nil && c.Postgresql.Spec.TDE.Enable {
		config += "pg-version-force=" + c.Spec.PgVersion + "\narchive-header-check=n\n"
	}
	config += "\n[global]\nlog-path = /home/postgres/pgdata/pgbackrest/log\nspool-path = /home/postgres/pgdata/pgbackrest/spool-path"

	if c.Postgresql.Spec.Backup != nil && c.Postgresql.Spec.Backup.Pgbackrest != nil {
		if global := c.Postgresql.Spec.Backup.Pgbackrest.Global; global != nil {
			for k, v := range global {
				config += fmt.Sprintf("\n%s = %s", k, v)
			}
		}
		repos := c.Postgresql.Spec.Backup.Pgbackrest.Repos

		if len(repos) >= 1 {
			for i, repo := range repos {
				switch repo.Storage {
				case "pvc":
					c.logger.Debugf("DEBUG_OUTPUT %s %s", c.clusterName().Name, c.Namespace)
					config += "\ntls-server-address=*"
					config += "\ntls-server-ca-file = /etc/pgbackrest/certs/pgbackrest.ca-roots"
					config += "\ntls-server-cert-file = /etc/pgbackrest/certs/pgbackrest-client.crt"
					config += "\ntls-server-key-file = /etc/pgbackrest/certs/pgbackrest-client.key"
					config += "\ntls-server-auth = " + c.clientCommonName() + "=*"
					config += "\nrepo" + fmt.Sprintf("%d", i+1) + "-host = " + c.clusterName().Name + "-pgbackrest-repo-host-0." + c.serviceName(ClusterPods) + "." + c.Namespace + ".svc." + c.OpConfig.ClusterDomain
					config += "\nrepo" + fmt.Sprintf("%d", i+1) + "-host-ca-file = /etc/pgbackrest/certs/pgbackrest.ca-roots"
					config += "\nrepo" + fmt.Sprintf("%d", i+1) + "-host-cert-file = /etc/pgbackrest/certs/pgbackrest-client.crt"
					config += "\nrepo" + fmt.Sprintf("%d", i+1) + "-host-key-file = /etc/pgbackrest/certs/pgbackrest-client.key"
					config += "\nrepo" + fmt.Sprintf("%d", i+1) + "-host-type = tls"
					config += "\nrepo" + fmt.Sprintf("%d", i+1) + "-host-user = postgres"

				case "s3":
					config += fmt.Sprintf("\n%s-%s-bucket = %s", repo.Name, repo.Storage, repo.Resource)
					config += fmt.Sprintf("\n%s-%s-endpoint = %s", repo.Name, repo.Storage, repo.Endpoint)
					config += fmt.Sprintf("\n%s-%s-region = %s", repo.Name, repo.Storage, repo.Region)
					config += fmt.Sprintf("\n%s-type = %s", repo.Name, repo.Storage)

				case "gcs":
					config += fmt.Sprintf("\n%s-%s-bucket = %s", repo.Name, repo.Storage, repo.Resource)
					config += fmt.Sprintf("\n%s-%s-key = /etc/pgbackrest/conf.d/%s", repo.Name, repo.Storage, repo.Key)
					config += fmt.Sprintf("\n%s-%s-key-type = %s", repo.Name, repo.Storage, repo.KeyType)
					config += fmt.Sprintf("\n%s-type = %s", repo.Name, repo.Storage)

				case "azure":
					config += fmt.Sprintf("\n%s-%s-container = %s", repo.Name, repo.Storage, repo.Resource)
					config += fmt.Sprintf("\n%s-%s-endpoint = %s", repo.Name, repo.Storage, repo.Endpoint)
					config += fmt.Sprintf("\n%s-%s-key = %s", repo.Name, repo.Storage, repo.Key)
					config += fmt.Sprintf("\n%s-%s-account = %s", repo.Name, repo.Storage, repo.Account)

					config += fmt.Sprintf("\n%s-type = %s", repo.Name, repo.Storage)
				default:
				}

			}
		}
	}

	data := map[string]string{"pgbackrest_instance.conf": config}
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      c.getPgbackrestConfigmapName(),
		},
		Data: data,
	}
	return configmap, nil
}

func (c *Cluster) generatePgbackrestRepoHostConfigmap() (*v1.ConfigMap, error) {
	// choose repo1 for log, because this will for sure exist, repo2 or an other not
	config := "[global]\nlog-path = /data/pgbackrest/repo1/log"
	config += "\ntls-server-address=*"
	config += "\ntls-server-ca-file = /etc/pgbackrest/certs/pgbackrest.ca-roots"
	config += "\ntls-server-cert-file = /etc/pgbackrest/certs/pgbackrest-repo-host.crt"
	config += "\ntls-server-key-file = /etc/pgbackrest/certs/pgbackrest-repo-host.key"
	config += "\ntls-server-auth = " + c.clientCommonName() + "=*"

	repos := c.Postgresql.Spec.Backup.Pgbackrest.Repos
	if len(repos) >= 1 {
		for i, repo := range repos {
			if repo.Storage == "pvc" {
				config += "\nrepo" + fmt.Sprintf("%d", i+1) + "-path = /data/pgbackrest/repo" + fmt.Sprintf("%d", i+1)
			}
		}
		config += "\n[db]"
		if c.Postgresql.Spec.Backup != nil && c.Postgresql.Spec.Backup.Pgbackrest != nil {
			if global := c.Postgresql.Spec.Backup.Pgbackrest.Global; global != nil {
				for k, v := range global {
					config += fmt.Sprintf("\n%s = %s", k, v)
				}
			}
			n := c.Postgresql.Spec.NumberOfInstances
			for j := int32(0); j < n; j++ {
				config += "\npg" + fmt.Sprintf("%d", j+1) + "-host = " + c.clusterName().Name + "-" + fmt.Sprintf("%d", j) + "." + c.clusterName().Name + "." + c.Namespace + ".svc." + c.OpConfig.ClusterDomain
				config += "\npg" + fmt.Sprintf("%d", j+1) + "-host-ca-file = /etc/pgbackrest/certs/pgbackrest.ca-roots"
				config += "\npg" + fmt.Sprintf("%d", j+1) + "-host-cert-file = /etc/pgbackrest/certs/pgbackrest-repo-host.crt"
				config += "\npg" + fmt.Sprintf("%d", j+1) + "-host-key-file = /etc/pgbackrest/certs/pgbackrest-repo-host.key"
				config += "\npg" + fmt.Sprintf("%d", j+1) + "-host-type = tls"
				config += "\npg" + fmt.Sprintf("%d", j+1) + "-path = /home/postgres/pgdata/pgroot/data"
			}
			if c.Postgresql.Spec.TDE != nil && c.Postgresql.Spec.TDE.Enable {
				config += "\npg-version-force=" + c.Spec.PgVersion + "\narchive-header-check=n\n"
			}
		}
	}

	data := map[string]string{"pgbackrest_instance.conf": config}
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      c.getPgbackrestRepoHostConfigmapName(),
		},
		Data: data,
	}
	return configmap, nil
}

func (c *Cluster) generatePgbackrestCloneConfigmap(clone *cpov1.CloneDescription) (*v1.ConfigMap, error) {
	config := map[string]map[string]string{
		"db": {
			"pg1-path":        "/home/postgres/pgdata/pgroot/data",
			"pg1-port":        "5432",
			"pg1-socket-path": "/var/run/postgresql/",
		},
		"global": {
			"log-path":   "/home/postgres/pgdata/pgbackrest/log",
			"spool-path": "/home/postgres/pgdata/pgbackrest/spool-path",
		},
	}

	if clone.Pgbackrest.Options != nil {
		maps.Copy(config["global"], clone.Pgbackrest.Options)
	}

	repo := clone.Pgbackrest.Repo
	repoName := "repo1"
	repoConf := func(conf map[string]string) {
		for k, v := range conf {
			config["global"][repoName+"-"+k] = v
		}
	}

	switch repo.Storage {
	case "pvc":
		// TODO: enable Cluster.serviceName to ask for other clusters services
		serviceName := fmt.Sprintf("%s-%s", clone.ClusterName, "clusterpods")
		// TODO: allow for cross namespace cloning
		repoConf(map[string]string{
			"host":           clone.ClusterName + "-pgbackrest-repo-host-0." + serviceName + "." + c.Namespace + ".svc." + c.OpConfig.ClusterDomain,
			"host-ca-file":   "/etc/pgbackrest/clone-certs/pgbackrest.ca-roots",
			"host-cert-file": "/etc/pgbackrest/clone-certs/pgbackrest-client.crt",
			"host-key-file":  "/etc/pgbackrest/clone-certs/pgbackrest-client.key",
			"host-type":      "tls",
			"host-user":      "postgres",
		})
	case "s3":
		repoConf(map[string]string{
			"type":        "s3",
			"s3-bucket":   repo.Resource,
			"s3-endpoint": repo.Endpoint,
			"s3-region":   repo.Region,
		})
	case "gcs":
		repoConf(map[string]string{
			"type":         "gcs",
			"gcs-bucket":   repo.Resource,
			"gcs-key":      fmt.Sprintf("/etc/pgbackrest/conf.d/%s", repo.Key),
			"gcs-key-type": repo.KeyType,
		})
	case "azure":
		repoConf(map[string]string{
			"type":            "azure",
			"azure-container": repo.Resource,
			"azure-endpoint":  repo.Endpoint,
			"azure-key":       repo.Key,
			"azure-account":   repo.Account,
		})
	default:
		return nil, fmt.Errorf("Invalid repository storage %s", repo.Storage)
	}

	confStr, err := renderPgbackrestConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Error rendering pgbackrest config: %v", err)
	}
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      c.getPgbackrestCloneConfigmapName(),
		},
		Data: map[string]string{
			"pgbackrest_clone.conf": confStr,
		},
	}
	return configmap, nil
}

func renderPgbackrestConfig(config map[string]map[string]string) (string, error) {
	var out bytes.Buffer
	tpl := template.Must(template.New("pgbackrest_instance.conf").Parse(`
{{- range $section, $config := . }}
[{{ $section }}]

{{- range $key, $value := . }}
{{ $key }} = {{ $value }}
{{ end -}}
{{ end -}}
`))
	if err := tpl.Execute(&out, config); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (c *Cluster) generatePgbackrestJob(backup *cpov1.Pgbackrest, repo *cpov1.Repo, backupType string, schedule string) (*batchv1.CronJob, error) {

	var (
		err                  error
		podTemplate          *v1.PodTemplateSpec
		resourceRequirements *v1.ResourceRequirements
	)

	// NB: a cron job creates standard batch jobs according to schedule; these batch jobs manage pods and clean-up

	c.logger.Debug("Generating pgbackrest pod template")

	// Using empty resources
	emptyResourceRequirements := v1.ResourceRequirements{}
	resourceRequirements = &emptyResourceRequirements

	envVars := c.generatePgbackrestBackupJobEnvVars(repo, backupType)
	pgbackrestContainer := generateContainer(
		constants.BackupContainerName,
		&c.Postgresql.Spec.Backup.Pgbackrest.Image,
		resourceRequirements,
		envVars,
		[]v1.VolumeMount{},
		c.OpConfig.SpiloPrivileged, // use same value as for normal DB pods
		c.OpConfig.SpiloAllowPrivilegeEscalation,
		c.OpConfig.Resources.ReadOnlyRootFilesystem,
		nil,
	)

	// Patch securityContext - readOnlyRootFilesystem
	pgbackrestContainer.SecurityContext.ReadOnlyRootFilesystem = util.True()

	podAffinityTerm := v1.PodAffinityTerm{
		LabelSelector: c.roleLabelsSelector(Master),
		TopologyKey:   "kubernetes.io/hostname",
	}
	podAffinity := v1.Affinity{
		PodAffinity: &v1.PodAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{{
				Weight:          1,
				PodAffinityTerm: podAffinityTerm,
			},
			},
		}}

	annotations := c.generatePodAnnotations(&c.Spec)

	// re-use the method that generates DB pod templates
	if podTemplate, err = c.generatePodTemplate(
		c.Namespace,
		c.labelsSetWithType(true, TYPE_BACKUP_JOB),
		annotations,
		pgbackrestContainer,
		[]v1.Container{},
		[]v1.Container{},
		util.False(),
		&[]v1.Toleration{},
		&[]v1.TopologySpreadConstraint{},
		nil,
		nil,
		nil,
		c.nodeAffinity(c.OpConfig.NodeReadinessLabel, nil),
		nil,
		int64(c.OpConfig.PodTerminateGracePeriod.Seconds()),
		c.OpConfig.PodServiceAccountName,
		c.OpConfig.KubeIAMRole,
		"",
		util.False(),
		false,
		"",
		false,
		c.OpConfig.AdditionalSecretMount,
		c.OpConfig.AdditionalSecretMountPath,
		[]cpov1.AdditionalVolume{c.generatePgbackrestConfigVolume(backup, false)},
		false); err != nil {
		return nil, fmt.Errorf("could not generate pod template for logical backup pod: %v", err)
	}

	// overwrite specific params of logical backups pods
	podTemplate.Spec.Affinity = &podAffinity
	podTemplate.Spec.RestartPolicy = "Never" // affects containers within a pod

	// configure a batch job

	jobSpec := batchv1.JobSpec{
		Template: *podTemplate,
	}

	// configure a cron job

	jobTemplateSpec := batchv1.JobTemplateSpec{
		Spec: jobSpec,
	}

	if schedule == "" {
		schedule = c.OpConfig.LogicalBackupSchedule
	}

	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.getPgbackrestJobName(repo.Name, backupType),
			Namespace:   c.Namespace,
			Labels:      c.labelsSetWithType(true, TYPE_BACKUP_JOB),
			Annotations: c.annotationsSet(nil),
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          schedule,
			JobTemplate:       jobTemplateSpec,
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
		},
	}

	return cronJob, nil
}

func (c *Cluster) generatePgbackrestBackupJobEnvVars(repo *cpov1.Repo, backupType string) []v1.EnvVar {
	selector := c.roleLabelsSet(false, Master).String()
	targetContainer := constants.PostgresContainerName
	if repo.Storage == "pvc" {
		// With a PVC based repo the backup command needs to run on the repository system
		// due to pgbackrest limitations
		selector = c.labelsSetWithType(false, TYPE_REPOSITORY).String()
		targetContainer = constants.RepoContainerName
	}

	envVars := []v1.EnvVar{
		{
			Name:  "USE_PGBACKREST",
			Value: "true",
		},
		{
			Name:  "MODE",
			Value: "backup",
		},
		{
			Name:  "COMMAND_OPTS",
			Value: fmt.Sprintf("--stanza=db --repo=%v --type=%s", repoNumberFromName(repo.Name), backupType),
		},
		{
			Name:  "CONTAINER",
			Value: targetContainer,
		},
		{
			Name:  "SELECTOR",
			Value: selector,
		},
	}
	return envVars
}

func repoNumberFromName(repoName string) int {
	repoNumber, err := strconv.Atoi(strings.TrimPrefix(repoName, "repo"))
	if err != nil {
		// CRD should be defining repo name to be ^repo[1-4]$
		panic("unexpected repo name " + repoName)
	}
	return repoNumber
}

// getLogicalBackupJobName returns the name; the job itself may not exists
func (c *Cluster) getPgbackrestJobName(repoName string, backupType string) (jobName string) {
	return trimCronjobName(fmt.Sprintf("%s-%s-%s-%s", "pgbackrest", c.clusterName().Name, repoName, backupType))
}

func (c *Cluster) generateMultisiteEnvVars() ([]v1.EnvVar, []cpov1.AdditionalVolume) {
	site, err := c.getPrimaryLoadBalancerIp()
	if err != nil {
		c.logger.Errorf("Error getting primary load balancer IP for %s: %s", c.Name, err)
		site = ""
	}
	clsConf := c.Spec.Multisite
	if clsConf == nil {
		clsConf = new(cpov1.Multisite)
	}

	envVars := []v1.EnvVar{
		{Name: "MULTISITE_SITE", Value: util.CoalesceStrPtr(clsConf.Site, c.OpConfig.Multisite.Site)},
		{Name: "MULTISITE_ETCD_HOSTS", Value: util.CoalesceStrPtr(clsConf.Etcd.Hosts, c.OpConfig.Multisite.Etcd.Hosts)},
		{Name: "MULTISITE_ETCD_USER", Value: util.CoalesceStrPtr(clsConf.Etcd.User, c.OpConfig.Multisite.Etcd.User)},
		{Name: "MULTISITE_ETCD_PASSWORD", Value: util.CoalesceStrPtr(clsConf.Etcd.Password, c.OpConfig.Multisite.Etcd.Password)},
		{Name: "MULTISITE_ETCD_PROTOCOL", Value: util.CoalesceStrPtr(clsConf.Etcd.Protocol, c.OpConfig.Multisite.Etcd.Protocol)},
		{Name: "MULTISITE_TTL", Value: strconv.Itoa(int(*util.CoalesceInt32(clsConf.TTL, c.OpConfig.Multisite.TTL)))},
		{Name: "MULTISITE_RETRY_TIMEOUT", Value: strconv.Itoa(int(*util.CoalesceInt32(clsConf.RetryTimeout, c.OpConfig.Multisite.RetryTimeout)))},
		{Name: "EXTERNAL_HOST", Value: site},
		{Name: "UPDATE_CRD", Value: c.Namespace + "." + c.Name},
		{Name: "CRD_UID", Value: string(c.UID)},
	}

	certSecretName := util.CoalesceStrPtr(clsConf.Etcd.CertSecretName, c.OpConfig.Multisite.Etcd.CertSecretName)
	volumes := make([]cpov1.AdditionalVolume, 0, 1)
	if certSecretName != "" {

		etcdCertMountPath := "/etcd-tls"
		envVars = append(envVars, v1.EnvVar{Name: "MULTISITE_ETCD_CA_CERT", Value: etcdCertMountPath + "/ca.crt"})
		envVars = append(envVars, v1.EnvVar{Name: "MULTISITE_ETCD_CERT", Value: etcdCertMountPath + "/tls.crt"})
		envVars = append(envVars, v1.EnvVar{Name: "MULTISITE_ETCD_KEY", Value: etcdCertMountPath + "/tls.key"})
		defaultMode := int32(0640)
		volumes = append(volumes, cpov1.AdditionalVolume{
			Name:      "multisite-etcd-certs",
			MountPath: etcdCertMountPath,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName:  certSecretName,
					DefaultMode: &defaultMode,
				},
			}})
	}

	return envVars, volumes
}

func (c *Cluster) getTlsConfigFromCertSecret(certSecretName string) (*tls.Config, error) {
	secret := &v1.Secret{}
	var notFoundErr error
	err := retryutil.Retry(c.OpConfig.ResourceCheckInterval, c.OpConfig.ResourceCheckTimeout,
		func() (bool, error) {
			var err error
			secret, err = c.KubeClient.Secrets(c.Namespace).Get(
				context.TODO(),
				certSecretName,
				metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					notFoundErr = err
					return false, nil
				}
				return false, err
			}
			return true, nil
		},
	)
	if notFoundErr != nil && err != nil {
		err = errors.Wrap(notFoundErr, err.Error())
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not get secret for TLS configuration")
	}

	var caPool *x509.CertPool
	if caCert, ok := secret.Data["ca.crt"]; ok {
		caPool = x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA cert from secret %s", certSecretName)
		}
	}
	clientCert, certPresent := secret.Data["tls.crt"]
	clientKey, keyPresent := secret.Data["tls.key"]
	var certificates []tls.Certificate
	if certPresent && keyPresent {
		cert, err := tls.X509KeyPair(clientCert, clientKey)
		if err != nil {
			c.logger.Warningf("failed to load TLS client certificate from secret %s: %s", certSecretName, err)
		} else {
			certificates = append(certificates, cert)
		}
	}

	tlsConfig := &tls.Config{
		RootCAs:      caPool,
		Certificates: certificates,
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}
