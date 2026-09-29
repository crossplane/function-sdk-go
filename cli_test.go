/*
Copyright 2026 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package function

import (
	"testing"
	"time"

	"github.com/alecthomas/kong"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/function-sdk-go/response"
)

func TestCLIParse(t *testing.T) {
	type want struct {
		cli            CLI
		maxRecvMsgSize int
		maxSendMsgSize int
	}

	defaults := CLI{
		Address:            ":9443",
		MaxRecvMessageSize: 4,
		Network:            "tcp",
		TTL:                response.DefaultTTL,
	}

	cases := map[string]struct {
		reason string
		args   []string
		env    map[string]string
		want   want
	}{
		"Defaults": {
			reason: "With no flags or environment variables we should use the defaults.",
			want: want{
				cli:            defaults,
				maxRecvMsgSize: 4 * 1024 * 1024,
				maxSendMsgSize: 4 * 1024 * 1024,
			},
		},
		"Flags": {
			reason: "We should accept the standard flag names.",
			args: []string{
				"--address=:8080", "--debug", "--insecure", "--network=unix",
				"--max-recv-message-size=8", "--max-send-message-size=2",
				"--ttl=30s", "--tls-certs-dir=/certs",
			},
			want: want{
				cli: CLI{
					Address:            ":8080",
					Debug:              true,
					Insecure:           true,
					MaxRecvMessageSize: 8,
					MaxSendMessageSize: 2,
					Network:            "unix",
					TTL:                30 * time.Second,
					TLSCertsDir:        "/certs",
				},
				maxRecvMsgSize: 8 * 1024 * 1024,
				maxSendMsgSize: 2 * 1024 * 1024,
			},
		},
		"Aliases": {
			reason: "We should accept the flag names existing functions already use.",
			args:   []string{"--insecure", "--max-grpc-message-size=16", "--tls-server-certs-dir=/certs"},
			want: want{
				cli: func() CLI {
					c := defaults
					c.Insecure = true
					c.MaxRecvMessageSize = 16
					c.TLSCertsDir = "/certs"
					return c
				}(),
				maxRecvMsgSize: 16 * 1024 * 1024,
				maxSendMsgSize: 16 * 1024 * 1024,
			},
		},
		"EnvironmentVariables": {
			reason: "We should accept the environment variables existing functions already use.",
			env: map[string]string{
				"INSECURE":              "true",
				"MAX_GRPC_MESSAGE_SIZE": "16",
				"TLS_SERVER_CERTS_DIR":  "/certs",
				"TTL":                   "0s",
			},
			want: want{
				cli: func() CLI {
					c := defaults
					c.Insecure = true
					c.MaxRecvMessageSize = 16
					c.TLSCertsDir = "/certs"
					c.TTL = 0
					return c
				}(),
				maxRecvMsgSize: 16 * 1024 * 1024,
				maxSendMsgSize: 16 * 1024 * 1024,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			c := &CLI{}
			p, err := kong.New(c, kongVars())
			if err != nil {
				t.Fatalf("\n%s\nkong.New(): unexpected error: %v", tc.reason, err)
			}
			if _, err := p.Parse(tc.args); err != nil {
				t.Fatalf("\n%s\nParse(%v): unexpected error: %v", tc.reason, tc.args, err)
			}
			if diff := cmp.Diff(tc.want.cli, *c); diff != "" {
				t.Errorf("\n%s\nParse(%v): -want, +got:\n%s", tc.reason, tc.args, diff)
			}

			so := &ServeOptions{}
			for _, o := range c.StandardOptions() {
				if err := o(so); err != nil {
					t.Fatalf("\n%s\nStandardOptions(): unexpected error: %v", tc.reason, err)
				}
			}
			if so.MaxRecvMsgSize != tc.want.maxRecvMsgSize {
				t.Errorf("\n%s\nStandardOptions(): MaxRecvMsgSize: want %d, got %d", tc.reason, tc.want.maxRecvMsgSize, so.MaxRecvMsgSize)
			}
			if so.MaxSendMsgSize != tc.want.maxSendMsgSize {
				t.Errorf("\n%s\nStandardOptions(): MaxSendMsgSize: want %d, got %d", tc.reason, tc.want.maxSendMsgSize, so.MaxSendMsgSize)
			}
		})
	}
}

func TestStandardOptionsInsecureIgnoresTLSCertsDir(t *testing.T) {
	c := &CLI{Insecure: true, TLSCertsDir: "/does/not/exist", MaxRecvMessageSize: 4}
	so := &ServeOptions{}
	for _, o := range c.StandardOptions() {
		if err := o(so); err != nil {
			t.Fatalf("StandardOptions() with --insecure: unexpected error: %v", err)
		}
	}
	if so.Credentials == nil {
		t.Fatal("StandardOptions() with --insecure: expected insecure credentials")
	}
}
