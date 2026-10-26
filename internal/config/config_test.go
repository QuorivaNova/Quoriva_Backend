package config

import (
	"strings"
	"testing"
)

func TestFromEnv(t *testing.T) {
	for _, tc := range []struct {
		name, address, database string
		valid                   bool
	}{
		{"valid default address", "", "postgres://user:password@db/quoriva", true},
		{"valid explicit address", "127.0.0.1:8080", "postgresql://db/quoriva", true},
		{"missing database", "", "", false},
		{"invalid scheme", "", "https://db/quoriva", false},
		{"missing database name", "", "postgres://db/", false},
		{"invalid address", "8080", "postgres://db/quoriva", false},
		{"invalid secret URL", "", "postgres://user:private-secret@%invalid/quoriva", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := FromEnv(func(key string) string {
				if key == "HTTP_ADDR" {
					return tc.address
				}
				return tc.database
			})
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, got error %v", tc.valid, err)
			}
			if err != nil && strings.Contains(err.Error(), "private-secret") {
				t.Fatal("configuration error exposes a credential")
			}
			if tc.valid && tc.address == "" && cfg.HTTPAddr != ":8080" {
				t.Fatalf("unexpected default address %q", cfg.HTTPAddr)
			}
		})
	}
}
