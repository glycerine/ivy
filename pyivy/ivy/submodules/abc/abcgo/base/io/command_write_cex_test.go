package io

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestWriteCexNamedAndRaw(t *testing.T) {
	n := abc.Abc_NtkCreateWithNode("11 1\n")
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	frame.ReplaceCurrentNetwork(n)
	frame.SetVerificationResult(0, 0, &abcmain.Counterexample{PIs: 2, Output: 0, Frame: 0, Bits: []uint64{3}})
	for _, tc := range []struct {
		options []string
		want    string
	}{{nil, "11# DONE\n"}, {[]string{"-n"}, "a@0=1\nb@0=1\n# DONE\n"}} {
		file := filepath.Join(t.TempDir(), "cex.txt")
		args := append([]string{"write_cex"}, tc.options...)
		args = append(args, file)
		if IoCommandWriteCex(frame, args) != 0 {
			t.Fatalf("command %v failed: %s", args, errors.String())
		}
		data, err := os.ReadFile(file)
		if err != nil || !strings.Contains(string(data), tc.want) {
			t.Fatalf("output %q: %v", data, err)
		}
	}
}
