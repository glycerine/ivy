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

func TestReadSFCommand(t *testing.T) {
	file := filepath.Join(t.TempDir(), "input.sf")
	if err := os.WriteFile(file, []byte("header\nSDF\n2 1 1\n11 1\nEND_SDF\n"), 0600); err != nil {
		t.Fatal(err)
	}
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	if IoCommandReadSF(frame, []string{"read_sf", file}) != 0 {
		t.Fatalf("read failed: %s", errors.String())
	}
	n := frame.Network().(*abc.Network)
	if abc.Abc_NtkPiNum(n) != 2 || abc.Abc_NtkPoNum(n) != 1 || n.Spec != file {
		t.Fatal("incorrect network")
	}
	var output bytes.Buffer
	if err := transformSFToPLA(strings.NewReader("SDF\n1 1 1\n1 1\nEND_SDF\n"), &output); err != nil || output.String() != ".type fd\n.i 1\n.o 1\n.p 1\n1 1\n.e\n" {
		t.Fatalf("PLA %q: %v", output.String(), err)
	}
}
