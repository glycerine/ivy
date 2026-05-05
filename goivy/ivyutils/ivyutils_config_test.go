package ivyutils

import (
	"testing"
)

func TestConfigDefaults(t *testing.T) {

	cfg := NewIvyUtilsConfig()

	// Verify correct defaults
	if !cfg.UseNumerals {
		t.Error("UseNumerals default should be true")
	}
	if cfg.UseNewUI {
		t.Error("UseNewUI default should be false")
	}
	//if ErrorWillPanic {
	//	t.Error("ErrorWillPanic default should be false")
	//}
	if cfg.DefaultUI != "cti" {
		t.Errorf("DefaultUI default = %q, want 'cti'", cfg.DefaultUI)
	}
	if cfg.EnableDebug {
		t.Error("EnableDebug default should be false")
	}
}
