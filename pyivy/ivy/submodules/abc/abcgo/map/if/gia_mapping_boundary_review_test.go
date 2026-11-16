package if_

import (
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func mappingConfigReviewManager(t *testing.T, name string, withAnd bool) (*If_Man_t, *If_Obj_t) {
	t.Helper()
	manager := If_DsdManAlloc(2, 2)
	truth := []word{0x8888888888888888}
	perm := make([]byte, 2)
	lit := If_DsdManCompute(manager, truth, 2, perm, "")
	Id_DsdManTuneStr1(manager, "{ab}", 0, 0)
	p := If_ManStart(&If_Par_t{NLutSize: 4, NCutsMax: 4, FTruth: 1, FUseDsd: 1, FUseDsdTune: 1, FDeriveLuts: 1})
	t.Cleanup(func() { If_ManStop(p); If_DsdManFree(manager, 0) })
	p.pIfDsdMan, p.pName = manager, name
	a, b := If_ManCreateCi(p), If_ManCreateCi(p)
	var root *If_Obj_t
	if withAnd {
		root = If_ManCreateAnd(p, a, b)
		id := vec.Vec_MemHashInsert(p.vTtMem[2], truth)
		root.CutBest = If_Cut_t{nLimit: 4, nLeaves: 2, pLeaves: []int32{a.Id, b.Id}, iCutFunc: int32(id * 2)}
		vec.Vec_IntSetEntry(p.vTtDsds[2], id, lit)
		vec.Vec_StrFillExtra(p.vTtPerms[2], (id+1)*6, 0)
		for i, pin := range perm {
			vec.Vec_StrWriteEntry(p.vTtPerms[2], id*6+i, pin)
		}
		If_ManCreateCo(p, root)
		If_ManCreateCo(p, If_Not(root))
	}
	If_ManCreateCo(p, a)
	If_ManCreateCo(p, If_ManConst1(p))
	return p, root
}

func TestMappingConfigurationFileOutput(t *testing.T) {
	for _, tc := range []struct{ label, name, path string }{
		{"plain", "plain.aig", "plain_configs.txt"},
		{"multiple_dots", "multiple.part.aig", "multiple.part_configs.txt"},
		{"no_extension", "plain", "plain_configs.txt"},
		{"quote", "quo\"ted.aig", "quo\"ted_configs.txt"},
		{"backslash", "back\\slash.aig", "back\\slash_configs.txt"},
		{"unnamed", "", "nameless__configs.txt"},
	} {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/fail%t", tc.label, fail), func(t *testing.T) {
				if runtime.GOOS == "windows" && strings.ContainsAny(tc.name, "\"\\") {
					t.Skip("literal filename characters are not supported on Windows")
				}
				t.Chdir(t.TempDir())
				if fail {
					if err := os.Mkdir(tc.path, 0700); err != nil {
						t.Fatal(err)
					}
				} else if err := os.WriteFile(tc.path, []byte(strings.Repeat("old contents", 1000)), 0600); err != nil {
					t.Fatal(err)
				}
				p, root := mappingConfigReviewManager(t, tc.name, true)
				var dst *Gia_Man_t
				output := dsdReportCapture(t, func() { dst = Gia_ManFromIfLogic(p) })
				if dst == nil {
					t.Fatal("file output failure discarded valid mapping")
				}
				defer Gia_ManStop(dst)
				expected := fmt.Sprintf("Finished dumping configs into file \"%s\".\n", tc.path)
				if fail {
					expected = fmt.Sprintf("Cannot open file \"%s\".\n", tc.path)
				}
				if output != expected {
					t.Errorf("report%q want%q", output, expected)
				}
				cell, configs := dst.CellConfiguration()
				data := vec.Vec_IntArray(configs)
				if cell != "{ab}" || len(data) != 3 || data[0] != 1 || data[1] != 1 {
					t.Fatalf("configuration header%v cell%q", data, cell)
				}
				if !fail {
					// Independently format the attached bits using the source low-bit
					// order, sign spacing, root ID and final output-polarity field.
					packed := uint32(data[2])
					var want strings.Builder
					fmt.Fprintf(&want, "%d ", root.iCopy>>1)
					for bit := 0; bit < 4; bit++ {
						fmt.Fprintf(&want, "%d", packed>>bit&1)
					}
					want.WriteString("  ")
					for pin := 0; pin < 2; pin++ {
						for bit := 0; bit < 3; bit++ {
							fmt.Fprintf(&want, "%d", packed>>(4+pin*3+bit)&1)
							if bit == 0 {
								want.WriteByte(' ')
							}
						}
						want.WriteString("  ")
					}
					fmt.Fprintf(&want, "%d\n", packed>>10&1)
					actual, err := os.ReadFile(tc.path)
					if err != nil || string(actual) != want.String() {
						t.Fatalf("file%q err%v want%q", actual, err, want.String())
					}
				}
				for x := 0; x < 4; x++ {
					input := []bool{x&1 != 0, x&2 != 0}
					got, err := dst.Evaluate(input)
					want := []bool{x == 3, x != 3, input[0], true}
					if err != nil || !slices.Equal(got, want) {
						t.Fatalf("assignment%d got%v err%v want%v", x, got, err, want)
					}
				}
			})
		}
	}
}

