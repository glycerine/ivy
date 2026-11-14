package lpk

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestLpkLutSizing(t *testing.T) {
	for _, tc := range []struct{ luts, k, wantVars, wantLuts int32 }{
		{1, 4, 4, 1}, {2, 4, 7, 2}, {3, 6, 16, 3}, {5, 3, 11, 5},
	} {
		if got := Lpk_LutNumVars(tc.luts, tc.k); got != tc.wantVars {
			t.Errorf("Lpk_LutNumVars(%d,%d)=%d want %d", tc.luts, tc.k, got, tc.wantVars)
		}
		if got := Lpk_LutNumLuts(tc.wantVars, tc.k); got != tc.wantLuts {
			t.Errorf("Lpk_LutNumLuts(%d,%d)=%d want %d", tc.wantVars, tc.k, got, tc.wantLuts)
		}
	}
}

func TestLpkSupportDelayAndVariables(t *testing.T) {
	delays := []int32{2, 5, 3, 9}
	if got := Lpk_SuppDelay(0b1010, delays); got != 10 {
		t.Fatalf("support delay=%d want 10", got)
	}
	var vars [4]byte
	if got := Lpk_SuppToVars(0b10101, vars[:]); got != 3 {
		t.Fatalf("support variable count=%d want 3", got)
	}
	if want := [4]byte{0, 2, 4, 0}; vars != want {
		t.Fatalf("support variables=%v want %v", vars, want)
	}
}

func TestLpkFunctionTruthStorage(t *testing.T) {
	fun := Lpk_FunAlloc(6)
	if fun == nil {
		t.Fatal("Lpk_FunAlloc(6) returned nil")
	}
	base, workspace0, workspace1 := Lpk_FunTruth(fun, 0), Lpk_FunTruth(fun, 1), Lpk_FunTruth(fun, 2)
	if len(base) != 2 || len(workspace0) != 2 || len(workspace1) != 2 {
		t.Fatalf("truth table word counts = %d/%d/%d, want 2 each", len(base), len(workspace0), len(workspace1))
	}
	base[0] = 0x1234
	workspace0[0] = 0x5678
	if base[0] != 0x1234 || workspace0[0] != 0x5678 || workspace1[0] != 0 {
		t.Fatalf("truth table workspaces alias: %x/%x/%x", base[0], workspace0[0], workspace1[0])
	}
}

func TestLpkFunctionCreateAndDuplicatePreserveSourceMetadata(t *testing.T) {
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	defer abc.Abc_NtkDelete(network)
	x, y := abc.Abc_NtkCreatePi(network), abc.Abc_NtkCreatePi(network)
	x.Level, y.Level = 2, 5
	leaves := vec.Vec_PtrAlloc(4)
	defer vec.Vec_PtrFree(leaves)
	vec.Vec_PtrPush(leaves, x)
	vec.Vec_PtrPush(leaves, y)

	fun := Lpk_FunCreate(network, leaves, []uint32{0x66666666}, 3, 4, 8)
	if fun == nil {
		t.Fatal("Lpk_FunCreate returned nil")
	}
	if fun.Id != 2 || fun.nVars != 2 || fun.uSupp != 3 || fun.pFanins[0] != 0 || fun.pFanins[1] != 1 || fun.pDelays[0] != 2 || fun.pDelays[1] != 5 {
		t.Fatalf("created function metadata id/vars/support/fanins/delays = %d/%d/%b/%v/%v", fun.Id, fun.nVars, fun.uSupp, fun.pFanins[:2], fun.pDelays[:2])
	}

	dup := Lpk_FunDup(fun, []uint32{0x88888888})
	if dup == nil {
		t.Fatal("Lpk_FunDup returned nil")
	}
	if dup.Id != 3 || dup.nVars != fun.nVars || dup.nLutK != fun.nLutK || dup.nAreaLim != fun.nAreaLim || dup.nDelayLim != fun.nDelayLim || dup.pFanins != fun.pFanins || dup.pDelays != fun.pDelays {
		t.Fatalf("duplicated metadata differs: %+v from %+v", dup, fun)
	}
	if Lpk_FunTruth(fun, 0)[0] != 0x66666666 || Lpk_FunTruth(dup, 0)[0] != 0x88888888 {
		t.Fatal("duplicate did not own and preserve its supplied truth table")
	}
	Lpk_FunFree(dup)
	Lpk_FunFree(fun)
}

