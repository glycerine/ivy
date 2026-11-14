package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

// Ccf_Man_t incrementally enumerates backward images by fixing frame PIs to
// a satisfying assignment while keeping the initial register CIs symbolic.
type Ccf_Man_t struct {
	pGia                                    *Gia_Man_t
	nFrameMax, nConfMax, nTimeMax, fVerbose int32
	pUnr                                    any
	pFrames                                 *Gia_Man_t
	vCopies                                 *Vec_Int_t
	pSat                                    *bsat.Solver
	encoded                                 int32
}

func Ccf_ManStart(src *Gia_Man_t, frames, conflicts, seconds, verbose int32) *Ccf_Man_t {
	if frames <= 0 || conflicts < 0 || seconds < 0 {
		panic("gia: invalid circuit-cofactoring limits")
	}
	parameters := &Gia_ParFra_t{NFrames: frames, fSaveLastLit: 1, FVerbose: verbose}
	p := &Ccf_Man_t{pGia: src, nFrameMax: frames, nConfMax: conflicts, nTimeMax: seconds, fVerbose: verbose, pUnr: Gia_ManUnrollStart(src, parameters), vCopies: Vec_IntAlloc(1000), pSat: bsat.NewSolver()}
	p.pSat.SetLimits(int64(conflicts), 0)
	if seconds > 0 {
		p.pSat.SetDeadline(time.Now().Add(time.Duration(seconds) * time.Second))
	}
	return p
}
func Ccf_ManStop(p *Ccf_Man_t) {
	if p == nil {
		return
	}
	Gia_ManUnrollStop(p.pUnr)
	if p.pFrames != nil {
		Gia_ManStop(p.pFrames)
	}
	*p = Ccf_Man_t{}
}
func Gia_ManCofExtendSolver(p *Ccf_Man_t) {
	if p.pFrames == nil {
		panic("gia: circuit cofactoring has no frames")
	}
	for p.pSat.NumVars() < p.pFrames.nObjs {
		p.pSat.NewVariable()
	}
	add := func(literals ...int32) { p.pSat.AddClause(literals...) }
	if p.encoded == 0 {
		add(1)
		p.encoded = 1
	} // Object zero is the false constant.
	for id := p.encoded; id < p.pFrames.nObjs; id++ {
		a, b, c, kind := p.pFrames.Gate(id)
		out := id << 1
		switch kind {
		case 2:
			add(out^1, a)
			add(out^1, b)
			add(out, a^1, b^1)
		case 3:
			add(a, b, out^1)
			add(a^1, b^1, out^1)
			add(a, b^1, out)
			add(a^1, b, out)
		case 4:
			add(a^1, b^1, out)
			add(a^1, b, out^1)
			add(a, c^1, out)
			add(a, c, out^1)
		case 5:
			add(out^1, a)
			add(out, a^1)
		}
	}
	p.encoded = p.pFrames.nObjs
}
func Gia_Obj0Copy(copies *Vec_Int_t, fanin, phase int32) int32 {
	return Vec_IntEntry(copies, int(fanin)) ^ phase
}
func Gia_Obj1Copy(copies *Vec_Int_t, fanin, phase int32) int32 {
	return Gia_Obj0Copy(copies, fanin, phase)
}
func Gia_ManCofOneDerive_rec(p *Ccf_Man_t, id int32) {
	if Vec_IntEntry(p.vCopies, int(id)) != -1 {
		return
	}
	obj := p.pFrames.pObjs[id]
	var result int32
	if Gia_ObjIsCi(obj) != 0 {
		if Gia_ObjCioId(obj) >= p.pGia.nRegs {
			value, ok := p.pSat.Value(id)
			if !ok {
				panic("gia: cofactoring requires satisfying SAT model")
			}
			result = giaBool(value)
		} else {
			result = id << 1
		}
	} else {
		a, b, c, kind := p.pFrames.Gate(id)
		copyLiteral := func(literal int32) int32 {
			Gia_ManCofOneDerive_rec(p, literal>>1)
			return Vec_IntEntry(p.vCopies, int(literal>>1)) ^ (literal & 1)
		}
		left := copyLiteral(a)
		switch kind {
		case 2:
			result = Gia_ManHashAnd(p.pFrames, left, copyLiteral(b))
		case 3:
			result = Gia_ManHashXor(p.pFrames, left, copyLiteral(b))
		case 4:
			result = Gia_ManHashMux(p.pFrames, left, copyLiteral(b), copyLiteral(c))
		case 5, 6:
			result = left
		default:
			panic("gia: unsupported circuit cofactoring object")
		}
	}
	Vec_IntWriteEntry(p.vCopies, int(id), result)
}
func Gia_ManCofOneDerive(p *Ccf_Man_t, property int32) int32 {
	Vec_IntFill(p.vCopies, int(p.pFrames.nObjs), -1)
	Vec_IntWriteEntry(p.vCopies, 0, 0)
	Gia_ManCofOneDerive_rec(p, property>>1)
	literal := Vec_IntEntry(p.vCopies, int(property>>1)) ^ (property & 1)
	Gia_ManAppendCo(p.pFrames, literal)
	Gia_ManCofExtendSolver(p)
	return literal ^ 1
}

