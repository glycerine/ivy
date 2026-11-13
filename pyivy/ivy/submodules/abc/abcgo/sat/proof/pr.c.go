package proof

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	proofChunkSize   = 1 << 16
	defaultProofFile = "proof.cnf_"
)

const LIT_UNDEF Lit = ^Lit(0)

// Pr_Cls_t stores one input clause or learned clause and its assigned proof ID.
type Pr_Cls_t struct {
	uTruth uint32
	pProof int32
	pNext  *Pr_Cls_t
	Id     int32
	fA     bool
	fRoot  bool
	fVisit bool
	pLits  []Lit
}

// Pr_Man_t_ owns a clause database and the state used to check RUP learned
// clauses and emit the corresponding resolution trace.
type Pr_Man_t_ struct {
	fProofWrite   int32
	fProofVerif   int32
	nVars         int32
	nVarsAB       int32
	nRoots        int32
	nClauses      int32
	nClausesA     int32
	activeClauses int32
	pHead         *Pr_Cls_t
	pTail         *Pr_Cls_t
	pLearnt       *Pr_Cls_t
	pEmpty        *Pr_Cls_t
	rootConflict  *Pr_Cls_t

	nRootSize  int32
	nTrailSize int32
	pTrail     []Lit
	pAssigns   []Lit
	pSeens     []bool
	pVarTypes  []int8
	pReasons   []*Pr_Cls_t
	nVarsAlloc int32

	Counter     int32
	clauses     []*Pr_Cls_t
	proofFile   string
	proofWriter io.Writer
	lastError   error
	proofSteps  []ProofStep
	startTime   time.Time
	timeBcp     time.Duration
	timeTrace   time.Duration
	timeRead    time.Duration
	timeTotal   time.Duration
	memoryUsed  int64
	interpolant uint32
	finished    bool
}

// ProofStep records one resolution of clause Left with clause Right.
type ProofStep struct {
	ID          int32
	Left, Right int32
	Literals    []Lit
}

