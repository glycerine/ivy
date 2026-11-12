package dsd

import bdd "github.com/glycerine/abcgo/bdd/cudd/core"

func Dsd_ManagerStart(dd *DdManager, nSuppMax, fVerbose int) *Dsd_Manager_t {
	if dd == nil || nSuppMax < 0 || nSuppMax > len(dd.Vars) {
		panic("dsd: invalid input count")
	}
	m := &Dsd_Manager_t{dd: dd, table: make(map[*DdNode]*Dsd_Node_t), verbose: fVerbose != 0, identity: make(map[identityKey]bool)}
	for _, v := range dd.Vars[:nSuppMax] {
		p := Dsd_TreeNodeCreate(int(DSD_NODE_BUF), 1, 0)
		p.G, p.S = v, v
		bdd.Cudd_Ref(v)
		bdd.Cudd_Ref(v)
		m.inputs = append(m.inputs, p)
		m.table[v] = p
	}
	m.constant = Dsd_TreeNodeCreate(int(DSD_NODE_CONST1), 0, 0)
	m.constant.G, m.constant.S = dd.One, dd.One
	bdd.Cudd_Ref(dd.One)
	bdd.Cudd_Ref(dd.One)
	m.table[dd.One] = m.constant
	return m
}
func Dsd_ManagerStop(m *Dsd_Manager_t) {
	if m == nil {
		return
	}
	seen := map[*Dsd_Node_t]bool{}
	for _, p := range m.table {
		r := Dsd_Regular(p)
		if !seen[r] {
			seen[r] = true
			Dsd_TreeNodeDelete(m.dd, r)
		}
	}
	m.table = nil
	m.inputs = nil
	m.roots = nil
	m.constant = nil
	m.identity = nil
}
