package abc

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuildPLA(t *testing.T) {
	covers := []string{"1000 1\n", "0110 1\n0011 1\n"}
	n, err := BuildPLA(covers, 4)
	if err != nil {
		t.Fatal(err)
	}
	defer Abc_NtkDelete(n)
	inputs, outputs := n.ObjectsOfKind(ObjectPI), n.ObjectsOfKind(ObjectPO)
	for mask := 0; mask < 16; mask++ {
		values := make(map[*Object]bool)
		for i, input := range inputs {
			values[input] = mask&(1<<i) != 0
		}
		want := []bool{mask == 1, mask == 6 || mask == 12}
		for i, output := range outputs {
			got, err := n.Evaluate(output.fanins[0], values)
			if err != nil || got != want[i] {
				t.Fatalf("mask=%d output=%d got=%v want=%v err=%v", mask, i, got, want[i], err)
			}
		}
	}
	var out bytes.Buffer
	if err := writePLATestBLIF(&out, n); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), ".model pla") || !strings.Contains(out.String(), ".end") {
		t.Fatal("invalid demonstration BLIF")
	}
	if _, err := BuildPLA([]string{"10x 1\n"}, 3); err == nil {
		t.Fatal("invalid cover accepted")
	}
}
