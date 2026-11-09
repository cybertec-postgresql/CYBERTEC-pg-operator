package patroni

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
)

const postgresTLSCAFileEnv = "POSTGRES_TLS_CA_FILE"

// podAPIScheme follows the effective configuration of the target Pod. This
// allows old HTTP Pods and new HTTPS Pods to coexist during an ordered rollout.
func podAPIScheme(pod *v1.Pod) string {
	for _, container := range pod.Spec.Containers {
		if container.Name != "postgres" {
			continue
		}

		for _, env := range container.Env {
			if env.Name != "SPILO_CONFIGURATION" {
				continue
			}

			var config struct {
				RestAPI struct {
					CertFile string `json:"certfile"`
				} `json:"restapi"`
			}
			if err := json.Unmarshal([]byte(env.Value), &config); err != nil {
				// Do not silently select plaintext for an unreadable declaration.
				return "https"
			}
			if config.RestAPI.CertFile != "" {
				return "https"
			}
		}
	}

	return "http"
}

type invalidTrustTransport struct {
	err error
}

func (transport invalidTrustTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, transport.err
}

func newVerifiedHTTPClient(serverName string) *http.Client {
	client := &http.Client{Timeout: timeout}
	caFile := os.Getenv(postgresTLSCAFileEnv)
	if caFile == "" {
		return client
	}

	ca, err := os.ReadFile(caFile)
	pool := x509.NewCertPool()
	if err != nil || !pool.AppendCertsFromPEM(ca) {
		client.Transport = invalidTrustTransport{
			err: fmt.Errorf("invalid PostgreSQL/Patroni CA file %q", caFile),
		}
		return client
	}
	if serverName == "" {
		client.Transport = invalidTrustTransport{
			err: fmt.Errorf("Patroni TLS requires the cluster server name"),
		}
		return client
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    pool,
		ServerName: serverName,
	}
	transport.TLSHandshakeTimeout = 10 * time.Second
	client.Transport = transport
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return client
}
