package goivy

// CompilerConfig holds per-session compiler state.
// Moved from compiler/phase6.go to break the import cycle
// between compiler and module.
type CompilerConfig struct {
	OptionVerifying bool
	PropIDCounter   int64
	ModCfg          *Config
}

// NewCompilerConfig creates a new CompilerConfig.
func NewCompilerConfig(modCfg *Config) *CompilerConfig {
	return &CompilerConfig{ModCfg: modCfg}
}

// FreshPropID generates a fresh unique ID for labeled formulas.
func (cc *CompilerConfig) FreshPropID() int64 {
	cc.PropIDCounter++
	return cc.PropIDCounter
}

// SetVerifying sets the verifying flag on the CompilerConfig.
// Corresponds to Python's option_verifying flag.
func (cc *CompilerConfig) SetVerifying(v bool) {
	cc.OptionVerifying = v
}

// GetVerifying returns the current value of the option_verifying flag.
func (cc *CompilerConfig) GetVerifying() bool {
	return cc.OptionVerifying
}

// SetVerifyingOnMod sets the verifying flag. Uses the CompilerConfig stored
// on the module if available, otherwise panics.
func SetVerifyingOnMod(mod *Module, v bool) {
	if mod != nil && mod.CompCfg != nil {
		mod.CompCfg.SetVerifying(v)
		return
	}
	panic("SetVerifyingOnMod: module has no CompCfg")
}

// GetVerifyingFromCfg returns the option_verifying flag from the given
// CompilerConfig. For backward compatibility, also accepts nil (returns false).
func GetVerifyingFromCfg(cc ...*CompilerConfig) bool {
	if len(cc) > 0 && cc[0] != nil {
		return cc[0].OptionVerifying
	}
	return false
}

// GetModCompCfg extracts the CompilerConfig from a module, or returns nil.
func GetModCompCfg(mod *Module) *CompilerConfig {
	if mod == nil || mod.CompCfg == nil {
		return nil
	}
	return mod.CompCfg
}

// GetModVerifying returns the verifying flag from the module's CompilerConfig.
func GetModVerifying(mod *Module) bool {
	cc := GetModCompCfg(mod)
	if cc != nil {
		return cc.OptionVerifying
	}
	return false
}

// GetModFreshPropID generates a fresh prop ID via the module's CompilerConfig.
// Panics if no CompilerConfig is available — callers must ensure one exists.
func GetModFreshPropID(mod *Module) int64 {
	cc := GetModCompCfg(mod)
	if cc != nil {
		return cc.FreshPropID()
	}
	panic("GetModFreshPropID: module has no CompilerConfig — ensure mod.CompCfg is initialized")
}
