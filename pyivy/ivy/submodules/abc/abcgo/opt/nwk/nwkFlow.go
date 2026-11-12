package nwk

import "fmt"

func Nwk_ObjPred(o *Nwk_Obj_t) *Nwk_Obj_t      { n, _ := o.Copy.(*Nwk_Obj_t); return n }
func Nwk_ObjSetPred(o, n *Nwk_Obj_t) int32     { o.Copy = n; return 1 }
func Nwk_ObjIsSink(o *Nwk_Obj_t) int32         { return bit(o.MarkA) }
func Nwk_ObjSetSink(o *Nwk_Obj_t)              { o.MarkA = true }
func Nwk_ObjHasFlow(o *Nwk_Obj_t) int32        { return bit(o.MarkB) }
func Nwk_ObjSetFlow(o *Nwk_Obj_t)              { o.MarkB = true }
func Nwk_ObjClearFlow(o *Nwk_Obj_t)            { o.MarkB = false }
func Nwk_ObjVisitedBotOnly(o *Nwk_Obj_t) int32 { return bit(o.TravID == o.Manager.TravID-2) }
func Nwk_ObjVisitedBot(o *Nwk_Obj_t) int32 {
	return bit(o.TravID == o.Manager.TravID-2 || o.TravID == o.Manager.TravID)
}
func Nwk_ObjVisitedTop(o *Nwk_Obj_t) int32 {
	return bit(o.TravID == o.Manager.TravID-1 || o.TravID == o.Manager.TravID)
}
func Nwk_ObjSetVisitedBot(o *Nwk_Obj_t) {
	if o.TravID < o.Manager.TravID-2 {
		o.TravID = o.Manager.TravID - 2
	} else if o.TravID == o.Manager.TravID-1 {
		o.TravID = o.Manager.TravID
	} else {
		panic("nwk: repeated bottom visit")
	}
}
func Nwk_ObjSetVisitedTop(o *Nwk_Obj_t) {
	if o.TravID < o.Manager.TravID-2 {
		o.TravID = o.Manager.TravID - 1
	} else if o.TravID == o.Manager.TravID-2 {
		o.TravID = o.Manager.TravID
	} else {
		panic("nwk: repeated top visit")
	}
}
func Nwk_ManIncrementTravIdFlow(p *Nwk_Man_t) {
	Nwk_ManIncrementTravId(p)
	Nwk_ManIncrementTravId(p)
	Nwk_ManIncrementTravId(p)
}
func Nwk_ManMarkTfiCone_rec(o *Nwk_Obj_t) {
	if o.MarkA {
		return
	}
	o.MarkA = true
	for _, n := range o.Fanins {
		Nwk_ManMarkTfiCone_rec(n)
	}
}
func Nwk_ManMarkTfoCone_rec(o *Nwk_Obj_t) {
	if o.MarkA {
		return
	}
	o.MarkA = true
	for _, n := range o.Fanouts {
		Nwk_ManMarkTfoCone_rec(n)
	}
}
func fastFlow(o, pred *Nwk_Obj_t, forward bool) int32 {
	if Nwk_ObjIsTravIdCurrent(o) != 0 {
		return 0
	}
	Nwk_ObjSetTravIdCurrent(o)
	if o.MarkB {
		return 0
	}
	if o.MarkA {
		o.MarkB = true
		return Nwk_ObjSetPred(o, pred)
	}
	neighbors := o.Fanins
	if forward {
		neighbors = o.Fanouts
	}
	for _, n := range neighbors {
		if fastFlow(n, o, forward) != 0 {
			o.MarkB = true
			return Nwk_ObjSetPred(o, pred)
		}
	}
	return 0
}
func Nwk_ManPushForwardFast_rec(o, pred *Nwk_Obj_t) int32  { return fastFlow(o, pred, true) }
func Nwk_ManPushBackwardFast_rec(o, pred *Nwk_Obj_t) int32 { return fastFlow(o, pred, false) }
func Nwk_ManPushForwardBot_rec(o, pred *Nwk_Obj_t) int32 {
	if Nwk_ObjVisitedBot(o) != 0 {
		return 0
	}
	Nwk_ObjSetVisitedBot(o)
	if o.MarkB {
		if previous := Nwk_ObjPred(o); previous != nil && Nwk_ManPushForwardTop_rec(previous, previous) != 0 {
			return Nwk_ObjSetPred(o, pred)
		}
	} else if Nwk_ManPushForwardTop_rec(o, o) != 0 {
		o.MarkB = true
		return Nwk_ObjSetPred(o, pred)
	}
	for _, n := range o.Fanins {
		if Nwk_ManPushForwardBot_rec(n, pred) != 0 {
			return 1
		}
	}
	return 0
}
func Nwk_ManPushForwardTop_rec(o, pred *Nwk_Obj_t) int32 {
	if Nwk_ObjVisitedTop(o) != 0 {
		return 0
	}
	Nwk_ObjSetVisitedTop(o)
	if o.MarkA {
		return 1
	}
	for _, n := range o.Fanouts {
		if Nwk_ManPushForwardBot_rec(n, pred) != 0 {
			return 1
		}
	}
	if o.MarkB && o.Type != NWK_OBJ_CI && Nwk_ManPushForwardBot_rec(o, Nwk_ObjPred(o)) != 0 {
		o.MarkB = false
		return Nwk_ObjSetPred(o, nil)
	}
	return 0
}
func Nwk_ManPushBackwardBot_rec(o, pred *Nwk_Obj_t) int32 {
	if Nwk_ObjVisitedBot(o) != 0 {
		return 0
	}
	Nwk_ObjSetVisitedBot(o)
	if o.MarkB {
		if previous := Nwk_ObjPred(o); previous != nil && Nwk_ManPushBackwardTop_rec(previous, previous) != 0 {
			return Nwk_ObjSetPred(o, pred)
		}
	} else if Nwk_ManPushBackwardTop_rec(o, o) != 0 {
		o.MarkB = true
		return Nwk_ObjSetPred(o, pred)
	}
	return 0
}
func Nwk_ManPushBackwardTop_rec(o, pred *Nwk_Obj_t) int32 {
	if Nwk_ObjVisitedTop(o) != 0 {
		return 0
	}
	Nwk_ObjSetVisitedTop(o)
	if o.MarkA {
		return 1
	}
	for _, n := range o.Fanins {
		if Nwk_ManPushBackwardBot_rec(n, pred) != 0 {
			return 1
		}
	}
	for _, n := range o.Fanouts {
		if o.Type != NWK_OBJ_CO && Nwk_ManPushBackwardTop_rec(n, pred) != 0 {
			return 1
		}
	}
	if o.MarkB && Nwk_ObjPred(o) != nil && Nwk_ManPushBackwardBot_rec(o, Nwk_ObjPred(o)) != 0 {
		o.MarkB = false
		return Nwk_ObjSetPred(o, nil)
	}
	return 0
}
func Nwk_ManVerifyCut_rec(o *Nwk_Obj_t) int32 {
	if o.MarkA {
		return 1
	}
	if Nwk_ObjIsLo(o) != 0 {
		return 0
	}
	if Nwk_ObjIsTravIdCurrent(o) != 0 {
		return 1
	}
	Nwk_ObjSetTravIdCurrent(o)
	for _, n := range o.Fanins {
		if Nwk_ManVerifyCut_rec(n) == 0 {
			return 0
		}
	}
	return 1
}
func Nwk_ManRetimeVerifyCutForward(p *Nwk_Man_t, v *Vec_Ptr_t) int32 {
	nodes := vectorNodes(v)
	for _, o := range nodes {
		o.MarkA = true
	}
	Nwk_ManIncrementTravId(p)
	valid := int32(1)
	for _, o := range p.Outputs {
		if Nwk_ManVerifyCut_rec(o) == 0 {
			valid = 0
		}
	}
	for _, o := range nodes {
		o.MarkA = false
	}
	return valid
}
func Nwk_ManRetimeVerifyCutBackward(p *Nwk_Man_t, v *Vec_Ptr_t) int32 {
	cut := map[*Nwk_Obj_t]bool{}
	for _, o := range vectorNodes(v) {
		cut[o] = true
	}
	seen := map[*Nwk_Obj_t]bool{}
	var visit func(*Nwk_Obj_t) bool
	visit = func(o *Nwk_Obj_t) bool {
		if cut[o] {
			return true
		}
		if o.Type == NWK_OBJ_CI || len(o.Fanins) == 0 || Nwk_ObjIsPo(o) != 0 {
			return false
		}
		if seen[o] {
			return true
		}
		seen[o] = true
		for _, n := range o.Fanins {
			if !visit(n) {
				return false
			}
		}
		return true
	}
	for _, o := range p.Outputs[p.TruePOs:] {
		if !visit(o.Fanins[0]) {
			return 0
		}
	}
	return 1
}
func retimeCut(p *Nwk_Man_t, latches, verbose int32, forward bool) *Vec_Ptr_t {
	if latches < 0 || latches > int32(min(len(p.Inputs), len(p.Outputs))) {
		panic("nwk: invalid register interface")
	}
	p.Latches = latches
	p.TruePIs = int32(len(p.Inputs)) - latches
	p.TruePOs = int32(len(p.Outputs)) - latches
	Nwk_ManCleanMarks(p)
	for _, o := range p.Objects {
		if o != nil {
			o.Copy = nil
		}
	}
	var sources []*Nwk_Obj_t
	if forward {
		for _, o := range p.Outputs {
			o.MarkA = true
		}
		for _, o := range p.Inputs[:p.TruePIs] {
			Nwk_ManMarkTfoCone_rec(o)
		}
		sources = p.Inputs[p.TruePIs:]
	} else {
		for _, o := range p.Inputs {
			o.MarkA = true
		}
		for _, o := range p.Outputs[:p.TruePOs] {
			Nwk_ManMarkTfiCone_rec(o)
		}
		for _, o := range p.Objects {
			if o != nil && o.Type == NWK_OBJ_NODE && len(o.Fanins) == 0 {
				o.MarkA = true
			}
		}
		for _, o := range p.Outputs[p.TruePOs:] {
			sources = append(sources, o.Fanins[0])
		}
	}
	fast, bot := Nwk_ManPushBackwardFast_rec, Nwk_ManPushBackwardBot_rec
	if forward {
		fast, bot = Nwk_ManPushForwardFast_rec, Nwk_ManPushForwardBot_rec
	}
	Nwk_ManIncrementTravIdFlow(p)
	flow := 0
	for _, o := range sources {
		if fast(o, nil) != 0 {
			Nwk_ManIncrementTravIdFlow(p)
			flow++
		}
	}
	Nwk_ManIncrementTravIdFlow(p)
	additional := 0
	for _, o := range sources {
		if bot(o, nil) != 0 {
			Nwk_ManIncrementTravIdFlow(p)
			additional++
		}
	}
	if additional > 0 {
		Nwk_ManIncrementTravIdFlow(p)
		for _, o := range sources {
			if bot(o, nil) != 0 {
				panic("nwk: augmenting path remains")
			}
		}
	}
	cut := []*Nwk_Obj_t{}
	for _, o := range p.Objects {
		if o != nil && Nwk_ObjVisitedBotOnly(o) != 0 {
			cut = append(cut, o)
		}
	}
	Nwk_ManCleanMarks(p)
	if verbose != 0 {
		fmt.Fprintf(writer(p), "Max-flow = %d. Min-cut = %d.\n", flow+additional, len(cut))
	}
	return nodesVector(cut)
}
func Nwk_ManRetimeCutForward(p *Nwk_Man_t, latches, verbose int32) *Vec_Ptr_t {
	return retimeCut(p, latches, verbose, true)
}
func Nwk_ManRetimeCutBackward(p *Nwk_Man_t, latches, verbose int32) *Vec_Ptr_t {
	return retimeCut(p, latches, verbose, false)
}
