package csat

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glycerine/abcgo/sat/bsat2"
)

const (
	ABC_DEFAULT_CONF_LIMIT int64 = 0
	ABC_DEFAULT_IMP_LIMIT  int64 = 0
)

type gate struct {
	name   string
	typeID GateType
	inputs []string
	value  int
}

type managerState struct {
	gates            []gate
	byName           map[string]int
	targetNames      []string
	targetValues     []int
	finalized        bool
	prepared         bool
	mode             int
	timeLimit        time.Duration
	conflictLimit    int64
	implicationLimit int64
	backtrackCap     uint64
	inspectCap       uint64
	backtracks       uint64
	inspects         uint64
	dumpFile         string
	result           *CSAT_Target_ResultT
	lastError        error
}

func ABC_InitManager() ABC_Manager {
	return &ABC_ManagerStruct_t{state: &managerState{
		byName: make(map[string]int), conflictLimit: -1, implicationLimit: -1,
	}}
}

func (m ABC_Manager) valid() bool { return m != nil && m.state != nil }

func (m ABC_Manager) Release() {
	if m == nil {
		return
	}
	m.state = nil
}

func ABC_ReleaseManager(m ABC_Manager) { m.Release() }

func (m ABC_Manager) LastError() error {
	if !m.valid() {
		return fmt.Errorf("csat: invalid manager")
	}
	return m.state.lastError
}

func (m ABC_Manager) setError(err error) {
	if m.valid() {
		m.state.lastError = err
	}
}

func ABC_SetSolveOption(m ABC_Manager, option CSAT_OptionT) {
	if !m.valid() {
		return
	}
	// The native solver has one learning strategy; the option remains accepted
	// for compatibility with clients of the original API.
	_ = option
}

func ABC_UseOnlyCoreSatSolver(m ABC_Manager) {
	if m.valid() {
		m.state.mode = 1
	}
}

// ABC_AddGate adds a uniquely named gate. Fanins must already exist, as in the
// original network-construction API. Invalid input returns 0 and sets LastError.
func ABC_AddGate(m ABC_Manager, typ GateType, name string, nofi int32, fanins []string, dcAttr int32) int32 {
	if !m.valid() {
		return 0
	}
	s := m.state
	s.lastError = nil
	if s.finalized {
		return failAdd(m, "cannot add gates after network finalization")
	}
	if name == "" {
		return failAdd(m, "gate name is empty")
	}
	if _, exists := s.byName[name]; exists {
		return failAdd(m, "duplicate gate name %q", name)
	}
	if int(nofi) != len(fanins) || nofi < 0 {
		return failAdd(m, "gate %q declares %d fanins, received %d", name, nofi, len(fanins))
	}
	for _, input := range fanins {
		if _, exists := s.byName[input]; !exists {
			return failAdd(m, "fanin %q of gate %q does not exist", input, name)
		}
	}
	if err := validateGate(typ, len(fanins), dcAttr); err != nil {
		m.setError(fmt.Errorf("gate %q: %w", name, err))
		return 0
	}
	g := gate{name: name, typeID: typ, inputs: append([]string(nil), fanins...)}
	if typ == CSAT_CONST {
		g.value = 1
	}
	s.byName[name] = len(s.gates)
	s.gates = append(s.gates, g)
	return 1
}

func failAdd(m ABC_Manager, format string, args ...any) int32 {
	m.setError(fmt.Errorf(format, args...))
	return 0
}

func validateGate(typ GateType, n int, dcAttr int32) error {
	switch typ {
	case CSAT_CONST, CSAT_BPI, CSAT_BPPI:
		if n != 0 {
			return fmt.Errorf("expected no fanins")
		}
	case CSAT_BAND, CSAT_BNAND, CSAT_BOR, CSAT_BNOR:
		if n < 1 {
			return fmt.Errorf("expected one or more fanins")
		}
	case CSAT_BXOR, CSAT_BXNOR:
		if n < 1 || n > 2 {
			return fmt.Errorf("expected one or two fanins")
		}
	case CSAT_BINV, CSAT_BBUF:
		if n != 1 {
			return fmt.Errorf("expected one fanin")
		}
	case CSAT_BPPO, CSAT_BPO:
		if n != 1 {
			return fmt.Errorf("expected one fanin")
		}
	case CSAT_BDC:
		return fmt.Errorf("don't-care gates are not supported by the original CSAT network builder")
	default:
		return fmt.Errorf("unsupported gate type %d", typ)
	}
	return nil
}

