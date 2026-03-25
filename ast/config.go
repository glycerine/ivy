package ast

import "sync/atomic"

// AstConfig holds per-session AST state that was previously stored in
// package-level globals. Each concurrent Ivy model gets its own AstConfig,
// enabling safe multi-tenancy on multi-core machines.
//
// Constructors are methods on *AstConfig (e.g. cfg.NewAtom("x")),
// which stores cfg on the created node's Base.Cfg field. Clone methods
// access cfg from their receiver's Base, so the Node interface is unchanged.
type AstConfig struct {
	// ReferenceLineno matches Python's reference_lineno from ivy_ast.py:13.
	// When set (non-zero), LinenoAddRef wraps cloned node linenos with this reference.
	ReferenceLineno Location

	// ChoiceActionCounter generates unique IDs for ChoiceAction nodes.
	ChoiceActionCounter int64

	// LocalActionCtr generates unique IDs for LocalAction nodes.
	LocalActionCtr int

	// CallActionCtr generates unique IDs for CallAction nodes.
	CallActionCtr int

	// LfCounter generates unique IDs for LabeledFormula nodes.
	LfCounter int64

	// AlwaysCloneWithFreshID mirrors Python's always_clone_with_fresh_id (ivy_ast.py:615).
	// When true, LabeledFormula.Clone() allocates a fresh ID instead of preserving the original.
	// Set to true during instMod (module instantiation).
	AlwaysCloneWithFreshID bool

	// LabelCounter generates unique label/mixer names across all parses in a session.
	// Python: label_counter (ivy_parser.py:495) — module-level global.
	LabelCounter int

	// CheckUnprovable matches Python's check_unprovable (ivy_actions.py:25).
	// When false (default), unprovable declarations are silently dropped.
	// When true, they are declared normally.
	CheckUnprovable bool
}

// NewAstConfig creates a fresh AstConfig with default values.
func NewAstConfig() *AstConfig {
	return &AstConfig{}
}

// SetReferenceLineno sets the reference lineno for LinenoAddRef.
// Matches Python set_reference_lineno() from ivy_ast.py:15.
func (cfg *AstConfig) SetReferenceLineno(lineno Location) {
	cfg.ReferenceLineno = lineno
}

// GetReferenceLineno returns the current reference lineno.
func (cfg *AstConfig) GetReferenceLineno() Location {
	return cfg.ReferenceLineno
}

// LinenoAddRef wraps a location with the current reference lineno.
// Matches Python lineno_add_ref (ivy_ast.py:19-22).
func (cfg *AstConfig) LinenoAddRef(loc Location) Location {
	if cfg.ReferenceLineno == (Location{}) {
		return loc
	}
	refCopy := loc
	return Location{
		Filename:  cfg.ReferenceLineno.Filename,
		Line:      cfg.ReferenceLineno.Line,
		Reference: &refCopy,
	}
}

// SetAlwaysCloneWithFreshID controls whether LabeledFormula.Clone()
// allocates fresh IDs. Python: set_always_clone_with_fresh_id() (ivy_ast.py:617-619).
func (cfg *AstConfig) SetAlwaysCloneWithFreshID(val bool) {
	cfg.AlwaysCloneWithFreshID = val
}

// NextLFID generates the next unique LabeledFormula ID.
func (cfg *AstConfig) NextLFID() int64 {
	return atomic.AddInt64(&cfg.LfCounter, 1) - 1
}
