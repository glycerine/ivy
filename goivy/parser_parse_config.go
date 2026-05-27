package goivy

// parseConfig is the grammar-neutral configuration for full-file Ivy parses.
// Version-specific parser adapters consume it instead of owning parse options.
type parseConfig struct {
	importer    ImporterFunc
	included    map[string]bool
	parentAccum *ivyAccum
	nested      bool
	filename    string
	astCfg      *AstConfig
}

func newParseConfig() *parseConfig {
	return &parseConfig{
		astCfg: NewAstConfig(),
	}
}

// ParseOption configures optional behavior for the full-file LALR parser.
type ParseOption func(*parseConfig)

// WithImporter sets the include-resolution callback.
func WithImporter(fn ImporterFunc) ParseOption {
	return func(cfg *parseConfig) {
		cfg.importer = fn
	}
}

// WithIncluded sets the already-included module set for nested parses.
func WithIncluded(inc map[string]bool) ParseOption {
	return func(cfg *parseConfig) {
		cfg.included = inc
	}
}

// WithParentAccum links a nested parser to the including parser's accumulator.
// Python keeps these accumulators on a single global stack while importing.
func WithParentAccum(parent *ivyAccum) ParseOption {
	return func(cfg *parseConfig) {
		cfg.parentAccum = parent
	}
}

// WithNested marks this as a nested include parse.
// Nested parses skip expand_autoinstances, matching Python behavior.
func WithNested() ParseOption {
	return func(cfg *parseConfig) {
		cfg.nested = true
	}
}

// WithFilename sets the source filename, matching Python's iu.filename.
// Used by location helpers to produce source locations for canon matching.
func WithFilename(name string) ParseOption {
	return func(cfg *parseConfig) {
		cfg.filename = name
	}
}

// WithAstConfig shares an existing AstConfig with this parse.
// Python uses module-level globals shared across all parses. This option keeps
// counters and flags in sync across nested/imported parses without globals.
func WithAstConfig(astCfg *AstConfig) ParseOption {
	return func(cfg *parseConfig) {
		cfg.astCfg = astCfg
	}
}

func (cfg *parseConfig) applyToParser17(lex *parser17LexAdapter) {
	lex.importer = cfg.importer
	if cfg.included != nil {
		lex.included = cfg.included
	}
	lex.accum = cfg.parentAccum
	lex.nested = cfg.nested
	lex.filename = cfg.filename
	if cfg.astCfg != nil {
		lex.astCfg = cfg.astCfg
	} else {
		lex.astCfg = NewAstConfig()
	}
}

func (cfg *parseConfig) applyToParser16(lex *parser16LexAdapter) {
	lex.importer = cfg.importer
	if cfg.included != nil {
		lex.included = cfg.included
	}
	lex.accum = cfg.parentAccum
	lex.nested = cfg.nested
	lex.filename = cfg.filename
	if cfg.astCfg != nil {
		lex.astCfg = cfg.astCfg
	} else {
		lex.astCfg = NewAstConfig()
	}
}
