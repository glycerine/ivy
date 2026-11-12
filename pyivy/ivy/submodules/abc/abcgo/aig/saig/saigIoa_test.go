package saig

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestBLIFRoundTripPreservesSequentialTruth(t *testing.T) {
	original := aig.NewGraph()
	pi0 := original.NewInput()
	pi1 := original.NewInput()
	lo := original.NewInput()
	po := original.AND(pi0.Not(), pi1)
	li := original.OR(lo, pi0)
	original.AddOutput(po)
	original.AddOutput(li)
	aig.Aig_ManSetRegNum(original, 1)

	fileName := filepath.Join(t.TempDir(), "roundtrip.blif")
	if err := WriteBLIF(fileName, original); err != nil {
		t.Fatal(err)
	}
	restored, err := ReadBLIF(fileName)
	if err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		want, err := original.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := restored.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("inputs %03b: restored outputs %v, want %v", mask, got, want)
		}
	}
}

func TestReadBLIFRejectsUnsupportedCovers(t *testing.T) {
	for _, content := range []string{
		".model x\n.inputs a\n.outputs y\n.names a y\n1 1\n0 1\n.end\n",
		".model x\n.inputs a\n.outputs y\n.foo bar\n.end\n",
	} {
		name := filepath.Join(t.TempDir(), "bad.blif")
		if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadBLIF(name); err == nil {
			t.Fatalf("accepted unsupported BLIF:\n%s", content)
		}
	}
}
