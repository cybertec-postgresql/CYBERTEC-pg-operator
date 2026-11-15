package cluster

import (
	"os"
	"path/filepath"
	"testing"

	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util/config"
	v1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"
)

func TestEphemeralStorageQuantityContract(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		fallback string
		want     string
		invalid  bool
	}{
		{name: "omitted"},
		{name: "default", fallback: "64Mi", want: "64Mi"},
		{name: "override", value: "128Mi", fallback: "64Mi", want: "128Mi"},
		{name: "explicit zero", value: "0", fallback: "64Mi", want: "0"},
		{name: "decimal", value: "1000000", want: "1000000"},
		{name: "invalid", value: "banana", invalid: true},
		{name: "negative", value: "-1Mi", invalid: true},
		{name: "invalid default", fallback: "bad", invalid: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := fillResourceList(
				cpov1.ResourceDescription{CPU: "100m", Memory: "128Mi", EphemeralStorage: test.value},
				cpov1.ResourceDescription{EphemeralStorage: test.fallback},
			)
			if (err != nil) != test.invalid {
				t.Fatalf("unexpected error: %v", err)
			}
			if err != nil {
				return
			}

			quantity, ok := got[v1.ResourceEphemeralStorage]
			if test.want == "" {
				if ok {
					t.Fatal("omitted ephemeral-storage changed the resource list")
				}
				return
			}
			if !ok || quantity.Cmp(resource.MustParse(test.want)) != 0 {
				t.Fatalf("unexpected ephemeral-storage quantity: %v", quantity)
			}
		})
	}
}

func TestEphemeralStorageRequestLimitAndRoundTrip(t *testing.T) {
	cluster := newMockCluster(config.Config{})
	defaults := cpov1.Resources{
		ResourceRequests: cpov1.ResourceDescription{CPU: "100m", Memory: "128Mi"},
		ResourceLimits:   cpov1.ResourceDescription{CPU: "1", Memory: "512Mi"},
	}
	configured := cpov1.Resources{
		ResourceRequests: cpov1.ResourceDescription{EphemeralStorage: "64Mi"},
		ResourceLimits:   cpov1.ResourceDescription{EphemeralStorage: "256Mi"},
	}

	generated, err := cluster.generateResourceRequirements(&configured, defaults, "sidecar")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := parseResourceRequirements(*generated)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.ResourceRequests.EphemeralStorage != "64Mi" || parsed.ResourceLimits.EphemeralStorage != "256Mi" {
		t.Fatalf("ephemeral-storage was lost during round-trip: %#v", parsed)
	}

	configured.ResourceRequests.EphemeralStorage = "512Mi"
	if _, err := cluster.generateResourceRequirements(&configured, defaults, "sidecar"); err == nil {
		t.Fatal("ephemeral-storage request greater than the limit was accepted")
	}

	configured.ResourceLimits.EphemeralStorage = ""
	if _, err := cluster.generateResourceRequirements(&configured, defaults, "sidecar"); err != nil {
		t.Fatalf("an unbounded ephemeral-storage request must remain supported: %v", err)
	}
}

func TestEphemeralStorageRegisteredInCRDSchema(t *testing.T) {
	spec := cpov1.PostgresCRDResourceValidation.OpenAPIV3Schema.Properties["spec"]
	assertEphemeralStorageSchema(t, spec)
}

func TestEphemeralStorageInCRDManifests(t *testing.T) {
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
			spec := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"]
			assertEphemeralStorageSchema(t, spec)
		})
	}
}

func assertEphemeralStorageSchema(t *testing.T, spec apiextv1.JSONSchemaProps) {
	t.Helper()
	resourceSchemas := []struct {
		name   string
		schema map[string]struct{}
	}{
		{
			name:   "postgres",
			schema: resourceFields(spec.Properties["resources"]),
		},
		{
			name:   "connection pooler",
			schema: resourceFields(spec.Properties["connectionPooler"].Properties["resources"]),
		},
		{
			name:   "pgBackRest",
			schema: resourceFields(spec.Properties["backup"].Properties["pgbackrest"].Properties["resources"]),
		},
	}

	for _, resourceSchema := range resourceSchemas {
		for _, field := range []string{"requests", "limits"} {
			if _, ok := resourceSchema.schema[field]; !ok {
				t.Fatalf("%s %s schema does not include ephemeral-storage", resourceSchema.name, field)
			}
		}
	}
}

func resourceFields(schema apiextv1.JSONSchemaProps) map[string]struct{} {
	result := make(map[string]struct{}, 2)
	for _, field := range []string{"requests", "limits"} {
		if schema.Properties[field].Properties["ephemeral-storage"].Type == "string" {
			result[field] = struct{}{}
		}
	}
	return result
}
