package patroni

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
)

func TestVerifiedPatroniTransport(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	caFile := filepath.Join(t.TempDir(), "ca.crt")
	writeCertificate(t, caFile, server.Certificate().Raw)
	t.Setenv(postgresTLSCAFileEnv, caFile)
	serverName := server.Certificate().DNSNames[0]

	response, err := newVerifiedHTTPClient(serverName).Get(server.URL)
	if err != nil {
		t.Fatalf("trusted server must succeed: %v", err)
	}
	response.Body.Close()

	if _, err := newVerifiedHTTPClient("wrong-server.invalid").Get(server.URL); err == nil {
		t.Fatal("wrong hostname accepted")
	}

	foreignCA := newCertificateAuthority(t)
	writeCertificate(t, caFile, foreignCA)
	if _, err := newVerifiedHTTPClient(serverName).Get(server.URL); err == nil {
		t.Fatal("untrusted CA accepted")
	}

	if err := os.WriteFile(caFile, []byte("invalid trust bundle"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := newVerifiedHTTPClient(serverName).Get(server.URL); err == nil {
		t.Fatal("invalid CA accepted")
	}

	t.Setenv(postgresTLSCAFileEnv, filepath.Join(t.TempDir(), "missing-ca.crt"))
	if _, err := newVerifiedHTTPClient(serverName).Get(server.URL); err == nil {
		t.Fatal("missing CA accepted")
	}
}

func TestVerifiedPatroniTransportRejectsMissingServerNameAndRedirects(t *testing.T) {
	redirect := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://example.invalid", http.StatusFound)
	}))
	t.Cleanup(redirect.Close)

	caFile := filepath.Join(t.TempDir(), "ca.crt")
	writeCertificate(t, caFile, redirect.Certificate().Raw)
	t.Setenv(postgresTLSCAFileEnv, caFile)
	serverName := redirect.Certificate().DNSNames[0]

	if _, err := newVerifiedHTTPClient("").Get(redirect.URL); err == nil {
		t.Fatal("missing server name accepted")
	}

	response, err := newVerifiedHTTPClient(serverName).Get(redirect.URL)
	if err != nil {
		t.Fatalf("redirect response failed: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusFound {
		t.Fatalf("redirect was followed: got status %d", response.StatusCode)
	}
}

func TestPatroniTransportFollowsTargetPodDeclaration(t *testing.T) {
	pod := &v1.Pod{
		Spec: v1.PodSpec{Containers: []v1.Container{{Name: "postgres"}}},
		Status: v1.PodStatus{
			PodIP: "192.0.2.8",
		},
	}

	url, err := apiURL(pod)
	if err != nil || url != "http://192.0.2.8:8008" {
		t.Fatalf("unexpected legacy endpoint: %s %v", url, err)
	}

	pod.Spec.Containers[0].Env = []v1.EnvVar{{
		Name:  "PATRONI_RESTAPI_CERTFILE",
		Value: "/tls/tls.crt",
	}}
	url, err = apiURL(pod)
	if err != nil || url != "http://192.0.2.8:8008" {
		t.Fatalf("unexpected env-only endpoint: %s %v", url, err)
	}

	pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, v1.EnvVar{
		Name:  "SPILO_CONFIGURATION",
		Value: `{"restapi":{"certfile":"/tls/tls.crt"}}`,
	})
	url, err = apiURL(pod)
	if err != nil || url != "https://192.0.2.8:8008" {
		t.Fatalf("unexpected TLS endpoint: %s %v", url, err)
	}

	pod.Spec.Containers[0].Env[1].Value = "not-json"
	url, err = apiURL(pod)
	if err != nil || url != "https://192.0.2.8:8008" {
		t.Fatalf("unreadable TLS declaration selected plaintext: %s %v", url, err)
	}
}

func writeCertificate(t *testing.T, path string, certificate []byte) {
	t.Helper()
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate,
	}), 0o600); err != nil {
		t.Fatal(err)
	}
}

func newCertificateAuthority(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(42),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
	}
	certificate, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return certificate
}
