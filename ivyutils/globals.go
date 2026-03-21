package ivyutils

import (
	"fmt"
)

// GlobalRegistry is the default parameter registry for ivy_utils parameters.
// Corresponds to Python's module-level parameter definitions in ivy_utils.py.
var GlobalRegistry = NewParameterRegistry()

// Global parameters corresponding to Python's ivy_utils.py module-level parameters.
var (
	// UseNumerals controls whether numerals are used in model display.
	// Corresponds to Python: use_numerals = BooleanParameter("use_numerals", True)
	UseNumerals = NewBooleanParameterOn(GlobalRegistry, "use_numerals", true)

	// UseNewUI controls whether the new UI is used.
	// Corresponds to Python: use_new_ui = BooleanParameter("new_ui", False)
	UseNewUI = NewBooleanParameterOn(GlobalRegistry, "new_ui", false)

	// Catch controls whether IvyError is caught or triggers an assertion.
	// Corresponds to Python: catch = BooleanParameter("catch", True)
	Catch = NewBooleanParameterOn(GlobalRegistry, "catch", true)

	// DefaultUI controls which UI module is loaded.
	// Corresponds to Python: default_ui = Parameter("ui", "cti")
	DefaultUI = NewParameterOn(GlobalRegistry, "ui", "cti")

	// EnableDebug controls debug output.
	// Corresponds to Python: enable_debug = BooleanParameter("debug", False)
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
