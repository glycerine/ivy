package int

import (
	"errors"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

type stoppedBackend struct {
	fixtureM114p
	closed bool
}

func TestMiniSatInterpolationStepRejectsPendingInterpolant(t *testing.T) {
	p := &Inter_Man_t{pInterNew: aig.NewGraph()}
	defer func() {
		if recover() == nil {
			t.Fatal("pending interpolant accepted")
		}
	}()
	Inter_ManPerformOneStepM114p(p, &stoppedBackend{}, false)
}

func (s *stoppedBackend) Close() { s.closed = true }

func TestMiniSatInterpolationStepPropagatesUnknownAndCloses(t *testing.T) {
	inter := aig.NewGraph()
	x := inter.NewInput()
	inter.AddOutput(x.Not())
	trans := aig.NewGraph()
	state := trans.NewInput()
	trans.AddOutput(state)
	aig.Aig_ManSetRegNum(trans, 1)
	frames := aig.NewGraph()
	next := frames.NewInput()
	frames.AddOutput(next)
	p := &Inter_Man_t{pInter: inter, pCnfInter: cnf.Cnf_Derive(inter, 0), pAigTrans: trans, pCnfAig: cnf.Cnf_Derive(trans, 1), pFrames: frames, pCnfFrames: cnf.Cnf_Derive(frames, 0)}
	if got, err := Inter_ManPerformOneStepM114p(p, nil, false); got != -1 || !errors.Is(err, ErrMiniSat114pUnavailable) {
		t.Fatalf("nil backend got %d err %v", got, err)
	}
	s := &stoppedBackend{}
	if got, err := Inter_ManPerformOneStepM114p(p, s, false); got != -1 || err != nil || !s.closed || p.pInterNew != nil {
		t.Fatalf("stopped backend got %d err %v closed=%v", got, err, s.closed)
	}
}

func TestInterpolantStepIdentityTransition(t *testing.T) {
	inter := aig.NewGraph()
	x := inter.NewInput()
	inter.AddOutput(x.Not())
	trans := aig.NewGraph()
	state := trans.NewInput()
	trans.AddOutput(state)
	aig.Aig_ManSetRegNum(trans, 1)
	frames := aig.NewGraph()
	next := frames.NewInput()
	frames.AddOutput(next)
	p := &Inter_Man_t{pInter: inter, pCnfInter: cnf.Cnf_Derive(inter, 0), pAigTrans: trans, pCnfAig: cnf.Cnf_Derive(trans, 1), pFrames: frames, pCnfFrames: cnf.Cnf_Derive(frames, 0), vVarsAB: vec.Vec_IntAlloc(1)}
	beforeInter := append([]int32(nil), p.pCnfInter.VarNums...)
	beforeTrans := append([]int32(nil), p.pCnfAig.VarNums...)
	result := Inter_ManPerformOneStep(p, 0, 0, 0)
	if result != 1 || p.pInterNew == nil {
		t.Fatalf("expected interpolant, got %d graph=%v", result, p.pInterNew)
	}
	for mask := 0; mask < 2; mask++ {
		output, err := p.pInterNew.Evaluate([]bool{mask != 0})
		if err != nil || len(output) != 1 || output[0] != (mask == 0) {
			t.Fatalf("input %d: %v %v", mask, output, err)
		}
	}
	for i, v := range beforeInter {
		if p.pCnfInter.VarNums[i] != v {
			t.Fatal("inter CNF was not restored")
		}
	}
	for i, v := range beforeTrans {
		if p.pCnfAig.VarNums[i] != v {
			t.Fatal("transition CNF was not restored")
		}
	}
	satFrames := aig.NewGraph()
	satState := satFrames.NewInput()
	satFrames.AddOutput(satState.Not())
	p.pFrames, p.pCnfFrames, p.pInterNew = satFrames, cnf.Cnf_Derive(satFrames, 0), nil
	if result := Inter_ManPerformOneStep(p, 1, 0, 0); result != 0 || p.pInterNew != nil {
		t.Fatalf("expected satisfiable interpolation query, got %d graph=%v", result, p.pInterNew)
	}
}

func TestMiniSatInterpolationStepIdentityTransition(t *testing.T) {
	for _, pudlak := range []bool{false, true} {
		inter := aig.NewGraph()
		x := inter.NewInput()
		inter.AddOutput(x.Not())
		trans := aig.NewGraph()
		state := trans.NewInput()
		trans.AddOutput(state)
		aig.Aig_ManSetRegNum(trans, 1)
		frames := aig.NewGraph()
		next := frames.NewInput()
		frames.AddOutput(next)
		p := &Inter_Man_t{pInter: inter, pCnfInter: cnf.Cnf_Derive(inter, 0), pAigTrans: trans, pCnfAig: cnf.Cnf_Derive(trans, 1), pFrames: frames, pCnfFrames: cnf.Cnf_Derive(frames, 0)}
		beforeInter := append([]int32(nil), p.pCnfInter.VarNums...)
		beforeTrans := append([]int32(nil), p.pCnfAig.VarNums...)
		got, err := Inter_ManPerformOneStepM114p(p, &miniSat114pBackend{}, pudlak)
		if err != nil || got != 1 || p.pInterNew == nil {
			t.Fatalf("pudlak=%v got=%d err=%v graph=%v", pudlak, got, err, p.pInterNew)
		}
		for i, want := range beforeInter {
			if p.pCnfInter.VarNums[i] != want {
				t.Fatalf("pudlak=%v inter CNF not restored", pudlak)
			}
		}
		for i, want := range beforeTrans {
			if p.pCnfAig.VarNums[i] != want {
				t.Fatalf("pudlak=%v transition CNF not restored", pudlak)
			}
		}
		for x := 0; x < 2; x++ {
			out, e := p.pInterNew.Evaluate([]bool{x != 0})
			if e != nil || out[0] != (x == 0) {
				t.Fatalf("pudlak=%v x=%d out=%v err=%v", pudlak, x, out, e)
			}
		}
	}
}

func TestMiniSatInterpolationStepReturnsSATWithoutInterpolant(t *testing.T) {
	inter:=aig.NewGraph();x:=inter.NewInput();inter.AddOutput(x.Not())
	trans:=aig.NewGraph();state:=trans.NewInput();trans.AddOutput(state);aig.Aig_ManSetRegNum(trans,1)
	frames:=aig.NewGraph();next:=frames.NewInput();frames.AddOutput(next.Not())
	p:=&Inter_Man_t{pInter:inter,pCnfInter:cnf.Cnf_Derive(inter,0),pAigTrans:trans,pCnfAig:cnf.Cnf_Derive(trans,1),pFrames:frames,pCnfFrames:cnf.Cnf_Derive(frames,0)}
	for _,pudlak:=range []bool{false,true} {
		got,err:=Inter_ManPerformOneStepM114p(p,&miniSat114pBackend{},pudlak)
		if got!=0 || err!=nil || p.pInterNew!=nil { t.Fatalf("pudlak=%v got=%d err=%v interpolant=%v",pudlak,got,err,p.pInterNew) }
	}
}
