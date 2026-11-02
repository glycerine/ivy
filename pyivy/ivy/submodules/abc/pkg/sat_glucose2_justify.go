package abc

// Port of the circuit-justification extension (CGLUCOSE_EXP) of src/sat/glucose2:
// the declarations in Solver.h and the definitions in CGlucoseCore.h (namespace Gluco2).
// The commented-out Solver::delVarFaninLits of CGlucoseCore.h is not ported.

// gluco2LitNone is toLit(~0), the "undefined" literal marker of the fanin tables.
var gluco2LitNone = Gluco2ToLit(-1)

// Gluco2SolverJustData is Solver::JustData.
type Gluco2SolverJustData struct {
	now       uint32 // bit-field of 1 bit
	act_fanin float64
}

// gluco2SolverMkJustData is Solver::mkJustData (protected static inline).
// C: static inline JustData mkJustData(bool n)
func gluco2SolverMkJustData(n bool) Gluco2SolverJustData {
	d := Gluco2SolverJustData{act_fanin: 0}
	if n {
		d.now = 1
	}
	return d
}

// Gluco2SolverJustOrderLt is Solver::JustOrderLt.
type Gluco2SolverJustOrderLt struct {
	pS *Gluco2Solver // const Solver *
}

// NewGluco2SolverJustOrderLt is JustOrderLt(const Solver * _pS).
// C: JustOrderLt(const Solver * _pS)
func NewGluco2SolverJustOrderLt(pS *Gluco2Solver) Gluco2SolverJustOrderLt {
	return Gluco2SolverJustOrderLt{pS: pS}
}

// Call is JustOrderLt::operator().
// C: bool operator()(Var x, Var y) const
func (l Gluco2SolverJustOrderLt) Call(x int, y int) bool {
	if l.pS.JustActivity(x) != l.pS.JustActivity(y) {
		return l.pS.JustActivity(x) > l.pS.JustActivity(y)
	}
	if l.pS.Level(x) != l.pS.Level(y) {
		return l.pS.Level(x) < l.pS.Level(y)
	}
	return x > y
}

// Gluco2SolverJustWatch is Solver::JustWatch.
type Gluco2SolverJustWatch struct {
	header_dir uint32 // header.dir: 1 bit (anonymous struct member `header`)
	head       Gluco2Var
	next       Gluco2Var
	prev       Gluco2Var
}

// gluco2SolverMkJustWatch is Solver::mkJustWatch (protected static inline).
// C: static inline JustWatch mkJustWatch()
func gluco2SolverMkJustWatch() Gluco2SolverJustWatch {
	return Gluco2SolverJustWatch{header_dir: 0, head: Gluco2VarUndef, next: Gluco2VarUndef, prev: Gluco2VarUndef}
}

// JustUsage is Solver::justUsage (public).
// C: int justUsage() const
func (s *Gluco2Solver) JustUsage() int { return s.Jftr }

// isTwoFanin is Solver::isTwoFanin: this var has two fanins.
// C: bool isTwoFanin(Var v) const
func (s *Gluco2Solver) isTwoFanin(v Gluco2Var) bool {
	if v >= s.NVars() {
		panic("Gluco2Solver.isTwoFanin: v >= nVars()")
	}
	if s.var2FaninLits.Size() <= (v<<1)+1 {
		return false
	}
	return *s.var2FaninLits.At((v << 1) + 0) != gluco2LitNone && *s.var2FaninLits.At((v << 1) + 1) != gluco2LitNone
}

// isAND is Solver::isAND.
// C: bool isAND(Var v) const
func (s *Gluco2Solver) isAND(v Gluco2Var) bool { return s.getFaninVar0(v) < s.getFaninVar1(v) }

// isJReason is Solver::isJReason.
// C: bool isJReason(Var v) const
func (s *Gluco2Solver) isJReason(v Gluco2Var) bool {
	return s.isTwoFanin(v) && (Gluco2LFalse == s.Value(v) || (!s.isAND(v) && Gluco2LUndef != s.Value(v)))
}

// getFaninLit0 is Solver::getFaninLit0.
// C: Lit getFaninLit0(Var v) const
func (s *Gluco2Solver) getFaninLit0(v Gluco2Var) Gluco2Lit { return *s.var2FaninLits.At((v << 1) + 0) }

// getFaninLit1 is Solver::getFaninLit1.
// C: Lit getFaninLit1(Var v) const
func (s *Gluco2Solver) getFaninLit1(v Gluco2Var) Gluco2Lit { return *s.var2FaninLits.At((v << 1) + 1) }