func ABC_Network_Finalize(m ABC_Manager) {
	if !m.valid() {
		return
	}
	m.state.finalized = true
	m.state.lastError = nil
}

// ABC_Check_Integrity reports whether every non-input internal gate reaches an
// output, and whether the network's references are well-formed.
func ABC_Check_Integrity(m ABC_Manager) int32 {
	if !m.valid() || len(m.state.gates) == 0 {
		return 0
	}
	s := m.state
	used := make([]bool, len(s.gates))
	var visit func(int)
	visit = func(i int) {
		if used[i] {
			return
		}
		used[i] = true
		for _, name := range s.gates[i].inputs {
			visit(s.byName[name])
		}
	}
	for _, g := range s.gates {
		if g.typeID == CSAT_BPO || g.typeID == CSAT_BPPO {
			visit(s.byName[g.name])
		}
	}
	for i, g := range s.gates {
		if g.typeID != CSAT_CONST && g.typeID != CSAT_BPI && g.typeID != CSAT_BPPI && !used[i] {
			return 0
		}
	}
	return 1
}

func ABC_SetTimeLimit(m ABC_Manager, seconds int32) {
	if m.valid() {
		if seconds <= 0 {
			m.state.timeLimit = 0
		} else {
			m.state.timeLimit = time.Duration(seconds) * time.Second
		}
	}
}

func ABC_SetLearnLimit(m ABC_Manager, n int32) {
	if m.valid() {
		if n <= 0 {
			m.state.conflictLimit = -1
		} else {
			m.state.conflictLimit = int64(n)
		}
	}
}
func ABC_SetLearnBacktrackLimit(m ABC_Manager, n int32) { ABC_SetLearnLimit(m, n) }
func ABC_SetSolveBacktrackLimit(m ABC_Manager, n int32) { ABC_SetLearnLimit(m, n) }
func ABC_SetSolveImplicationLimit(m ABC_Manager, n int32) {
	if m.valid() {
		m.state.implicationLimit = int64(n)
		if n <= 0 {
			m.state.implicationLimit = -1
		}
	}
}
func ABC_SetTotalBacktrackLimit(m ABC_Manager, n ABC_UINT64_T) {
	if m.valid() {
		m.state.backtrackCap = n
	}
}
func ABC_SetTotalInspectLimit(m ABC_Manager, n ABC_UINT64_T) {
	if m.valid() {
		m.state.inspectCap = n
	}
}
func ABC_GetTotalBacktracksMade(m ABC_Manager) ABC_UINT64_T {
	if !m.valid() {
		return 0
	}
	return m.state.backtracks
}
func ABC_GetTotalInspectsMade(m ABC_Manager) ABC_UINT64_T {
	if !m.valid() {
		return 0
	}
	return m.state.inspects
}
func ABC_EnableDump(m ABC_Manager, file string) {
	if m.valid() {
		m.state.dumpFile = file
	}
}

func ABC_AddTarget(m ABC_Manager, n int32, names []string, values []int32) int32 {
	if !m.valid() {
		return 0
	}
	s := m.state
	s.lastError = nil
	if n < 1 || int(n) != len(names) || len(names) != len(values) {
		m.setError(fmt.Errorf("target requires matching non-empty name and value slices"))
		return 0
	}
	newNames := append([]string(nil), names...)
	newValues := make([]int, len(values))
	for i, name := range names {
		if _, exists := s.byName[name]; !exists {
			m.setError(fmt.Errorf("target gate %q does not exist", name))
			return 0
		}
		if values[i] != 0 && values[i] != 1 {
			m.setError(fmt.Errorf("target value for %q must be 0 or 1", name))
			return 0
		}
		newValues[i] = int(values[i])
	}
	s.targetNames, s.targetValues = newNames, newValues
	s.prepared, s.result = false, nil
	return 1
}

func ABC_SolveInit(m ABC_Manager) {
	if !m.valid() {
		return
	}
	s := m.state
	s.prepared = len(s.targetNames) > 0 && len(s.targetNames) == len(s.targetValues)
	s.result = nil
	if !s.prepared {
		s.lastError = fmt.Errorf("target is not specified")
	}
}

func ABC_AnalyzeTargets(m ABC_Manager) { _ = m }

