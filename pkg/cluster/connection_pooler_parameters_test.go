package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeConnectionPoolerParameters(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]string
		want    string
		wantErr string
	}{
		{name: "nil", params: nil, want: ""},
		{name: "empty", params: map[string]string{}, want: ""},
		{
			name:   "sorted and stable",
			params: map[string]string{"server_idle_timeout": "300", "query_timeout": "60"},
			want:   `{"query_timeout":"60","server_idle_timeout":"300"}`,
		},
		{
			name:   "comma separated value",
			params: map[string]string{"ignore_startup_parameters": "extra_float_digits,options"},
			want:   `{"ignore_startup_parameters":"extra_float_digits,options"}`,
		},
		{
			name:    "forbidden key",
			params:  map[string]string{"auth_file": "/tmp/x"},
			wantErr: "managed by the operator",
		},
		{
			name:    "forbidden tls file",
			params:  map[string]string{"client_tls_key_file": "/tmp/x"},
			wantErr: "managed by the operator",
		},
		{
			name:    "invalid key",
			params:  map[string]string{"Bad-Key": "1"},
			wantErr: "invalid connection pooler parameter name",
		},
		{
			name:    "section injection via key",
			params:  map[string]string{"[databases]": "1"},
			wantErr: "invalid connection pooler parameter name",
		},
		{
			name:    "line break in value",
			params:  map[string]string{"query_timeout": "60\n[databases]\n* = host=evil"},
			wantErr: "line breaks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodeConnectionPoolerParameters(tt.params)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
