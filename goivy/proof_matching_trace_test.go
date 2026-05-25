package goivy

import (
	"strings"
	"testing"
)

func TestSetupSchemaMatchingTraceOrderMatchesPython(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	pc := NewProofChecker(nil, mod, nil, nil, nil, cfg)

	out := captureActionUpdateStdout(t, func() {
		schema := cfg.NewLabeledFormula(cfg.NewAtom("schema"), True)
		decl := cfg.NewLabeledFormula(cfg.NewAtom("decl"), True)
		if _, _, err := pc.SetupSchemaMatchingRaw(decl, nil, nil, schema, false); err != nil {
			t.Fatalf("SetupSchemaMatchingRaw failed: %v", err)
		}
	})

	wrapperIdx := strings.Index(out, "XTRACE: proof.SetupSchemaMatching ENTER schemaLabel=schema declLabel=decl allowWitness=False\n")
	if wrapperIdx < 0 {
		t.Fatalf("missing SetupSchemaMatching wrapper trace:\n%s", out)
	}
	rawIdx := strings.Index(out, "XTRACE: proof.SetupSchemaMatchingRaw ENTER schemaLabel=schema declLabel=decl nmatches=0 allowWitness=False\n")
	if rawIdx < 0 {
		t.Fatalf("missing SetupSchemaMatchingRaw trace:\n%s", out)
	}
	if rawIdx < wrapperIdx {
		t.Fatalf("raw trace preceded wrapper trace; want Python order:\n%s", out)
	}
}

func TestSetupMatchingSuccessTraceOmitsNilErrLikePython(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	pc := NewProofChecker(nil, mod, nil, nil, nil, cfg)

	out := captureActionUpdateStdout(t, func() {
		schema := cfg.NewLabeledFormula(cfg.NewAtom("schema"), True)
		decl := cfg.NewLabeledFormula(cfg.NewAtom("decl"), True)
		pc.Schemata.Set("schema", schema)
		proof := cfg.NewSchemaInstantiation(cfg.NewAtom("schema"), nil)
		if _, _, err := pc.SetupMatching(decl, proof, mod); err != nil {
			t.Fatalf("SetupMatching failed: %v", err)
		}
	})

	if !strings.Contains(out, "XTRACE: proof.SetupMatching EXIT npmatch=0\n") {
		t.Fatalf("missing Python-shaped SetupMatching success trace:\n%s", out)
	}
	if strings.Contains(out, "XTRACE: proof.SetupMatching EXIT err=<nil>") {
		t.Fatalf("SetupMatching success trace included Go-only nil err:\n%s", out)
	}
}
