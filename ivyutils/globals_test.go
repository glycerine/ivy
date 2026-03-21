package ivyutils

import (
	"testing"
)

func TestGlobalParameters(t *testing.T) {
	// Verify all global parameters exist and have correct defaults
	if !UseNumerals.GetBool() {
		t.Error("UseNumerals default should be true")
	}
	if UseNewUI.GetBool() {
		t.Error("UseNewUI default should be false")
	}
	if !Catch.GetBool() {
		t.Error("Catch default should be true")
	}
	if DefaultUI.GetString() != "cti" {
		t.Errorf("DefaultUI default = %q, want 'cti'", DefaultUI.GetString())
	}
	if EnableDebug.GetBool() {
		t.Error("EnableDebug default should be false")
	}
}

func TestGlobalParameterKeys(t *testing.T) {
	// Verify parameter keys match Python
	if UseNumerals.Key != "use_numerals" {
		t.Errorf("UseNumerals.Key = %q, want 'use_numerals'", UseNumerals.Key)
	}
	if UseNewUI.Key != "new_ui" {
		t.Errorf("UseNewUI.Key = %q, want 'new_ui'", UseNewUI.Key)
	}
	if Catch.Key != "catch" {
		t.Errorf("Catch.Key = %q, want 'catch'", Catch.Key)
	}
	if DefaultUI.Key != "ui" {
		t.Errorf("DefaultUI.Key = %q, want 'ui'", DefaultUI.Key)
	}
	if EnableDebug.Key != "debug" {
		t.Errorf("EnableDebug.Key = %q, want 'debug'", EnableDebug.Key)
	}
}

func TestGlobalRegistry(t *testing.T) {
	// All parameters should be in the GlobalRegistry
	for _, key := range []string{"use_numerals", "new_ui", "catch", "ui", "debug"} {
		if _, ok := GlobalRegistry.Get(key); !ok {
			t.Errorf("GlobalRegistry missing parameter %q", key)
		}
	}
}

func TestDbgPanicsWhenDisabled(t *testing.T) {
	EnableDebug.Value = false
	defer func() {
		if r := recover(); r == nil {
			t.Error("Dbg should panic when debug is disabled")
		}
	}()
	Dbg("x", 42)
}

func TestDbgWhenEnabled(t *testing.T) {
	EnableDebug.Value = true
	defer func() { EnableDebug.Value = false }()
	// Should not panic
	Dbg("x", 42, "y", "hello")
}

func TestGetDefaultUIModule(t *testing.T) {
	// No modules registered: should return nil
	mod := GetDefaultUIModule()
	if mod != nil {
		t.Error("GetDefaultUIModule with no registrations should return nil")
	}

	// Register a module
	RegisterUIModule("ivy_ui_cti", &UIModule{Name: "cti"})
	mod = GetDefaultUIModule()
	if mod == nil {
		t.Fatal("GetDefaultUIModule should find 'ivy_ui_cti'")
	}
	if mod.Name != "cti" {
		t.Errorf("module name = %q, want 'cti'", mod.Name)
	}

	// Clean up
	delete(uiModules, "ivy_ui_cti")
}

func TestGetDefaultUIModuleArt(t *testing.T) {
	// When default_ui is "art", module name should be "ivy_ui" (not "ivy_ui_art")
	DefaultUI.Value = "art"
	defer func() { DefaultUI.Value = "cti" }()

	RegisterUIModule("ivy_ui", &UIModule{Name: "art_ui"})
	mod := GetDefaultUIModule()
	if mod == nil {
		t.Fatal("GetDefaultUIModule should find 'ivy_ui' when default is 'art'")
	}
	if mod.Name != "art_ui" {
		t.Errorf("module name = %q, want 'art_ui'", mod.Name)
	}
	delete(uiModules, "ivy_ui")
}
