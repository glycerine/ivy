package dau

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
func Dss_ManPrintOne(w io.Writer, p *Dss_Man_t, lit int32, perm []int32) error {
	count := int32(0)
	return Dss_ManPrint_rec(w, p, Dss_Lit2Obj(p.Objects, lit), perm, &count)
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
		fmt.Fprintf(&b, "%d: ", o.ID)
		Dss_ManPrintOne(&b, p, o.ID*2, nil)
		b.WriteByte('\n')
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
		sort.SliceStable(children, func(i, j int) bool { return Dss_ObjCompare(p.Objects, children[i].obj, children[j].obj) < 0 })
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