func TestLpkManagerTruthWorkspaceLayoutAndCleanup(t *testing.T) {
	pars := &Lpk_Par_t{nLutsMax: 2, nVarsMax: 5, nLutSize: 3}
	man := Lpk_ManStart(pars)
	if man == nil {
		t.Fatal("Lpk_ManStart returned nil")
	}
	wordCount := int(kit.Kit_TruthWordNum(pars.nVarsMax))
	for level := range man.ppTruths {
		for index := 0; index < 1<<level; index++ {
			if got := len(man.ppTruths[level][index]); got != wordCount {
				t.Fatalf("truth workspace [%d][%d] has %d words, want %d", level, index, got, wordCount)
			}
		}
	}
	first, second := man.ppTruths[2][0], man.ppTruths[2][1]
	first[0] = 0x12345678
	if second[0] != 0 {
		t.Fatal("truth workspace entries overlap")
	}
	Lpk_ManStop(man)
	if man.vTtNodes != nil || man.vTtElems != nil || man.vSets[0] != nil || man.pIfMan != nil || man.pDsdMan != nil {
		t.Fatal("manager stop did not release owned workspace references")
	}
}

func TestLpkMapPrimeReusesTruthGraphScratch(t *testing.T) {
	pars := &Lpk_Par_t{nLutsMax: 1, nVarsMax: 2, nLutSize: 2}
	man := Lpk_ManStart(pars)
	if man == nil {
		t.Fatal("Lpk_ManStart returned nil")
	}
	defer Lpk_ManStop(man)
	Lpk_IfManStart(man)
	leaves := []*If_Obj_t{ifpkg.If_ManCreateCi(man.pIfMan), ifpkg.If_ManCreateCi(man.pIfMan)}
	if result := Lpk_MapPrime(man, []uint32{0x88888888}, 2, leaves); result == nil {
		t.Fatal("Lpk_MapPrime returned nil for x0 AND x1")
	}
	if len(man.graphMemory) == 0 {
		t.Fatal("Lpk_MapPrime did not retain its reusable truth-graph scratch")
	}
}

func TestLpkFunctionSupportMinimizeCompactsFanins(t *testing.T) {
	fun := Lpk_FunAlloc(3)
	fun.nVars = 3
	fun.pFanins = [16]byte{4, 7, 9}
	fun.pDelays = [16]int32{2, 3, 8}
	// The function is x2, whose truth table repeats 00001111.
	Lpk_FunTruth(fun, 0)[0] = 0xF0F0F0F0
	fun.uSupp = kit.Kit_TruthSupport(Lpk_FunTruth(fun, 0), 3)

	if got := Lpk_FunSuppMinimize(fun); got != 1 {
		t.Fatalf("support minimization = %d, want 1", got)
	}
	if fun.nVars != 1 || fun.uSupp != 1 {
		t.Fatalf("minimized vars/support = %d/%b, want 1/1", fun.nVars, fun.uSupp)
	}
	if got := Lpk_FunTruth(fun, 0)[0]; got != 0xAAAAAAAA {
		t.Fatalf("minimized truth = %#08x, want variable truth %#08x", got, uint32(0xAAAAAAAA))
	}
	if fun.pFanins[0] != 9 || fun.pDelays[0] != 8 {
		t.Fatalf("retained fanin metadata = %d/%d, want 9/8", fun.pFanins[0], fun.pDelays[0])
	}
}

