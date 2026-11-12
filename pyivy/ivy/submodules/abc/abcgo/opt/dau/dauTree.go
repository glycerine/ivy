package dau

/**CFile****************************************************************

  FileName    [dauTree.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Canonical DSD package.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauTree.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// These size helpers report the original packed format's word counts; Go's
// in-memory objects use slices and never rely on those ABI offsets.
func Dss_EntWordNum(p *Dss_Ent_t) int32 { return 3 + int32((len(p.Shared)+7)/8) }
func Dss_FunWordNum(p *Dss_Fun_t) int32 { return int32((len(p.Fans) + 4 + 7) / 8) }
func Dss_ObjWordNum(fans int32) int32   { return 1 + (fans+1)/2 }
func Dss_NtkPrint_rec(p *Dss_Ntk_t, obj *Dss_Obj_t) string {
	var b strings.Builder
	dssWriteNode(&b, p.Objects, obj, func(o *Dss_Obj_t) string { return string(rune('a' + o.Var)) })
	return b.String()
}
func Dss_NtkPrint(p *Dss_Ntk_t) string { return Dss_NtkPrint_rec(p, p.Root) + "\n" }
func Dss_ObjCompare(objects []*Dss_Obj_t, a, b *Dss_Obj_t) int32 {
	x, y := Dss_Regular(a), Dss_Regular(b)
	if x.Type < y.Type {
		return -1
	}
	if x.Type > y.Type {
		return 1
	}
	if x.Type < DAU_DSD_AND {
		return 0
	}
	if len(x.Fans) < len(y.Fans) {
		return -1
	}
	if len(x.Fans) > len(y.Fans) {
		return 1
	}
	for i := range x.Fans {
		if cmp := Dss_ObjCompare(objects, Dss_Lit2Obj(objects, x.Fans[i]), Dss_Lit2Obj(objects, y.Fans[i])); cmp != 0 {
			return cmp
		}
	}
	return Dss_IsComplement(a) - Dss_IsComplement(b)
}
func Dss_ObjSort(objects []*Dss_Obj_t, nodes []*Dss_Obj_t, n int32, perm []int32) {
	for i := int32(0); i < n-1; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			if Dss_ObjCompare(objects, nodes[best], nodes[j]) > 0 {
				best = j
			}
		}
		nodes[i], nodes[best] = nodes[best], nodes[i]
		if perm != nil {
			perm[i], perm[best] = perm[best], perm[i]
		}
	}
}
func Dss_NtkCheck(p *Dss_Ntk_t) {
	for _, o := range p.Objects {
		for i, lit := range o.Fans {
			f := p.Objects[lit/2]
			if o.Type == DAU_DSD_AND && f.Type == DAU_DSD_AND && lit&1 == 0 {
				panic("unflattened DSD AND")
			}
			if o.Type == DAU_DSD_XOR && f.Type == DAU_DSD_XOR {
				panic("unflattened DSD XOR")
			}
			if o.Type == DAU_DSD_MUX && i == 0 && lit&1 != 0 {
				panic("complemented MUX control")
			}
		}
	}
}
func Dss_NtkCollectPerm_rec(p *Dss_Ntk_t, obj *Dss_Obj_t, perm []int32, count *int32) int32 {
	phase := Dss_IsComplement(obj)
	o := Dss_Regular(obj)
	if o.Type == DAU_DSD_VAR {
		perm[*count] = o.Var*2 + phase
		o.Var = *count
		*count++
		return phase
	}
	for i, lit := range o.Fans {
		if Dss_NtkCollectPerm_rec(p, Dss_Lit2Obj(p.Objects, lit), perm, count) != 0 {
			o.Fans[i] &= ^int32(1)
		}
	}
	return 0
}
func Dss_NtkTransform(p *Dss_Ntk_t, perm []int32) {
	if Dss_ObjType(p.Root) == DAU_DSD_CONST0 {
		return
	}
	for _, o := range p.Objects {
		if o.Type != DAU_DSD_AND && o.Type != DAU_DSD_XOR {
			continue
		}
		children := make([]*Dss_Obj_t, len(o.Fans))
		for i, lit := range o.Fans {
			children[i] = Dss_Lit2Obj(p.Objects, lit)
		}
		Dss_ObjSort(p.Objects, children, int32(len(children)), nil)
		for i, c := range children {
			o.Fans[i] = Dss_Obj2Lit(c)
		}
	}
	count := int32(0)
	if Dss_NtkCollectPerm_rec(p, p.Root, perm, &count) != 0 {
		p.Root = Dss_Regular(p.Root)
	}
	if count != Dss_ObjSuppSize(p.Root) {
		panic("DSD transform support mismatch")
	}
}
func Dss_ObjAlloc(p *Dss_Man_t, kind Dau_DsdType_t, fans, truthVars int32) *Dss_Obj_t {
	return dssAllocate(&p.Objects, kind, fans, truthVars)
}
func Dss_ObjCreate(p *Dss_Man_t, kind Dau_DsdType_t, fans []int32, truth []uint64) *Dss_Obj_t {
	truthVars := int32(0)
	if kind == DAU_DSD_PRIME {
		truthVars = int32(len(fans))
	}
	o := Dss_ObjAlloc(p, kind, int32(len(fans)), truthVars)
	copy(o.Fans, fans)
	if truthVars > 0 {
		if len(truth) < len(o.Truth) {
			panic("prime truth is too short")
		}
		copy(o.Truth, truth)
	}
	for _, lit := range fans {
		o.support += p.Objects[lit/2].support
	}
	return o
}
func Dss_ObjFindOrAdd(p *Dss_Man_t, kind Dau_DsdType_t, fans []int32, truth []uint64) *Dss_Obj_t {
	if kind != DAU_DSD_PRIME {
		truth = nil
	} else {
		truth = truth[:dssWords(int32(len(fans)))]
	}
	key := dssNodeKey(kind, fans, truth)
	if id, ok := p.unique[key]; ok {
		return p.Objects[id]
	}
	o := Dss_ObjCreate(p, kind, fans, truth)
	p.unique[key] = o.ID
	return o
}
func Dss_ObjHashKey(p *Dss_Man_t, kind Dau_DsdType_t, fans []int32, truth []uint64) uint32 {
	if kind != DAU_DSD_PRIME {
		truth = nil
	} else {
		truth = truth[:dssWords(int32(len(fans)))]
	}
	key := dssNodeKey(kind, fans, truth)
	h := uint32(2166136261)
	for _, v := range []byte(key) {
		h = (h ^ uint32(v)) * 16777619
	}
	return h
}
func Dss_ObjHashLookup(p *Dss_Man_t, kind Dau_DsdType_t, fans []int32, truth []uint64) *Dss_Obj_t {
	if kind != DAU_DSD_PRIME {
		truth = nil
	} else {
		truth = truth[:dssWords(int32(len(fans)))]
	}
	id, ok := p.unique[dssNodeKey(kind, fans, truth)]
	if !ok {
		return nil
	}
	return p.Objects[id]
}
func Dss_ManAlloc(n, limit int32) *Dss_Man_t {
	dssWords(n)
	p := &Dss_Man_t{NVars: n, NonDecLimit: limit, unique: make(map[string]int32)}
	Dss_ObjAlloc(p, DAU_DSD_CONST0, 0, 0)
	Dss_ObjAlloc(p, DAU_DSD_VAR, 0, 0).support = 1
	return p
}
func Dss_ManFree(p *Dss_Man_t) {
	if p != nil {
		*p = Dss_Man_t{}
	}
}
func Dss_ManCacheAlloc(p *Dss_Man_t) { p.cache = make(map[string]*Dss_Ent_t) }
func Dss_ManCacheFree(p *Dss_Man_t)  { p.cache = nil }
func dssCacheKey(e *Dss_Ent_t) string {
	return fmt.Sprintf("%d/%d/", e.Dsd0, e.Dsd1) + string(e.Shared)
}
func Dss_ManCacheLookup(p *Dss_Man_t, e *Dss_Ent_t) *Dss_Ent_t {
	entry := p.cache[dssCacheKey(e)]
	which := 0
	if len(e.Shared) > 0 {
		which = 1
	}
	if entry != nil {
		p.CacheHits[which]++
	} else {
		p.CacheMisses[which]++
	}
	return entry
}
func Dss_ManCacheCreate(p *Dss_Man_t, e *Dss_Ent_t, f *Dss_Fun_t) *Dss_Ent_t {
	copyEnt := &Dss_Ent_t{Dsd0: e.Dsd0, Dsd1: e.Dsd1, Shared: append([]byte(nil), e.Shared...), Func: &Dss_Fun_t{Dsd: f.Dsd, Fans: append([]byte(nil), f.Fans...)}}
	if p.cache != nil {
		p.cache[dssCacheKey(e)] = copyEnt
	}
	return copyEnt
}
func Dss_ManHashProfile(p *Dss_Man_t) string { return fmt.Sprintf("Nodes = %d.\n", len(p.unique)) }
func Dss_ManCacheProfile(p *Dss_Man_t) string {
	return fmt.Sprintf("Entries = %d. Hits = %v. Misses = %v.\n", len(p.cache), p.CacheHits, p.CacheMisses)
}
func Dss_ManPrint_rec(w io.Writer, p *Dss_Man_t, obj *Dss_Obj_t, perm []int32, count *int32) error {
	return dssWriteNode(w, p.Objects, obj, func(o *Dss_Obj_t) string {
		lit := *count * 2
		if perm != nil {
			lit = perm[*count]
		}
		*count++
		prefix := ""
		if lit&1 != 0 {
			prefix = "!"
		}
		return prefix + string(rune('a'+lit/2))
	})
}
func Dss_ManExpression(p *Dss_Man_t, lit int32, perm []int32) string {
	var b strings.Builder
	count := int32(0)
	Dss_ManPrint_rec(&b, p, Dss_Lit2Obj(p.Objects, lit), perm, &count)
	return b.String()
}
func Dss_ManPrintOne(w io.Writer, p *Dss_Man_t, lit int32, perm []int32) error {
	_, err := fmt.Fprintf(w, "%6d : %2d %s\n", lit/2, Dss_VecLitSuppSize(p.Objects, lit), Dss_ManExpression(p, lit, perm))
	return err
}

func Dss_ManCheckNonDec_rec(p *Dss_Man_t, obj *Dss_Obj_t) int32 {
	o := Dss_Regular(obj)
	max := int32(0)
	if o.Type == DAU_DSD_PRIME {
		max = int32(len(o.Fans))
	}
	for _, lit := range o.Fans {
		if n := Dss_ManCheckNonDec_rec(p, Dss_Lit2Obj(p.Objects, lit)); n > max {
			max = n
		}
	}
	return max
}
func Dss_ManDump(p *Dss_Man_t) string {
	var b strings.Builder
	for _, o := range p.Objects {
		Dss_ManPrintOne(&b, p, o.ID*2, nil)
	}
	return b.String()
}
func Dss_ManPrint(name string, p *Dss_Man_t) error {
	if name == "" {
		_, err := io.WriteString(os.Stdout, Dss_ManDump(p))
		return err
	}
	return os.WriteFile(name, []byte(Dss_ManDump(p)), 0666)
}
func dssEvalManager(p *Dss_Man_t, obj *Dss_Obj_t, bitsValue int, perm []int32, count *int32) bool {
	phase := Dss_IsComplement(obj) != 0
	o := Dss_Regular(obj)
	value := false
	switch o.Type {
	case DAU_DSD_CONST0:
		value = false
	case DAU_DSD_VAR:
		lit := *count * 2
		if perm != nil {
			lit = perm[*count]
		}
		*count++
		value = (bitsValue&(1<<uint(lit/2)) != 0) != (lit&1 != 0)
	case DAU_DSD_AND:
		value = true
		for _, lit := range o.Fans {
			child := dssEvalManager(p, Dss_Lit2Obj(p.Objects, lit), bitsValue, perm, count)
			value = value && child
		}
	case DAU_DSD_XOR:
		for _, lit := range o.Fans {
			value = value != dssEvalManager(p, Dss_Lit2Obj(p.Objects, lit), bitsValue, perm, count)
		}
	case DAU_DSD_MUX:
		if len(o.Fans) != 3 {
			panic("invalid mux")
		}
		a := dssEvalManager(p, Dss_Lit2Obj(p.Objects, o.Fans[0]), bitsValue, perm, count)
		b := dssEvalManager(p, Dss_Lit2Obj(p.Objects, o.Fans[1]), bitsValue, perm, count)
		c := dssEvalManager(p, Dss_Lit2Obj(p.Objects, o.Fans[2]), bitsValue, perm, count)
		value = a && b || !a && c
	case DAU_DSD_PRIME:
		index := 0
		for i, lit := range o.Fans {
			if dssEvalManager(p, Dss_Lit2Obj(p.Objects, lit), bitsValue, perm, count) {
				index |= 1 << uint(i)
			}
		}
		value = dssBit(o.Truth, index)
	default:
		panic("invalid DSD node")
	}
	return value != phase
}
func Dss_ManComputeTruth_rec(p *Dss_Man_t, obj *Dss_Obj_t, n int32, result []uint64, perm []int32, count *int32) {
	start := *count
	clear(result[:dssWords(n)])
	for bit := 0; bit < 64*dssWords(n); bit++ {
		current := start
		if dssEvalManager(p, obj, bit, perm, &current) {
			dssSet(result, bit)
		}
		*count = current
	}
}
func Dss_ManComputeTruth(p *Dss_Man_t, lit, n int32, perm []int32) []uint64 {
	result := make([]uint64, dssWords(n))
	count := int32(0)
	Dss_ManComputeTruth_rec(p, Dss_Lit2Obj(p.Objects, lit), n, result, perm, &count)
	if count != Dss_VecLitSuppSize(p.Objects, lit) {
		panic("DSD truth support mismatch")
	}
	return result
}
func Dss_NtkRebuild_rec(p *Dss_Man_t, ntk *Dss_Ntk_t, obj *Dss_Obj_t) int32 {
	phase := Dss_IsComplement(obj)
	o := Dss_Regular(obj)
	if o.Type == DAU_DSD_CONST0 {
		return phase
	}
	if o.Type == DAU_DSD_VAR {
		return 2 + phase
	}
	fans := make([]int32, len(o.Fans))
	for i, lit := range o.Fans {
		fans[i] = Dss_NtkRebuild_rec(p, ntk, Dss_Lit2Obj(ntk.Objects, lit))
		if o.Type == DAU_DSD_XOR {
			phase ^= fans[i] & 1
			fans[i] &= ^int32(1)
		}
	}
	if o.Type == DAU_DSD_MUX {
		if fans[0]&1 != 0 {
			fans[0] ^= 1
			fans[1], fans[2] = fans[2], fans[1]
		}
		if fans[1]&1 != 0 {
			fans[1] ^= 1
			fans[2] ^= 1
			phase ^= 1
		}
	}
	node := Dss_ObjFindOrAdd(p, o.Type, fans, o.Truth)
	return node.ID*2 + phase
}
func Dss_NtkRebuild(p *Dss_Man_t, ntk *Dss_Ntk_t) int32 {
	if p.NVars != ntk.NVars {
		panic("DSD manager variable limit differs")
	}
	return Dss_NtkRebuild_rec(p, ntk, ntk.Root)
}
func Dss_ManOperation(p *Dss_Man_t, kind Dau_DsdType_t, lits []int32, n int32, perm []byte, truth []uint64) int32 {
	type child struct {
		obj        *Dss_Obj_t
		begin, end int32
		phase      int32
	}
	var children []child
	phase := int32(0)
	offset := int32(0)
	for _, lit := range lits[:n] {
		o := Dss_Lit2Obj(p.Objects, lit)
		regular := Dss_Regular(o)
		if kind == DAU_DSD_XOR {
			phase ^= lit & 1
			o = regular
		}
		flatten := (kind == DAU_DSD_AND && lit&1 == 0 || kind == DAU_DSD_XOR) && regular.Type == kind
		parts := []*Dss_Obj_t{o}
		if flatten {
			parts = nil
			for _, fan := range regular.Fans {
				parts = append(parts, Dss_Lit2Obj(p.Objects, fan))
			}
		}
		for _, c := range parts {
			support := Dss_ObjSuppSize(c)
			leafPhase := int32(0)
			if perm != nil && kind == DAU_DSD_AND && Dss_ObjType(c) == DAU_DSD_VAR {
				leafPhase = Dss_IsComplement(c)
				c = Dss_Regular(c)
			}
			children = append(children, child{c, offset, offset + support, leafPhase})
			offset += support
		}
	}
	if kind == DAU_DSD_AND || kind == DAU_DSD_XOR {
		for i := 0; i < len(children)-1; i++ {
			best := i
			for j := i + 1; j < len(children); j++ {
				if Dss_ObjCompare(p.Objects, children[best].obj, children[j].obj) > 0 {
					best = j
				}
			}
			children[i], children[best] = children[best], children[i]
		}
	}
	if kind == DAU_DSD_MUX {
		if len(children) != 3 {
			panic("mux requires three fanins")
		}
		if Dss_IsComplement(children[0].obj) != 0 {
			children[0].obj = Dss_Not(children[0].obj)
			children[1], children[2] = children[2], children[1]
		}
		if Dss_IsComplement(children[1].obj) != 0 {
			children[1].obj = Dss_Not(children[1].obj)
			children[2].obj = Dss_Not(children[2].obj)
			phase ^= 1
		}
	}
	fans := make([]int32, len(children))
	position := 0
	for i, c := range children {
		fans[i] = Dss_Obj2Lit(c.obj)
		if perm != nil {
			for k := c.begin; k < c.end; k++ {
				perm[position] = byte(k*2 + c.phase)
				position++
			}
		}
	}
	node := Dss_ObjFindOrAdd(p, kind, fans, truth)
	return node.ID*2 + phase
}
func Dss_ManOperationFun(p *Dss_Man_t, lits []int32, total int32) *Dss_Fun_t {
	f := &Dss_Fun_t{Fans: make([]byte, total)}
	f.Dsd = Dss_ManOperation(p, DAU_DSD_AND, lits, 2, f.Fans, nil)
	return f
}

func Dss_EntPrint(e *Dss_Ent_t, f *Dss_Fun_t) string {
	return fmt.Sprintf("%d %d shared=%v result=%d perm=%v", e.Dsd0, e.Dsd1, e.Shared, f.Dsd, f.Fans)
}
func Dss_ManBooleanAnd(p *Dss_Man_t, e *Dss_Ent_t, counter int32) *Dss_Fun_t {
	n0, n1 := Dss_VecLitSuppSize(p.Objects, e.Dsd0), Dss_VecLitSuppSize(p.Objects, e.Dsd1)
	mapping := make([]int32, 0, n0+n1)
	perm0 := make([]int32, n0)
	for i := int32(0); i < n0; i++ {
		perm0[i] = i * 2
		mapping = append(mapping, i)
	}
	first := Dss_ManComputeTruth(p, e.Dsd0, p.NVars, perm0)
	perm1 := make([]int32, n1)
	for i := range perm1 {
		perm1[i] = -1
	}
	for i := 0; i < len(e.Shared); i += 2 {
		local, lit := int(e.Shared[i]), int32(e.Shared[i+1])
		if local >= len(perm1) || lit/2 >= n0 {
			panic("invalid shared DSD map")
		}
		perm1[local] = lit
	}
	support := n0
	for i := int32(0); i < n1; i++ {
		if perm1[i] < 0 {
			perm1[i] = support * 2
			mapping = append(mapping, n0+i)
			support++
		}
	}
	if support > p.NVars {
		return nil
	}
	second := Dss_ManComputeTruth(p, e.Dsd1, p.NVars, perm1)
	for i := range first {
		first[i] &= second[i]
	}
	formula, nondec := DecomposeDSD(first, support)
	if nondec < 0 || p.NonDecLimit != 0 && nondec > p.NonDecLimit {
		return nil
	}
	ntk, err := ParseDssNetwork(formula, p.NVars, nil)
	if err != nil {
		panic(err)
	}
	permDsd := make([]int32, p.NVars)
	Dss_NtkTransform(ntk, permDsd)
	lit := Dss_NtkRebuild(p, ntk)
	fans := make([]byte, Dss_VecLitSuppSize(p.Objects, lit))
	for i := range fans {
		source := permDsd[i]
		fans[i] = byte(mapping[source/2]*2 + source&1)
	}
	return &Dss_Fun_t{Dsd: lit, Fans: fans}
}
func Dss_ManSharedMap(p *Dss_Man_t, lits, fanCounts []int32, fans [][]int32, shared uint32) *Dss_Ent_t {
	e := &Dss_Ent_t{Dsd0: lits[0], Dsd1: lits[1]}
	first := make(map[int32]int32)
	for i, v := range fans[0][:fanCounts[0]] {
		first[v/2] = int32(i)*2 + (v & 1)
	}
	for i, v := range fans[1][:fanCounts[1]] {
		if shared>>uint(v/2)&1 == 0 {
			continue
		}
		local, ok := first[v/2]
		if !ok {
			panic("shared variable is absent from first function")
		}
		e.Shared = append(e.Shared, byte(i), byte(local^(v&1)))
	}
	return e
}
func Dss_ManSharedMapDerive(p *Dss_Man_t, a, b int32, shared []byte) *Dss_Ent_t {
	if len(shared)%2 != 0 {
		panic("odd shared-map length")
	}
	return &Dss_Ent_t{Dsd0: a, Dsd1: b, Shared: append([]byte(nil), shared...)}
}
func dssGetAnd(p *Dss_Man_t, e *Dss_Ent_t, total int32) *Dss_Fun_t {
	if p.cache != nil {
		if hit := Dss_ManCacheLookup(p, e); hit != nil {
			return hit.Func
		}
	}
	var f *Dss_Fun_t
	if len(e.Shared) == 0 {
		f = Dss_ManOperationFun(p, []int32{e.Dsd0, e.Dsd1}, total)
	} else {
		f = Dss_ManBooleanAnd(p, e, 0)
	}
	if f != nil && p.cache != nil {
		f = Dss_ManCacheCreate(p, e, f).Func
	}
	return f
}
func Dss_ManMerge(p *Dss_Man_t, lits, fanCounts []int32, fans [][]int32, shared uint32, lutSize int32, resultPerm []byte, truth []uint64) int32 {
	if lutSize > DAU_MAX_VAR || lutSize > p.NVars {
		return -1
	}
	if len(lits) < 2 || len(fanCounts) < 2 || len(fans) < 2 {
		panic("two DSD operands required")
	}
	if lits[0] > lits[1] {
		panic("DSD operands must be ordered")
	}
	for i, lit := range lits[:2] {
		if fanCounts[i] != Dss_VecLitSuppSize(p.Objects, lit) {
			panic("operand support count mismatch")
		}
	}
	if lits[0] == 0 || lits[1] == 0 {
		return 0
	}
	if lits[0] == 1 || lits[1] == 1 {
		which := 0
		if lits[0] == 1 {
			which = 1
		}
		if fanCounts[which] > lutSize {
			return -1
		}
		for i, v := range fans[which][:fanCounts[which]] {
			resultPerm[i] = byte(v)
		}
		return lits[which]
	}
	entry := Dss_ManSharedMap(p, lits, fanCounts, fans, shared)
	total := fanCounts[0] + fanCounts[1] - int32(len(entry.Shared)/2)
	if total > lutSize {
		return -1
	}
	f := dssGetAnd(p, entry, fanCounts[0]+fanCounts[1])
	if f == nil || int32(len(f.Fans)) > lutSize {
		return -1
	}
	for i, v := range f.Fans {
		if int32(v) < fanCounts[0]*2 {
			resultPerm[i] = byte(Dss_Lit2Lit(fans[0], int32(v)))
		} else {
			resultPerm[i] = byte(Dss_Lit2Lit(fans[1], int32(v)-fanCounts[0]*2))
		}
	}
	if shared != 0 && len(f.Fans) > 1 {
		variables := make([]int, 0, len(f.Fans))
		present := make(map[int]bool)
		for _, v := range resultPerm[:len(f.Fans)] {
			variable := int(v / 2)
			if !present[variable] {
				present[variable] = true
				variables = append(variables, variable)
			}
		}
		sort.Ints(variables)
		position := make(map[int]byte)
		for i, v := range variables {
			position[v] = byte(i)
		}
		for i, v := range resultPerm[:len(f.Fans)] {
			resultPerm[i] = position[int(v/2)]*2 + v&1
		}
	}
	return f.Dsd
}
func Mpm_FuncCompute(p *Dss_Man_t, a, b int32, shared []byte, perm []int32, leaves *int32) int32 {
	if *leaves > DAU_MAX_VAR {
		return -1
	}
	entry := Dss_ManSharedMapDerive(p, a, b, shared)
	f := dssGetAnd(p, entry, *leaves)
	if f == nil || len(f.Fans) > int(*leaves) {
		return -1
	}
	*leaves = int32(len(f.Fans))
	for i, v := range f.Fans {
		perm[i] = int32(v)
	}
	return f.Dsd
}
func Dss_ObjCheckTransparent(p *Dss_Man_t, obj *Dss_Obj_t) int32 {
	o := Dss_Regular(obj)
	switch o.Type {
	case DAU_DSD_VAR:
		return 1
	case DAU_DSD_AND, DAU_DSD_CONST0, DAU_DSD_PRIME:
		return 0
	case DAU_DSD_XOR:
		for _, lit := range o.Fans {
			if Dss_ObjCheckTransparent(p, Dss_Lit2Obj(p.Objects, lit)) != 0 {
				return 1
			}
		}
		return 0
	case DAU_DSD_MUX:
		return Dss_ObjCheckTransparent(p, Dss_Lit2Obj(p.Objects, o.Fans[1])) & Dss_ObjCheckTransparent(p, Dss_Lit2Obj(p.Objects, o.Fans[2]))
	}
	panic("unknown DSD node")
}

// Dss_ManTtElems returns independent elementary tables, including the source's
// final scratch-table slot. It avoids mutable process-global scratch storage.
func Dss_ManTtElems() [][]uint64 {
	tables := make([][]uint64, DAU_MAX_VAR+1)
	for v := 0; v <= DAU_MAX_VAR; v++ {
		tables[v] = make([]uint64, dssWords(DAU_MAX_VAR))
		if v < DAU_MAX_VAR {
			for bit := 0; bit < 64*len(tables[v]); bit++ {
				if bit&(1<<uint(v)) != 0 {
					dssSet(tables[v], bit)
				}
			}
		}
	}
	return tables
}
func Dau_DsdTest__() { ntk := Dss_NtkCreate("(!(a!(bh))[cde]!(fg))", 8, nil); Dss_NtkFree(ntk) }

// The original enumeration driver returns before its body (dauTree.c:1852).
func Dau_DsdTest() {}
func Dau_DsdTest444() {
	p := Dss_ManAlloc(6, 0)
	inner := Dss_ManOperation(p, DAU_DSD_AND, []int32{2, 2}, 2, nil, nil) ^ 1
	left := Dss_ManOperation(p, DAU_DSD_AND, []int32{inner, inner}, 2, nil, nil) ^ 1
	right := Dss_ManOperation(p, DAU_DSD_AND, []int32{2, 2, 2}, 3, nil, nil)
	fans := [][]int32{{0, 2, 5, 6}, {2, 9, 10}}
	perm := make([]byte, 6)
	result := Dss_ManMerge(p, []int32{left, right}, []int32{4, 3}, fans, 2, 6, perm, nil)
	Dss_ManPrintOne(os.Stdout, p, left, fans[0])
	Dss_ManPrintOne(os.Stdout, p, right, fans[1])
	if result >= 0 {
		mapped := make([]int32, Dss_VecLitSuppSize(p.Objects, result))
		for i := range mapped {
			mapped[i] = int32(perm[i])
		}
		Dss_ManPrintOne(os.Stdout, p, result, mapped)
	}
}
