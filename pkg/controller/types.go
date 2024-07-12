package controller

import (
	"time"

	"k8s.io/apimachinery/pkg/types"

	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
)

// EventType contains type of the events for the TPRs and Pods received from Kubernetes
type EventType string

// Possible values for the EventType
const (
	EventAdd    EventType = "ADD"
	EventUpdate EventType = "UPDATE"
	EventDelete EventType = "DELETE"
	EventSync   EventType = "SYNC"
	EventRepair EventType = "REPAIR"
)

// ClusterEvent carries the payload of the Cluster TPR events.
type ClusterEvent struct {
	EventTime time.Time
	UID       types.UID
	EventType EventType
	OldSpec   *cpov1.Postgresql
	NewSpec   *cpov1.Postgresql
	WorkerID  uint32
}
