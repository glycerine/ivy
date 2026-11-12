package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/ioa"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// This executes the real original C SAT-sweeping command, rather than a
// reimplementation of its logic. Gate IDs/scheduling may differ; the ordered
// output functions and removal of redundant cones must agree exhaustively.
func TestNativeSweepAgainstOriginalABC(t *testing.T) {
	binary := os.Getenv("ABC_SOURCE_BINARY")
	if binary == "" {
		binary = filepath.Join("..", "..", "..", "abc")
	}
	binary, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	original, _, _ := sweepFixture()
	params := new(Cec_ParFra_t)
	Cec_ManFraSetDefaultParams(params)
	reduced, err := SatSweep(original, params)
	if err != nil {
		t.Fatal(err)
	}
	inputModel, err := gia.Gia_ManToAigSimple(original).AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	nativeModel, err := gia.Gia_ManToAigSimple(reduced).AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.aig"), filepath.Join(dir, "output.aig")
	if err = ioa.Ioa_WriteAiger(inputModel, input, 0, 0); err != nil {
		t.Fatal(err)
	}
	command := "&read " + input + "; &fraig -J 0; &write " + output
	if log, err := exec.Command(binary, "-q", command).CombinedOutput(); err != nil {
		t.Fatalf("original C sweep: %v\n%s", err, log)
	}
	reference, err := ioa.Ioa_ReadAiger(output, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(reference.ANDs) >= len(inputModel.ANDs) || len(nativeModel.ANDs) >= len(inputModel.ANDs) {
		t.Fatal("reference/native failed to remove redundant cone")
	}
	for mask := 0; mask < 1<<len(inputModel.Inputs); mask++ {
		before := sweepEvaluateAiger(inputModel, mask)
		after, ref := sweepEvaluateAiger(nativeModel, mask), sweepEvaluateAiger(reference, mask)
		for i := range before {
			if before[i] != after[i] || before[i] != ref[i] {
				t.Fatalf("C/native difference input %d output %d", mask, i)
			}
		}
	}
}
func sweepEvaluateAiger(model *ioa.Model, mask int) []bool {
	values := make([]bool, int(model.MaxVar)+1)
	for i, lit := range model.Inputs {
		values[lit>>1] = mask&(1<<i) != 0
	}
	value := func(lit uint32) bool { return values[lit>>1] != (lit&1 != 0) }
	for _, gate := range model.ANDs {
		values[gate.Literal>>1] = value(gate.Left) && value(gate.Right)
	}
	outputs := make([]bool, len(model.Outputs))
	for i, lit := range model.Outputs {
		outputs[i] = value(lit)
	}
	return outputs
}
