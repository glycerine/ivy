package core

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestZddCoverPathToStringEncodings(t *testing.T) {
	dd := NewBDDManager(4)
	dd.ZddSize = 8
	dd.PermZ = []int{0, 1, 2, 3, 4, 5, 6, 7}
	dd.InvPermZ = []int{0, 1, 2, 3, 4, 5, 6, 7}
	got := Cudd_zddCoverPathToString(dd, []int{0, 0, 0, 1, 1, 0, 2, 2}, nil)
	if string(got[:4]) != "-01-" || got[4] != 0 {
		t.Fatalf("cover path = %q", got)
	}
	if Cudd_zddCoverPathToString(dd, []int{0}, nil) != nil {
		t.Fatal("short path accepted")
	}
}

func TestCuddNodeSizeForPointerNames(t *testing.T) {
	word := uintptr(4)
	if ^uintptr(0)>>32 != 0 {
		word = 8
	}
	want := uintptr(6) * word
	if word == 8 {
		want = 5 * word
	}
	if got := Cudd_zddNodeSize(); got != want {
		t.Fatalf("C DdNode size = %d, want %d", got, want)
	}
}

func TestZddDumpDotSupportRanksAndArcStyles(t *testing.T) {
	dd := NewBDDManager(2)
	dd.ZddSize = 2
	dd.PermZ = []int{0, 1}
	dd.InvPermZ = []int{0, 1}
	y := dd.zddNode(1, dd.One, dd.Zero)
	root := dd.zddNode(0, y, dd.Zero)
	var out bytes.Buffer
	if Cudd_zddDumpDot(dd, []*DdNode{root}, []string{"x", "y"}, []string{"out"}, &out) != 1 {
		t.Fatal("dump failed")
	}
	got := out.String()
	for _, want := range []string{"digraph \"ZDD\" {", "\" x \" ->", "\" y \" ->", "\"  out  \"", "[style = dashed]", "[label = \"0\"]"} {
		if !strings.Contains(got, want) {
			t.Errorf("DOT output missing %q:\n%s", want, got)
		}
	}
	if c := strings.Count(got, "[style = dashed]"); c == 0 {
		t.Fatal("else arcs missing")
	}
	// Addresses are runtime-specific; their placement and hexadecimal form are stable.
	if !regexp.MustCompile(`"0x[0-9a-f]+" -> "0x[0-9a-f]+"`).MatchString(got) {
		t.Fatalf("node IDs not in C pointer format:\n%s", got)
	}
}

func TestZddSymmetryAndGroupSiftingFlow(t *testing.T) {
	dd := NewBDDManager(2)
	dd.ZddSize = 2
	dd.PermZ = []int{0, 1}
	dd.InvPermZ = []int{0, 1}
	y := dd.zddNode(1, dd.One, dd.One)
	f := dd.zddNode(0, y, y)
	if !zddSymmCheck(dd, 0, 1) {
		t.Fatal("equivalent adjacent variables were not recognized as symmetric")
	}
	if Cudd_zddReduceHeap(dd, CUDD_REORDER_SIFT, 0) == 0 {
		t.Fatal("symmetric sift failed")
	}
	if !validZddOrder(dd) {
		t.Fatal("symmetric sift corrupted ZDD order")
	}
	_ = f
}

func TestZddDumpDotExactTextLayout(t *testing.T) {
	dd := NewBDDManager(1)
	dd.ZddSize = 1
	dd.PermZ = []int{0}
	dd.InvPermZ = []int{0}
	root := dd.zddNode(0, dd.One, dd.Zero)
	var out bytes.Buffer
	if Cudd_zddDumpDot(dd, []*DdNode{root}, nil, nil, &out) != 1 {
		t.Fatal("dump failed")
	}
	normalize := regexp.MustCompile(`0x[0-9a-f]+`).ReplaceAllString(out.String(), "NODE")
	want := `digraph "ZDD" {
size = "7.5,10"
center = true;
edge [dir = none];
{ node [shape = plaintext];
  edge [style = invis];
  "CONST NODES" [style = invis];
" 0 " -> "CONST NODES"; 
}
{ rank = same; node [shape = box]; edge [style = invis];
"F0"; }
{ rank = same; " 0 ";
"NODE";
}
{ rank = same; "CONST NODES";
{ node [shape = box]; "NODE";
"NODE";
}
}
"F0" -> "NODE" [style = solid];
"NODE" -> "NODE";
"NODE" -> "NODE" [style = dashed];
"NODE" [label = "1"];
"NODE" [label = "0"];
}
`
	if normalize != want {
		t.Fatalf("DOT text mismatch\n--- got ---\n%s\n--- want ---\n%s", normalize, want)
	}
}
