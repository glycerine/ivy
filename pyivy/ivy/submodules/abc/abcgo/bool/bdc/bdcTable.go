package bdc

func Bdc_TableCheckContainment(_ *Bdc_Man_t, isf *Bdc_Isf_t, t []uint32) int32 {
	return bdcFlag(ttImpl(isf.On, t) && ttDisjoint(t, isf.Off))
}
func Bdc_TableLookup(p *Bdc_Man_t, f *Bdc_Isf_t) *Bdc_Fun_t {
	if f.Supp == 0 {
		if ttFull(f.On) {
			return p.nodes[0]
		}
		if !ttFull(f.Off) {
			panic("bdc: invalid constant ISF")
		}
		return Bdc_Not(p.nodes[0])
	}
	for i := len(p.table[f.Supp]) - 1; i >= 0; i-- {
		node := p.table[f.Supp][i]
		if Bdc_TableCheckContainment(p, f, node.Func) != 0 {
			return node
		}
	}
	Bdc_IsfNot(f)
	defer Bdc_IsfNot(f)
	for i := len(p.table[f.Supp]) - 1; i >= 0; i-- {
		node := p.table[f.Supp][i]
		if Bdc_TableCheckContainment(p, f, node.Func) != 0 {
			return Bdc_Not(node)
		}
	}
	return nil
}
func Bdc_TableAdd(p *Bdc_Man_t, f *Bdc_Fun_t) {
	p.table[f.Supp] = append(p.table[f.Supp], f)
}
func Bdc_TableClear(p *Bdc_Man_t) { clear(p.table) }
