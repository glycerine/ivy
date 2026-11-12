package hop

// These methods satisfy the generic factoring package's Network interfaces
// without requiring dec to depend on a concrete graph implementation.
func (p *Hop_Man_t) ConstantOne() any     { return Hop_ManConst1(p) }
func (p *Hop_Man_t) Complement(a any) any { return Hop_Not(a.(*Hop_Obj_t)) }
func (p *Hop_Man_t) And(a, b any) any     { return Hop_And(p, a.(*Hop_Obj_t), b.(*Hop_Obj_t)) }
func (p *Hop_Man_t) Variable(i int) any   { return Hop_IthVar(p, int32(i)) }
