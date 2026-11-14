package if_

// The translated ABC command code lives outside this package. These accessors
// expose the corresponding If_Par_t_ fields without exporting the struct's
// storage directly.
func (p *If_Par_t_) CutMinFlag() int32      { return p.fCutMin }
func (p *If_Par_t_) SetCutMinFlag(v int32)  { p.fCutMin = v }
func (p *If_Par_t_) VerboseFlag() int32     { return p.fVerbose }
func (p *If_Par_t_) SetVerboseFlag(v int32) { p.fVerbose = v }
func (p *If_Par_t_) ExpRedFlag() int32      { return p.fExpRed }
func (p *If_Par_t_) SetExpRedFlag(v int32)  { p.fExpRed = v }

func (p *If_Par_t_) LUTLibrary() *If_LibLut_t     { return p.pLutLib }
func (p *If_Par_t_) SetLUTLibrary(v *If_LibLut_t) { p.pLutLib = v }

func (p *If_Par_t_) SetLUTStructure(v string) { p.pLutStruct = v }
