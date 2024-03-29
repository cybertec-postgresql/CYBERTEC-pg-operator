package controller

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/spec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	True  = true
	False = false
)

func newPostgresqlTestController() *Controller {
	controller := NewController(&spec.ControllerConfig{}, "postgresql-test")
	return controller
}

var postgresqlTestController = newPostgresqlTestController()

func TestControllerOwnershipOnPostgresql(t *testing.T) {
	tests := []struct {
		name  string
		pg    *cpov1.Postgresql
		owned bool
		error string
	}{
		{
			"Postgres cluster with defined ownership of mocked controller",
			&cpov1.Postgresql{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"cpo.opensource.cybertec.at/controller": "postgresql-test"},
				},
			},
			True,
			"Postgres cluster should be owned by operator, but controller says no",
		},
		{
			"Postgres cluster with defined ownership of another controller",
			&cpov1.Postgresql{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"cpo.opensource.cybertec.at/controller": "stups-test"},
				},
			},
			False,
			"Postgres cluster should be owned by another operator, but controller say yes",
		},
		{
			"Test Postgres cluster without defined ownership",
			&cpov1.Postgresql{},
			False,
			"Postgres cluster should be owned by operator with empty controller ID, but controller says yes",
		},
	}
	for _, tt := range tests {
		if postgresqlTestController.hasOwnership(tt.pg) != tt.owned {
			t.Errorf("%s: %v", tt.name, tt.error)
		}
	}
}

func TestMergeDeprecatedPostgreSQLSpecParameters(t *testing.T) {
	tests := []struct {
		name  string
		in    *cpov1.PostgresSpec
		out   *cpov1.PostgresSpec
		error string
	}{
		{
			"Check that old parameters propagate values to the new ones",
			&cpov1.PostgresSpec{UseLoadBalancer: &True, ReplicaLoadBalancer: &True},
			&cpov1.PostgresSpec{UseLoadBalancer: nil, ReplicaLoadBalancer: nil,
				EnableMasterLoadBalancer: &True, EnableReplicaLoadBalancer: &True},
			"New parameters should be set from the values of old ones",
		},
		{
			"Check that new parameters are not set when both old and new ones are present",
			&cpov1.PostgresSpec{UseLoadBalancer: &True, EnableMasterLoadBalancer: &False},
			&cpov1.PostgresSpec{UseLoadBalancer: nil, EnableMasterLoadBalancer: &False},
			"New parameters should remain unchanged when both old and new are present",
		},
	}
	for _, tt := range tests {
		result := postgresqlTestController.mergeDeprecatedPostgreSQLSpecParameters(tt.in)
		if !reflect.DeepEqual(result, tt.out) {
			t.Errorf("%s: %v", tt.name, tt.error)
		}
	}
}

func TestMeetsClusterDeleteAnnotations(t *testing.T) {
	// set delete annotations in configuration
	postgresqlTestController.opConfig.DeleteAnnotationDateKey = "delete-date"
	postgresqlTestController.opConfig.DeleteAnnotationNameKey = "delete-clustername"

	currentTime := time.Now()
	today := currentTime.Format("2006-01-02") // go's reference date
	clusterName := "acid-test-cluster"

	tests := []struct {
		name  string
		pg    *cpov1.Postgresql
		error string
	}{
		{
			"Postgres cluster with matching delete annotations",
			&cpov1.Postgresql{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterName,
					Annotations: map[string]string{
						"delete-date":        today,
						"delete-clustername": clusterName,
					},
				},
			},
			"",
		},
		{
			"Postgres cluster with violated delete date annotation",
			&cpov1.Postgresql{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterName,
					Annotations: map[string]string{
						"delete-date":        "2020-02-02",
						"delete-clustername": clusterName,
					},
				},
			},
			fmt.Sprintf("annotation delete-date not matching the current date: got 2020-02-02, expected %s", today),
		},
		{
			"Postgres cluster with violated delete cluster name annotation",
			&cpov1.Postgresql{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterName,
					Annotations: map[string]string{
						"delete-date":        today,
						"delete-clustername": "acid-minimal-cluster",
					},
				},
			},
			fmt.Sprintf("annotation delete-clustername not matching the cluster name: got acid-minimal-cluster, expected %s", clusterName),
		},
		{
			"Postgres cluster with missing delete annotations",
			&cpov1.Postgresql{
				ObjectMeta: metav1.ObjectMeta{
					Name:        clusterName,
					Annotations: map[string]string{},
				},
			},
			"annotation delete-date not set in manifest to allow cluster deletion",
		},
		{
			"Postgres cluster with missing delete cluster name annotation",
			&cpov1.Postgresql{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterName,
					Annotations: map[string]string{
						"delete-date": today,
					},
				},
			},
			"annotation delete-clustername not set in manifest to allow cluster deletion",
		},
	}
	for _, tt := range tests {
		if err := postgresqlTestController.meetsClusterDeleteAnnotations(tt.pg); err != nil {
			if !reflect.DeepEqual(err.Error(), tt.error) {
				t.Errorf("Expected error %q, got: %v", tt.error, err)
			}
		}
	}
}