func ABC_Solve(m ABC_Manager) CSAT_StatusT {
	if !m.valid() || !m.state.prepared {
		return UNDETERMINED
	}
	s := m.state
	started := time.Now()
	res := &CSAT_Target_ResultT{}
	vars := make([]bsat2.Var, len(s.gates))
	var solver interface {
		NewVariable() bsat2.Var
		AddClause(...bsat2.Lit) bool
		SolveLimited(...bsat2.Lit) bsat2.LBool
		ModelValue(bsat2.Lit) bsat2.LBool
	}
	var core *bsat2.Solver
	if s.mode == 1 {
		core = bsat2.NewSolver()
		solver = core
	} else {
		simplifier := bsat2.NewSimpSolver()
		core = simplifier.Solver
		solver = simplifier
	}
	for i := range vars {
		vars[i] = solver.NewVariable()
	}
	add := func(lits ...bsat2.Lit) bool { return solver.AddClause(lits...) }
	lit := func(i int, neg bool) bsat2.Lit { return bsat2.MakeLit(vars[i], neg) }
	constantOK := true
	for i, g := range s.gates {
		out := lit(i, false)
		in := make([]bsat2.Lit, len(g.inputs))
		for j, name := range g.inputs {
			in[j] = lit(s.byName[name], false)
		}
		switch g.typeID {
		case CSAT_CONST:
			constantOK = constantOK && add(out)
		case CSAT_BPI, CSAT_BPPI:
		case CSAT_BAND:
			for _, x := range in {
				constantOK = constantOK && add(out.Not(), x)
			}
			clause := append([]bsat2.Lit{out}, negate(in)...)
			constantOK = constantOK && add(clause...)
		case CSAT_BNAND:
			for _, x := range in {
				constantOK = constantOK && add(out, x)
			}
			clause := append([]bsat2.Lit{out.Not()}, negate(in)...)
			constantOK = constantOK && add(clause...)
		case CSAT_BOR:
			for _, x := range in {
				constantOK = constantOK && add(out, x.Not())
			}
			clause := append([]bsat2.Lit{out.Not()}, in...)
			constantOK = constantOK && add(clause...)
		case CSAT_BNOR:
			for _, x := range in {
				constantOK = constantOK && add(out.Not(), x.Not())
			}
			clause := append([]bsat2.Lit{out}, in...)
			constantOK = constantOK && add(clause...)
		case CSAT_BXOR, CSAT_BXNOR:
			x := in[0]
			if len(in) == 1 {
				if g.typeID == CSAT_BXNOR {
					x = x.Not()
				}
				constantOK = constantOK && equivalent(solver, out, x)
			} else {
				y := in[1]
				if g.typeID == CSAT_BXNOR {
					out = out.Not()
				}
				constantOK = constantOK && xor(solver, out, x, y)
			}
		case CSAT_BINV:
			constantOK = constantOK && equivalent(solver, out, in[0].Not())
		case CSAT_BBUF, CSAT_BPPO, CSAT_BPO:
			constantOK = constantOK && equivalent(solver, out, in[0])
		case CSAT_BDC:
			// The legacy two-input don't-care primitive leaves its output
			// unconstrained; this preserves the relation used by ABC's SOP gate.
		}
	}
	if !constantOK {
		res.Status, res.Time = UNSATISFIABLE, time.Since(started).Seconds()
		s.result = res
		return res.Status
	}
	for i, name := range s.targetNames {
		targetLit := lit(s.byName[name], s.targetValues[i] == 0)
		if !add(targetLit) {
			res.Status, res.Time = UNSATISFIABLE, time.Since(started).Seconds()
			s.result = res
			return res.Status
		}
	}
	if s.conflictLimit > 0 {
		core.SetConflictBudget(s.conflictLimit)
	}
	if s.backtrackCap > 0 {
		remaining := int64(0)
		if s.backtracks < s.backtrackCap {
			remaining = int64(s.backtrackCap - s.backtracks)
		}
		if s.conflictLimit <= 0 || remaining < s.conflictLimit {
			core.SetConflictBudget(remaining)
		}
	}
	propagationLimit := s.implicationLimit
	if s.inspectCap > 0 {
		remaining := int64(0)
		if s.inspects < s.inspectCap {
			remaining = int64(s.inspectCap - s.inspects)
		}
		if propagationLimit < 0 || remaining < propagationLimit {
			propagationLimit = remaining
		}
	}
	if propagationLimit >= 0 {
		core.SetPropagationBudget(propagationLimit)
	}
	if s.timeLimit > 0 {
		timer := time.AfterFunc(s.timeLimit, core.Interrupt)
		defer timer.Stop()
	}
	answer := solver.SolveLimited()
	res.NumDec = int32(core.Decisions)
	res.NumImp = int32(core.Propagations)
	res.NumCftg = int32(core.Conflicts)
	res.NumCfts = int32(core.ClauseLiterals)
	res.Time = time.Since(started).Seconds()
	s.backtracks += core.Conflicts
	s.inspects += core.Propagations
	if s.backtrackCap > 0 && s.backtracks > s.backtrackCap || s.inspectCap > 0 && s.inspects > s.inspectCap {
		res.Status = TIME_OUT
	} else {
		switch answer {
		case bsat2.True:
			res.Status = SATISFIABLE
			for i, g := range s.gates {
				if g.typeID == CSAT_BPI || g.typeID == CSAT_BPPI {
					res.Names = append(res.Names, g.name)
					v := int32(0)
					if solver.ModelValue(lit(i, false)) == bsat2.True {
						v = 1
					}
					res.Values = append(res.Values, v)
				}
			}
			res.NoSig = int32(len(res.Names))
		case bsat2.False:
			res.Status = UNSATISFIABLE
		default:
			res.Status = TIME_OUT
		}
	}
	s.result = res
	if s.dumpFile != "" {
		_ = dumpBENCH(s, s.dumpFile)
	}
	return res.Status
}