func TestMappingConfigurationEmptyFile(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("empty_configs.txt", []byte("old contents"), 0600); err != nil {
		t.Fatal(err)
	}
	p, _ := mappingConfigReviewManager(t, "empty.aig", false)
	var dst *Gia_Man_t
	output := dsdReportCapture(t, func() { dst = Gia_ManFromIfLogic(p) })
	if dst == nil {
		t.Fatal("terminal-only configuration mapping failed")
	}
	defer Gia_ManStop(dst)
	if output != "Finished dumping configs into file \"empty_configs.txt\".\n" {
		t.Fatalf("report%q", output)
	}
	_, configs := dst.CellConfiguration()
	if !slices.Equal(vec.Vec_IntArray(configs), []int32{0, 1}) {
		t.Fatalf("empty configs%v", vec.Vec_IntArray(configs))
	}
	if data, err := os.ReadFile("empty_configs.txt"); err != nil || len(data) != 0 {
		t.Fatalf("file%q err%v", data, err)
	}
}

func TestMappingRecordArrayGrowsBeyondIfGraph(t *testing.T) {
	p := If_ManStart(&If_Par_t{NLutSize: 3, NCutsMax: 4, FTruth: 1, FDeriveLuts: 1})
	defer If_ManStop(p)
	inputs := make([]*If_Obj_t, 4)
	for i := range inputs {
		inputs[i] = If_ManCreateCi(p)
	}
	shared := If_ManCreateAnd(p, inputs[0], inputs[1])
	roots := []*If_Obj_t{If_ManCreateAnd(p, shared, inputs[2]), If_ManCreateAnd(p, shared, inputs[3])}
	shared.nRefs = 0
	shared.iCopy = -99
	id := vec.Vec_MemHashInsert(p.vTtMem[3], []word{0x8080808080808080})
	for i, root := range roots {
		root.CutBest = If_Cut_t{nLimit: 3, nLeaves: 3, pLeaves: []int32{inputs[0].Id, inputs[1].Id, inputs[i+2].Id}, iCutFunc: int32(id * 2)}
		If_ManCreateCo(p, root)
	}
	dst := Gia_ManFromIfLogic(p)
	if dst == nil {
		t.Fatal("shared-cone reconstruction failed")
	}
	defer Gia_ManStop(dst)
	if Gia_ManObjNum(dst) <= If_ManObjNum(p) {
		t.Fatalf("fixture did not grow: GIA%d IF%d", Gia_ManObjNum(dst), If_ManObjNum(p))
	}
	if shared.iCopy != -99 {
		t.Fatal("bypassed shared node reconstructed separately")
	}
	data := vec.Vec_IntArray(dst.VMapping)
	for _, root := range roots {
		node := root.iCopy >> 1
		at := data[node]
		if at < Gia_ManObjNum(dst) || at+4 >= int32(len(data)) || data[at] != 3 || data[at+4] != node {
			t.Fatalf("root%d mapping%v", node, data)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(dst); i++ {
		if id := gia.Gia_ObjId(dst, Gia_ManCo(dst, i)); data[id] != 0 {
			t.Fatalf("CO%d has an offset", id)
		}
	}
	for x := 0; x < 16; x++ {
		input := make([]bool, 4)
		for i := range input {
			input[i] = x>>i&1 != 0
		}
		want := []bool{input[0] && input[1] && input[2], input[0] && input[1] && input[3]}
		got, err := dst.Evaluate(input)
		if err != nil || !slices.Equal(got, want) {
			t.Fatalf("assignment%d got%v err%v want%v", x, got, err, want)
		}
	}
}
