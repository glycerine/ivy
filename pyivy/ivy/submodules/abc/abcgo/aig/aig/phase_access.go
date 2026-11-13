package aig

// Aig_ObjSetPhase sets the simulation phase on an AIG object.
func Aig_ObjSetPhase(g *Graph, e Edge, phase bool) { g.node(e).Phase = phase }