func TestLpkFunctionCofactorSupports(t *testing.T) {
	fun := Lpk_FunAlloc(2)
	fun.nVars = 2
	// XOR: each cofactor of x0 is x1.
	Lpk_FunTruth(fun, 0)[0] = 0x66666666
	Lpk_FunComputeCofSupps(fun)
	if fun.fSupports == 0 || fun.puSupps[0] != 2 || fun.puSupps[1] != 2 {
		t.Fatalf("x0 cofactor supports = %b/%b (computed=%d), want 10/10", fun.puSupps[0], fun.puSupps[1], fun.fSupports)
	}
}

func TestLpkFunctionCofactorSupportsRejectsRepeatedComputation(t *testing.T) {
	fun := Lpk_FunAlloc(1)
	fun.nVars = 1
	Lpk_FunComputeCofSupps(fun)
	defer func() {
		if recover() == nil {
			t.Fatal("recomputing cached cofactor supports did not trigger the source assertion")
		}
	}()
	Lpk_FunComputeCofSupps(fun)
}

func TestLpkComputeSetsEnumeratesAndBoundaries(t *testing.T) {
	truth := []uint32{0x88888888} // a AND b
	network := kit.Kit_DsdDecompose(truth, 2)
	defer kit.Kit_DsdNtkFree(network)
	sets := vec.Vec_IntAlloc(8)
	defer vec.Vec_IntFree(sets)
	if support := Lpk_ComputeSets(network, sets); support != 3 {
		t.Fatalf("root support = %02b, want both variables", support)
	}
	want := []int32{3 << 16, (1 << 16) | 2, (2 << 16) | 1, 3}
	if got := vec.Vec_IntArray(sets); len(got) != len(want) {
		t.Fatalf("set count = %d, want %d (%v)", len(got), len(want), got)
	} else {
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("set %d = %#x, want %#x", i, got[i], want[i])
			}
		}
	}
}

func TestLpkPackTruthWords(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []uint32
		want []uint64
	}{
		{"empty", nil, []uint64{}},
		{"one half", []uint32{0x89abcdef}, []uint64{0x0000000089abcdef}},
		{"paired halves", []uint32{0x01234567, 0x89abcdef}, []uint64{0x89abcdef01234567}},
		{"odd tail", []uint32{1, 2, 3}, []uint64{0x0000000200000001, 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := lpkPackTruthWords(tc.in); !equalUint64s(got, tc.want) {
				t.Fatalf("lpkPackTruthWords(%08x) = %016x, want %016x", tc.in, got, tc.want)
			}
		})
	}
}

