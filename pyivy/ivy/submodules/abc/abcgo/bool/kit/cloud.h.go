package kit

// Cloud is ABC's reduced ordered BDD manager. Complemented edges use a paired
// Go node wrapper rather than tagged pointers; no unsafe pointer arithmetic.
type CloudVar = uint32
type CloudSign = uint32
type CloudOper int32

const (
	CLOUD_OPER_AND CloudOper = iota
	CLOUD_OPER_XOR
	CLOUD_OPER_BDIFF
	CLOUD_OPER_LEQ
)
const (
	CLOUD_NODE_BITS          = 23
	CLOUD_CONST_INDEX uint32 = 0x0fffffff
	CLOUD_MARK_ON     uint32 = 0x10000000
	CLOUD_MARK_OFF    uint32 = 0xefffffff
)
const (
	DD_P1 = 12582917
	DD_P2 = 4256249
	DD_P3 = 741457
	DD_P4 = 1618033999
)

type cloudKey struct {
	v    uint32
	t, e *CloudNode
}
type cloudPair struct{ f, g *CloudNode }
type CloudManager struct {
	nVars, bitsNode, nNodesLimit                         int32
	cacheLimit                                           int
	nSignCur                                             CloudSign
	nUniqueHits, nUniqueMisses, nCacheHits, nCacheMisses int32
	one, zero                                            *CloudNode
	vars, ppNodes, tUnique                               []*CloudNode
	unique                                               map[cloudKey]*CloudNode
	andCache                                             map[cloudPair]*CloudNode
}
type CloudNode struct {
	s                 CloudSign
	v                 CloudVar
	e, t              *CloudNode
	regular, opposite *CloudNode
	complement        bool
	owner             *CloudManager
	id                uint32
}
type CloudCacheEntry1 struct {
	s    CloudSign
	a, r *CloudNode
}

// Cloud_One and Cloud_Var return nodes owned by the supplied manager.
func Cloud_One(dd *CloudManager) *CloudNode          { return dd.one }
func Cloud_Var(dd *CloudManager, i int32) *CloudNode { return dd.vars[i] }

type CloudCacheEntry2 struct {
	s       CloudSign
	a, b, r *CloudNode
}
type CloudCacheEntry3 struct {
	s          CloudSign
	a, b, c, r *CloudNode
}

func Cloud_Regular(p *CloudNode) *CloudNode {
	if p == nil {
		return nil
	}
	return p.regular
}
func Cloud_Not(p *CloudNode) *CloudNode {
	if p == nil {
		return nil
	}
	return p.opposite
}
func Cloud_NotCond(p *CloudNode, c bool) *CloudNode {
	if c {
		return Cloud_Not(p)
	}
	return p
}
func Cloud_IsComplement(p *CloudNode) bool { return p != nil && p.complement }
func Cloud_IsConstant(p *CloudNode) bool {
	return p != nil && (p.regular.v&CLOUD_MARK_OFF) == CLOUD_CONST_INDEX
}
func cloudIsConstant(p *CloudNode) bool   { return Cloud_IsConstant(p) }
func Cloud_V(p *CloudNode) CloudVar       { return p.regular.v & CLOUD_MARK_OFF }
func Cloud_E(p *CloudNode) *CloudNode     { return p.regular.e }
func Cloud_T(p *CloudNode) *CloudNode     { return p.regular.t }
func cloudV(p *CloudNode) CloudVar        { return Cloud_V(p) }
func cloudE(p *CloudNode) *CloudNode      { return p.e }
func cloudT(p *CloudNode) *CloudNode      { return p.t }
func cloudNodeMark(p *CloudNode)          { p.regular.v |= CLOUD_MARK_ON }
func cloudNodeUnmark(p *CloudNode)        { p.regular.v &= CLOUD_MARK_OFF }
func cloudNodeIsMarked(p *CloudNode) bool { return p.regular.v&CLOUD_MARK_ON != 0 }
func cloudCacheLookup1(p *CloudCacheEntry1, s CloudSign, f *CloudNode) *CloudNode {
	if p.s == s && p.a == f {
		return p.r
	}
	return nil
}
func cloudCacheLookup2(p *CloudCacheEntry2, s CloudSign, f, g *CloudNode) *CloudNode {
	if p.s == s && p.a == f && p.b == g {
		return p.r
	}
	return nil
}
func cloudCacheLookup3(p *CloudCacheEntry3, s CloudSign, f, g, h *CloudNode) *CloudNode {
	if p.s == s && p.a == f && p.b == g && p.c == h {
		return p.r
	}
	return nil
}
func cloudCacheInsert1(p *CloudCacheEntry1, s CloudSign, f, r *CloudNode) {
	*p = CloudCacheEntry1{s, f, r}
}
func cloudCacheInsert2(p *CloudCacheEntry2, s CloudSign, f, g, r *CloudNode) {
	*p = CloudCacheEntry2{s, f, g, r}
}
func cloudCacheInsert3(p *CloudCacheEntry3, s CloudSign, f, g, h, r *CloudNode) {
	*p = CloudCacheEntry3{s, f, g, h, r}
}