// getFaninC0 is Solver::getFaninC0.
// C: bool getFaninC0(Var v) const
func (s *Gluco2Solver) getFaninC0(v Gluco2Var) bool { return Gluco2Sign(s.getFaninLit0(v)) }

// getFaninC1 is Solver::getFaninC1.
// C: bool getFaninC1(Var v) const
func (s *Gluco2Solver) getFaninC1(v Gluco2Var) bool { return Gluco2Sign(s.getFaninLit1(v)) }

// getFaninVar0 is Solver::getFaninVar0.
// C: Var getFaninVar0(Var v) const
func (s *Gluco2Solver) getFaninVar0(v Gluco2Var) Gluco2Var { return Gluco2LitVar(s.getFaninLit0(v)) }

// getFaninVar1 is Solver::getFaninVar1.
// C: Var getFaninVar1(Var v) const
func (s *Gluco2Solver) getFaninVar1(v Gluco2Var) Gluco2Var { return Gluco2LitVar(s.getFaninLit1(v)) }

// getFaninPlt0 is Solver::getFaninPlt0.
// C: Lit getFaninPlt0(Var v) const
func (s *Gluco2Solver) getFaninPlt0(v Gluco2Var) Gluco2Lit {
	return Gluco2MkLit(s.getFaninVar0(v), *s.polarity.At(s.getFaninVar0(v)) == 1)
}

// getFaninPlt1 is Solver::getFaninPlt1.
// C: Lit getFaninPlt1(Var v) const
func (s *Gluco2Solver) getFaninPlt1(v Gluco2Var) Gluco2Lit {
	return Gluco2MkLit(s.getFaninVar1(v), *s.polarity.At(s.getFaninVar1(v)) == 1)
}

// maxActiveLit is Solver::maxActiveLit.
// C: Lit maxActiveLit(Lit lit0, Lit lit1) const
func (s *Gluco2Solver) maxActiveLit(lit0 Gluco2Lit, lit1 Gluco2Lit) Gluco2Lit {
	if *s.activity.At(Gluco2LitVar(lit0)) < *s.activity.At(Gluco2LitVar(lit1)) {
		return lit1
	}
	return lit0
}

// Var2CRef is Solver::Var2CRef.
// C: CRef Var2CRef(Var v) const
func (s *Gluco2Solver) Var2CRef(v Gluco2Var) Gluco2CRef { return uint32(v) | (1 << 31) }

// CRef2Var is Solver::CRef2Var.
// C: Var CRef2Var(CRef cr) const
func (s *Gluco2Solver) CRef2Var(cr Gluco2CRef) Gluco2Var { return Gluco2Var(cr &^ (1 << 31)) }

// isGateCRef is Solver::isGateCRef.
// C: bool isGateCRef(CRef cr) const
func (s *Gluco2Solver) isGateCRef(cr Gluco2CRef) bool {
	return Gluco2CRefUndef != cr && 0 != (cr&(1<<31))
}

// SetNewRound is Solver::setNewRound (public).
// C: int setNewRound()
func (s *Gluco2Solver) SetNewRound() int {
	r := int(s.travId)
	s.travId++
	return r
}

// IsRoundWatch is Solver::isRoundWatch (public).
// C: bool isRoundWatch(Var v) const
func (s *Gluco2Solver) IsRoundWatch(v Gluco2Var) bool { return s.travId == int64(*s.var2TravId.At(v)) }

// GetJustData is Solver::getJustData (public).
// C: const JustData& getJustData(int v) const
func (s *Gluco2Solver) GetJustData(v int) *Gluco2SolverJustData { return s.jdata.At(v) }

// VarActivity is Solver::varActivity (public).
// C: double varActivity(int v) const
func (s *Gluco2Solver) VarActivity(v int) float64 { return *s.activity.At(v) }

// JustActivity is Solver::justActivity (public).
// C: double justActivity(int v) const
func (s *Gluco2Solver) JustActivity(v int) float64 { return s.jdata.At(v).act_fanin }

// VarPolarity is Solver::varPolarity (public).
// C: int varPolarity(int v)
func (s *Gluco2Solver) VarPolarity(v int) int {
	if *s.polarity.At(v) != 0 {
		return 1
	}
	return 0
}

