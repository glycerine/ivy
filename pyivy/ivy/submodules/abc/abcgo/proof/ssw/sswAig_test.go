package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func sswTestFrameManager() (*Ssw_Man_t, aig.Edge, aig.Edge) {
	g := aig.NewGraph()
	g.Name = "transition"
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(g.XOR(pi, lo))
	aig.Aig_ManSetRegNum(g, 1)
	p := &Ssw_Man_t{pAig: g, pPars: &Ssw_Pars_t{nFramesK: 2}, nFrames: 3, pNodeToFrames: make([]aig.Edge, int(aig.Aig_ManObjNumMax(g))*3)}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	return p, pi, lo
}
func TestSswFramesWithClassesUnrollsTransition(t *testing.T) {
	p, _, _ := sswTestFrameManager()
	p.vInits = vec.Vec_IntAllocArrayCopy([]int32{1, 0, 0}, 3)
	frames := Ssw_FramesWithClasses(p)
	inputs := frames.Inputs()
	if len(inputs) != 3 || !aig.Aig_ObjPhase(frames, inputs[1]) || aig.Aig_ObjPhase(frames, inputs[2]) {
		t.Fatal("unexpected input order or initial phases")
	}
	for m := 0; m < 8; m++ {
		in := []bool{m&1 != 0, m&2 != 0, m&4 != 0}
		out, err := frames.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		want := in[0] != in[1] != in[2]
		if out[0] != want {
			t.Fatalf("inputs %v: got %v want %v", in, out, want)
		}
	}
}
func TestSswSpeculativeReductionPreservesLatchTransition(t *testing.T) {
	p, _, _ := sswTestFrameManager()
	g := Ssw_SpeculativeReduction(p)
	if g.Name != "transition" || g.RegisterCount() != 1 {
		t.Fatal("name or register count lost")
	}
	for m := 0; m < 4; m++ {
		in := []bool{m&1 != 0, m&2 != 0}
		out, err := g.Evaluate(in)
		if err != nil || out[0] != (in[0] != in[1]) {
			t.Fatalf("inputs %v: outputs %v, error %v", in, out, err)
		}
	}
}
func TestSswSpeculativeReductionClassConstraint(t *testing.T) {
	p, pi, lo := sswTestFrameManager()
	aig.Aig_ObjSetRepr(p.pAig, lo, pi)
	g := Ssw_SpeculativeReduction(p)
	if p.nConstrTotal != 1 || p.nConstrReduced != 1 {
		t.Fatal("equivalence constraint not created")
	}
	for m := 0; m < 4; m++ {
		in := []bool{m&1 != 0, m&2 != 0}
		out, err := g.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != 2 || out[0] != (in[0] != in[1]) || out[1] {
			t.Fatalf("inputs %v: got %v", in, out)
		}
	}
}
func TestSswFrameManagerInitialMapping(t *testing.T) {
	p, _, _ := sswTestFrameManager()
	f := Ssw_FrmStart(p.pAig)
	defer Ssw_FrmStop(f)
	if f.nObjs != aig.Aig_ManObjNumMax(p.pAig) || vec.Vec_PtrSize(f.vAig2Frm) != int(2*f.nObjs) {
		t.Fatal("incorrect frame mapping allocation")
	}
}

func TestFrameConstructionPreservesBufferedTransition(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	buffer := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{pi.Not(), aig.InvalidEdge}})
	g.AddOutput(lo)
	g.AddOutput(buffer)
	aig.Aig_ManSetRegNum(g, 1)
	p := &Ssw_Man_t{pAig: g, pPars: &Ssw_Pars_t{nFramesK: 1}, nFrames: 2, pNodeToFrames: make([]aig.Edge, len(g.Nodes())*2)}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	frames := Ssw_FramesWithClasses(p)
	for _, value := range []bool{false, true} {
		outputs, err := frames.Evaluate([]bool{false, value})
		if err != nil || len(outputs) != 1 || outputs[0] == value {
			t.Fatalf("input %v: output %v error %v", value, outputs, err)
		}
	}
}
