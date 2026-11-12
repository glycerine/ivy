package bdc

// Existentially quantify a support variable only when doing so cannot merge
// an onset minterm with an offset minterm.
func Bdc_SuppMinimize(p *Bdc_Man_t, f *Bdc_Isf_t) {
	f.Supp = 0
	for v := int32(0); v < p.nVars; v++ {
		if !ttHasVar(f.On, v) && !ttHasVar(f.Off, v) {
			continue
		}
		on, off := ttQuant(f.On, v, false), ttQuant(f.Off, v, false)
		if ttDisjoint(on, off) {
			f.On, f.Off = on, off
		} else {
			f.Supp |= 1 << uint(v)
		}
	}
}
func Bdc_SuppMinimize2(p *Bdc_Man_t, f *Bdc_Isf_t) { Bdc_SuppMinimize(p, f) }
func Bdc_DecomposeUpdateRight(p *Bdc_Man_t, f, l, r *Bdc_Isf_t, left *Bdc_Fun_t, kind int32) int32 {
	truth := funTruth(left)
	if kind == BDC_TYPE_OR {
		r.On = ttExistSet(ttBinary(f.On, truth, '-'), l.Uniq)
		r.Off = ttExistSet(f.Off, l.Uniq)
		return bdcFlag(ttZero(r.On))
	}
	if kind == BDC_TYPE_AND {
		r.Off = ttExistSet(ttBinary(f.Off, truth, '&'), l.Uniq)
		r.On = ttExistSet(f.On, l.Uniq)
		return bdcFlag(ttZero(r.Off))
	}
	return 0
}
func Bdc_DecomposeGetCost(p *Bdc_Man_t, l, r int32) int32 {
	return BDC_SCALE * (p.nVars*min(l, r) + max(l, r))
}
func Bdc_DecomposeFindInitialVarSet(p *Bdc_Man_t, f, l, r *Bdc_Isf_t) int32 {
	var vars []int32
	for v := int32(0); v < p.nVars; v++ {
		if f.Supp&(1<<uint(v)) != 0 {
			vars = append(vars, v)
		}
	}
	for beg, a := range vars {
		x := ttQuant(f.Off, a, false)
		for end := len(vars) - 1; end > beg; end-- {
			y := ttQuant(f.Off, vars[end], false)
			if ttDisjoint3(f.On, x, y) {
				l.Uniq = 1 << uint(a)
				r.Uniq = 1 << uint(vars[end])
				return 1
			}
		}
	}
	return 0
}
func Bdc_DecomposeWeakOr(p *Bdc_Man_t, f, l, r *Bdc_Isf_t) int32 {
	best, bestCost := int32(-1), int32(0)
	for v := int32(0); v < p.nVars; v++ {
		if f.Supp&(1<<uint(v)) == 0 {
			continue
		}
		offset := ttQuant(f.Off, v, false)
		if ttImpl(f.On, offset) {
			continue
		}
		cost := ttCount(ttQuant(f.On, v, true))
		if cost == 0 {
			cost = 1
		}
		if cost > bestCost {
			best, bestCost = v, cost
		}
	}
	if best < 0 {
		return 0
	}
	l.On = ttBinary(f.On, ttQuant(f.Off, best, false), '&')
	l.Off = append([]uint32(nil), f.Off...)
	l.Uniq = 1 << uint(best)
	r.Uniq = 0
	return max(1, bestCost*BDC_SCALE/(1<<uint(p.nVars)))
}
func Bdc_DecomposeOr(p *Bdc_Man_t, f, l, r *Bdc_Isf_t) int32 {
	Bdc_IsfStart(p, l)
	Bdc_IsfStart(p, r)
	if Bdc_DecomposeFindInitialVarSet(p, f, l, r) == 0 {
		return Bdc_DecomposeWeakOr(p, f, l, r)
	}
	left, right := int32(1), int32(1)
	offL, offR := ttExistSet(f.Off, l.Uniq), ttExistSet(f.Off, r.Uniq)
	remaining := f.Supp &^ l.Uniq &^ r.Uniq
	for v := int32(0); v < p.nVars; v++ {
		if remaining&(1<<uint(v)) == 0 {
			continue
		}
		nextL, nextR := ttQuant(offL, v, false), ttQuant(offR, v, false)
		addLeft := func() bool {
			if ttDisjoint3(f.On, nextL, offR) {
				l.Uniq |= 1 << uint(v)
				left++
				offL = nextL
				return true
			}
			return false
		}
		addRight := func() bool {
			if ttDisjoint3(f.On, nextR, offL) {
				r.Uniq |= 1 << uint(v)
				right++
				offR = nextR
				return true
			}
			return false
		}
		if left < right {
			if !addLeft() {
				addRight()
			}
		} else {
			if !addRight() {
				addLeft()
			}
		}
	}
	l.On = ttExistSet(ttBinary(f.On, offL, '&'), r.Uniq)
	l.Off = append([]uint32(nil), offR...)
	r.On = ttExistSet(ttBinary(f.On, offR, '&'), l.Uniq)
	r.Off = append([]uint32(nil), offL...)
	return Bdc_DecomposeGetCost(p, left, right)
}
func Bdc_DecomposeStep(p *Bdc_Man_t, f, l, r *Bdc_Isf_t) Bdc_Type_t {
	Bdc_IsfClean(&p.ol)
	Bdc_IsfClean(&p.or)
	Bdc_IsfClean(&p.al)
	Bdc_IsfClean(&p.ar)
	weightOr := Bdc_DecomposeOr(p, f, &p.ol, &p.or)
	Bdc_IsfNot(f)
	weightAnd := Bdc_DecomposeOr(p, f, &p.al, &p.ar)
	Bdc_IsfNot(f)
	Bdc_IsfNot(&p.al)
	Bdc_IsfNot(&p.ar)
	if weightOr == 0 && weightAnd == 0 {
		Bdc_IsfCopy(l, &p.ol)
		Bdc_IsfCopy(r, &p.or)
		return BDC_TYPE_MUX
	}
	reuseOr, reuseAnd := int32(0), int32(0)
	for _, branch := range []struct {
		weight int32
		l, r   *Bdc_Isf_t
		reuse  *int32
	}{{weightOr, &p.ol, &p.or, &reuseOr}, {weightAnd, &p.al, &p.ar, &reuseAnd}} {
		if branch.weight == 0 {
			continue
		}
		for _, part := range []*Bdc_Isf_t{branch.l, branch.r} {
			if part.Uniq != 0 {
				Bdc_SuppMinimize(p, part)
				if Bdc_TableLookup(p, part) != nil {
					(*branch.reuse)++
				}
			}
		}
	}
	kind := int32(BDC_TYPE_AND)
	if reuseOr > reuseAnd || reuseOr == reuseAnd && weightOr > weightAnd {
		kind = BDC_TYPE_OR
	}
	if reuseOr != reuseAnd {
		p.numReuse++
	} else {
		weight := weightAnd
		if kind == BDC_TYPE_OR {
			weight = weightOr
		}
		if weight < BDC_SCALE {
			p.numWeaks++
		}
	}
	if kind == BDC_TYPE_OR {
		p.numOrs++
		Bdc_IsfCopy(l, &p.ol)
		Bdc_IsfCopy(r, &p.or)
	} else {
		p.numAnds++
		Bdc_IsfCopy(l, &p.al)
		Bdc_IsfCopy(r, &p.ar)
	}
	return kind
}
func Bdc_DecomposeStepMux(p *Bdc_Man_t, f, l, r *Bdc_Isf_t) int32 {
	// The original implementation stops at the first support variable that
	// improves its initial bound; keep this tie/order policy.
	for v := int32(0); v < p.nVars; v++ {
		if f.Supp&(1<<uint(v)) == 0 {
			continue
		}
		l.On, l.Off = ttCof(f.On, v, false), ttCof(f.Off, v, false)
		r.On, r.Off = ttCof(f.On, v, true), ttCof(f.Off, v, true)
		Bdc_SuppMinimize(p, l)
		Bdc_SuppMinimize(p, r)
		return v
	}
	return -1
}
func Bdc_ManNodeVerify(p *Bdc_Man_t, f *Bdc_Isf_t, node *Bdc_Fun_t) int32 {
	return Bdc_TableCheckContainment(p, f, funTruth(node))
}
func Bdc_ManCreateGate(p *Bdc_Man_t, left, right *Bdc_Fun_t, kind Bdc_Type_t) *Bdc_Fun_t {
	f := Bdc_FunNew(p)
	if f == nil {
		return nil
	}
	f.Type = BDC_TYPE_AND
	f.Fan0, f.Fan1 = left, right
	switch kind {
	case BDC_TYPE_AND:
		f.Func = ttBinary(funTruth(left), funTruth(right), '&')
	case BDC_TYPE_OR:
		f.Fan0, f.Fan1 = Bdc_Not(left), Bdc_Not(right)
		f.Func = ttNot(ttBinary(funTruth(left), funTruth(right), '|'))
	default:
		panic("bdc: unsupported gate type")
	}
	f.Supp = ttSupport(f.Func, p.nVars)
	Bdc_TableAdd(p, f)
	if kind == BDC_TYPE_OR {
		return Bdc_Not(f)
	}
	return f
}
func Bdc_ManDecompose_rec(p *Bdc_Man_t, f *Bdc_Isf_t) *Bdc_Fun_t {
	if !ttDisjoint(f.On, f.Off) {
		panic("bdc: overlapping ISF")
	}
	if cached := Bdc_TableLookup(p, f); cached != nil {
		return cached
	}
	var l, r Bdc_Isf_t
	kind := Bdc_DecomposeStep(p, f, &l, &r)
	if kind == BDC_TYPE_MUX {
		v := Bdc_DecomposeStepMux(p, f, &l, &r)
		if v < 0 {
			panic("bdc: nonconstant function has empty support")
		}
		p.numMuxes++
		left, right := Bdc_ManDecompose_rec(p, &l), Bdc_ManDecompose_rec(p, &r)
		if left == nil || right == nil {
			return nil
		}
		selector := Bdc_FunWithId(p, v+1)
		left = Bdc_ManCreateGate(p, Bdc_Not(selector), left, BDC_TYPE_AND)
		right = Bdc_ManCreateGate(p, selector, right, BDC_TYPE_AND)
		if left == nil || right == nil {
			return nil
		}
		return Bdc_ManCreateGate(p, left, right, BDC_TYPE_OR)
	}
	left := Bdc_ManDecompose_rec(p, &l)
	if left == nil {
		return nil
	}
	if Bdc_DecomposeUpdateRight(p, f, &l, &r, left, kind) != 0 {
		p.nNodesNew--
		return left
	}
	Bdc_SuppMinimize(p, &r)
	right := Bdc_ManDecompose_rec(p, &r)
	if right == nil {
		return nil
	}
	return Bdc_ManCreateGate(p, left, right, kind)
}