// gateJustFanin is Solver::gateJustFanin (l_Undef=satisfied, 0/1 = fanin0/fanin1 requires justify).
// C: Lit gateJustFanin(Var v) const
func (s *Gluco2Solver) gateJustFanin(v Gluco2Var) Gluco2Lit {
	if v >= s.NVars() || !s.isJReason(v) {
		panic("Gluco2Solver.gateJustFanin: precondition")
	}
	val0 := s.ValueLit(s.getFaninLit0(v))
	val1 := s.ValueLit(s.getFaninLit1(v))
	if s.isAND(v) {
		if !(s.Value(v) != Gluco2LFalse || Gluco2LTrue != val0 || Gluco2LTrue != val1) {
			panic("Gluco2Solver.gateJustFanin: AND gate is not justified")
		}
		if val0 == Gluco2LFalse || val1 == Gluco2LFalse {
			return Gluco2LitUndef
		}
		if val0 == Gluco2LTrue {
			return Gluco2LitNot(s.getFaninLit1(v))
		}
		if val1 == Gluco2LTrue {
			return Gluco2LitNot(s.getFaninLit0(v))
		}
		return s.maxActiveLit(Gluco2LitNot(s.getFaninLit0(v)), Gluco2LitNot(s.getFaninLit1(v)))
	}
	// xor scope
	if !(s.Value(v) == Gluco2LUndef || s.Value(v) != Gluco2LFalse || val0 == val1) {
		panic("Gluco2Solver.gateJustFanin: XOR gate is not justified")
	}
	if Gluco2LUndef != val0 && Gluco2LUndef != val1 {
		return Gluco2LitUndef
	}
	if !(Gluco2LUndef == val0 && Gluco2LUndef == val1) {
		panic("Gluco2Solver.gateJustFanin: XOR fanins")
	}
	return s.maxActiveLit(s.getFaninPlt0(v), s.getFaninPlt1(v))
}

// pushJustQueue is Solver::pushJustQueue.
// C: void pushJustQueue(Var v)
func (s *Gluco2Solver) pushJustQueue(v Gluco2Var) {
	if v >= s.NVars() || !s.isJReason(v) {
		panic("Gluco2Solver.pushJustQueue: precondition")
	}
	if !s.IsRoundWatch(v) {
		return
	}
	a0 := *s.activity.At(s.getFaninVar0(v))
	a1 := *s.activity.At(s.getFaninVar1(v))
	if a0 > a1 {
		s.jdata.At(v).act_fanin = a0
	} else {
		s.jdata.At(v).act_fanin = a1
	}
	s.jdata.At(v).now = 1
	s.jheap.Update(v)
}

// resetJustData is Solver::ResetJustData.
// C: void ResetJustData(bool fCleanMemory)
func (s *Gluco2Solver) resetJustData(fCleanMemory bool) {
	s.jstack.Shrink_(s.jstack.Size())
	s.jheap.Clear(fCleanMemory)
}

// pickJustLit is Solver::pickJustLit.
// C: Lit pickJustLit(Var& j_reason)
func (s *Gluco2Solver) pickJustLit(jReason *Gluco2Var) Gluco2Lit {
	next := Gluco2VarUndef
	jlit := Gluco2LitUndef
	for i := 0; i < s.jstack.Size(); i++ {
		x := *s.jstack.At(i)
		if Gluco2LUndef != s.ValueLit(s.getFaninLit0(x)) || Gluco2LUndef != s.ValueLit(s.getFaninLit1(x)) {
			continue
		}
		s.pushJustQueue(x)
	}
	s.jstack.Shrink_(s.jstack.Size())
	for !s.jheap.Empty() {
		next = s.jheap.RemoveMin()
		if Gluco2VarUndef == next {
			break
		}
		if s.jdata.At(next).now == 0 {
			continue
		}
		if !s.isJReason(next) {
			panic("Gluco2Solver.pickJustLit: not a justification reason")
		}
		jlit = s.gateJustFanin(next)
		if Gluco2LitUndef != jlit {
			break
		}
		s.gateAddJwatch(next)
	}
	*jReason = next
	return jlit
}

// gateAddJwatch is Solver::gateAddJwatch.
// C: void gateAddJwatch(Var v)
func (s *Gluco2Solver) gateAddJwatch(v Gluco2Var) {
	if v >= s.NVars() || !s.isJReason(v) {
		panic("Gluco2Solver.gateAddJwatch: precondition")
	}
	if Gluco2VarUndef != s.jwatch.At(v).prev { // already in jwatch
		return
	}
	var0 := s.getFaninVar0(v)
	var1 := s.getFaninVar1(v)
	val0 := s.ValueLit(s.getFaninLit0(v))
	val1 := s.ValueLit(s.getFaninLit1(v))
	if !(!s.isAND(v) || Gluco2LFalse == val0 || Gluco2LFalse == val1) || !(s.isAND(v) || (Gluco2LUndef != val0 && Gluco2LUndef != val1)) {
		panic("Gluco2Solver.gateAddJwatch: gate values")
	}
	if (val0 == Gluco2LFalse && val1 == Gluco2LFalse) || !s.isAND(v) {
		if s.vardata.At(var0).level < s.vardata.At(var1).level {
			s.addJwatch(var0, v)
		} else {
			s.addJwatch(var1, v)
		}
		return
	}
	if Gluco2LFalse == val0 {
		s.addJwatch(var0, v)
	} else {
		s.addJwatch(var1, v)
	}
}

