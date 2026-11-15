package cluster

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/config"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/k8sutil"
	v1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

func TestSupplementalGroupsPolicyGeneration(t *testing.T) {
	tests := []struct {
		name    string
		policy  *v1.SupplementalGroupsPolicy
		want    *v1.SupplementalGroupsPolicy
		invalid bool
	}{
		{name: "omitted"},
		{name: "merge", policy: supplementalGroupsPolicy(v1.SupplementalGroupsPolicyMerge), want: supplementalGroupsPolicy(v1.SupplementalGroupsPolicyMerge)},
		{name: "strict", policy: supplementalGroupsPolicy(v1.SupplementalGroupsPolicyStrict), want: supplementalGroupsPolicy(v1.SupplementalGroupsPolicyStrict)},
		{name: "invalid", policy: supplementalGroupsPolicy(v1.SupplementalGroupsPolicy("Unknown")), invalid: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cluster := supplementalGroupsTestCluster()
			spec := supplementalGroupsTestSpec(test.policy)

			first, err := cluster.generateStatefulSet(&spec)
			if (err != nil) != test.invalid {
				t.Fatalf("unexpected error: %v", err)
			}
			if err != nil {
				return
			}

			got := first.Spec.Template.Spec.SecurityContext.SupplementalGroupsPolicy
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("unexpected supplemental groups policy: got %v, want %v", got, test.want)
			}

			second, err := cluster.generateStatefulSet(&spec)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(first, second) {
				t.Fatal("identical input generated a different StatefulSet")
			}
		})
	}
}

func TestSupplementalGroupsPolicyAPIRoundTripAndDeepCopy(t *testing.T) {
	policy := v1.SupplementalGroupsPolicyStrict
	original := cpov1.Postgresql{
		Spec: cpov1.PostgresSpec{SpiloSupplementalGroupsPolicy: &policy},
	}

	raw, err := yaml.Marshal(&original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded cpov1.Postgresql
	if err := yaml.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Spec.SpiloSupplementalGroupsPolicy == nil || *decoded.Spec.SpiloSupplementalGroupsPolicy != policy {
		t.Fatalf("policy was lost during API round trip: %#v", decoded.Spec.SpiloSupplementalGroupsPolicy)
	}

	copy := original.DeepCopy()
	if copy.Spec.SpiloSupplementalGroupsPolicy == original.Spec.SpiloSupplementalGroupsPolicy {
		t.Fatal("deep copy retained the original policy pointer")
	}
	*copy.Spec.SpiloSupplementalGroupsPolicy = v1.SupplementalGroupsPolicyMerge
	if *original.Spec.SpiloSupplementalGroupsPolicy != v1.SupplementalGroupsPolicyStrict {
		t.Fatal("changing the copy changed the original policy")
	}
}

func TestSupplementalGroupsPolicyRepoHostGeneration(t *testing.T) {
	policy := v1.SupplementalGroupsPolicyStrict
	spec := supplementalGroupsTestSpec(&policy)
	spec.Backup = &cpov1.Backup{
		Pgbackrest: &cpov1.Pgbackrest{
			Image: "pgbackrest:test",
			Repos: []cpov1.Repo{{
				Name:    "repo1",
				Storage: "pvc",
				Volume:  cpov1.Volume{Size: "1G"},
			}},
			Resources: &cpov1.Resources{
				ResourceRequests: cpov1.ResourceDescription{CPU: "1", Memory: "10"},
				ResourceLimits:   cpov1.ResourceDescription{CPU: "1", Memory: "10"},
			},
		},
	}

	cluster := supplementalGroupsTestCluster()
	cluster.Postgresql.Spec = spec
	statefulSet, err := cluster.generateRepoHostStatefulSet(&spec)
	if err != nil {
		t.Fatal(err)
	}
	got := statefulSet.Spec.Template.Spec.SecurityContext.SupplementalGroupsPolicy
	if got == nil || *got != policy {
		t.Fatalf("policy was not propagated to the repository host: %v", got)
	}
}

func TestSupplementalGroupsPolicyCRDSchemas(t *testing.T) {
	assertSupplementalGroupsPolicySchema(t, cpov1.PostgresCRDResourceValidation.OpenAPIV3Schema.Properties["spec"])

	for _, relativePath := range []string{
		"../../manifests/postgresql.crd.yaml",
		"../../charts/postgres-operator/crds/postgresqls.yaml",
	} {
		t.Run(filepath.Base(relativePath), func(t *testing.T) {
			raw, err := os.ReadFile(relativePath)
			if err != nil {
				t.Fatal(err)
			}
			var crd apiextv1.CustomResourceDefinition
			if err := yaml.Unmarshal(raw, &crd); err != nil {
				t.Fatal(err)
			}
			if len(crd.Spec.Versions) == 0 || crd.Spec.Versions[0].Schema == nil || crd.Spec.Versions[0].Schema.OpenAPIV3Schema == nil {
				t.Fatal("CRD does not contain an OpenAPI schema")
			}
			assertSupplementalGroupsPolicySchema(t, crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"])
		})
	}
}

func supplementalGroupsPolicy(policy v1.SupplementalGroupsPolicy) *v1.SupplementalGroupsPolicy {
	return &policy
}

func supplementalGroupsTestCluster() *Cluster {
	return New(
		Config{OpConfig: config.Config{
			PodManagementPolicy: "ordered_ready",
			ProtectedRoles:      []string{"admin"},
			Resources: config.Resources{
				MinInstances: -1,
				MaxInstances: -1,
			},
			Auth: config.Auth{
				SuperUsername:       superUserName,
				ReplicationUsername: replicationUserName,
			},
		}},
		k8sutil.KubernetesClient{},
		cpov1.Postgresql{},
		logger,
		eventRecorder,
	)
}

func supplementalGroupsTestSpec(policy *v1.SupplementalGroupsPolicy) cpov1.PostgresSpec {
	return cpov1.PostgresSpec{
		TeamID:                        "myapp",
		NumberOfInstances:             1,
		SpiloSupplementalGroupsPolicy: policy,
		PostgresqlParam:               cpov1.PostgresqlParam{PgVersion: "18"},
		Resources: &cpov1.Resources{
			ResourceRequests: cpov1.ResourceDescription{CPU: "1", Memory: "10"},
			ResourceLimits:   cpov1.ResourceDescription{CPU: "1", Memory: "10"},
		},
		Volume: cpov1.Volume{Size: "1G"},
	}
}

func assertSupplementalGroupsPolicySchema(t *testing.T, spec apiextv1.JSONSchemaProps) {
	t.Helper()
	schema, ok := spec.Properties["spiloSupplementalGroupsPolicy"]
	if !ok || schema.Type != "string" {
		t.Fatal("spiloSupplementalGroupsPolicy string schema is missing")
	}

	values := make(map[string]bool, len(schema.Enum))
	for _, value := range schema.Enum {
		values[string(value.Raw)] = true
	}
	if len(values) != 2 || !values[`"Merge"`] || !values[`"Strict"`] {
		t.Fatalf("unexpected policy enum: %#v", values)
	}
}
