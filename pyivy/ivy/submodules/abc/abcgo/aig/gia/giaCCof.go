package gia

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/sat/bsat"
)

// Ccf_Man_t incrementally enumerates backward images by fixing frame PIs to
// a satisfying assignment while keeping the initial register CIs symbolic.
type Ccf_Man_t struct {
	pGia               *Gia_Man_t
	nConfMax, fVerbose int32
	pUnr               any
	pFrames            *Gia_Man_t
	vCopies            *Vec_Int_t
	pSat               *bsat.Solver
	initialSolver      *bsat.Solver
}

func Ccf_ManStart(src *Gia_Man_t, frames, conflicts, seconds, verbose int32) *Ccf_Man_t {
	if frames <= 0 || conflicts < 0 || seconds < 0 {
		panic("gia: invalid circuit-cofactoring limits")
	}
	parameters := &Gia_ParFra_t{NFrames: frames, fSaveLastLit: 1, FVerbose: verbose}
	p := &Ccf_Man_t{
		pGia: src, nConfMax: conflicts, fVerbose: verbose,
		pUnr:    Gia_ManUnrollStart(src, parameters),
		vCopies: Vec_IntAlloc(0),
		pSat:    bsat.NewSolver(),
	}
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
	if p.pSat != nil {
		p.pSat.Close()
	}
	if p.initialSolver != nil {
		p.initialSolver.Close()
	}
	*p = Ccf_Man_t{}
}
func Gia_ManCofExtendSolver(p *Ccf_Man_t) {
	if p.pFrames == nil {
		panic("gia: circuit cofactoring has no frames")
	}
	first := p.pSat.NumVars()
	if first < p.pFrames.nObjs {
		p.pSat.SetNumVars(p.pFrames.nObjs)
	}
	add := func(literals ...int32) { p.pSat.AddClause(literals...) }
	if first == 0 {
		add(1)
		first = 1
	} // Object zero is the false constant.
	for id := first; id < p.pFrames.nObjs; id++ {
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
	var previous int32
	var conflicts int64
	for count := 0; status == bsat.Satisfiable; count++ {
		var started time.Time
		if p.fVerbose != 0 {
			started = time.Now()
		}
		blocked := Gia_ManCofOneDerive(p, property)
		// A false blocking literal makes the whole solver inconsistent. This is a
		// completed image, not the source's debug-only add-clause assertion failure.
		p.pSat.AddClause(blocked)
		status = p.pSat.Solve(property)
		if p.fVerbose != 0 {
			currentConflicts := p.pSat.Statistics().Conflicts
			fmt.Printf("%3d : AIG =%7d  Conf =%7d.  Time = %8.2f sec\n", count, p.pFrames.nObjs-previous, currentConflicts-conflicts, time.Since(started).Seconds())
			previous, conflicts = p.pFrames.nObjs, currentConflicts
		}
	}
	if status == bsat.Unknown {
		return -1
	}
	return 0
}

// C assumes zero-initialized registers. Go also retains AIGER one/unknown
// resets, which must be intersected with the accumulated bad-state images.
func giaCofInitialFailure(p *Ccf_Man_t) bsat.Status {
	inits := p.pGia.aigerLatchInits
	unknown := false
	for _, reset := range inits {
		unknown = unknown || reset == -1
	}
	if !unknown {
		for _, obj := range p.pFrames.pObjs {
			if Gia_ObjIsCi(obj) != 0 {
				obj.fPhase = 0
				if i := Gia_ObjCioId(obj); int(i) < len(inits) {
					obj.fPhase = uint8(inits[i])
				}
			} else {
				Gia_ObjSetPhase(p.pFrames, obj)
			}
		}
		for _, id := range Vec_IntArray(p.pFrames.vCos) {
			if p.pFrames.pObjs[id].fPhase != 0 {
				return bsat.Satisfiable
			}
		}
		return bsat.Unsatisfiable
	}
	// The image-enumeration solver contains blocking clauses. Use the same
	// gate encoding without those clauses to test intersection with reset states.
	if p.initialSolver == nil {
		p.initialSolver = bsat.NewSolver()
	}
	check := &Ccf_Man_t{pFrames: p.pFrames, pSat: p.initialSolver}
	check.pSat.SetLimits(int64(p.nConfMax), 0)
	clock, deadline := p.pSat.RuntimeDeadline()
	check.pSat.SetRuntimeDeadline(clock, deadline)
	Gia_ManCofExtendSolver(check)
	assumptions := make([]int32, 0, len(inits)+1)
	for i, reset := range inits {
		if reset != -1 {
			assumptions = append(assumptions, Gia_ObjToLit(p.pFrames, Gia_ManCi(p.pFrames, int32(i)))^(reset^1))
		}
	}
	for _, id := range Vec_IntArray(p.pFrames.vCos) {
		property := Gia_ObjFaninLit0p(p.pFrames, p.pFrames.pObjs[id])
		if status := check.pSat.Solve(append(assumptions, property)...); status != bsat.Unsatisfiable {
			return status
		}
	}
	return bsat.Unsatisfiable
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
			fmt.Printf("ITER %3d :\n", frame)
		}
		p.pFrames = Gia_ManUnrollAdd(p.pUnr, frame+1).(*Gia_Man_t)
		Gia_ManCofExtendSolver(p)
		status = Gia_ManCofGetReachable(p, Gia_ManUnrollLastLit(p.pUnr))
		if status != 0 {
			break
		}
		initial := giaCofInitialFailure(p)
		if initial == bsat.Unknown {
			status = -1
			break
		}
		if initial == bsat.Satisfiable {
			fmt.Printf("Property failed in frame %d.\n", frame)
			failed = true
			break
		}
	}
	switch {
	case seconds > 0 && time.Since(started) >= time.Duration(seconds)*time.Second:
		fmt.Printf("Runtime limit (%d sec) is reached after %d frames.  ", seconds, min(frame+1, frames))
	case frame == frames:
		fmt.Printf("Completed %d frames without converging.  ", frame)
	case status == 1:
		fmt.Printf("Backward reachability converged after %d iterations.  ", frame+1)
	case status == -1:
		fmt.Printf("Conflict limit or timeout is reached after %d frames.  ", frame+1)
	}
	fmt.Printf("Runtime = %8.2f sec\n", time.Since(started).Seconds())
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
