package gia

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGigValidationRejectsMalformedRecords(t *testing.T) {
	for _, tc := range []struct{ name, text string }{
		{"missing-equals", "w2 PI\n"}, {"reserved-zero", "w0 = PI\n"}, {"reserved-one", "w1 = PI\n"},
		{"overflow-id", "w2147483648 = PI\n"}, {"negative-id", "w-2 = PI\n"}, {"duplicate", "w2 = PI\nw2 = PI\n"},
		{"unknown-type", "w2 = Magic(0)\n"}, {"PI-arguments", "w2 = PI(0)\n"}, {"PO-arguments", "w2 = PO(0,1)\n"},
		{"LUT4-arity", "w2 = Lut4(0,0,0) [0]\n"}, {"LUT6-arity", "w2 = Lut6(0,0,0,0) [0]\n"},
		{"truth-empty", "w2 = Lut4(0,0,0,0) []\n"}, {"truth-invalid", "w2 = Lut4(0,0,0,0) [QQ]\n"},
		{"truth-overflow", "w2 = Lut6(0,0,0,0,0,0) [10000000000000000]\n"}, {"missing-close", "w2 = Bar(0\n"},
		{"truth-no-bracket", "w2 = Lut4(0,0,0,0) 8888\n"}, {"truth-no-close", "w2 = Lut4(0,0,0,0) [8888\n"},
		{"box-arity", "w2 = Box(0,0,0)\n"}, {"box-too-many", "w2 = Box(0,0,0,0,0,0)\n"},
		{"delay-empty", "w2 = Del() [3]\n"}, {"delay-missing", "w2 = Del(0)\n"}, {"delay-invalid", "w2 = Del(0) [x]\n"},
		{"delay-overflow", "w2 = Del(0) [2147483648]\n"}, {"missing-parent", "w2 = Sel(w99)\n"},
		{"select-box", "w2 = Box(0,0,0,0)\nw3 = Sel(w2)\n"}, {"select-seq", "w2 = Seq(0)\nw3 = Sel(w2)\n"},
		{"select-select", "w2 = Del(0) [1]\nw3 = Sel(w2)\nw4 = Sel(w3)\n"},
		{"undefined-LUT", "w2 = Lut4(w99,0,0,0) [8]\n"}, {"undefined-bar", "w2 = Bar(w99)\n"},
		{"undefined-output", "w2 = PO(w99)\n"}, {"undefined-seq", "w2 = Seq(w99)\n"},
		{"undefined-delay-input", "w2 = Del(w99) [3]\n"}, {"missing-non-LUT", "w2 = PO(-)\n"},
		{"self-LUT", "w2 = Lut4(w2,0,0,0) [8]\n"},
		{"selected-delay-reference", "w2 = Del(0) [3]\nw3 = Sel(w2)\nw4 = PO(w2)\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := gigReviewWrite(t, tc.text)
			p, report := gigReviewRead(t, file)
			if p != nil {
				Gia_ManStop(p)
				t.Fatal("malformed record silently accepted")
			}
			if !strings.Contains(report, file) || !strings.Contains(report, "Cannot read GIG file") {
				t.Fatalf("missing file/error context: %s", report)
			}
		})
	}
	for _, path := range []string{filepath.Join(t.TempDir(), "missing.gig"), t.TempDir()} {
		p, report := gigReviewRead(t, path)
		if p != nil {
			Gia_ManStop(p)
			t.Fatal("unreadable input returned graph")
		}
		if report == "" {
			t.Fatal("missing read error")
		}
	}
}

func TestGigSafeWhitespaceSparseAndForwardDelayExtensions(t *testing.T) {
	exe := gigReviewCompile(t)
	cases := []struct {
		name, text, sourceFailure string
		inputs                    int
		want                      func(int) []bool
	}{
		{"sparse", "w1000000 = PI\nw2000000 = Lut4(w1000000,1,-,-) [A]\nw2147483647 = PO(w2000000)\n", "Vec_StrWriteEntry", 1, func(in int) []bool { return []bool{in&1 != 0} }},
		{"whitespace", "\n\tw2\t =\t PI \r\n\n w3 = PO( w2 )\r\n", "", 1, func(in int) []bool { return []bool{in&1 != 0} }},
		{"forward-delay", "w2 = PI\nw3 = Sel(w4)\nw4 = Del(w2) [2]\nw5 = PO(w3)\n", "Gls_ManParse", 2, func(in int) []bool { return []bool{in&1 != 0, in&2 != 0} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := gigReviewWrite(t, tc.text)
			if tc.sourceFailure != "" {
				out, err := exec.Command(exe, "--raw", path).CombinedOutput()
				if err == nil || !strings.Contains(string(out), tc.sourceFailure) {
					t.Fatalf("source precondition changed: %v %s", err, out)
				}
			}
			p, report := gigReviewRead(t, path)
			if p == nil {
				t.Fatalf("existing safe Go input rejected: %s", report)
			}
			defer Gia_ManStop(p)
			if Gia_ManCiNum(p) != int32(tc.inputs) {
				t.Fatal("input count changed")
			}
			for input := 0; input < 1<<tc.inputs; input++ {
				values := cofReviewValues(p, input)
				for out, want := range tc.want(input) {
					if values[Gia_ManCo(p, int32(out)).id] != want {
						t.Fatalf("input%d output%d", input, out)
					}
				}
			}
		})
	}
}

func TestGigConstructionOwnsGraphAndPreservesParsedRecords(t *testing.T) {
	var tc gigReviewCase
	for _, candidate := range gigReviewCases() {
		if candidate.name == "mixed-cuts" {
			tc = candidate
			break
		}
	}
	path := gigReviewWrite(t, tc.text)
	nodes, err := gigParse(path)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := append([]gigNode(nil), nodes...)
	for i := range snapshot {
		snapshot[i].args = append([]int32(nil), nodes[i].args...)
	}
	p, err := gigConstruct(nodes, path)
	if err != nil {
		t.Fatal(err)
	}
	q, err := gigConstruct(nodes, path)
	if err != nil {
		t.Fatal(err)
	}
	defer Gia_ManStop(q)
	if !reflect.DeepEqual(snapshot, nodes) {
		t.Fatal("construction modified caller's parsed records")
	}
	p.SetIdentity("changed", "")
	Gia_ManCo(p, 0).fCompl0 ^= 1
	for i := range nodes {
		if len(nodes[i].args) > 0 {
			nodes[i].args[0] = 0
		}
		nodes[i].truth = 0
	}
	Gia_ManStop(p)
	if q.pName != path || q.PSpec != path || q.nRegs != 0 || q.PManTime != nil {
		t.Fatal("identity/cut metadata changed")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	for input := 0; input < 1<<tc.inputs; input++ {
		values := cofReviewValues(q, input)
		for out, want := range tc.want(input) {
			if values[Gia_ManCo(q, int32(out)).id] != want {
				t.Fatalf("result not independently owned input%d output%d", input, out)
			}
		}
	}
}
