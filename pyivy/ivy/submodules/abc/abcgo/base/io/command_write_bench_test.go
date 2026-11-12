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

func TestWriteBenchCommandModes(t *testing.T) {
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	frame.ReplaceCurrentNetwork(abc.Abc_NtkCreateWithNode("11 1\n"))
	for _, options := range [][]string{nil, {"-l"}} {
		filename := filepath.Join(t.TempDir(), "out.bench")
		args := append([]string{"write_bench"}, options...)
		args = append(args, filename)
		if IoCommandWriteBench(frame, args) != 0 {
			t.Fatalf("%v: %s", options, errors.String())
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		if options == nil && !strings.Contains(string(data), "LUT 0x8") {
			t.Fatalf("expected AND LUT: %s", data)
		}
		if _, err := ReadBench(filename, true); err != nil {
			t.Fatalf("invalid BENCH %v: %v\n%s", options, err, data)
		}
	}
}

func TestWriteBenchInverterLut(t *testing.T) {
	frame := abcmain.NewFrame()
	frame.ReplaceCurrentNetwork(abc.Abc_NtkCreateWithNode("0 1\n"))
	filename := filepath.Join(t.TempDir(), "inv.bench")
	if IoCommandWriteBench(frame, []string{"write_bench", filename}) != 0 {
		t.Fatal("write failed")
	}
	data, err := os.ReadFile(filename)
	if err != nil || !strings.Contains(string(data), "LUT 0x1") {
		t.Fatalf("inverter output %q: %v", data, err)
	}
}
