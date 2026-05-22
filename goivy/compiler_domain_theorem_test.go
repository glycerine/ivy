package goivy

import "testing"

func TestDomainSetupTheoremCompilesSchemaBody(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	body := cfg.NewSchemaBody(True)
	def := cfg.NewDefinition(cfg.NewAtom("extensionality"), body)
	schema := cfg.NewSchema(def)

	if err := d.Theorem(schema); err != nil {
		t.Fatalf("Theorem: %v", err)
	}
	if len(c.Module.LabeledProps) != 1 {
		t.Fatalf("LabeledProps len = %d, want 1", len(c.Module.LabeledProps))
	}
	prop := c.Module.LabeledProps[0]
	if prop.LabelName() != "extensionality" {
		t.Fatalf("property label = %q, want extensionality", prop.LabelName())
	}
	if d.LastFact != prop {
		t.Fatal("LastFact was not updated to the theorem property")
	}
	if got := c.Module.Theorems["extensionality"]; got == nil {
		t.Fatal("theorem was not registered")
	}
}
