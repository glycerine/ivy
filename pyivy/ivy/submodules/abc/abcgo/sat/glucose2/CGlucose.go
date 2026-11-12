package glucose2

// Circuit reasoning is available in this port, as in CGLUCOSE_EXP builds.
const CircuitGlucoseEnabled = true

type JustData struct {
	Now         bool
	ActiveFanin float64
}
type justWatch struct {
	direction        bool
	head, next, prev Var
}
type circuitState struct {
	fanins               [][2]Lit
	fanouts              [][]Lit // each edge encodes its gate variable and fanin index
	refs                 []*Clause
	data                 []JustData
	watches              []justWatch
	marks                []uint64
	round, previousRound uint64
	heap                 *Heap
	stack                []Var
}

func (s *Solver) ensureCircuit() *circuitState {
	if s.circuit == nil {
		c := new(circuitState)
		s.circuit = c
		c.heap = NewHeap(func(a, b int) bool {
			x, y := c.data[a].ActiveFanin, c.data[b].ActiveFanin
			if x != y {
				return x > y
			}
			if s.levels[a] != s.levels[b] {
				return s.levels[a] < s.levels[b]
			}
			return a > b
		})
	}
	c := s.circuit
	for len(c.fanins) < s.NVars() {
		v := Var(len(c.fanins))
		c.fanins = append(c.fanins, [2]Lit{LitError, LitError})
		c.fanouts = append(c.fanouts, nil)
		c.refs = append(c.refs, &Clause{gateReason: true, gate: v})
		c.data = append(c.data, JustData{})
		c.watches = append(c.watches, justWatch{head: VarUndef, next: VarUndef, prev: VarUndef})
		c.marks = append(c.marks, 0)
	}
	return c
}

// SetVarFaninLits encodes AND by ascending fanin variable order and XOR by
// descending order. Literal phases are independent of this gate-kind encoding.
func (s *Solver) SetVarFaninLits(v Var, a, b Lit) {
	if v < 0 || int(v) >= s.NVars() || a < 0 || b < 0 || int(a.Var()) >= s.NVars() || int(b.Var()) >= s.NVars() || a.Var() == b.Var() || v == a.Var() || v == b.Var() {
		panic("invalid two-fanin gate")
	}
	c := s.ensureCircuit()
	old := c.fanins[v]
	// Replacing an installed gate must unlink its old fanouts; the original C
	// forward-only list corrupts itself if the public setter is called twice.
	if old[0] != LitError {
		for i, p := range old {
			edges := c.fanouts[p.Var()]
			for k, e := range edges {
				if e == MkLit(v, i != 0) {
					copy(edges[k:], edges[k+1:])
					edges = edges[:len(edges)-1]
					break
				}
			}
			c.fanouts[p.Var()] = edges
		}
	}
	c.fanins[v] = [2]Lit{a, b}
	for i, p := range c.fanins[v] {
		c.fanouts[p.Var()] = append([]Lit{MkLit(v, i != 0)}, c.fanouts[p.Var()]...)
	}
}
func (s *Solver) SetNewRound() uint64 { c := s.ensureCircuit(); old := c.round; c.round++; return old }
func (s *Solver) MarkCone(v Var) {
	c := s.ensureCircuit()
	stack := []Var{v}
	for len(stack) > 0 {
		x := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if x < 0 || int(x) >= s.NVars() {
			panic("cone variable out of range")
		}
		if c.marks[x] >= c.round {
			continue
		}
		c.marks[x] = c.round
		if s.isTwoFanin(x) {
			stack = append(stack, c.fanins[x][0].Var(), c.fanins[x][1].Var())
		}
	}
}
func (s *Solver) IsRoundWatch(v Var) bool {
	return s.circuit != nil && int(v) < len(s.circuit.marks) && s.circuit.round == s.circuit.marks[v]
}
func (s *Solver) isTwoFanin(v Var) bool {
	return s.circuit != nil && v >= 0 && int(v) < len(s.circuit.fanins) && s.circuit.fanins[v][0] != LitError && s.circuit.fanins[v][1] != LitError
}
func (s *Solver) isAND(v Var) bool { f := s.circuit.fanins[v]; return f[0].Var() < f[1].Var() }
func (s *Solver) isJReason(v Var) bool {
	return s.isTwoFanin(v) && (s.ValueVar(v) == LFalse || !s.isAND(v) && !s.ValueVar(v).IsUndef())
}
func (s *Solver) VarActivity(v Var) float64  { return s.activity[v] }
func (s *Solver) JustActivity(v Var) float64 { return s.ensureCircuit().data[v].ActiveFanin }
func (s *Solver) GetJustData(v Var) JustData { return s.ensureCircuit().data[v] }
func (s *Solver) VarPolarity(v Var) int {
	if s.polarity[v] {
		return 1
	}
	return 0
}
func (s *Solver) JustUsage() int                      { return s.JustificationMode }
func (s *Solver) SatSolverSetVarFaninLit(v, a, b int) { s.SetVarFaninLits(Var(v), Lit(a), Lit(b)) }
func (s *Solver) SatSolverStartNewRound()             { s.SetNewRound() }
func (s *Solver) SatSolverMarkCone(v int)             { s.MarkCone(Var(v)) }
