package glucose2

// The circuit reasoning algorithms were developed by He-Teng Zhang (NTU).
// Gate reasons are tagged Go objects, replacing the original high-bit CRef tag.
func (s *Solver) gateClause(v, target Var, conflict bool) *Clause {
	f := s.circuit.fanins[v]
	a, b := f[0], f[1]
	out := MkLit(v, false)
	var lits []Lit
	if s.isAND(v) {
		if conflict {
			if s.ValueVar(v) == LFalse {
				lits = []Lit{out, a.Not(), b.Not()}
			} else {
				pick := b
				if s.ValueLit(a) == LFalse {
					pick = a
					if s.ValueLit(b) == LFalse && s.levels[b.Var()] <= s.levels[a.Var()] {
						pick = b
					}
				}
				lits = []Lit{out.Not(), pick}
			}
		} else if target == v {
			if s.ValueVar(v) == LFalse {
				pick := b
				if s.ValueLit(a) == LFalse {
					pick = a
					if s.ValueLit(b) == LFalse && s.circuit.watches[v].direction {
						pick = b
					}
				}
				lits = []Lit{out.Not(), pick}
			} else {
				lits = []Lit{out, a.Not(), b.Not()}
			}
		} else if s.ValueVar(v) == LFalse {
			lits = []Lit{a.Not(), b.Not(), out}
			if target == b.Var() {
				lits[0], lits[1] = lits[1], lits[0]
			}
		} else {
			pick := a
			if target == b.Var() {
				pick = b
			}
			lits = []Lit{pick, out.Not()}
		}
	} else {
		if conflict {
			lits = []Lit{MkLit(v, s.ValueVar(v) == LTrue), MkLit(a.Var(), s.ValueVar(a.Var()) == LTrue), MkLit(b.Var(), s.ValueVar(b.Var()) == LTrue)}
		} else if target == v {
			lits = []Lit{MkLit(v, s.ValueVar(v) == LFalse), MkLit(a.Var(), s.ValueVar(a.Var()) == LTrue), MkLit(b.Var(), s.ValueVar(b.Var()) == LTrue)}
		} else {
			first, second := a.Var(), b.Var()
			if target == second {
				first, second = second, first
			}
			lits = []Lit{MkLit(first, s.ValueVar(first) == LFalse), MkLit(second, s.ValueVar(second) == LTrue), MkLit(v, s.ValueVar(v) == LTrue)}
		}
	}
	return &Clause{Lits: lits}
}
func (s *Solver) reasonClause(p Lit) *Clause {
	r := s.reasons[p.Var()]
	if r != nil && r.gateReason {
		return s.gateClause(r.gate, p.Var(), false)
	}
	return r
}
func (s *Solver) conflictClause(r *Clause) *Clause {
	if r != nil && r.gateReason {
		return s.gateClause(r.gate, VarUndef, true)
	}
	return r
}
func (s *Solver) gatePropagateCheckThis(v Var) *Clause {
	c := s.circuit
	f := c.fanins[v]
	a, b := f[0], f[1]
	va, vb, vo := s.ValueLit(a), s.ValueLit(b), s.ValueVar(v)
	r := c.refs[v]
	if s.isAND(v) {
		if vo == LFalse {
			if va == LTrue && vb == LTrue {
				return r
			}
			if va == LFalse || vb == LFalse {
				return nil
			}
			if va == LTrue {
				s.uncheckedEnqueue(b.Not(), r)
			}
			if vb == LTrue {
				s.uncheckedEnqueue(a.Not(), r)
			}
		} else {
			var conflict *Clause
			if va == LFalse || vb == LFalse {
				conflict = r
			}
			if va.IsUndef() {
				s.uncheckedEnqueue(a, r)
			}
			if vb.IsUndef() {
				s.uncheckedEnqueue(b, r)
			}
			return conflict
		}
	} else {
		if va.IsUndef() && vb.IsUndef() {
			return nil
		}
		if va.IsUndef() {
			s.uncheckedEnqueue(a.Not().Xor((vb == LTrue) != (vo == LTrue)), r)
		} else if vb.IsUndef() {
			s.uncheckedEnqueue(b.Not().Xor((va == LTrue) != (vo == LTrue)), r)
		} else if vo.Xor(va.Equal(vb)) == LFalse {
			return r
		}
	}
	return nil
}
func (s *Solver) gatePropagateCheckFanout(v Var, edge Lit) *Clause {
	c := s.circuit
	gate := edge.Var()
	f := c.fanins[gate]
	a, b := f[0], f[1]
	if edge.Sign() {
		a, b = b, a
	}
	va, vb, vo := s.ValueLit(a), s.ValueLit(b), s.ValueVar(gate)
	r := c.refs[gate]
	if s.isAND(gate) {
		if va == LFalse {
			if vo == LFalse {
				return nil
			}
			if vo == LTrue {
				return r
			}
			c.watches[gate].direction = edge.Sign()
			s.uncheckedEnqueue(MkLit(gate, true), r)
		} else {
			if vo == LTrue {
				return nil
			}
			if vo == LFalse {
				if vb == LFalse {
					return nil
				}
				if vb == LTrue {
					return r
				}
				s.uncheckedEnqueue(b.Not(), r)
			} else if vb == LTrue {
				s.uncheckedEnqueue(MkLit(gate, false), r)
			}
		}
	} else {
		if vb.IsUndef() && vo.IsUndef() {
			return nil
		}
		if vb.IsUndef() {
			s.uncheckedEnqueue(b.Not().Xor((va == LTrue) != (vo == LTrue)), r)
		} else if vo.IsUndef() {
			s.uncheckedEnqueue(MkLit(gate, (va == LTrue) != (vb == LTrue)).Not(), r)
		} else if vo.Xor(va.Equal(vb)) == LFalse {
			return r
		}
	}
	return nil
}
func (s *Solver) gatePropagate(p Lit) *Clause {
	if s.JustificationMode < 2 || !s.IsRoundWatch(p.Var()) {
		return nil
	}
	var conflict *Clause
	if s.isTwoFanin(p.Var()) {
		conflict = s.gatePropagateCheckThis(p.Var())
		if conflict != nil && s.ValueVar(p.Var()) == LTrue {
			return conflict
		}
	}
	for _, edge := range s.circuit.fanouts[p.Var()] {
		if !s.IsRoundWatch(edge.Var()) {
			continue
		}
		if r := s.gatePropagateCheckFanout(p.Var(), edge); r != nil {
			conflict = r
			if s.ValueVar(edge.Var()) == LTrue {
				return conflict
			}
		}
	}
	return conflict
}
func (s *Solver) gateJustFanin(v Var) Lit {
	f := s.circuit.fanins[v]
	a, b := f[0], f[1]
	va, vb := s.ValueLit(a), s.ValueLit(b)
	if s.isAND(v) {
		if va == LFalse || vb == LFalse {
			return LitUndef
		}
		if va == LTrue {
			return b.Not()
		}
		if vb == LTrue {
			return a.Not()
		}
		a, b = a.Not(), b.Not()
	} else {
		if !va.IsUndef() && !vb.IsUndef() {
			return LitUndef
		}
		a, b = MkLit(a.Var(), s.polarity[a.Var()]), MkLit(b.Var(), s.polarity[b.Var()])
	}
	if s.activity[a.Var()] < s.activity[b.Var()] {
		return b
	}
	return a
}
func (s *Solver) pushJustQueue(v Var) {
	if !s.IsRoundWatch(v) {
		return
	}
	c := s.circuit
	f := c.fanins[v]
	a, b := s.activity[f[0].Var()], s.activity[f[1].Var()]
	if b > a {
		a = b
	}
	c.data[v] = JustData{true, a}
	c.heap.Update(int(v))
}
func (s *Solver) addJwatch(host, member Var) {
	c := s.circuit
	if c.watches[member].next != VarUndef || c.watches[member].prev != VarUndef {
		panic("duplicate justification watch")
	}
	head := c.watches[host].head
	if head != VarUndef {
		c.watches[head].prev = member
	}
	c.watches[member].next = head
	c.watches[member].prev = host
	c.watches[host].head = member
}
func (s *Solver) delJwatch(member Var) {
	c := s.circuit
	prev, next := c.watches[member].prev, c.watches[member].next
	if prev == VarUndef {
		panic("missing justification watch")
	}
	if c.watches[prev].next == member {
		c.watches[prev].next = next
	} else {
		c.watches[prev].head = next
	}
	if next != VarUndef {
		c.watches[next].prev = prev
	}
	c.watches[member].prev = VarUndef
	c.watches[member].next = VarUndef
}
func (s *Solver) gateAddJwatch(v Var) {
	c := s.circuit
	if c.watches[v].prev != VarUndef {
		return
	}
	f := c.fanins[v]
	a, b := f[0], f[1]
	va, vb := s.ValueLit(a), s.ValueLit(b)
	host := a.Var()
	if va == LFalse && vb == LFalse || !s.isAND(v) {
		if s.levels[a.Var()] >= s.levels[b.Var()] {
			host = b.Var()
		}
	} else if va != LFalse {
		host = b.Var()
	}
	s.addJwatch(host, v)
}
func (s *Solver) gateClearJwatch(v Var, backtrack int) {
	c := s.circuit
	for member := c.watches[v].head; member != VarUndef; {
		next := c.watches[member].next
		s.delJwatch(member)
		if s.levels[member] <= backtrack {
			s.pushJustQueue(member)
		}
		member = next
	}
}
func (s *Solver) pickJustLit() (Lit, Var) {
	c := s.circuit
	for _, v := range c.stack {
		f := c.fanins[v]
		if s.ValueLit(f[0]).IsUndef() && s.ValueLit(f[1]).IsUndef() {
			s.pushJustQueue(v)
		}
	}
	c.stack = c.stack[:0]
	next := VarUndef
	for !c.heap.Empty() {
		next = Var(c.heap.RemoveMin())
		if !c.data[next].Now {
			continue
		}
		if p := s.gateJustFanin(next); p != LitUndef {
			return p, next
		}
		s.gateAddJwatch(next)
	}
	return LitUndef, next
}
func (s *Solver) updateJustActivity(v Var) {
	if s.circuit == nil {
		return
	}
	c := s.circuit
	for _, edge := range c.fanouts[v] {
		x := edge.Var()
		if c.data[x].Now && c.heap.InHeap(int(x)) {
			f := c.fanins[x]
			a, b := s.activity[f[0].Var()], s.activity[f[1].Var()]
			if b > a {
				a = b
			}
			c.data[x].ActiveFanin = a
			c.heap.Update(int(x))
		}
	}
}
func (s *Solver) ResetJustData() {
	if s.circuit == nil {
		return
	}
	s.circuit.stack = s.circuit.stack[:0]
	s.circuit.heap.Clear()
}
func (s *Solver) JustCheck() bool {
	for _, p := range s.trail {
		v := p.Var()
		if s.IsRoundWatch(v) && s.isJReason(v) && s.gateJustFanin(v) != LitUndef {
			return false
		}
	}
	return true
}
