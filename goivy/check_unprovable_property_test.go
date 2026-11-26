package goivy

import (
	"strings"
	"testing"
)

func TestUnprovablePropertiesAreNeverAxioms(t *testing.T) {
	const source = `#lang ivy1.8
unprovable property [intentional] false
property [ordinary] true
`
	for _, tc := range []struct {
		name, source string
		unprovable   bool
		fails        bool
	}{
		{"ordinary mode", source, false, false},
		{"reserved property fails as intended", source, true, false},
		{"provable property rejected", strings.Replace(source, "[intentional] false", "[intentional] true", 1), true, true},
		{"reserved property is not assumed", strings.Replace(source, "[ordinary] true", "[ordinary] false", 1), false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.OnlyCheckUnprovable = tc.unprovable
			err := StartSourceWithConfig("unprovable_property.ivy", tc.source, cfg)
			if tc.fails {
				if err == nil || !strings.Contains(err.Error(), "failed checks: 1") {
					t.Fatalf("error = %v, want one failed check", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
