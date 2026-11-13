package live

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
)

// Compare source commands with ABC on multiple liveness and safety POs and
// both general and one-step simulation; all four-frame input traces are used.
func TestSourceMultiOutputAndSimulationAgainstABC(t *testing.T) {
	binary := os.Getenv("ABC_SOURCE_BINARY")
	if binary == "" {
		_, file, _, _ := runtime.Caller(0)
		binary = filepath.Join(filepath.Dir(file), "../../../abc")
	}
	binary, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	for _, command := range []string{"l2s", "l2s -1", "l2ssim", "l2ssim -1"} {
		t.Run(command, func(t *testing.T) {
			g := aig.Aig_ManStart(30)
			in, state := aig.Aig_ObjCreateCi(g), aig.Aig_ObjCreateCi(g)
			for _, e := range []aig.Edge{in, state, in.Not(), state.Not(), state} {
				aig.Aig_ObjCreateCo(g, e)
			}
			aig.Aig_ManSetRegNum(g, 1)
			n := &Network{Graph: g, InputNames: []string{"in", "state"}, OutputNames: []string{"assert_fair_a", "assert_fair_b", "assume_fair_a", "assert_safety_a", "next"}}
			input, output := filepath.Join(t.TempDir(), "in.aig"), filepath.Join(t.TempDir(), "out.aig")
			model, err := g.AigerModel()
			if err != nil {
				t.Fatal(err)
			}
			model.Symbols = make(map[ioa.Symbol]string)
			for i := 0; i < trueOutputs(g); i++ {
				model.Symbols[ioa.Symbol{Kind: 'o', Index: uint32(i)}] = n.OutputNames[i]
			}
			if err := ioa.Ioa_WriteAiger(model, input, 1, 0); err != nil {
				t.Fatal(err)
			}
			script := "read_aiger " + input + "; " + command + "; write_aiger " + output
			if log, err := exec.Command(binary, "-q", script).CombinedOutput(); err != nil {
				t.Fatalf("ABC: %s %v", log, err)
			}
			reference, err := ioa.Ioa_ReadAiger(output, 1)
			if err != nil {
				t.Fatal(err)
			}
			frame := &Frame{Network: n}
			args := strings.Fields(command)
			var status int32
			if args[0] == "l2ssim" {
				status = Abc_CommandAbcLivenessToSafetySim(frame, int32(len(args)), args)
			} else {
				status = Abc_CommandAbcLivenessToSafety(frame, int32(len(args)), args)
			}
			if status != 0 {
				t.Fatalf("Go command status %d", status)
			}
			got, err := frame.Network.Graph.AigerModel()
			if err != nil {
				t.Fatal(err)
			}
			if len(reference.Inputs) != len(got.Inputs) || len(reference.Latches) != len(got.Latches) || len(reference.Outputs) != len(got.Outputs) {
				t.Fatalf("C/Go interfaces differ: %d/%d inputs %d/%d latches %d/%d outputs", len(reference.Inputs), len(got.Inputs), len(reference.Latches), len(got.Latches), len(reference.Outputs), len(got.Outputs))
			}
			for pattern := 0; pattern < 256; pattern++ {
				if !slices.Equal(evalAiger(reference, pattern, 4), evalAiger(got, pattern, 4)) {
					t.Fatalf("trace differs pattern %d", pattern)
				}
			}
			if args[0] == "l2ssim" {
				live, fair := populateLivenessVector(n, g), populateFairnessVector(n, g)
				var direct *aig.Graph
				if len(args) > 1 {
					direct = LivenessToSafetyTransformationOneStepLoopSim(n, g, live, fair)
				} else {
					direct = LivenessToSafetyTransformationSim(n, g, live, fair)
				}
				directModel, err := direct.AigerModel()
				if err != nil {
					t.Fatal(err)
				}
				for pattern := 0; pattern < 256; pattern++ {
					if !slices.Equal(evalAiger(reference, pattern, 4), evalAiger(directModel, pattern, 4)) {
						t.Fatalf("direct simulation trace differs pattern %d", pattern)
					}
				}
			}
		})
	}
}