// gateClearJwatch is Solver::gateClearJwatch.
// C: void gateClearJwatch(Var v, int backtrack_level)
func (s *Gluco2Solver) gateClearJwatch(v Gluco2Var, backtrackLevel int) {
	if Gluco2VarUndef == s.jwatch.At(v).head {
		return
	}
	member := s.jwatch.At(v).head
	for Gluco2VarUndef != member {
		next := s.jwatch.At(member).next
		s.delJwatch(member)
		if s.vardata.At(member).level <= backtrackLevel {
			s.pushJustQueue(member)
		}
		member = next
	}
}

// updateJustActivity is Solver::updateJustActivity.
// C: void updateJustActivity(Var v)
func (s *Gluco2Solver) updateJustActivity(v Gluco2Var) {
	for lfo := *s.var2Fanout0.At(v); lfo != Gluco2ToLit(^0); lfo = *s.var2FanoutN.At(Gluco2ToIntLit(lfo)) {
		x := Gluco2LitVar(lfo)
		if s.jdata.At(x).now != 0 && s.jheap.InHeap(x) {
			a0 := *s.activity.At(s.getFaninVar0(x))
			a1 := *s.activity.At(s.getFaninVar1(x))
			if a0 > a1 {
				s.jdata.At(x).act_fanin = a0
			} else {
				s.jdata.At(x).act_fanin = a1
			}
			s.jheap.Update(x)
		}
	}
}

// addJwatch is Solver::addJwatch.
// C: void addJwatch(Var host, Var member)
func (s *Gluco2Solver) addJwatch(host Gluco2Var, member Gluco2Var) {
	if !(Gluco2VarUndef == s.jwatch.At(member).next && Gluco2VarUndef == s.jwatch.At(member).prev) {
		panic("Gluco2Solver.addJwatch: member is already in a list")
	}
	if Gluco2VarUndef != s.jwatch.At(host).head {
		s.jwatch.At(s.jwatch.At(host).head).prev = member
	}
	s.jwatch.At(member).next = s.jwatch.At(host).head
	s.jwatch.At(member).prev = host
	s.jwatch.At(host).head = member
}

// delJwatch is Solver::delJwatch.
// C: void delJwatch(Var member)
func (s *Gluco2Solver) delJwatch(member Gluco2Var) {
	prev := s.jwatch.At(member).prev
	next := s.jwatch.At(member).next
	if Gluco2VarUndef == prev { // must be in a list to be deleted
		panic("Gluco2Solver.delJwatch: member is not in a list")
	}
	if !(s.jwatch.At(prev).next == member || s.jwatch.At(prev).head == member) {
		panic("Gluco2Solver.delJwatch: broken list")
	}
	if s.jwatch.At(prev).next == member {
		s.jwatch.At(prev).next = next
	} else {
		s.jwatch.At(prev).head = next
	}
	if Gluco2VarUndef != next {
		s.jwatch.At(next).prev = prev
	}
	s.jwatch.At(member).prev = Gluco2VarUndef
	s.jwatch.At(member).next = Gluco2VarUndef
}

// justCheck is Solver::justCheck.
// C: void justCheck()
func (s *Gluco2Solver) justCheck() {
	nJustFail := 0
	for i := 0; i < s.trail.Size(); i++ {
		x := Gluco2LitVar(*s.trail.At(i))
		if !s.IsRoundWatch(x) {
			continue
		}
		if !s.isJReason(x) {
			continue
		}
		if Gluco2LitUndef != s.gateJustFanin(x) {
			t := 0
			if Gluco2LTrue == s.Value(x) {
				t = 1
			}
			println("\t", x, "is not justified (value=", t, ", level=", s.vardata.At(x).level, ")")
			nJustFail++
			panic("Gluco2Solver.justCheck: assert(false)")
		}
	}
	if nJustFail != 0 {
		println(nJustFail, "just-fails")
	}
}

