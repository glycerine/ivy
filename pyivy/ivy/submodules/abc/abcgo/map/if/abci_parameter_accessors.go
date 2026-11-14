package if_

// The translated ABC command code lives outside this package. These accessors
// expose the corresponding If_Par_t_ fields without exporting the struct's
// storage directly.
func (p *If_Par_t_) CutMinFlag() int32      { return p.FCutMin }
func (p *If_Par_t_) SetCutMinFlag(v int32)  { p.FCutMin = v }
func (p *If_Par_t_) VerboseFlag() int32     { return p.FVerbose }
func (p *If_Par_t_) SetVerboseFlag(v int32) { p.FVerbose = v }
func (p *If_Par_t_) ExpRedFlag() int32      { return p.FExpRed }
func (p *If_Par_t_) SetExpRedFlag(v int32)  { p.FExpRed = v }

func (p *If_Par_t_) LUTLibrary() *If_LibLut_t     { return p.PLutLib }
func (p *If_Par_t_) SetLUTLibrary(v *If_LibLut_t) { p.PLutLib = v }

func (p *If_Par_t_) SetLUTStructure(v string) { p.PLutStruct = v }
