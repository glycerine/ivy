package io

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestReadTruthCommand(t *testing.T) {
	file := filepath.Join(t.TempDir(), "truth.txt")
	if err := os.WriteFile(file, []byte("1000\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"read_truth", "8"}, {"read_truth", "-x", "1000"}, {"read_truth", "-xf", file}} {
		frame := abcmain.NewFrame()
		var errors bytes.Buffer
		frame.Err = &errors
		if IoCommandReadTruth(frame, args) != 0 {
			t.Fatalf("%v: %s", args, errors.String())
		}
		n := frame.Network().(*abc.Network)
		nodes := n.ObjectsOfKind(abc.ObjectNode)
		if len(nodes) != 1 || abc.Abc_SopToTruth(nodes[0].Data.(string), 2)&15 != 8 {
			t.Fatalf("%v: wrong function", args)
		}
	}
	frame := abcmain.NewFrame()
	frame.Err = &bytes.Buffer{}
	if IoCommandReadTruth(frame, []string{"read_truth", "-x", "bad"}) != 1 {
		t.Fatal("accepted invalid truth")
	}
}