// SetVarFaninLits is Solver::setVarFaninLits (public).
// C: void setVarFaninLits(Var v, Lit lit1, Lit lit2)
func (s *Gluco2Solver) SetVarFaninLits(v Gluco2Var, lit1 Gluco2Lit, lit2 Gluco2Lit) {
	if Gluco2LitVar(lit1) == Gluco2LitVar(lit2) {
		panic("Gluco2Solver.SetVarFaninLits: equal fanin variables")
	}
	mincap := Gluco2LitVar(lit1)
	if Gluco2LitVar(lit1) < Gluco2LitVar(lit2) {
		mincap = Gluco2LitVar(lit2)
	}
	if v >= mincap {
		mincap = v
	}
	mincap++
	if s.var2FaninLits.Size() < (mincap << 1) {
		s.var2FaninLits.GrowToPad(mincap<<1, gluco2LitNone)
	}
	*s.var2FaninLits.At((v << 1) + 0) = lit1
	*s.var2FaninLits.At((v << 1) + 1) = lit2
	if s.var2FanoutN.Size() < (mincap << 1) {
		s.var2FanoutN.GrowToPad(mincap<<1, gluco2LitNone)
	}
	if s.var2Fanout0.Size() < mincap {
		s.var2Fanout0.GrowToPad(mincap, gluco2LitNone)
	}
	*s.var2FanoutN.At((v << 1) + 0) = *s.var2Fanout0.At(Gluco2LitVar(lit1))
	*s.var2FanoutN.At((v << 1) + 1) = *s.var2Fanout0.At(Gluco2LitVar(lit2))
	*s.var2Fanout0.At(Gluco2LitVar(lit1)) = Gluco2ToLit((v << 1) + 0)
	*s.var2Fanout0.At(Gluco2LitVar(lit2)) = Gluco2ToLit((v << 1) + 1)
}

// gatePropagate is Solver::gatePropagate: returns the last conflict encountered (the minisat concept).
// C: CRef gatePropagate(Lit p)
func (s *Gluco2Solver) gatePropagate(p Gluco2Lit) Gluco2CRef {
	confl := Gluco2CRefUndef
	if s.JustUsage() < 2 || s.var2FaninLits.Size() <= Gluco2LitVar(p) {
		return Gluco2CRefUndef
	}
	if !s.IsRoundWatch(Gluco2LitVar(p)) {
		return Gluco2CRefUndef
	}
	if s.isTwoFanin(Gluco2LitVar(p)) {
		// check fanin consistency
		if stats := s.gatePropagateCheckThis(Gluco2LitVar(p)); Gluco2CRefUndef != stats {
			confl = stats
			if Gluco2LTrue == s.Value(Gluco2LitVar(p)) {
				return confl
			}
		}
	}
	// check_fanout: check fanout consistency
	if Gluco2LitVar(p) >= s.var2Fanout0.Size() {
		panic("Gluco2Solver.gatePropagate: var(p) >= var2Fanout0.size()")
	}
	for lfo := *s.var2Fanout0.At(Gluco2LitVar(p)); lfo != gluco2LitNone; lfo = *s.var2FanoutN.At(Gluco2ToIntLit(lfo)) {
		if !s.IsRoundWatch(Gluco2LitVar(lfo)) {
			continue
		}
		if stats := s.gatePropagateCheckFanout(Gluco2LitVar(p), lfo); Gluco2CRefUndef != stats {
			confl = stats
			if Gluco2LTrue == s.Value(Gluco2LitVar(lfo)) {
				return confl
			}
		}
	}
	return confl
}