func negate(in []bsat2.Lit) []bsat2.Lit {
	out := make([]bsat2.Lit, len(in))
	for i, l := range in {
		out[i] = l.Not()
	}
	return out
}

type clauseAdder interface {
	AddClause(...bsat2.Lit) bool
}

func equivalent(s clauseAdder, a, b bsat2.Lit) bool {
	return s.AddClause(a.Not(), b) && s.AddClause(a, b.Not())
}
func xor(s clauseAdder, z, a, b bsat2.Lit) bool {
	return s.AddClause(a, b, z.Not()) && s.AddClause(a.Not(), b.Not(), z.Not()) && s.AddClause(a, b.Not(), z) && s.AddClause(a.Not(), b, z)
}

func ABC_Get_Target_Result(m ABC_Manager, targetID int32) *CSAT_Target_ResultT {
	if !m.valid() {
		return nil
	}
	_ = targetID // The original API ignores this argument and returns its last result.
	return cloneResult(m.state.result)
}
func cloneResult(r *CSAT_Target_ResultT) *CSAT_Target_ResultT {
	if r == nil {
		return nil
	}
	copyResult := *r
	copyResult.Names = append([]string(nil), r.Names...)
	copyResult.Values = append([]int32(nil), r.Values...)
	return &copyResult
}
func ABC_TargetResAlloc(nVars int32) *CSAT_Target_ResultT {
	if nVars < 0 {
		nVars = 0
	}
	return &CSAT_Target_ResultT{NoSig: nVars, Names: make([]string, 0, nVars), Values: make([]int32, 0, nVars)}
}
func ABC_TargetResFree(p *CSAT_Target_ResultT) {
	if p == nil {
		return
	}
	p.Names, p.Values = nil, nil
}

func ABC_Dump_Bench_File(m ABC_Manager) {
	if !m.valid() {
		return
	}
	file := m.state.dumpFile
	if file == "" {
		file = "abc_test.bench"
	}
	if err := dumpBENCH(m.state, file); err != nil {
		m.state.lastError = err
	}
}
func dumpBENCH(s *managerState, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	for _, g := range s.gates {
		switch g.typeID {
		case CSAT_BPI, CSAT_BPPI:
			_, err = fmt.Fprintf(w, "INPUT(%s)\n", g.name)
		case CSAT_BPO, CSAT_BPPO:
			_, err = fmt.Fprintf(w, "OUTPUT(%s)\n", g.inputs[0])
		case CSAT_CONST:
			_, err = fmt.Fprintf(w, "%s = 1\n", g.name)
		default:
			op := benchOp(g.typeID)
			if op == "" {
				continue
			}
			_, err = fmt.Fprintf(w, "%s = %s(%s)\n", g.name, op, strings.Join(g.inputs, ", "))
		}
		if err != nil {
			break
		}
	}
	if flushErr := w.Flush(); err == nil {
		err = flushErr
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}
func benchOp(typ GateType) string {
	switch typ {
	case CSAT_CONST:
		return "CONST1"
	case CSAT_BAND:
		return "AND"
	case CSAT_BNAND:
		return "NAND"
	case CSAT_BOR:
		return "OR"
	case CSAT_BNOR:
		return "NOR"
	case CSAT_BXOR:
		return "XOR"
	case CSAT_BXNOR:
		return "XNOR"
	case CSAT_BINV:
		return "NOT"
	case CSAT_BBUF:
		return "BUFF"
	default:
		return ""
	}
}