func equalUint64s(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLpkMuxSplitPreservesFunctionForBothPolarities(t *testing.T) {
	for _, polarity := range []int32{0, 1} {
		// f = mux(x0, x1 XOR x3, x1 AND x2). Each cofactor uses two
		// variables, leaving a slot for the decomposed component.
		truth := make([]uint32, kit.Kit_TruthWordNum(4))
		for assignment := int32(0); assignment < 16; assignment++ {
			x0, x1 := assignment&1 != 0, assignment&2 != 0
			x2, x3 := assignment&4 != 0, assignment&8 != 0
			value := x1 && x2
			if x0 {
				value = x1 != x3
			}
			if value {
				kit.Kit_TruthSetBit(truth, assignment)
			}
		}
		leaves := vec.Vec_PtrAlloc(5)
		for i := 0; i < 4; i++ {
			vec.Vec_PtrPush(leaves, int32(i))
		}
		fun := Lpk_FunAlloc(4)
		fun.Id, fun.vNodes, fun.nVars = 4, leaves, 4
		fun.nLutK, fun.nAreaLim, fun.nDelayLim = 3, 4, 5
		fun.uSupp = kit.Kit_TruthSupport(truth, 4)
		copy(Lpk_FunTruth(fun, 0), truth)
		for i := 0; i < 4; i++ {
			fun.pFanins[i] = byte(i)
		}
		vec.Vec_PtrPush(leaves, fun)
		child := Lpk_MuxSplit(nil, fun, 0, polarity)
		if child == nil {
			t.Fatalf("polarity %d: split failed", polarity)
		}
		for assignment := int32(0); assignment < 16; assignment++ {
			childInputs := int32(0)
			for i := int32(0); i < int32(child.nVars); i++ {
				if assignment&(1<<fun.pFanins[i]) != 0 {
					childInputs |= 1 << i
				}
			}
			childValue := kit.Kit_TruthHasBit(Lpk_FunTruth(child, 0), childInputs) != 0
			parentInputs := int32(0)
			for i := int32(0); i < int32(fun.nVars); i++ {
				fanin := fun.pFanins[i]
				value := assignment&(1<<fanin) != 0
				if uint32(fanin) == child.Id {
					value = childValue
				}
				if value {
					parentInputs |= 1 << i
				}
			}
			got := kit.Kit_TruthHasBit(Lpk_FunTruth(fun, 0), parentInputs) != 0
			want := kit.Kit_TruthHasBit(truth, assignment) != 0
			if got != want {
				t.Fatalf("polarity %d assignment %04b: got %t, want %t", polarity, assignment, got, want)
			}
		}
		Lpk_FunFree(child)
		Lpk_FunFree(fun)
		vec.Vec_PtrFree(leaves)
	}
}

func TestLpkImplementRecBuildsEachComponentOnce(t *testing.T) {
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_AIG, 1)
	defer abc.Abc_NtkDelete(network)
	x, y := abc.Abc_NtkCreatePi(network), abc.Abc_NtkCreatePi(network)
	leaves := vec.Vec_PtrAlloc(4)
	vec.Vec_PtrPush(leaves, x)
	vec.Vec_PtrPush(leaves, y)

	child := Lpk_FunAlloc(2)
	child.Id, child.vNodes, child.nVars, child.nLutK = 3, leaves, 2, 3
	child.pFanins[0], child.pFanins[1] = 0, 1
	Lpk_FunTruth(child, 0)[0] = 0x66666666 // x XOR y

	root := Lpk_FunAlloc(2)
	root.Id, root.vNodes, root.nVars, root.nLutK = 2, leaves, 2, 3
	root.pFanins[0], root.pFanins[1] = 0, byte(child.Id)
	Lpk_FunTruth(root, 0)[0] = 0x88888888 // x AND child
	vec.Vec_PtrPush(leaves, root)
	vec.Vec_PtrPush(leaves, child)

	result := Lpk_Implement(nil, network, leaves, 2)
	if result == nil {
		t.Fatal("recursive implementation returned nil")
	}
	if vec.Vec_PtrSize(leaves) != 2 {
		t.Fatalf("leaf vector size after implementation = %d, want 2", vec.Vec_PtrSize(leaves))
	}
	if got := abc.Abc_NtkNodeNum(network); got != 2 {
		t.Fatalf("created node count = %d, want child and root exactly once", got)
	}
	if got := len(result.Fanins()); got != 2 {
		t.Fatalf("root fanin count = %d, want 2", got)
	}
	childNode := result.Fanins()[1].Object
	if got := len(childNode.Fanins()); got != 2 {
		t.Fatalf("decomposition fanin count = %d, want 2", got)
	}
	if result.Level != 2 || childNode.Level != 1 {
		t.Fatalf("component levels root/child = %d/%d, want 2/1", result.Level, childNode.Level)
	}
	vec.Vec_PtrFree(leaves)
}

