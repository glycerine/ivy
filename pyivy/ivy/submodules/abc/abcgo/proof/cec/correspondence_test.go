package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func TestCorrespondencePacking(t *testing.T) {
	info := vec.Vec_PtrAllocSimInfo(3, 1)
	presence := vec.Vec_PtrAllocSimInfo(3, 1)
	if Cec_ManLoadCounterExamplesTry(info, presence, 1, []int32{0, 3}, 2) != 1 {
		t.Fatal("first pack")
	}
	if Cec_ManLoadCounterExamplesTry(info, presence, 1, []int32{1}, 1) != 0 {
		t.Fatal("conflicting pack")
	}
	if cecBit(cecInfo(info, 0), 1) != 1 || cecBit(cecInfo(info, 1), 1) != 0 {
		t.Fatal("cube literal phase")
	}
}
func TestLatchCorrespondenceEqualTransitions(t *testing.T) {
	p := gia.Gia_ManStart(20)
	pi := gia.Gia_ManAppendCi(p)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a)
	gia.Gia_ManAppendCo(p, b)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManSetRegNum(p, 2)
	var pars Cec_ParCor_t
	Cec_ManCorSetDefaultParams(&pars)
	pars.FLatchCorr = 1
	pars.nWords = 1
	result := Cec_ManLSCorrespondence(p, &pars)
	if gia.Gia_ManRegNum(result) != 1 {
		t.Fatalf("registers %d", gia.Gia_ManRegNum(result))
	}
}

func corrFixture(mode int32) *Gia_Man_t {
	p := gia.Gia_ManStart(30)
	pi := gia.Gia_ManAppendCi(p)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	c := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a)
	gia.Gia_ManAppendCo(p, b)
	gia.Gia_ManAppendCo(p, c)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManAppendCo(p, pi)
	if mode == 0 {
		gia.Gia_ManAppendCo(p, 0)
	} else {
		gia.Gia_ManAppendCo(p, pi^1)
	}
	gia.Gia_ManSetRegNum(p, 3)
	return p
}
func corrTrace(p *Gia_Man_t, pattern int, frames int) []bool {
	values := make([]bool, gia.Gia_ManObjNum(p))
	out := []bool{}
	literal := func(lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
	for f := 0; f < frames; f++ {
		for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
			values[gia.Gia_ObjId(p, gia.Gia_ManPi(p, i))] = pattern&(1<<(f*int(gia.Gia_ManPiNum(p))+int(i))) != 0
		}
		for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 2:
				values[id] = literal(a) && literal(b)
			case 3:
				values[id] = literal(a) != literal(b)
			case 4:
				if literal(a) {
					values[id] = literal(b)
				} else {
					values[id] = literal(c)
				}
			case 5, 6:
				values[id] = literal(a)
			}
		}
		outputs := p.OutputLiterals()
		for _, lit := range outputs[:gia.Gia_ManPoNum(p)] {
			out = append(out, literal(lit))
		}
		next := make([]bool, gia.Gia_ManRegNum(p))
		for r := range next {
			next[r] = literal(outputs[int(gia.Gia_ManPoNum(p))+r])
		}
		for r, value := range next {
			values[gia.Gia_ObjId(p, gia.Gia_ManRo(p, int32(r)))] = value
		}
	}
	return out
}
func TestCorrespondenceModesPreserveSequentialFunction(t *testing.T) {
	for _, latch := range []int32{0, 1} {
		for _, rings := range []int32{0, 1} {
			for mode := int32(0); mode < 2; mode++ {
				original := corrFixture(mode)
				var pars Cec_ParCor_t
				Cec_ManCorSetDefaultParams(&pars)
				pars.FLatchCorr = latch
				pars.FUseRings = rings
				pars.nWords = 1
				pars.NBTLimit = 10000
				reduced := Cec_ManLSCorrespondence(original, &pars)
				if gia.Gia_ManRegNum(reduced) > 2 {
					t.Fatalf("failed to merge registers latch%d rings%d mode%d", latch, rings, mode)
				}
				for pattern := 0; pattern < 32; pattern++ {
					if !slices.Equal(corrTrace(original, pattern, 5), corrTrace(reduced, pattern, 5)) {
						t.Fatalf("trace changed latch%d rings%d mode%d pattern%d", latch, rings, mode, pattern)
					}
				}
			}
		}
	}
}