// gatePropagateCheckFanout is Solver::gatePropagateCheckFanout.
// C: CRef gatePropagateCheckFanout(Var v, Lit lfo)
func (s *Gluco2Solver) gatePropagateCheckFanout(v Gluco2Var, lfo Gluco2Lit) Gluco2CRef {
	var faninLit Gluco2Lit
	if Gluco2Sign(lfo) {
		faninLit = s.getFaninLit1(Gluco2LitVar(lfo))
	} else {
		faninLit = s.getFaninLit0(Gluco2LitVar(lfo))
	}
	if Gluco2LitVar(faninLit) != v {
		panic("Gluco2Solver.gatePropagateCheckFanout: var(faninLit) != v")
	}
	lv := Gluco2LitVar(lfo)
	if s.isAND(lv) {
		if Gluco2LFalse == s.ValueLit(faninLit) {
			if Gluco2LFalse == s.Value(lv) {
				return Gluco2CRefUndef
			}
			if Gluco2LTrue == s.Value(lv) {
				return s.Var2CRef(lv)
			}
			if Gluco2Sign(lfo) {
				s.jwatch.At(lv).header_dir = 1
			} else {
				s.jwatch.At(lv).header_dir = 0
			}
			s.uncheckedEnqueue(Gluco2LitNot(Gluco2MkLit(lv, false)), s.Var2CRef(lv))
		} else {
			if Gluco2LTrue != s.ValueLit(faninLit) {
				panic("Gluco2Solver.gatePropagateCheckFanout: fanin is not true")
			}
			if Gluco2LTrue == s.Value(lv) {
				return Gluco2CRefUndef
			}
			// check value of the other fanin
			var faninLitP Gluco2Lit
			if Gluco2Sign(lfo) {
				faninLitP = s.getFaninLit0(lv)
			} else {
				faninLitP = s.getFaninLit1(lv)
			}
			if Gluco2LFalse == s.Value(lv) {
				if Gluco2LFalse == s.ValueLit(faninLitP) {
					return Gluco2CRefUndef
				}
				if Gluco2LTrue == s.ValueLit(faninLitP) {
					return s.Var2CRef(lv)
				}
				s.uncheckedEnqueue(Gluco2LitNot(faninLitP), s.Var2CRef(lv))
			} else if Gluco2LTrue == s.ValueLit(faninLitP) {
				s.uncheckedEnqueue(Gluco2MkLit(lv, false), s.Var2CRef(lv))
			}
		}
	} else { // xor scope
		// check value of the other fanin
		var faninLitP Gluco2Lit
		if Gluco2Sign(lfo) {
			faninLitP = s.getFaninLit0(lv)
		} else {
			faninLitP = s.getFaninLit1(lv)
		}
		val0 := s.ValueLit(faninLit)
		val1 := s.ValueLit(faninLitP)
		valo := s.Value(lv)
		if Gluco2LUndef == val1 && Gluco2LUndef == valo {
			return Gluco2CRefUndef
		} else if Gluco2LUndef == val1 {
			s.uncheckedEnqueue(Gluco2LitXor(Gluco2LitNot(faninLitP), (Gluco2LTrue == val0) != (Gluco2LTrue == valo)), s.Var2CRef(lv))
		} else if Gluco2LUndef == valo {
			s.uncheckedEnqueue(Gluco2LitNot(Gluco2MkLit(lv, (Gluco2LTrue == val0) != (Gluco2LTrue == val1))), s.Var2CRef(lv))
		} else if Gluco2LFalse == valo.Xor(val0 == val1) {
			return s.Var2CRef(lv)
		}
	}
	return Gluco2CRefUndef
}

// gatePropagateCheckThis is Solver::gatePropagateCheckThis.
// C: CRef gatePropagateCheckThis(Var v)
func (s *Gluco2Solver) gatePropagateCheckThis(v Gluco2Var) Gluco2CRef {
	confl := Gluco2CRefUndef
	if s.isAND(v) {
		if Gluco2LFalse == s.Value(v) {
			if Gluco2LTrue == s.ValueLit(s.getFaninLit0(v)) && Gluco2LTrue == s.ValueLit(s.getFaninLit1(v)) {
				return s.Var2CRef(v)
			}
			if Gluco2LFalse == s.ValueLit(s.getFaninLit0(v)) || Gluco2LFalse == s.ValueLit(s.getFaninLit1(v)) {
				return Gluco2CRefUndef
			}
			if Gluco2LTrue == s.ValueLit(s.getFaninLit0(v)) {
				s.uncheckedEnqueue(Gluco2LitNot(s.getFaninLit1(v)), s.Var2CRef(v))
			}
			if Gluco2LTrue == s.ValueLit(s.getFaninLit1(v)) {
				s.uncheckedEnqueue(Gluco2LitNot(s.getFaninLit0(v)), s.Var2CRef(v))
			}
		} else {
			if Gluco2LTrue != s.Value(v) {
				panic("Gluco2Solver.gatePropagateCheckThis: value(v) != l_True")
			}
			if Gluco2LFalse == s.ValueLit(s.getFaninLit0(v)) || Gluco2LFalse == s.ValueLit(s.getFaninLit1(v)) {
				confl = s.Var2CRef(v)
			}
			if Gluco2LUndef == s.ValueLit(s.getFaninLit0(v)) {
				s.uncheckedEnqueue(s.getFaninLit0(v), s.Var2CRef(v))
			}
			if Gluco2LUndef == s.ValueLit(s.getFaninLit1(v)) {
				s.uncheckedEnqueue(s.getFaninLit1(v), s.Var2CRef(v))
			}
		}
	} else { // xor scope
		val0 := s.ValueLit(s.getFaninLit0(v))
		val1 := s.ValueLit(s.getFaninLit1(v))
		valo := s.Value(v)
		if Gluco2LUndef == val0 && Gluco2LUndef == val1 {
			return Gluco2CRefUndef
		}
		if Gluco2LUndef == val0 {
			s.uncheckedEnqueue(Gluco2LitXor(Gluco2LitNot(s.getFaninLit0(v)), (Gluco2LTrue == val1) != (Gluco2LTrue == valo)), s.Var2CRef(v))
		} else if Gluco2LUndef == val1 {
			s.uncheckedEnqueue(Gluco2LitXor(Gluco2LitNot(s.getFaninLit1(v)), (Gluco2LTrue == val0) != (Gluco2LTrue == valo)), s.Var2CRef(v))
		} else if Gluco2LFalse == valo.Xor(val0 == val1) {
			return s.Var2CRef(v)
		}
	}
	return confl
}

