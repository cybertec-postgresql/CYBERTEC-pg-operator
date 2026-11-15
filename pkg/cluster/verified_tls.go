package cluster

import (
	"fmt"
	"os"
	"strings"
)

const postgresTLSCAFileEnv = "POSTGRES_TLS_CA_FILE"

func postgresTLSOptions() string {
	caFile := os.Getenv(postgresTLSCAFileEnv)
	if caFile == "" {
		return "sslmode=require"
	}

	caFile = strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(caFile)
	return fmt.Sprintf("sslmode=verify-full sslrootcert='%s'", caFile)
}
