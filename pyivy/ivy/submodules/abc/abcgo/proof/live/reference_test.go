package live

import (
	"github.com/glycerine/abcgo/aig/ioa"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func evalAiger(model *ioa.Model, pattern int, frames int) []bool {
	values := make([]bool, model.MaxVar+1)
	result := []bool{}
	literal := func(e uint32) bool { return values[e>>1] != (e&1 != 0) }
	for frame := 0; frame < frames; frame++ {
		for i, e := range model.Inputs {
			values[e>>1] = pattern&(1<<uint(frame*len(model.Inputs)+i)) != 0
		}
		for _, node := range model.ANDs {
			values[node.Literal>>1] = literal(node.Left) && literal(node.Right)
		}
		for _, e := range model.Outputs {
			result = append(result, literal(e))
		}
		next := make([]bool, len(model.Latches))
		for i, node := range model.Latches {
			next[i] = literal(node.Next)
		}
		for i, node := range model.Latches {
			values[node.Literal>>1] = next[i]
		}
	}
	return result
}
func TestBiereAgainstOriginalABC(t *testing.T) {
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
	for _, command := range []string{"l2s", "l2s -1", "l2s -l", "l2s -s", "l2s -s -1", "l2ssim", "l2ssim -1"} {
		n := liveFixture(true)
		model, err := n.Graph.AigerModel()
		if err != nil {
			t.Fatal(err)
		}
		model.Symbols = make(map[ioa.Symbol]string)
		for i := 0; i < trueOutputs(n.Graph); i++ {
			model.Symbols[ioa.Symbol{Kind: 'o', Index: uint32(i)}] = n.OutputNames[i]
		}
		dir := t.TempDir()
		input, output := filepath.Join(dir, "input.aig"), filepath.Join(dir, "output.aig")
		if err := ioa.Ioa_WriteAiger(model, input, 1, 0); err != nil {
			t.Fatal(err)
		}
		text := "read_aiger " + input + "; " + command + "; write_aiger " + output
		if log, err := exec.Command(binary, "-q", text).CombinedOutput(); err != nil {
			t.Fatalf("C live transform %v: %s", err, log)
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
			t.Fatalf("Go source command %s returned %d", command, status)
		}
		nativeModel, err := frame.Network.Graph.AigerModel()
		if err != nil {
			t.Fatal(err)
		}
		if len(reference.Inputs) != len(nativeModel.Inputs) || len(reference.Latches) != len(nativeModel.Latches) || len(reference.Outputs) != len(nativeModel.Outputs) {
			t.Fatalf("C/native interface %s: %d/%d PI %d/%d regs %d/%d PO", command, len(reference.Inputs), len(nativeModel.Inputs), len(reference.Latches), len(nativeModel.Latches), len(reference.Outputs), len(nativeModel.Outputs))
		}
		for pattern := 0; pattern < 256; pattern++ {
			if !slices.Equal(evalAiger(reference, pattern, 8), evalAiger(nativeModel, pattern, 8)) {
				t.Fatalf("C/native trace differs %s pattern%d", command, pattern)
			}
		}
	}
}
