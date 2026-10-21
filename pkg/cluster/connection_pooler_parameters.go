package cluster

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	v1 "k8s.io/api/core/v1"
)

// connectionPoolerParametersEnv is the name of the environment variable which
// carries user defined pgBouncer parameters (JSON object, key -> value) into
// the pooler container. The container merges them into the [pgbouncer]
// section of the generated pgbouncer.ini; they take precedence over the values
// generated from the other CONNECTION_POOLER_* variables.
const connectionPoolerParametersEnv = "CONNECTION_POOLER_PARAMETERS"

var poolerParameterNameRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Parameters managed by the operator / container image. Changing them would
// break authentication, TLS, networking or the lookup of the pooler user.
var poolerForbiddenParameters = map[string]struct{}{
	"auth_file":            {},
	"auth_type":            {},
	"auth_user":            {},
	"auth_query":           {},
	"auth_dbname":          {},
	"auth_hba_file":        {},
	"admin_users":          {},
	"stats_users":          {},
	"listen_addr":          {},
	"listen_port":          {},
	"unix_socket_dir":      {},
	"unix_socket_mode":     {},
	"unix_socket_group":    {},
	"logfile":              {},
	"pidfile":              {},
	"syslog":               {},
	"syslog_ident":         {},
	"syslog_facility":      {},
	"user":                 {},
	"include":              {},
	"client_tls_key_file":  {},
	"client_tls_cert_file": {},
	"client_tls_ca_file":   {},
	"server_tls_key_file":  {},
	"server_tls_cert_file": {},
	"server_tls_ca_file":   {},
}

// validateConnectionPoolerParameter checks a single user supplied pgBouncer
// parameter and returns a descriptive error if it must not be used.
func validateConnectionPoolerParameter(key, value string) error {
	if !poolerParameterNameRegexp.MatchString(key) {
		return fmt.Errorf("invalid connection pooler parameter name %q: must match %s",
			key, poolerParameterNameRegexp.String())
	}
	if _, forbidden := poolerForbiddenParameters[key]; forbidden {
		return fmt.Errorf("connection pooler parameter %q is managed by the operator and cannot be set", key)
	}
	if strings.ContainsAny(value, "\n\r\x00") {
		return fmt.Errorf("invalid value for connection pooler parameter %q: must not contain line breaks or NUL", key)
	}
	return nil
}

// splitConnectionPoolerParameters separates usable parameters from rejected
// ones. Rejected parameters are reported as errors (sorted by key, so the
// output is deterministic) and must be ignored by the caller.
func splitConnectionPoolerParameters(params map[string]string) (map[string]string, []error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	valid := make(map[string]string, len(params))
	var rejected []error
	for _, k := range keys {
		if err := validateConnectionPoolerParameter(k, params[k]); err != nil {
			rejected = append(rejected, err)
			continue
		}
		valid[k] = params[k]
	}
	return valid, rejected
}

// encodeConnectionPoolerParameters serializes the usable parameters to JSON
// and returns the rejected ones as errors. Rejected parameters are skipped,
// they do not make the encoding fail. The JSON string is empty if there is
// nothing to pass on. json.Marshal sorts map keys, so the output is stable
// and does not trigger needless rollouts. The final error is only set if
// the JSON encoding itself fails.
func encodeConnectionPoolerParameters(params map[string]string) (string, []error, error) {
	valid, rejected := splitConnectionPoolerParameters(params)
	if len(valid) == 0 {
		return "", rejected, nil
	}
	b, err := json.Marshal(valid)
	if err != nil {
		return "", rejected, fmt.Errorf("could not encode connection pooler parameters: %v", err)
	}
	return string(b), rejected, nil
}

// needSyncConnectionPoolerParameters compares the parameters requested in the
// spec with the ones in the env of the running pooler container. In contrast
// to the spec diff this works without the previous spec, e.g. after a restart
// of the operator. It returns one reason per changed parameter.
func needSyncConnectionPoolerParameters(spec *cpov1.ConnectionPooler, container v1.Container) (bool, []string) {
	var requested map[string]string
	if spec != nil {
		requested = spec.Parameters
	}
	// rejected parameters are ignored here, they are reported when the
	// deployment is generated
	want, _, err := encodeConnectionPoolerParameters(requested)
	if err != nil {
		return false, nil
	}

	have := ""
	for _, env := range container.Env {
		if env.Name == connectionPoolerParametersEnv {
			have = env.Value
			break
		}
	}

	if have == want {
		return false, nil
	}

	var running, wanted map[string]string
	if have != "" {
		// an unreadable value is treated as empty, so everything is re-applied
		_ = json.Unmarshal([]byte(have), &running)
	}
	if want != "" {
		_ = json.Unmarshal([]byte(want), &wanted)
	}

	reasons := diffConnectionPoolerParameters(running, wanted)
	if len(reasons) == 0 {
		reasons = []string{"connection pooler parameters changed"}
	}
	return true, reasons
}

// diffConnectionPoolerParameters describes what changed between two parameter
// sets, one entry per parameter, sorted by name.
func diffConnectionPoolerParameters(have, want map[string]string) []string {
	names := make(map[string]struct{})
	for k := range have {
		names[k] = struct{}{}
	}
	for k := range want {
		names[k] = struct{}{}
	}
	sorted := make([]string, 0, len(names))
	for k := range names {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	var reasons []string
	for _, k := range sorted {
		oldVal, hadOld := have[k]
		newVal, hasNew := want[k]
		switch {
		case !hadOld:
			reasons = append(reasons, fmt.Sprintf("parameter %s added: '%s'", k, newVal))
		case !hasNew:
			reasons = append(reasons, fmt.Sprintf("parameter %s removed (was '%s')", k, oldVal))
		case oldVal != newVal:
			reasons = append(reasons, fmt.Sprintf("parameter %s changed from '%s' to '%s'", k, oldVal, newVal))
		}
	}
	return reasons
}

// withValidConnectionPoolerParameters returns a shallow copy of the pooler
// spec which only contains the parameters that are actually applied. It is
// used to compare specs, so that ignored parameters do not show up as changes.
// The given spec is not modified.
func withValidConnectionPoolerParameters(spec *cpov1.ConnectionPooler) *cpov1.ConnectionPooler {
	if spec == nil {
		return nil
	}
	filtered := *spec
	valid, _ := splitConnectionPoolerParameters(spec.Parameters)
	if len(valid) == 0 {
		valid = nil
	}
	filtered.Parameters = valid
	return &filtered
}