func TestCorrespondenceOriginalABC(t *testing.T) {
	binary := os.Getenv("ABC_SOURCE_BINARY")
	if binary == "" {
		t.Skip("original ABC executable unavailable")
	}
	for _, signal := range []bool{false, true} {
		for mode := int32(0); mode < 2; mode++ {
			original := corrFixture(mode)
			model, err := gia.Gia_ManToAigSimple(original).AigerModel()
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			input, output := filepath.Join(dir, "input.aig"), filepath.Join(dir, "output.aig")
			if err := ioa.Ioa_WriteAiger(model, input, 0, 0); err != nil {
				t.Fatal(err)
			}
			command := "&read " + input + "; &lcorr -c; &write " + output
			if signal {
				command = "&read " + input + "; &scorr -c; &write " + output
			}
			if log, err := exec.Command(binary, "-q", command).CombinedOutput(); err != nil {
				t.Fatalf("C correspondence %v: %s", err, log)
			}
			reference, err := ioa.Ioa_ReadAiger(output, 1)
			if err != nil {
				t.Fatal(err)
			}
			var pars Cec_ParCor_t
			Cec_ManCorSetDefaultParams(&pars)
			pars.nWords = 1
			if !signal {
				pars.FLatchCorr = 1
			}
			native := Cec_ManLSCorrespondence(original, &pars)
			if int(gia.Gia_ManRegNum(native)) != len(reference.Latches) {
				t.Fatalf("C/native latch count %d/%d", len(reference.Latches), gia.Gia_ManRegNum(native))
			}
			for pattern := 0; pattern < 32; pattern++ {
				if !slices.Equal(corrAigerTrace(reference, pattern, 5), corrTrace(native, pattern, 5)) {
					t.Fatalf("C/native trace differs signal%v mode%d pattern%d", signal, mode, pattern)
				}
			}
		}
	}
}
func corrAigerTrace(model *ioa.Model, pattern, frames int) []bool {
	values := make([]bool, model.MaxVar+1)
	out := []bool{}
	literal := func(lit uint32) bool { return values[lit>>1] != (lit&1 != 0) }
	for f := 0; f < frames; f++ {
		for i, lit := range model.Inputs {
			values[lit>>1] = pattern&(1<<(f*len(model.Inputs)+i)) != 0
		}
		for _, gate := range model.ANDs {
			values[gate.Literal>>1] = literal(gate.Left) && literal(gate.Right)
		}
		for _, lit := range model.Outputs {
			out = append(out, literal(lit))
		}
		next := make([]bool, len(model.Latches))
		for i, latch := range model.Latches {
			next[i] = literal(latch.Next)
		}
		for i, latch := range model.Latches {
			values[latch.Literal>>1] = next[i]
		}
	}
	return out
}

func TestCorrespondenceRandomSequentialNetworks(t *testing.T) {
	random := rand.New(rand.NewPCG(573, 117))
	for trial := 0; trial < 60; trial++ {
		p := gia.Gia_ManStart(50)
		literals := []int32{0, 1}
		pis := []int32{gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)}
		literals = append(literals, pis...)
		registers := []int32{gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)}
		literals = append(literals, registers...)
		for gate := 0; gate < 6; gate++ {
			a, b := literals[random.IntN(len(literals))], literals[random.IntN(len(literals))]
			if random.IntN(2) == 0 {
				a ^= 1
			}
			if random.IntN(2) == 0 {
				b ^= 1
			}
			literals = append(literals, gia.Gia_ManAppendAnd2(p, a, b))
		}
		gia.Gia_ManAppendCo(p, registers[0])
		gia.Gia_ManAppendCo(p, registers[1])
		gia.Gia_ManAppendCo(p, literals[len(literals)-1])
		transition := literals[random.IntN(len(literals))]
		gia.Gia_ManAppendCo(p, transition)
		gia.Gia_ManAppendCo(p, transition)
		gia.Gia_ManAppendCo(p, literals[random.IntN(len(literals))])
		gia.Gia_ManSetRegNum(p, 3)
		var pars Cec_ParCor_t
		Cec_ManCorSetDefaultParams(&pars)
		pars.nWords = 1
		pars.NFrames = 1 + int32(trial%2)
		pars.FLatchCorr = int32(trial % 2)
		pars.FUseRings = 0
		if trial%3 == 0 {
			pars.FUseRings = 1
		}
		pars.NBTLimit = 10000
		result := Cec_ManLSCorrespondence(p, &pars)
		for sample := 0; sample < 40; sample++ {
			pattern := random.IntN(1 << 12)
			if !slices.Equal(corrTrace(p, pattern, 6), corrTrace(result, pattern, 6)) {
				t.Fatalf("trial %d sequential trace differs pattern%d", trial, pattern)
			}
		}
	}
}

func TestSequentialSATPatternExpansion(t *testing.T) {
	p := gia.Gia_ManStart(100)
	input := gia.Gia_ManAppendCi(p)
	for i := 0; i < 70; i++ {
		gia.Gia_ManAppendCo(p, input)
	}
	info := vec.Vec_PtrAllocSimInfo(2, 1)
	var pars Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&pars)
	var patterns int32
	status := Cec_ManSatSolveSeq(info, p, &pars, 1, &patterns)
	if vec.Vec_PtrReadWordsSimInfo(info) < 4 || patterns < 70 {
		t.Fatalf("did not grow SAT witness storage: words%d patterns%d", vec.Vec_PtrReadWordsSimInfo(info), patterns)
	}
	for i, value := range vec.Vec_StrArray(status) {
		if value != 0 {
			t.Fatalf("SAT output%d", i)
		}
	}
	for bit := int32(1); bit <= patterns; bit++ {
		if bit%32 == 0 {
			continue
		}
		if cecBit(cecInfo(info, 1), bit) != 1 {
			t.Fatalf("lost SAT witness bit%d", bit)
		}
	}
}

