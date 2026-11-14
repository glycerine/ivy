package abci

import (
	"testing"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/fpga"
)

func fpgaTestAig(t *testing.T) (*baseabc.Network, []*baseabc.Object, baseabc.Edge) {
	t.Helper()
	n := baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_STRASH, baseabc.ABC_FUNC_AIG, 1)
	manager, ok := n.FunctionManager.(*baseabc.Abc_Aig_t)
	if !ok {
		t.Fatal("strash network has no local AIG manager")
	}
	inputs := []*baseabc.Object{baseabc.Abc_NtkCreatePi(n), baseabc.Abc_NtkCreatePi(n), baseabc.Abc_NtkCreatePi(n)}
	for i, input := range inputs {
		baseabc.Abc_ObjAssignName(input, []string{"a", "b", "c"}[i], "")
	}
	root := baseabc.Abc_AigAnd(manager, inputs[0], baseabc.Abc_ObjNot(inputs[1]))
	root = baseabc.Abc_AigAnd(manager, root, inputs[2])
	po := baseabc.Abc_NtkCreatePo(n)
	baseabc.Abc_ObjAssignName(po, "out", "")
	baseabc.Abc_ObjAddFanin(po, root.Not())
	baseabc.Abc_NtkOrderCisCos(n)
	return n, inputs, root.Not()
}

func TestAbcNtkFpgaPreservesLogicAndComplements(t *testing.T) {
	for _, switching := range []int32{0, 1} {
		n, inputs, original := fpgaTestAig(t)
		mapped := Abc_NtkFpga(n, -1, 1, switching, 0, 0)
		if mapped == nil {
			t.Fatalf("FPGA mapping returned nil with switching=%d", switching)
		}
		if mapped.Type != uint8(baseabc.ABC_NTK_LOGIC) || mapped.Function != uint8(baseabc.ABC_FUNC_BDD) {
			t.Fatalf("mapped network type/function = %d/%d", mapped.Type, mapped.Function)
		}
		if baseabc.Abc_NtkCheck(mapped) == 0 {
			t.Fatalf("mapped network failed the source network check")
		}
		po := baseabc.Abc_NtkPo(mapped, 0)
		mappedDriver := baseabc.Abc_ObjFanin0(po)
		mappedInputs := []*baseabc.Object{baseabc.Abc_NtkPi(mapped, 0), baseabc.Abc_NtkPi(mapped, 1), baseabc.Abc_NtkPi(mapped, 2)}
		for assignment := 0; assignment < 1<<len(inputs); assignment++ {
			values := make(map[*baseabc.Object]bool, len(inputs))
			mappedValues := make(map[*baseabc.Object]bool, len(inputs))
			for i, input := range inputs {
				value := assignment&(1<<i) != 0
				values[input] = value
				mappedValues[mappedInputs[i]] = value
			}
			want, err := n.Evaluate(original, values)
			if err != nil {
				t.Fatal(err)
			}
			got, err := mapped.Evaluate(mappedDriver.Signal(), mappedValues)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("switching=%d assignment=%03b: mapped %t, want %t", switching, assignment, got, want)
			}
		}
		baseabc.Abc_NtkDelete(mapped)
		baseabc.Abc_NtkDelete(n)
	}
}

func TestAbcNtkToFpgaMapsChoiceChain(t *testing.T) {
	n, inputs, _ := fpgaTestAig(t)
	manager, ok := n.FunctionManager.(*baseabc.Abc_Aig_t)
	if !ok {
		t.Fatal("strash network has no local AIG manager")
	}
	alternative := baseabc.Abc_AigAnd(manager, inputs[0], baseabc.Abc_AigAnd(manager, baseabc.Abc_ObjNot(inputs[1]), inputs[2]))
	root := baseabc.Abc_ObjFanin0(baseabc.Abc_NtkPo(n, 0))
	root.Choice = alternative.Object
	fpgaMan := Abc_NtkToFpga(n, 0, nil, 0, 0)
	if fpgaMan == nil {
		t.Fatal("FPGA import returned nil")
	}
	defer fpga.Fpga_ManFree(fpgaMan)
	if fpga.Fpga_Mapping(fpgaMan) == 0 {
		t.Fatalf("FPGA mapping failed: %v", fpgaMan.LastError)
	}
	if fpgaMan.Choices == 0 {
		t.Fatal("equivalent choice alternative was not recorded")
	}
	baseabc.Abc_NtkDelete(n)
}

func TestFpgaSwitchingRandomWordMatchesCSequence(t *testing.T) {
	rng := newGlibcRand(1)
	words := make([]uint32, 2)
	abcFpgaSetRandom(words, rng)
	if words[0] != 0xb100f869 {
		t.Fatalf("first SIM_RANDOM_UNSIGNED word = %#08x, want 0xb100f869", words[0])
	}
	if words[1] == words[0] {
		t.Fatalf("consecutive simulation words unexpectedly match: %#08x", words[0])
	}
}

func TestAbcNtkFpgaReconstructsConstantAndBufferOutputs(t *testing.T) {
	for _, constant := range []bool{false, true} {
		n := baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_STRASH, baseabc.ABC_FUNC_AIG, 1)
		input := baseabc.Abc_NtkCreatePi(n)
		baseabc.Abc_ObjAssignName(input, "input", "")
		po := baseabc.Abc_NtkCreatePo(n)
		baseabc.Abc_ObjAssignName(po, "output", "")
		driver := input.Signal()
		if constant {
			driver = baseabc.Abc_AigConst1(n)
		}
		baseabc.Abc_ObjAddFanin(po, driver)
		baseabc.Abc_NtkOrderCisCos(n)
		mapped := Abc_NtkFpga(n, -1, 0, 0, 0, 0)
		if mapped == nil {
			t.Fatalf("mapping returned nil for constant=%t", constant)
		}
		defer baseabc.Abc_NtkDelete(mapped)
		out := baseabc.Abc_ObjFanin0(baseabc.Abc_NtkPo(mapped, 0))
		for _, value := range []bool{false, true} {
			got, err := mapped.Evaluate(out.Signal(), map[*baseabc.Object]bool{baseabc.Abc_NtkPi(mapped, 0): value})
			if err != nil {
				t.Fatal(err)
			}
			want := value
			if constant {
				want = true
			}
			if got != want {
				t.Fatalf("constant=%t input=%t: got %t, want %t", constant, value, got, want)
			}
		}
		baseabc.Abc_NtkDelete(n)
	}
}