// Source return convention: 1 means already converged on entry; 0 means this
// call enumerated and blocked states; -1 means a conflict/deadline limit.
func Gia_ManCofGetReachable(p *Ccf_Man_t, property int32) int32 {
	status := p.pSat.Solve(property)
	if status == bsat.Unsatisfiable {
		return 1
	}
	for count := 0; status == bsat.Satisfiable; count++ {
		previous := p.pFrames.nObjs
		conflicts := p.pSat.Statistics().Conflicts
		started := time.Now()
		blocked := Gia_ManCofOneDerive(p, property)
		// A false blocking literal makes the whole solver inconsistent. This is a
		// completed image, not the source's debug-only add-clause assertion failure.
		p.pSat.AddClause(blocked)
		status = p.pSat.Solve(property)
		if p.fVerbose != 0 {
			fmt.Printf("%3d: AIG=%7d Conf=%7d Time=%s\n", count, p.pFrames.nObjs-previous, p.pSat.Statistics().Conflicts-conflicts, time.Since(started))
		}
	}
	if status == bsat.Unknown {
		return -1
	}
	return 0
}
func Gia_ManCofTest(src *Gia_Man_t, frames, conflicts, seconds, verbose int32) *Gia_Man_t {
	if Gia_ManPoNum(src) != 1 {
		panic("gia: circuit cofactoring requires one property")
	}
	started := time.Now()
	p := Ccf_ManStart(src, frames, conflicts, seconds, verbose)
	var status int32 = -1
	var frame int32
	failed := false
	for frame = 0; frame < frames; frame++ {
		if verbose != 0 {
			fmt.Printf("ITER %3d:\n", frame)
		}
		p.pFrames = Gia_ManUnrollAdd(p.pUnr, frame+1).(*Gia_Man_t)
		Gia_ManCofExtendSolver(p)
		status = Gia_ManCofGetReachable(p, Gia_ManUnrollLastLit(p.pUnr))
		if status != 0 {
			break
		}
		Gia_ManSetPhase(p.pFrames)
		for i := int32(0); i < Gia_ManPoNum(p.pFrames); i++ {
			if Gia_ObjPhase(Gia_ManPo(p.pFrames, i)) != 0 {
				fmt.Printf("Property failed in frame %d.\n", frame)
				failed = true
				break
			}
		}
		if failed {
			break
		}
	}
	switch {
	case seconds > 0 && time.Since(started) >= time.Duration(seconds)*time.Second:
		fmt.Printf("Runtime limit (%d sec) reached after %d frames. ", seconds, frame)
	case frame == frames:
		fmt.Printf("Completed %d frames without converging. ", frame)
	case status == 1:
		fmt.Printf("Backward reachability converged after %d iterations. ", frame-1)
	case status == -1:
		fmt.Printf("Conflict limit or timeout reached after %d frames. ", frame-1)
	}
	fmt.Printf("Runtime=%s. ", time.Since(started))
	if !failed && status == 1 {
		fmt.Println("Property holds.")
	} else if !failed {
		fmt.Println("Property is undecided.")
	}
	result := p.pFrames
	p.pFrames = nil
	Gia_ManHashStop(result)
	Ccf_ManStop(p)
	cleaned := Gia_ManCleanup(result)
	Gia_ManStop(result)
	return cleaned
}