func TestCorrespondenceSATRefinesRareTransition(t *testing.T) {
	p := gia.Gia_ManStart(50)
	conjunction := int32(1)
	for i := 0; i < 12; i++ {
		input := gia.Gia_ManAppendCi(p)
		conjunction = gia.Gia_ManAppendAnd2(p, conjunction, input)
	}
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a)
	gia.Gia_ManAppendCo(p, b)
	gia.Gia_ManAppendCo(p, conjunction)
	gia.Gia_ManAppendCo(p, 0)
	gia.Gia_ManSetRegNum(p, 2)
	for _, latch := range []int32{0, 1} {
		var pars Cec_ParCor_t
		Cec_ManCorSetDefaultParams(&pars)
		pars.nWords = 1
		pars.nRounds = 1
		pars.FLatchCorr = latch
		pars.NBTLimit = 10000
		result := Cec_ManLSCorrespondence(p, &pars)
		// One latch is truly zero, while the rare conjunction's latch needs the
		// incremental SAT witness rather than relying on random simulation alone.
		if gia.Gia_ManRegNum(result) != 1 {
			t.Fatalf("rare transition latch count%d", gia.Gia_ManRegNum(result))
		}
		for _, pattern := range []int{0, (1 << 12) - 1, (1 << 24) - 1} {
			if !slices.Equal(corrTrace(p, pattern, 2), corrTrace(result, pattern, 2)) {
				t.Fatalf("SAT refinement lost rare transition pattern%d", pattern)
			}
		}
	}
}

func TestCorrespondencePrefixAndSemiformal(t *testing.T) {
	p := corrFixture(1)
	var pars Cec_ParCor_t
	Cec_ManCorSetDefaultParams(&pars)
	pars.NPrefix = 1
	pars.FLatchCorr = 1
	pars.nWords = 1
	result := Cec_ManLSCorrespondence(p, &pars)
	if gia.Gia_ManRegNum(result) > 2 {
		t.Fatal("prefix failed to merge equal transitions")
	}
	p = corrFixture(1)
	var simPars Cec_ParSim_t
	Cec_ManSimSetDefaultParams(&simPars)
	simPars.NWords = 1
	simPars.FSeqSimulate = 1
	sim := Cec_ManSimStart(p, &simPars)
	Cec_ManSimClassesPrepare(sim, -1)
	var semiformal Cec_ParSmf_t
	Cec_ManSmfSetDefaultParams(&semiformal)
	semiformal.NWords = 1
	semiformal.NFrames = 3
	semiformal.NRounds = 2
	if got := Cec_ManSeqSemiformal(p, &semiformal); got != 0 {
		t.Fatalf("unexpected semiformal result%d", got)
	}
}

func TestCorrespondenceCreatesInternalChoices(t *testing.T) {
	p := gia.Gia_ManStart(40)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	state := gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManAppendAnd2(p, a, gia.Gia_ManAppendAnd2(p, b, c))
	y := gia.Gia_ManAppendAnd2(p, gia.Gia_ManAppendAnd2(p, a, b), c)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManAppendCo(p, y)
	gia.Gia_ManAppendCo(p, state)
	gia.Gia_ManSetRegNum(p, 1)
	var pars Cec_ParCor_t
	Cec_ManCorSetDefaultParams(&pars)
	pars.nWords = 1
	pars.FMakeChoices = 1
	pars.NBTLimit = 10000
	result := Cec_ManLSCorrespondence(p, &pars)
	choices := 0
	for id := int32(1); id < gia.Gia_ManObjNum(result); id++ {
		if result.Next(id) > 0 {
			choices++
		}
	}
	if choices == 0 {
		t.Fatal("equivalent internal implementations were not retained as choices")
	}
	for pattern := 0; pattern < 64; pattern++ {
		if !slices.Equal(corrTrace(p, pattern, 2), corrTrace(result, pattern, 2)) {
			t.Fatal("choices changed trace", pattern)
		}
	}
}

func TestCorrespondenceAutonomousOscillator(t *testing.T) {
	p := gia.Gia_ManStart(10)
	state := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, state)
	gia.Gia_ManAppendCo(p, state^1)
	gia.Gia_ManSetRegNum(p, 1)
	var pars Cec_ParCor_t
	Cec_ManCorSetDefaultParams(&pars)
	pars.FLatchCorr = 1
	pars.nWords = 1
	result := Cec_ManLSCorrespondence(p, &pars)
	if gia.Gia_ManRegNum(result) != 1 || !slices.Equal(corrTrace(p, 0, 8), corrTrace(result, 0, 8)) {
		t.Fatal("autonomous oscillator collapsed")
	}
}
