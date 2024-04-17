package cluster

import (
	"time"

	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/types"
)

// PostgresRole describes role of the node
type PostgresRole string

const (
	// spilo roles
	Master  PostgresRole = "master"
	Replica PostgresRole = "replica"

	// roles returned by Patroni cluster endpoint
	Leader        PostgresRole = "leader"
	StandbyLeader PostgresRole = "standby_leader"
	SyncStandby   PostgresRole = "sync_standby"

	// clusterrole for service
	ClusterPods PostgresRole = "clusterpods"

)

// PodEventType represents the type of a pod-related event
type PodEventType string

// Possible values for the EventType
const (
	PodEventAdd    PodEventType = "ADD"
	PodEventUpdate PodEventType = "UPDATE"
	PodEventDelete PodEventType = "DELETE"
)

// PodEvent describes the event for a single Pod
type PodEvent struct {
	ResourceVersion string
	PodName         types.NamespacedName
	PrevPod         *v1.Pod
	CurPod          *v1.Pod
	EventType       PodEventType
}

// Process describes process of the cluster
type Process struct {
	Name      string
	StartTime time.Time
}

// WorkerStatus describes status of the worker
type WorkerStatus struct {
	CurrentCluster types.NamespacedName
	CurrentProcess Process
}

// ClusterStatus describes status of the cluster
type ClusterStatus struct {
	Team                string
	Cluster             string
	Namespace           string
	MasterService       *v1.Service
	ReplicaService      *v1.Service
	ClusterPodsService	*v1.Service
	MasterEndpoint      *v1.Endpoints
	ReplicaEndpoint     *v1.Endpoints
	StatefulSet         *appsv1.StatefulSet
	PodDisruptionBudget *policyv1.PodDisruptionBudget

	CurrentProcess Process
	Worker         uint32
	Status         cpov1.PostgresStatus
	Spec           cpov1.PostgresSpec
	Error          error
}

type TemplateParams map[string]interface{}

type InstallFunction func(schema string, user string) error

type SyncReason []string

// no sync happened, empty value
var NoSync SyncReason = []string{}
