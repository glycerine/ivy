package gia

// Tas2_Var_t is the byte-store solver's variable record from giaCTas2.c.
// It is separate from giaCTas.c's solver, which stores assignment state on
// Gia_Obj_t values directly.
type Tas2_Var_t struct {
	fTerm, fVal, fValOld, fAssign uint32
	fJQueue, fCompl0, fCompl1     uint32
	fMark0, fMark1, fPhase        uint32
	Level                         uint32
	Id, IdAig                     int32
	Reason0, Reason1              int32
	Diff0, Diff1                  int32
	Watch0, Watch1                int32
}

// Tas2_Man_t owns stable variable pointers, analogous to the resizable byte
// store in the original C manager.
type Tas2_Man_t struct {
	pVars []*Tas2_Var_t
}

// Tas2_ManCreateVar appends a zero-valued variable and assigns its unique ID.
func Tas2_ManCreateVar(p *Tas2_Man_t) *Tas2_Var_t {
	if p == nil {
		return nil
	}
	variable := &Tas2_Var_t{Id: int32(len(p.pVars))}
	p.pVars = append(p.pVars, variable)
	return variable
}
