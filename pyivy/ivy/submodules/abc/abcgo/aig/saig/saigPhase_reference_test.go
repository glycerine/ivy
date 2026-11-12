package saig

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPhaseAbstractionAgainstOriginalABC(t *testing.T) {
	binary, err := filepath.Abs("../../../abc")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	source := phaseFixture()
	native := Saig_ManPhaseAbstract(source, nil, 2, 0, 0, 0, 0)
	if native == nil {
		t.Fatal("native abstraction refused")
	}
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.aig"), filepath.Join(dir, "output.aig")
	if err = ioa.Ioa_WriteAiger(source, input, 0, 0); err != nil {
		t.Fatal(err)
	}
	if log, err := exec.Command(binary, "-q", "read_aiger "+input+"; phase -F 2; write_aiger "+output).CombinedOutput(); err != nil {
		t.Fatalf("original phase: %v\n%s", err, log)
	}
	reference, err := ioa.Ioa_ReadAiger(output, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(reference.Latches) != int(aig.Aig_ManRegNum(native)) {
		t.Fatalf("latch count native=%d original=%d", aig.Aig_ManRegNum(native), len(reference.Latches))
	}
	for pattern := 0; pattern < 256; pattern++ {
		nativeState := make([]bool, aig.Aig_ManRegNum(native))
		referenceState := make([]bool, len(reference.Latches))
		for frame := 0; frame < 4; frame++ {
			inputs := []bool{pattern&(1<<uint(2*frame)) != 0, pattern&(1<<uint(2*frame+1)) != 0}
			outputs, err := native.Evaluate(append(append([]bool(nil), inputs...), nativeState...))
			if err != nil {
				t.Fatal(err)
			}
			expected, next := simulateReferenceAiger(reference, inputs, referenceState)
			for i, want := range expected {
				if outputs[i] != want {
					t.Fatalf("phase mismatch input %d frame %d output %d", pattern, frame, i)
				}
			}
			copy(nativeState, outputs[len(outputs)-len(nativeState):])
			referenceState = next
		}
	}
}
func simulateReferenceAiger(model *ioa.Model, inputs, state []bool) ([]bool, []bool) {
	values := make([]bool, int(model.MaxVar)+1)
	for i, lit := range model.Inputs {
		values[lit>>1] = inputs[i]
	}
	for i, latch := range model.Latches {
		values[latch.Literal>>1] = state[i]
	}
	value := func(lit uint32) bool { return values[lit>>1] != (lit&1 != 0) }
	for _, gate := range model.ANDs {
		values[gate.Literal>>1] = value(gate.Left) && value(gate.Right)
	}
	outputs := make([]bool, len(model.Outputs))
	for i, lit := range model.Outputs {
		outputs[i] = value(lit)
	}
	next := make([]bool, len(model.Latches))
	for i, latch := range model.Latches {
		next[i] = value(latch.Next)
	}
	return outputs, next
}
func TestPhaseLogicAndHashAgainstOriginalC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	source, err := os.ReadFile("../../../src/aig/saig/saigPhase.c")
	if err != nil {
		t.Skip("original phase source unavailable")
	}
	var program strings.Builder
	program.WriteString("#include <stdio.h>\n#include <assert.h>\n#define SAIG_XVS0 1\n#define SAIG_XVS1 2\n#define SAIG_XVSX 3\n")
	for _, name := range []string{"Saig_XsimInv", "Saig_XsimAnd", "Saig_TsiStateHash"} {
		program.WriteString(extractSimulationC(t, string(source), name))
	}
	program.WriteString("int main(void){for(int a=1;a<=3;a++)for(int b=1;b<=3;b++)printf(\"%d \",Saig_XsimAnd(a,b));for(int a=1;a<=3;a++)printf(\"%d \",Saig_XsimInv(a));unsigned x[257];for(int i=0;i<257;i++)x[i]=0xdeadbeefu*(i+1);printf(\"%d\",Saig_TsiStateHash(x,257,5003));}\n")
	dir := t.TempDir()
	file, binary := filepath.Join(dir, "phase.c"), filepath.Join(dir, "phase")
	if err = os.WriteFile(file, []byte(program.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if log, err := exec.Command(cc, "-std=c99", file, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile phase source: %v %s", err, log)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	actual := make([]string, 0, 13)
	for a := int32(1); a <= 3; a++ {
		for b := int32(1); b <= 3; b++ {
			actual = append(actual, fmt.Sprint(Saig_XsimAnd(a, b)))
		}
	}
	for a := int32(1); a <= 3; a++ {
		actual = append(actual, fmt.Sprint(Saig_XsimInv(a)))
	}
	words := make([]uint32, 257)
	for i := range words {
		words[i] = 0xdeadbeef * uint32(i+1)
	}
	actual = append(actual, fmt.Sprint(Saig_TsiStateHash(words, 257, 5003)))
	if strings.Join(actual, " ") != strings.TrimSpace(string(output)) {
		t.Fatalf("ternary/hash native=%v C=%s", actual, output)
	}
}
