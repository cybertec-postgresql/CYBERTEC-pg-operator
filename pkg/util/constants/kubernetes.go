package constants

import "time"

// General kubernetes-related constants
const (
	PostgresContainerName = "postgres"
	RepoContainerName     = "pgbackrest"
	BackupContainerName   = "pgbackrest-backup"
	RestoreContainerName  = "pgbackrest-restore"
	K8sAPIPath            = "/apis"

	QueueResyncPeriodPod  = 5 * time.Minute
	QueueResyncPeriodTPR  = 5 * time.Minute
	QueueResyncPeriodNode = 5 * time.Minute
)
