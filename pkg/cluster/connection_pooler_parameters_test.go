package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeConnectionPoolerParameters(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]string
		want         string
		wantRejected []string // substrings expected in the rejection messages, in order
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
			name:         "forbidden key is skipped",
			params:       map[string]string{"auth_type": "md5"},
			want:         "",
			wantRejected: []string{"managed by the operator"},
		},
		{
			name:         "forbidden tls file is skipped",
			params:       map[string]string{"client_tls_key_file": "/tmp/x"},
			want:         "",
			wantRejected: []string{"managed by the operator"},
		},
		{
			name:         "invalid key is skipped",
			params:       map[string]string{"Bad-Key": "1"},
			want:         "",
			wantRejected: []string{"invalid connection pooler parameter name"},
		},
		{
			name:         "section injection via key is skipped",
			params:       map[string]string{"[databases]": "1"},
			want:         "",
			wantRejected: []string{"invalid connection pooler parameter name"},
		},
		{
			name:         "line break in value is skipped",
			params:       map[string]string{"query_timeout": "60\n[databases]\n* = host=evil"},
			want:         "",
			wantRejected: []string{"line breaks"},
		},
		{
			name: "invalid parameters do not break valid ones",
			params: map[string]string{
				"query_timeout":   "60",
				"auth_type":       "md5",
				"max_client_conn": "5000",
				"listen_port":     "6433",
			},
			want:         `{"max_client_conn":"5000","query_timeout":"60"}`,
			wantRejected: []string{`"auth_type"`, `"listen_port"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, rejected, err := encodeConnectionPoolerParameters(tt.params)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			if assert.Len(t, rejected, len(tt.wantRejected)) {
				for i, sub := range tt.wantRejected {
					assert.ErrorContains(t, rejected[i], sub)
				}
			}
		})
	}
}