// gatePropagateCheck is declared in Solver.h but never defined in CGlucoseCore.h or
// Glucose2.cpp (calling it is a link error in C++); it is not ported.

// setItpcSize is Solver::setItpcSize (sz <= 3): sets the size of the interpretation clause.
// C: void setItpcSize(int sz)
func (s *Gluco2Solver) setItpcSize(sz int) {
	if sz > 3 || Gluco2CRefUndef == s.itpc {
		panic("Gluco2Solver.setItpcSize: precondition")
	}
	s.ca.At(s.itpc).setSize(sz)
}

// getConfClause is Solver::getConfClause.
// C: CRef getConfClause(CRef r)
func (s *Gluco2Solver) getConfClause(r Gluco2CRef) Gluco2CRef {
	if !s.isGateCRef(r) {
		return r
	}
	v := s.CRef2Var(r)
	if !s.isTwoFanin(v) {
		panic("Gluco2Solver.getConfClause: not a gate")
	}
	if s.isAND(v) {
		if Gluco2LFalse == s.Value(v) {
			s.setItpcSize(3)
			c := s.ca.At(s.itpc)
			*c.At(0) = Gluco2MkLit(v, false)
			*c.At(1) = Gluco2LitNot(s.getFaninLit0(v))
			*c.At(2) = Gluco2LitNot(s.getFaninLit1(v))
		} else {
			s.setItpcSize(2)
			c := s.ca.At(s.itpc)
			*c.At(0) = Gluco2LitNot(Gluco2MkLit(v, false))
			val0 := s.ValueLit(s.getFaninLit0(v))
			val1 := s.ValueLit(s.getFaninLit1(v))
			if Gluco2LFalse == val0 && Gluco2LFalse == val1 {
				if s.level(s.getFaninVar0(v)) < s.level(s.getFaninVar1(v)) {
					*c.At(1) = s.getFaninLit0(v)
				} else {
					*c.At(1) = s.getFaninLit1(v)
				}
			} else if Gluco2LFalse == val0 {
				*c.At(1) = s.getFaninLit0(v)
			} else {
				*c.At(1) = s.getFaninLit1(v)
			}
		}
	} else { // xor scope
		s.setItpcSize(3)
		c := s.ca.At(s.itpc)
		*c.At(0) = Gluco2MkLit(v, Gluco2LTrue == s.Value(v))
		*c.At(1) = Gluco2MkLit(s.getFaninVar0(v), Gluco2LTrue == s.Value(s.getFaninVar0(v)))
		*c.At(2) = Gluco2MkLit(s.getFaninVar1(v), Gluco2LTrue == s.Value(s.getFaninVar1(v)))
	}
	return s.itpc
}

