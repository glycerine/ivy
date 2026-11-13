package live

import (
	"bytes"
	"strings"
	"testing"
)

func TestSourceCommandUsageOutput(t *testing.T) {
	cases := []struct {
		name                string
		run                 func(*Frame) int32
		prefix, description string
	}{
		{"l2s", func(f *Frame) int32 { return Abc_CommandAbcLivenessToSafety(f, 2, []string{"l2s", "-h"}) }, "usage: l2s [-1lsh]\n", "-1 : no shadow logic"},
		{"l3s", func(f *Frame) int32 { return Abc_CommandAbcLivenessToSafetyWithLTL(f, 2, []string{"l3s", "-h"}) }, "usage: l3s [-1lsh]\n", "-s : ignore safety"},
		{"kcs", func(f *Frame) int32 { return Abc_CommandCS_kLiveness(f, 2, []string{"kcs", "-h"}) }, "usage: kcs [-cmgCh]\n", "-g : verification"},
		{"nck", func(f *Frame) int32 { return Abc_CommandNChooseK(f, 2, []string{"nck", "-h"}) }, "usage: nck [-cmgCh]\n", "generates combinatorial"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			frame := &Frame{Network: liveFixture(false), Out: &out}
			if status := tc.run(frame); status != 1 {
				t.Fatalf("status %d", status)
			}
			if text := out.String(); !strings.HasPrefix(text, tc.prefix) || !strings.Contains(text, tc.description) {
				t.Fatalf("usage text %q", text)
			}
		})
	}
}

func TestSourceUsagePrecedesEmptyNetworkExceptSimulation(t *testing.T) {
	var out, errout bytes.Buffer
	frame := &Frame{Out: &out, Err: &errout}
	if status := Abc_CommandAbcLivenessToSafety(frame, 2, []string{"l2s", "-h"}); status != 1 || !strings.HasPrefix(out.String(), "usage: l2s") || errout.Len() != 0 {
		t.Fatalf("l2s help: status=%d out=%q err=%q", status, out.String(), errout.String())
	}
	out.Reset()
	errout.Reset()
	if status := Abc_CommandAbcLivenessToSafety(frame, 1, []string{"l2s"}); status != 1 || errout.String() != "Empty network.\n" {
		t.Fatalf("l2s empty: status=%d err=%q", status, errout.String())
	}
	out.Reset()
	errout.Reset()
	if status := Abc_CommandAbcLivenessToSafetySim(frame, 2, []string{"l2ssim", "-h"}); status != 1 || errout.String() != "Empty network.\n" || out.Len() != 0 {
		t.Fatalf("simulation empty: status=%d out=%q err=%q", status, out.String(), errout.String())
	}
}
