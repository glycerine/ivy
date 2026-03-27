package ivylogic

import (
	"strings"

	lg "github.com/glycerine/goivy/logic"
	iu "github.com/glycerine/goivy/ivyutils"
)

// Standard polymorphic type variables.
var (
	Alpha = &lg.TopSort{Name: "alpha"}
	Beta  = &lg.TopSort{Name: "beta"}
	Gamma = &lg.TopSort{Name: "gamma"}
)

// polymorphicSymbolsDef defines the built-in polymorphic symbols.
// Each entry is (name, list of sorts including range).
var polymorphicSymbolsDef = []struct {
	Name  string
	Sorts []lg.Sort
}{
	{"<", []lg.Sort{Alpha, Alpha, lg.Boolean}},
	{"<=", []lg.Sort{Alpha, Alpha, lg.Boolean}},
	{">", []lg.Sort{Alpha, Alpha, lg.Boolean}},
	{">=", []lg.Sort{Alpha, Alpha, lg.Boolean}},
	{"+", []lg.Sort{Alpha, Alpha, Alpha}},
	{"*", []lg.Sort{Alpha, Alpha, Alpha}},
	{"-", []lg.Sort{Alpha, Alpha, Alpha}},
	{"/", []lg.Sort{Alpha, Alpha, Alpha}},
	{"*>", []lg.Sort{Alpha, Beta, lg.Boolean}},
	{"bvand", []lg.Sort{Alpha, Alpha, Alpha}},
	{"bvor", []lg.Sort{Alpha, Alpha, Alpha}},
	{"bvnot", []lg.Sort{Alpha, Alpha}},
	// For liveness to safety reduction:
	{"l2s_waiting", []lg.Sort{lg.Boolean}},
	{"l2s_frozen", []lg.Sort{lg.Boolean}},
	{"l2s_saved", []lg.Sort{lg.Boolean}},
	{"l2s_d", []lg.Sort{Alpha, lg.Boolean}},
	{"l2s_a", []lg.Sort{Alpha, lg.Boolean}},
	{"cast", []lg.Sort{Alpha, Beta}},
	{"arrsel", []lg.Sort{Alpha, Beta, Gamma}},
	{"arrupd", []lg.Sort{Alpha, Beta, Gamma, Alpha}},
	{"arrcst", []lg.Sort{Alpha}},
}

// UninterpretedPolymorphicSymbols lists polymorphic symbols that should not
// be treated as interpreted even though they are polymorphic.
var UninterpretedPolymorphicSymbols = map[string]bool{
	"l2s_waiting": true,
	"l2s_frozen":  true,
	"l2s_saved":   true,
	"l2s_d":       true,
	"l2s_a":       true,
}

// polymorphicSymbols maps names to their Const definitions.
// Initialized at package load time — no init() needed.
var polymorphicSymbols = buildPolymorphicSymbols()

// buildPolymorphicSymbols creates the polymorphic symbols map from the
// definition table. Used by both init() and IvyLogicConfig.NewConfig().
func buildPolymorphicSymbols() map[string]*lg.Symbol {
	m := make(map[string]*lg.Symbol, len(polymorphicSymbolsDef))
	for _, def := range polymorphicSymbolsDef {
		var sort lg.Sort
		if len(def.Sorts) > 1 {
			sort, _ = lg.NewFunctionSort(def.Sorts...)
		} else {
			sort = def.Sorts[0]
		}
		m[def.Name] = lg.NewSymbol(def.Name, sort)
	}
	return m
}

// IvyLogicConfig holds per-session ivylogic state.
type IvyLogicConfig struct {
	IuCfg              *iu.IvyUtilsConfig // per-session version flags
	PolymorphicSymbols map[string]*lg.Symbol
	DefaultSort        lg.Sort
	AllowUnsorted      bool
	ReasonText         string
	Equals             *lg.Symbol
}

// NewIvyLogicConfigOn creates a new IvyLogicConfig with the given per-session config.
func NewIvyLogicConfigOn(iuCfg *iu.IvyUtilsConfig) *IvyLogicConfig {
	return &IvyLogicConfig{
		IuCfg:              iuCfg,
		PolymorphicSymbols: buildPolymorphicSymbols(),
		Equals:             lg.NewSymbol("=", RelationSort([]lg.Sort{lg.TopS, lg.TopS})),
	}
}

// NewIvyLogicConfig creates a new IvyLogicConfig with a fresh default config. For tests.
func NewIvyLogicConfig() *IvyLogicConfig {
	return NewIvyLogicConfigOn(iu.NewIvyUtilsConfig())
}

// FindPolymorphicSymbolOn looks up a polymorphic symbol by name on this config.
func (cfg *IvyLogicConfig) FindPolymorphicSymbolOn(name string) (*lg.Symbol, bool) {
	if cfg.IuCfg == nil || !cfg.IuCfg.HavePolymorphism {
		return nil, false
	}
	if c, ok := cfg.PolymorphicSymbols[name]; ok {
		return c, true
	}
	if strings.HasPrefix(name, "bfe[") {
		sort, _ := lg.NewFunctionSort(Alpha, Beta)
		c := lg.NewSymbol(name, sort)
		cfg.PolymorphicSymbols[name] = c
		return c, true
	}
	if len(name) > 0 && (name[0] >= '0' && name[0] <= '9' || name[0] == '"') {
		return lg.NewSymbol(name, Alpha), true
	}
	return nil, false
}

// FindPolymorphicSymbol looks up a polymorphic symbol by name.
// For "bfe[...]" symbols, creates them on demand.
// Returns false when IvyHavePolymorphism is disabled (language version <= 1.2).
// Matches Python ivy_logic.py find_polymorphic_symbol.
func FindPolymorphicSymbol(name string, iuCfg *iu.IvyUtilsConfig) (*lg.Symbol, bool) {
	if iuCfg == nil || !iuCfg.HavePolymorphism {
		return nil, false
	}
	if c, ok := polymorphicSymbols[name]; ok {
		return c, true
	}
	// Dynamic bfe[...] symbols
	if strings.HasPrefix(name, "bfe[") {
		sort, _ := lg.NewFunctionSort(Alpha, Beta)
		c := lg.NewSymbol(name, sort)
		polymorphicSymbols[name] = c
		return c, true
	}
	// Numerals and string literals get polymorphic sort alpha.
	// Matches Python ivy_logic.py:358-359.
	if len(name) > 0 && (name[0] >= '0' && name[0] <= '9' || name[0] == '"') {
		return lg.NewSymbol(name, Alpha), true
	}
	return nil, false
}

// PolymorphicMacrosMap maps comparison operators to their canonical forms.
var PolymorphicMacrosMap = map[string]string{
	"<=": "<",
	">":  "<",
	">=": "<",
}

// InfixSymbols is the set of symbols printed in infix notation.
var InfixSymbols = map[string]bool{
	"<": true, "<=": true, ">": true, ">=": true,
	"+": true, "-": true, "*": true, "/": true,
}

// PrecSymbols maps infix symbol names to their precedence for printing.
var PrecSymbols = map[string]int{
	"<": 7, "<=": 7, ">": 7, ">=": 7,
	"+": 12, "-": 13, "*": 14, "/": 15,
}

// SymbolIsPolymorphic returns true if the symbol name is in the polymorphic table.
func SymbolIsPolymorphic(name string) bool {
	_, ok := polymorphicSymbols[name]
	if ok {
		return true
	}
	return strings.HasPrefix(name, "bfe[")
}