// interpret is Solver::interpret: get the gate-clause on the implication graph.
// C: CRef interpret(Var v, Var t)
func (s *Gluco2Solver) interpret(v Gluco2Var, t Gluco2Var) Gluco2CRef {
	if !s.isTwoFanin(v) {
		panic("Gluco2Solver.interpret: not a gate")
	}
	var0 := s.getFaninVar0(v)
	var1 := s.getFaninVar1(v)
	val0 := s.Value(var0)
	val1 := s.Value(var1)
	valo := s.Value(v)

	// phased values
	if Gluco2LUndef != val0 {
		val0 = val0.Xor(s.getFaninC0(v))
	}
	if Gluco2LUndef != val1 {
		val1 = val1.Xor(s.getFaninC1(v))
	}

	if s.isAND(v) {
		if v == t { // tracing output
			if Gluco2LFalse == valo {
				s.setItpcSize(2)
				c := s.ca.At(s.itpc)
				*c.At(0) = Gluco2LitNot(Gluco2MkLit(v, false))
				if Gluco2LFalse == val0 && Gluco2LFalse == val1 {
					if 0 == s.jwatch.At(v).header_dir {
						*c.At(1) = s.getFaninLit0(v)
					} else {
						*c.At(1) = s.getFaninLit1(v)
					}
				} else if Gluco2LFalse == val0 {
					*c.At(1) = s.getFaninLit0(v)
				} else {
					*c.At(1) = s.getFaninLit1(v)
				}
			} else {
				s.setItpcSize(3)
				c := s.ca.At(s.itpc)
				*c.At(0) = Gluco2MkLit(v, false)
				*c.At(1) = Gluco2LitNot(s.getFaninLit0(v))
				*c.At(2) = Gluco2LitNot(s.getFaninLit1(v))
			}
		} else {
			if !(t == var0 || t == var1) {
				panic("Gluco2Solver.interpret: t is not a fanin")
			}
			if Gluco2LFalse == valo {
				s.setItpcSize(3)
				c := s.ca.At(s.itpc)
				*c.At(0) = Gluco2LitNot(s.getFaninLit0(v))
				*c.At(1) = Gluco2LitNot(s.getFaninLit1(v))
				*c.At(2) = Gluco2MkLit(v, false)
				if t == var1 {
					*c.At(0), *c.At(1) = *c.At(1), *c.At(0)
				}
			} else {
				s.setItpcSize(2)
				c := s.ca.At(s.itpc)
				if t == var0 {
					*c.At(0) = s.getFaninLit0(v)
				} else {
					*c.At(0) = s.getFaninLit1(v)
				}
				*c.At(1) = Gluco2LitNot(Gluco2MkLit(v, false))
			}
		}
	} else { // xor scope
		s.setItpcSize(3)
		c := s.ca.At(s.itpc)
		if v == t {
			*c.At(0) = Gluco2MkLit(v, Gluco2LFalse == s.Value(v))
			*c.At(1) = Gluco2MkLit(var0, Gluco2LTrue == s.Value(var0))
			*c.At(2) = Gluco2MkLit(var1, Gluco2LTrue == s.Value(var1))
		} else {
			if t == var0 {
				*c.At(0) = Gluco2MkLit(var0, Gluco2LFalse == s.Value(var0))
				*c.At(1) = Gluco2MkLit(var1, Gluco2LTrue == s.Value(var1))
			} else {
				*c.At(1) = Gluco2MkLit(var0, Gluco2LTrue == s.Value(var0))
				*c.At(0) = Gluco2MkLit(var1, Gluco2LFalse == s.Value(var1))
			}
			*c.At(2) = Gluco2MkLit(v, Gluco2LTrue == s.Value(v))
		}
	}
	return s.itpc
}

// castCRef is Solver::castCRef: interpret a gate into a clause.
// C: CRef castCRef(Lit p)
func (s *Gluco2Solver) castCRef(p Gluco2Lit) Gluco2CRef {
	cr := s.reason(Gluco2LitVar(p))
	if Gluco2CRefUndef == cr {
		return cr
	}
	if !s.isGateCRef(cr) {
		return cr
	}
	v := s.CRef2Var(cr)
	return s.interpret(v, Gluco2LitVar(p))
}

// MarkCone is Solver::markCone (public).
// C: void markCone(Var v)
func (s *Gluco2Solver) MarkCone(v Gluco2Var) {
	if int64(*s.var2TravId.At(v)) >= s.travId {
		return
	}
	*s.var2TravId.At(v) = uint32(s.travId)
	if !s.isTwoFanin(v) {
		return
	}
	s.MarkCone(s.getFaninVar0(v))
	s.MarkCone(s.getFaninVar1(v))
}

// Sat_solver_set_var_fanin_lit is Solver::sat_solver_set_var_fanin_lit.
// C: void sat_solver_set_var_fanin_lit(int, int, int)
func (s *Gluco2Solver) Sat_solver_set_var_fanin_lit(v int, lit0 int, lit1 int) {
	s.SetVarFaninLits(Gluco2Var(v), Gluco2ToLit(lit0), Gluco2ToLit(lit1))
}

// Sat_solver_start_new_round is Solver::sat_solver_start_new_round.
// C: void sat_solver_start_new_round()
func (s *Gluco2Solver) Sat_solver_start_new_round() { s.SetNewRound() }

// Sat_solver_mark_cone is Solver::sat_solver_mark_cone.
// C: void sat_solver_mark_cone(int)
func (s *Gluco2Solver) Sat_solver_mark_cone(v int) { s.MarkCone(v) }

// GetCex is Solver::getCex: the counter-example of a justification run (the
// literals of JustModel as ints; element 0 is the number of literals).
// C: int * getCex() const
func (s *Gluco2Solver) GetCex() []int {
	cex := make([]int, s.JustModel.Size())
	for i := range cex {
		cex[i] = Gluco2ToIntLit(*s.JustModel.At(i))
	}
	return cex
}

// Level is Solver::level (public in src/sat/glucose2, moved to compile JustOrderLt).
// C: int level(Var x) const
func (s *Gluco2Solver) Level(x Gluco2Var) int { return s.level(x) }
