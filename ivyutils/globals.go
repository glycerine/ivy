package ivyutils

import (
	"fmt"
)

// GlobalRegistry is a transitional global; use IvyUtilsConfig.Registry instead.
var GlobalRegistry = NewParameterRegistry()

// Transitional globals; use IvyUtilsConfig fields instead.
var (
	UseNumerals = NewBooleanParameterOn(GlobalRegistry, "use_numerals", true)
	UseNewUI    = NewBooleanParameterOn(GlobalRegistry, "new_ui", false)
	Catch       = NewBooleanParameterOn(GlobalRegistry, "catch", true)
	DefaultUI   = NewParameterOn(GlobalRegistry, "ui", "cti")
	EnableDebug = NewBooleanParameterOn(GlobalRegistry, "debug", false)
)

// Dbg prints key-value debug information.
// Corresponds to Python's dbg(*exprs) which evaluates expressions in the caller's frame.
// Since Go has no runtime eval, callers must pass key-value pairs: Dbg("x", x, "y", y).
func Dbg(kvs ...interface{}) {
	if !EnableDebug.GetBool() {
		panic("must use debug=true to enable debug output")
	}
	for i := 0; i+1 < len(kvs); i += 2 {
		fmt.Printf("%v:%v\n", kvs[i], kvs[i+1])
	}
}

// UIModule represents a UI module with an IvyUI class and compile_kwargs.
// Corresponds to Python's dynamically imported ivy_ui_* modules.
type UIModule struct {
	Name          string
	NewIvyUI      func() interface{}
	CompileKwargs map[string]interface{}
}

// uiModules is a transitional global; use IvyUtilsConfig.UIModules instead.
var uiModules = make(map[string]*UIModule)

// RegisterUIModule registers a UI module by name.
func RegisterUIModule(name string, mod *UIModule) {
	uiModules[name] = mod
}

// GetDefaultUIModule returns the UI module selected by the DefaultUI parameter.
// Corresponds to Python's get_default_ui_module().
func GetDefaultUIModule() *UIModule {
	defui := DefaultUI.GetString()
	name := defui
	if defui == "art" {
		name = "ivy_ui"
	} else {
		name = "ivy_ui_" + defui
	}
	mod, ok := uiModules[name]
	if !ok {
		return nil
	}
	return mod
}

// GetDefaultUIClass returns the IvyUI constructor from the default UI module.
// Corresponds to Python's get_default_ui_class().
func GetDefaultUIClass() func() interface{} {
	mod := GetDefaultUIModule()
	if mod == nil {
		return nil
	}
	return mod.NewIvyUI
}

// GetDefaultUICompileKwargs returns compile kwargs from the default UI module.
// Corresponds to Python's get_default_ui_compile_kwargs().
func GetDefaultUICompileKwargs() map[string]interface{} {
	mod := GetDefaultUIModule()
	if mod == nil {
		return nil
	}
	return mod.CompileKwargs
}
