package gia

import (
	"os/exec"
	"strings"
	"testing"
)

// Keep the original balancing policy observable separately from the generic
// mapper, which intentionally clears cell configurations and traces timing.
func TestMappingNativeBalanceRetainsMetadata(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_balance_metadata_review.c")
	want := "config-owner 1 retained-cell source-cleared 1\nidentity balance-review source-spec\nci-arrivals 7 11 13\nco-required 30 co-arrival 91\ntruth 11111011\n"
	for _, mode := range []string{"0", "1", "2"} {
		output, err := exec.Command(exe, mode).CombinedOutput()
		if err != nil {
			t.Fatalf("source balance mode%s: %v %s", mode, err, output)
		}
		expected := want
		if mode == "2" {
			expected = strings.ReplaceAll(expected, "config-owner 1 retained-cell", "config-owner 0 -")
			expected = strings.ReplaceAll(expected, "co-arrival 91", "co-arrival 14")
		}
		if string(output) != expected {
			t.Fatalf("source balance mode%s policy changed: %q want%q", mode, output, expected)
		}
	}
}

func TestMappingNativeIfSourceNameOwnership(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_balance_metadata_review.c")
	output, err := exec.Command(exe, "3").CombinedOutput()
	if err != nil {
		t.Fatalf("source IF naming: %v %s", err, output)
	}
	if string(output) != "if-name balance-review\nafter-source-rename balance-review\n" {
		t.Fatalf("original IF name copy changed: %s", output)
	}
}