func TestLpkDecomposeCleanDropsTemporaryFunctions(t *testing.T) {
	leaves := vec.Vec_PtrAlloc(4)
	vec.Vec_PtrPush(leaves, "leaf 0")
	vec.Vec_PtrPush(leaves, "leaf 1")
	for i := 0; i < 2; i++ {
		fun := Lpk_FunAlloc(1)
		fun.Id, fun.vNodes, fun.nVars = uint32(2+i), leaves, 1
		vec.Vec_PtrPush(leaves, fun)
	}
	Lpk_DecomposeClean(leaves, 2)
	if got := vec.Vec_PtrSize(leaves); got != 2 {
		t.Fatalf("leaf vector size after cleanup = %d, want 2", got)
	}
	if vec.Vec_PtrEntry(leaves, 0) != "leaf 0" || vec.Vec_PtrEntry(leaves, 1) != "leaf 1" {
		t.Fatal("cleanup changed preexisting leaves")
	}
	vec.Vec_PtrFree(leaves)
}

func TestLpkBestCofactorMatchesSourceVector(t *testing.T) {
	data, err := os.ReadFile("test_vectors/cofactor_choice.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name             string   `json:"name"`
			NVars            int32    `json:"n_vars"`
			LutSize          int32    `json:"lut_size"`
			TruthWords       []string `json:"truth_words"`
			ExpectedVariable int32    `json:"expected_variable"`
			ExpectedSupports []int32  `json:"expected_cofactor_supports"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, tc := range fixture.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			truth := make([]uint32, len(tc.TruthWords))
			for i, word := range tc.TruthWords {
				value, err := strconv.ParseUint(word, 16, 32)
				if err != nil {
					t.Fatal(err)
				}
				truth[i] = uint32(value)
			}
			p := &Lpk_Man_t{pPars: &Lpk_Par_t{nLutSize: tc.LutSize}}
			cof0 := make([]uint32, len(truth))
			cof1 := make([]uint32, len(truth))
			got := Lpk_MapTreeBestCofVar(p, truth, tc.NVars, cof0, cof1)
			if got != tc.ExpectedVariable {
				t.Fatalf("best cofactor variable = %d, want %d", got, tc.ExpectedVariable)
			}
			gotSupports := []int32{kit.Kit_TruthSupportSize(cof0, tc.NVars), kit.Kit_TruthSupportSize(cof1, tc.NVars)}
			if len(gotSupports) != len(tc.ExpectedSupports) {
				t.Fatalf("cofactor support sizes = %v, want %v", gotSupports, tc.ExpectedSupports)
			}
			for i := range gotSupports {
				if gotSupports[i] != tc.ExpectedSupports[i] {
					t.Fatalf("cofactor support sizes = %v, want %v", gotSupports, tc.ExpectedSupports)
				}
			}
		})
	}
}

func TestAbcNtkLutMinDecomposePreservesComplementedEdges(t *testing.T) {
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_AIG, 1)
	defer abc.Abc_NtkDelete(network)
	x, y := abc.Abc_NtkCreatePi(network), abc.Abc_NtkCreatePi(network)
	leaves := vec.Vec_PtrAlloc(2)
	defer vec.Vec_PtrFree(leaves)
	vec.Vec_PtrPush(leaves, x)
	vec.Vec_PtrPush(leaves, y)
	truth := []uint32{0x66666666} // x XOR y; its AIG factoring uses complemented edges.
	root := Abc_NtkLutMinDecompose(network, leaves, truth, 2, 2)
	if root == nil {
		t.Fatal("Abc_NtkLutMinDecompose rejected the two-input XOR")
	}
	for assignment := int32(0); assignment < 4; assignment++ {
		inputs := map[*abc.Object]bool{x: assignment&1 != 0, y: assignment&2 != 0}
		got, err := network.Evaluate(root.Signal(), inputs)
		if err != nil {
			t.Fatalf("evaluate assignment %02b: %v", assignment, err)
		}
		want := kit.Kit_TruthHasBit(truth, assignment) != 0
		if got != want {
			t.Fatalf("decomposition at assignment %02b = %t, want %t", assignment, got, want)
		}
	}
}
