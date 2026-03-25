package lalr_full

import "github.com/glycerine/goivy/ast"

// ParserConfig holds per-parse state that was previously stored in
// package-level globals. Each concurrent parse gets its own ParserConfig.
type ParserConfig struct {
	// LabelCounter generates unique label/mixer names within a parse.
	// Python: label_counter (ivy_parser.py)
	LabelCounter int

	// CheckUnprovable matches Python's check_unprovable thread-local from ivy_actions.py.
	// When false (default), unprovable declarations are silently dropped.
	// When true, they are declared normally.
	CheckUnprovable bool

	// AstCfg is the ast config for this parse session.
	AstCfg *ast.AstConfig
}

// NewParserConfig creates a fresh ParserConfig for a new parse session.
// The AstConfig must be provided — Python's ast globals (lf_counter,
// always_clone_with_fresh_id) are shared across all parses in a session.
func NewParserConfig(astCfg *ast.AstConfig) *ParserConfig {
	return &ParserConfig{
		AstCfg: astCfg,
	}
}
