package aig

// Aig_ManHasFanout reports whether the graph currently owns a fanout index.
func Aig_ManHasFanout(p *Graph) bool { return p != nil && p.fanouts != nil }
