package nwk

func Nwk_ManCreateObj(p *Nwk_Man_t, fanins, fanouts int32) *Nwk_Obj_t {
	if fanins < 0 || fanouts < 0 {
		panic("nwk: negative adjacency capacity")
	}
	o := &Nwk_Obj_t{Manager: p, ID: int32(len(p.Objects)), Fanins: make([]*Nwk_Obj_t, 0, fanins), Fanouts: make([]*Nwk_Obj_t, 0, fanouts)}
	p.Objects = append(p.Objects, o)
	return o
}
func Nwk_ManCreateCi(p *Nwk_Man_t, fanouts int32) *Nwk_Obj_t {
	o := Nwk_ManCreateObj(p, 0, fanouts)
	o.Type = NWK_OBJ_CI
	o.PioID = int32(len(p.Inputs))
	p.Inputs = append(p.Inputs, o)
	return o
}
func Nwk_ManCreateCo(p *Nwk_Man_t) *Nwk_Obj_t {
	o := Nwk_ManCreateObj(p, 1, 0)
	o.Type = NWK_OBJ_CO
	o.PioID = int32(len(p.Outputs))
	p.Outputs = append(p.Outputs, o)
	return o
}
func Nwk_ManCreateLatch(p *Nwk_Man_t) *Nwk_Obj_t {
	o := Nwk_ManCreateObj(p, 1, 1)
	o.Type = NWK_OBJ_LATCH
	return o
}
func Nwk_ManCreateNode(p *Nwk_Man_t, fanins, fanouts int32) *Nwk_Obj_t {
	o := Nwk_ManCreateObj(p, fanins, fanouts)
	o.Type = NWK_OBJ_NODE
	return o
}
func Nwk_ManDeleteNode(o *Nwk_Obj_t) {
	if len(o.Fanouts) != 0 {
		panic("nwk: deleting referenced node")
	}
	if o.Type == NWK_OBJ_CI || o.Type == NWK_OBJ_CO {
		panic("nwk: cannot delete terminal through node API")
	}
	p := o.Manager
	for len(o.Fanins) > 0 {
		Nwk_ObjDeleteFanin(o, o.Fanins[len(o.Fanins)-1])
	}
	p.Objects[o.ID] = nil
	if o.Type == NWK_OBJ_CI || o.Type == NWK_OBJ_CO {
		panic("nwk: terminals require interface reconstruction")
	}
	*o = Nwk_Obj_t{ID: -1}
}
func Nwk_ManDeleteNode_rec(o *Nwk_Obj_t) {
	if o.Type == NWK_OBJ_CI {
		panic("nwk: deleting input cone")
	}
	fanins := append([]*Nwk_Obj_t(nil), o.Fanins...)
	Nwk_ManDeleteNode(o)
	for _, fanin := range fanins {
		if fanin.Type == NWK_OBJ_NODE && len(fanin.Fanouts) == 0 {
			Nwk_ManDeleteNode_rec(fanin)
		}
	}
}
