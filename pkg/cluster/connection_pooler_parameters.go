package cluster

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
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

// validateConnectionPoolerParameters checks user supplied pgBouncer
// parameters and returns a descriptive error for the first problem found
// (keys are checked in sorted order so the message is deterministic).
func validateConnectionPoolerParameters(params map[string]string) error {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if !poolerParameterNameRegexp.MatchString(k) {
			return fmt.Errorf("invalid connection pooler parameter name %q: must match %s",
				k, poolerParameterNameRegexp.String())
		}
		if _, forbidden := poolerForbiddenParameters[k]; forbidden {
			return fmt.Errorf("connection pooler parameter %q is managed by the operator and cannot be set", k)
		}
		if strings.ContainsAny(params[k], "\n\r\x00") {
			return fmt.Errorf("invalid value for connection pooler parameter %q: must not contain line breaks or NUL", k)
		}
	}
	return nil
}

// encodeConnectionPoolerParameters validates the parameters and serializes
// them to JSON. It returns an empty string if there is nothing to pass on.
// json.Marshal sorts map keys, so the output is stable and does not trigger
// needless rollouts.
func encodeConnectionPoolerParameters(params map[string]string) (string, error) {
	if len(params) == 0 {
		return "", nil
	}
	if err := validateConnectionPoolerParameters(params); err != nil {
		return "", err
	}
	b, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("could not encode connection pooler parameters: %v", err)
	}
	return string(b), nil
}