func ToLit(v int32) Lit {
	if v < 0 || uint64(v) > uint64(^uint32(0)>>1) {
		panic("proof: variable out of range")
	}
	return Lit(uint32(v) << 1)
}
func LitFromDIMACS(v int32) Lit {
	if v == 0 || v == -1<<31 {
		panic("proof: invalid DIMACS literal")
	}
	l := ToLit(abs(v) - 1)
	if v < 0 {
		return l ^ 1
	}
	return l
}
func LitToDIMACS(l Lit) int32 {
	v := int64(l>>1) + 1
	if l&1 != 0 {
		v = -v
	}
	if v < -1<<31 || v > 1<<31-1 {
		panic("proof: literal out of range")
	}
	return int32(v)
}
func LitVar(l Lit) int32           { return int32(l >> 1) }
func LitSign(l Lit) bool           { return l&1 != 0 }
func LitNeg(l Lit) Lit             { return l ^ 1 }
func LitValid(l Lit, n int32) bool { return l != LIT_UNDEF && LitVar(l) < n }
func toLit(v int32) lit            { return ToLit(v) }
func toLitCond(v int32, c int32) lit {
	l := ToLit(v)
	if c != 0 {
		return l ^ 1
	}
	return l
}
func lit_neg(l lit) lit   { return LitNeg(l) }
func lit_var(l lit) int32 { return LitVar(l) }
func lit_sign(l lit) int32 {
	if LitSign(l) {
		return 1
	}
	return 0
}
func lit_print(l lit) int32 { return LitToDIMACS(l) }
func lit_read(v int32) lit  { return LitFromDIMACS(v) }
func lit_check(l lit, n int32) int32 {
	if LitValid(l, n) {
		return 1
	}
	return 0
}
func abs(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func Pr_ManAlloc(nVarsAlloc int32) *Pr_Man_t {
	if nVarsAlloc < 0 {
		nVarsAlloc = 0
	}
	p := &Pr_Man_t{proofFile: defaultProofFile, startTime: time.Now()}
	Pr_ManResize(p, nVarsAlloc)
	return p
}
func Pr_ManResize(p *Pr_Man_t, nVarsNew int32) {
	if p == nil || nVarsNew < 0 {
		return
	}
	if nVarsNew > p.nVarsAlloc {
		alloc := p.nVarsAlloc
		if alloc == 0 {
			alloc = 1
		}
		for alloc < nVarsNew {
			alloc *= 2
		}
		old := int(p.nVarsAlloc)
		p.pAssigns = append(p.pAssigns, make([]Lit, int(alloc)-len(p.pAssigns))...)
		p.pSeens = append(p.pSeens, make([]bool, int(alloc)-len(p.pSeens))...)
		p.pVarTypes = append(p.pVarTypes, make([]int8, int(alloc)-len(p.pVarTypes))...)
		p.pReasons = append(p.pReasons, make([]*Pr_Cls_t, int(alloc)-len(p.pReasons))...)
		for i := old; i < int(alloc); i++ {
			p.pAssigns[i] = LIT_UNDEF
		}
		p.nVarsAlloc = alloc
	}
	if p.nVars < nVarsNew {
		p.nVars = nVarsNew
	}
}
func Pr_ManFree(p *Pr_Man_t) {
	if p == nil {
		return
	}
	p.pTrail, p.pAssigns, p.pSeens, p.pVarTypes, p.pReasons = nil, nil, nil, nil, nil
	p.clauses, p.proofSteps = nil, nil
	p.pHead, p.pTail, p.pLearnt, p.pEmpty = nil, nil, nil, nil
	p.proofWriter = nil
}
func (p *Pr_Man_t) SetProofOutput(file string) {
	if p == nil {
		return
	}
	p.fProofWrite = 1
	if file == "" {
		file = defaultProofFile
	}
	p.proofFile = file
}
func (p *Pr_Man_t) DisableProofOutput() {
	if p != nil {
		p.fProofWrite = 0
	}
}
func (p *Pr_Man_t) SetProofVerification(enabled bool) {
	if p == nil {
		return
	}
	if enabled {
		p.fProofVerif = 1
	} else {
		p.fProofVerif = 0
	}
}
func (p *Pr_Man_t) LastError() error {
	if p == nil {
		return errors.New("proof: nil manager")
	}
	return p.lastError
}
func (p *Pr_Man_t) Clauses() [][]Lit {
	if p == nil {
		return nil
	}
	out := make([][]Lit, len(p.clauses))
	for i, c := range p.clauses {
		out[i] = append([]Lit(nil), c.pLits...)
	}
	return out
}
func (p *Pr_Man_t) ProofSteps() []ProofStep {
	if p == nil {
		return nil
	}
	out := make([]ProofStep, len(p.proofSteps))
	for i, s := range p.proofSteps {
		out[i] = s
		out[i].Literals = append([]Lit(nil), s.Literals...)
	}
	return out
}
func (p *Pr_Man_t) Counts() (vars, roots, clauses int32) {
	if p == nil {
		return 0, 0, 0
	}
	return p.nVars, p.nRoots, p.nClauses
}

// Pr_ManAddClause appends a clause. Root clauses must precede learned clauses;
// A-partition clauses must be roots. It sorts literals by packed value and
// rejects repeated variables or out-of-range flags instead of relying on C asserts.
func Pr_ManAddClause(p *Pr_Man_t, literals []Lit, fRoot, fClauseA int32) int32 {
	if p == nil {
		return 0
	}
	p.lastError = nil
	if p.pEmpty != nil {
		p.lastError = errors.New("proof: clause follows terminal empty clause")
		return 0
	}
	if fRoot < 0 || fRoot > 1 || fClauseA < 0 || fClauseA > 1 || (fClauseA != 0 && fRoot == 0) {
		p.lastError = errors.New("proof: invalid root/A flags")
		return 0
	}
	if p.pLearnt != nil && fRoot != 0 {
		p.lastError = errors.New("proof: root clause follows learned clause")
		return 0
	}
	clause := append([]Lit(nil), literals...)
	sortLits(clause)
	maxVar := int32(-1)
	for i, l := range clause {
		if l == LIT_UNDEF {
			p.lastError = errors.New("proof: undefined literal in clause")
			return 0
		}
		if v := LitVar(l); v > maxVar {
			maxVar = v
		}
		if i > 0 && LitVar(clause[i-1]) == LitVar(l) {
			p.lastError = fmt.Errorf("proof: repeated variable %d in clause", LitVar(l))
			return 0
		}
	}
	if maxVar >= 0 {
		Pr_ManResize(p, maxVar+1)
	}
	if fClauseA != 0 && p.nClausesA != int32(len(p.clauses)) {
		p.lastError = errors.New("proof: A clauses must form the initial root prefix")
		return 0
	}
	c := &Pr_Cls_t{Id: p.nClauses, fRoot: fRoot != 0, fA: fClauseA != 0, pLits: clause}
	if p.pHead == nil {
		p.pHead = c
	}
	if p.pTail != nil {
		p.pTail.pNext = c
	}
	p.pTail = c
	if p.pLearnt == nil && !c.fRoot {
		p.pLearnt = c
	}
	if len(clause) == 0 {
		p.pEmpty = c
	}
	p.clauses = append(p.clauses, c)
	p.nClauses++
	if c.fRoot {
		p.nRoots++
	}
	if c.fA {
		p.nClausesA++
	}
	return 1
}
func sortLits(lits []Lit) {
	for i := 1; i < len(lits); i++ {
		for j := i; j > 0 && lits[j-1] > lits[j]; j-- {
			lits[j-1], lits[j] = lits[j], lits[j-1]
		}
	}
}

func Pr_ManMemoryReport(p *Pr_Man_t) int32 {
	if p == nil {
		return 0
	}
	n := int64(cap(p.clauses))*8 + int64(cap(p.pAssigns))*16 + int64(cap(p.pTrail))*4
	for _, c := range p.clauses {
		n += int64(cap(c.pLits))*4 + 48
	}
	if n > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(n)
}
func Pr_ManMemoryStop(p *Pr_Man_t) {
	if p != nil {
		p.memoryUsed = 0
	}
}

func Pr_ManEnqueue(p *Pr_Man_t, l Lit, reason *Pr_Cls_t) int32 {
	if p == nil || !LitValid(l, p.nVars) {
		return 0
	}
	v := LitVar(l)
	if p.pAssigns[v] != LIT_UNDEF {
		if p.pAssigns[v] == l {
			return 1
		}
		return 0
	}
	p.pAssigns[v] = l
	p.pReasons[v] = reason
	p.pTrail = append(p.pTrail, l)
	p.nTrailSize = int32(len(p.pTrail))
	return 1
}
func Pr_ManCancelUntil(p *Pr_Man_t, level int32) {
	if p == nil {
		return
	}
	if level < 0 {
		level = 0
	}
	if level > int32(len(p.pTrail)) {
		level = int32(len(p.pTrail))
	}
	for _, l := range p.pTrail[level:] {
		v := LitVar(l)
		p.pAssigns[v], p.pReasons[v] = LIT_UNDEF, nil
	}
	p.pTrail = p.pTrail[:level]
	p.nTrailSize = level
}

// Pr_ManPropagateOne propagates the current trail to a fixpoint and returns a
// conflicting clause if one is found. The database is small and immutable per
// proof step, so this native implementation scans clauses instead of storing
// C-style pointer-linked watch lists.
func Pr_ManPropagateOne(p *Pr_Man_t, _ Lit) *Pr_Cls_t {
	if p == nil {
		return nil
	}
	conflict, _ := propagate(p, len(p.pTrail))
	return conflict
}
func Pr_ManPropagate(p *Pr_Man_t, start int32) *Pr_Cls_t {
	if p == nil {
		return nil
	}
	began := time.Now()
	conflict, _ := propagate(p, int(start))
	p.timeBcp += time.Since(began)
	return conflict
}
func propagate(p *Pr_Man_t, start int) (*Pr_Cls_t, int) {
	if start < 0 {
		start = 0
	}
	if start > len(p.pTrail) {
		start = len(p.pTrail)
	}
	for {
		changed := false
		for _, clause := range p.clauses[:p.activeClauses] {
			trueCount, undefCount, last := 0, 0, Lit(0)
			for _, l := range clause.pLits {
				value := p.pAssigns[LitVar(l)]
				if value == LIT_UNDEF {
					undefCount++
					last = l
				} else if value == l {
					trueCount++
					break
				}
			}
			if trueCount > 0 {
				continue
			}
			if undefCount == 0 {
				return clause, start
			}
			if undefCount == 1 {
				if Pr_ManEnqueue(p, last, clause) == 0 {
					return clause, start
				}
				changed = true
			}
		}
		if !changed {
			return nil, start
		}
	}
}

func Pr_ManProofWriteOne(p *Pr_Man_t, c *Pr_Cls_t) {
	if p == nil || c == nil {
		return
	}
	p.Counter++
	c.pProof = p.Counter
	if p.proofWriter != nil {
		writeClauseLine(p.proofWriter, c.pProof, c.pLits)
	}
}

func Pr_ManProofTraceOne(p *Pr_Man_t, conflict, final *Pr_Cls_t) int32 {
	if p == nil || conflict == nil || final == nil {
		return 0
	}
	start := time.Now()
	resolvent := append([]Lit(nil), conflict.pLits...)
	seen := make([]bool, p.nVars)
	for _, l := range resolvent {
		seen[LitVar(l)] = true
	}
	if p.nClausesA > 0 {
		final.uTruth = conflict.uTruth
	}
	prev := conflict.pProof
	for i := len(p.pTrail) - 1; i >= 0; i-- {
		assigned := p.pTrail[i]
		v := LitVar(assigned)
		if !seen[v] {
			continue
		}
		seen[v] = false
		reason := p.pReasons[v]
		if reason == nil {
			continue
		}
		resolvent, p.lastError = resolveOn(resolvent, reason.pLits, v)
		if p.lastError != nil {
			p.timeTrace += time.Since(start)
			return 0
		}
		if p.nClausesA > 0 {
			if p.pVarTypes[v] == 1 {
				final.uTruth |= reason.uTruth
			} else {
				final.uTruth &= reason.uTruth
			}
		}
		p.Counter++
		step := ProofStep{ID: p.Counter, Left: prev, Right: reason.pProof, Literals: append([]Lit(nil), resolvent...)}
		p.proofSteps = append(p.proofSteps, step)
		if p.proofWriter != nil {
			fmt.Fprintf(p.proofWriter, "%d * %d %d 0\n", step.ID, step.Left, step.Right)
		}
		prev = p.Counter
		for _, l := range reason.pLits {
			if LitVar(l) != v {
				seen[LitVar(l)] = true
			}
		}
	}
	if !subsetByLiteral(resolvent, final.pLits) {
		p.lastError = fmt.Errorf("proof: derived resolvent %v does not imply learned clause %v", resolvent, final.pLits)
		p.timeTrace += time.Since(start)
		return 0
	}
	if p.nClausesA > 0 {
		p.interpolant = final.uTruth
	}
	p.timeTrace += time.Since(start)
	return prev
}
func resolveOn(current, reason []Lit, variable int32) ([]Lit, error) {
	pivot := -1
	for i, l := range current {
		if LitVar(l) == variable {
			pivot = i
			break
		}
	}
	if pivot < 0 {
		return nil, fmt.Errorf("proof: pivot variable %d missing from resolvent", variable)
	}
	assignedComplement := LitNeg(current[pivot])
	if !containsLit(reason, assignedComplement) {
		return nil, fmt.Errorf("proof: reason lacks complement of pivot %d", variable)
	}
	out := append([]Lit(nil), current[:pivot]...)
	out = append(out, current[pivot+1:]...)
	for _, l := range reason {
		if LitVar(l) == variable {
			continue
		}
		found := false
		for _, old := range out {
			if LitVar(old) == LitVar(l) {
				if old != l {
					return nil, fmt.Errorf("proof: tautological resolvent on variable %d", LitVar(l))
				}
				found = true
				break
			}
		}
		if !found {
			out = append(out, l)
		}
	}
	return out, nil
}
func containsLit(list []Lit, needle Lit) bool {
	for _, l := range list {
		if l == needle {
			return true
		}
	}
	return false
}
func subsetByLiteral(a, b []Lit) bool {
	for _, l := range a {
		if !containsLit(b, l) {
			return false
		}
	}
	return true
}

func Pr_ManProofRecordOne(p *Pr_Man_t, clause *Pr_Cls_t) int32 {
	if p == nil || clause == nil || clause.fRoot || len(clause.pLits) == 0 || p.pEmpty == nil {
		if p != nil {
			p.lastError = errors.New("proof: learned clauses must be non-empty and have a final empty clause")
		}
		return 0
	}
	if len(p.pTrail) != int(p.nRootSize) {
		p.lastError = errors.New("proof: trail is not at root level")
		return 0
	}
	for _, l := range clause.pLits {
		if Pr_ManEnqueue(p, LitNeg(l), nil) == 0 {
			Pr_ManCancelUntil(p, p.nRootSize)
			p.lastError = errors.New("proof: learned clause contradicted by root assignment")
			return 0
		}
	}
	conflict := Pr_ManPropagate(p, p.nRootSize)
	if conflict == nil {
		Pr_ManCancelUntil(p, p.nRootSize)
		p.lastError = fmt.Errorf("proof: learned clause %d is not RUP", clause.Id)
		return 0
	}
	proofID := Pr_ManProofTraceOne(p, conflict, clause)
	Pr_ManCancelUntil(p, p.nRootSize)
	if proofID == 0 {
		return 0
	}
	clause.pProof = proofID
	if len(clause.pLits) == 1 {
		if Pr_ManEnqueue(p, clause.pLits[0], clause) == 0 {
			p.lastError = errors.New("proof: learned unit contradicts root assignment")
			return 0
		}
		conflict = Pr_ManPropagate(p, p.nRootSize)
		if conflict != nil {
			if p.pEmpty == nil {
				p.lastError = errors.New("proof: no terminal empty clause")
				return 0
			}
			p.pEmpty.pProof = Pr_ManProofTraceOne(p, conflict, p.pEmpty)
			if p.pEmpty.pProof == 0 {
				return 0
			}
			p.finished = true
			return 1
		}
		p.nRootSize = p.nTrailSize
	}
	p.activeClauses++
	return 1
}

func Pr_ManProcessRoots(p *Pr_Man_t) int32 {
	if p == nil {
		return 0
	}
	if len(p.clauses) == 0 || len(p.clauses[len(p.clauses)-1].pLits) != 0 {
		p.lastError = errors.New("proof: final clause must be empty")
		return 0
	}
	for i, c := range p.clauses {
		if c.fRoot != (int32(i) < p.nRoots) || c.fA != (int32(i) < p.nClausesA) {
			p.lastError = errors.New("proof: roots and A clauses must form ordered prefixes")
			return 0
		}
	}
	p.pTrail = nil
	p.nTrailSize = 0
	p.rootConflict = nil
	for i := range p.pAssigns {
		p.pAssigns[i] = LIT_UNDEF
		p.pReasons[i] = nil
	}
	p.activeClauses = p.nRoots
	for _, c := range p.clauses[:p.nRoots] {
		if len(c.pLits) == 0 {
			p.rootConflict = c
			return 1
		}
		if len(c.pLits) != 1 {
			continue
		}
		if !LitValid(c.pLits[0], p.nVars) || Pr_ManEnqueue(p, c.pLits[0], c) == 0 {
			p.rootConflict = c
			return 1
		}
	}
	if conflict := Pr_ManPropagate(p, 0); conflict != nil {
		p.rootConflict = conflict
		return 1
	}
	p.nRootSize = p.nTrailSize
	return 1
}

func Pr_ManPrepareInter(p *Pr_Man_t) {
	if p == nil || p.nClausesA == 0 {
		return
	}
	shared := make([]bool, p.nVars)
	for _, c := range p.clauses[:p.nClausesA] {
		for _, l := range c.pLits {
			p.pVarTypes[LitVar(l)] = 1
		}
	}
	for _, c := range p.clauses[:p.nRoots] {
		if c.fA {
			continue
		}
		for _, l := range c.pLits {
			v := LitVar(l)
			if p.pVarTypes[v] == 1 {
				shared[v] = true
			}
		}
	}
	common := []int32{}
	for v := int32(0); v < p.nVars; v++ {
		if shared[v] {
			p.pVarTypes[v] = -1 - int8(len(common))
			common = append(common, v)
		}
	}
	if len(common) > 5 {
		p.lastError = fmt.Errorf("proof: interpolation supports at most five shared variables, got %d", len(common))
		return
	}
	truths := []uint32{0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000}
	for _, c := range p.clauses[:p.nRoots] {
		if !c.fA {
			c.uTruth = ^uint32(0)
			continue
		}
		c.uTruth = 0
		for _, l := range c.pLits {
			v := LitVar(l)
			if p.pVarTypes[v] < 0 {
				idx := int(-p.pVarTypes[v] - 1)
				term := truths[idx]
				if LitSign(l) {
					term = ^term
				}
				c.uTruth |= term
			}
		}
	}
	p.nVarsAB = int32(len(common))
}

func Pr_ManProofWrite(p *Pr_Man_t) int32 {
	if p == nil {
		return 0
	}
	p.lastError = nil
	p.proofSteps = nil
	p.Counter = 0
	p.finished = false
	started := time.Now()
	defer func() { p.timeTotal = time.Since(started) }()
	if Pr_ManProcessRoots(p) == 0 {
		return 0
	}
	if p.nClausesA > 0 {
		Pr_ManPrepareInter(p)
		if p.lastError != nil {
			return 0
		}
	}
	var output *os.File
	if p.fProofWrite != 0 {
		f, err := os.Create(p.proofFile)
		if err != nil {
			p.lastError = err
			return 0
		}
		output = f
		p.proofWriter = f
	}
	defer func() {
		if output != nil {
			if err := output.Close(); err != nil && p.lastError == nil {
				p.lastError = err
			}
		}
		p.proofWriter = nil
	}()
	for _, c := range p.clauses[:p.nRoots] {
		Pr_ManProofWriteOne(p, c)
	}
	if p.rootConflict != nil {
		p.pEmpty.pProof = Pr_ManProofTraceOne(p, p.rootConflict, p.pEmpty)
		if p.pEmpty.pProof == 0 {
			return 0
		}
		p.finished = true
		return 1
	}
	for _, c := range p.clauses[p.nRoots:] {
		if len(c.pLits) == 0 {
			continue
		}
		if Pr_ManProofRecordOne(p, c) == 0 {
			if p.lastError == nil {
				p.lastError = fmt.Errorf("proof: failed to record clause %d", c.Id)
			}
			return 0
		}
		if p.finished {
			break
		}
	}
	if p.pEmpty != nil && p.pEmpty.pProof == 0 {
		p.lastError = errors.New("proof: learned clauses did not derive the terminal empty clause")
		return 0
	}
	return 1
}

func ReadProofLog(fileName string) (*Pr_Man_t, error) {
	started := time.Now()
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 4<<20)
	var p *Pr_Man_t
	var wantClauses, wantRoots, wantA int32
	var seen int32
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "c") {
			continue
		}
		fields := strings.Fields(line)
		if strings.HasPrefix(line, "p") {
			if p != nil || len(fields) < 4 || len(fields) > 5 {
				return nil, fmt.Errorf("proof: line %d: malformed header", lineNo)
			}
			vals := make([]int64, len(fields)-1)
			for i, s := range fields[1:] {
				vals[i], err = strconv.ParseInt(s, 10, 32)
				if err != nil || vals[i] < 0 {
					return nil, fmt.Errorf("proof: line %d: invalid header number", lineNo)
				}
			}
			p = Pr_ManAlloc(int32(vals[0]))
			wantClauses, wantRoots = int32(vals[1]), int32(vals[2])
			if len(vals) == 4 {
				wantA = int32(vals[3])
			}
			if wantRoots > wantClauses || wantA > wantRoots {
				return nil, fmt.Errorf("proof: line %d: inconsistent counts", lineNo)
			}
			continue
		}
		if p == nil {
			return nil, fmt.Errorf("proof: line %d: clause before header", lineNo)
		}
		vals := make([]int32, 0, len(fields))
		terminated := false
		for _, field := range fields {
			v, e := strconv.ParseInt(field, 10, 32)
			if e != nil {
				return nil, fmt.Errorf("proof: line %d: invalid literal %q", lineNo, field)
			}
			if v == 0 {
				terminated = true
				break
			}
			vals = append(vals, int32(v))
		}
		if !terminated {
			return nil, fmt.Errorf("proof: line %d: missing zero clause terminator", lineNo)
		}
		lits := make([]Lit, len(vals))
		for i, v := range vals {
			if int64(v) > int64(p.nVars) || int64(v) < -int64(p.nVars) {
				return nil, fmt.Errorf("proof: line %d: literal %d exceeds declared variables", lineNo, v)
			}
			lits[i] = LitFromDIMACS(v)
		}
		root := int32(0)
		if seen < wantRoots {
			root = 1
		}
		partA := int32(0)
		if seen < wantA {
			partA = 1
		}
		if Pr_ManAddClause(p, lits, root, partA) == 0 {
			return nil, fmt.Errorf("proof: line %d: %w", lineNo, p.LastError())
		}
		seen++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("proof: missing problem header")
	}
	if seen != wantClauses {
		return nil, fmt.Errorf("proof: header declares %d clauses, read %d", wantClauses, seen)
	}
	p.timeRead = time.Since(started)
	return p, nil
}
func Pr_ManProofRead(fileName string) *Pr_Man_t {
	p, err := ReadProofLog(fileName)
	if err != nil {
		return nil
	}
	return p
}
func Pr_ManProofTest(fileName string) int32 {
	p, err := ReadProofLog(fileName)
	if err != nil {
		return 0
	}
	defer Pr_ManFree(p)
	if Pr_ManProofWrite(p) == 0 {
		return 0
	}
	return 1
}

