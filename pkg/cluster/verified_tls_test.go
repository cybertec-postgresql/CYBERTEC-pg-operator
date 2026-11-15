package cluster

import "testing"

func TestPostgresTLSOptions(t *testing.T) {
	t.Run("keeps the current mode without a CA", func(t *testing.T) {
		t.Setenv(postgresTLSCAFileEnv, "")
		if got := postgresTLSOptions(); got != "sslmode=require" {
			t.Fatalf("unexpected TLS options: %q", got)
		}
	})

	t.Run("enables full verification with a CA", func(t *testing.T) {
		t.Setenv(postgresTLSCAFileEnv, `/public/CA's bundle.crt`)
		want := `sslmode=verify-full sslrootcert='/public/CA\'s bundle.crt'`
		if got := postgresTLSOptions(); got != want {
			t.Fatalf("unexpected TLS options: %q", got)
		}
	})
}