func Extra_PrintBinary_(w io.Writer, sign []uint32, nBits int32) {
	if w == nil || nBits <= 0 {
		return
	}
	for bit := nBits - 1; bit >= 0; bit-- {
		word := int(bit / 32)
		off := uint(bit % 32)
		ch := '0'
		if word < len(sign) && sign[word]&(uint32(1)<<off) != 0 {
			ch = '1'
		}
		_, _ = fmt.Fprintf(w, "%c", ch)
	}
}
func Pr_ManPrintInterOne(p *Pr_Man_t, c *Pr_Cls_t) {
	if p != nil && c != nil {
		fmt.Printf("Clause %2d :  %032b\n", c.Id, c.uTruth)
	}
}
func Pr_ManPrintClause(c *Pr_Cls_t) {
	if c == nil {
		return
	}
	fmt.Printf("Clause ID = %d. Proof = %d. %v\n", c.Id, c.pProof, c.pLits)
}
func Pr_ManPrintResolvent(lits []Lit) { fmt.Printf("Resolvent: %v\n", lits) }

func writeClauseLine(w io.Writer, id int32, lits []Lit) {
	fmt.Fprintf(w, "%d", id)
	for _, l := range lits {
		fmt.Fprintf(w, " %d", LitToDIMACS(l))
	}
	fmt.Fprintln(w, " 0 0")
}
