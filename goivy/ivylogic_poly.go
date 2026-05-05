package goivy

import (
	"strings"
)

// Standard polymorphic type variables.
var (
	Alpha = &TopSort{Name: "alpha"}
	Beta  = &TopSort{Name: "beta"}
	Gamma = &TopSort{Name: "gamma"}
)

// polymorphicSymbolsDef defines the built-in polymorphic symbols.
// Each entry is (name, list of sorts including range).
var polymorphicSymbolsDef = []struct {
	Name  string
	Sorts []Sort
}{
	{"<", []Sort{Alpha, Alpha, Boolean}},
	{"<=", []Sort{Alpha, Alpha, Boolean}},
	{">", []Sort{Alpha, Alpha, Boolean}},
	{">=", []Sort{Alpha, Alpha, Boolean}},
	{"+", []Sort{Alpha, Alpha, Alpha}},
	{"*", []Sort{Alpha, Alpha, Alpha}},
	{"-", []Sort{Alpha, Alpha, Alpha}},
	{"/", []Sort{Alpha, Alpha, Alpha}},
	{"*>", []Sort{Alpha, Beta, Boolean}},
	{"bvand", []Sort{Alpha, Alpha, Alpha}},
	{"bvor", []Sort{Alpha, Alpha, Alpha}},
	{"bvnot", []Sort{Alpha, Alpha}},
	// For liveness to safety reduction:
	{"l2s_waiting", []Sort{Boolean}},
	{"l2s_frozen", []Sort{Boolean}},
	{"l2s_saved", []Sort{Boolean}},
	{"l2s_d", []Sort{Alpha, Boolean}},
	{"l2s_a", []Sort{Alpha, Boolean}},
	{"cast", []Sort{Alpha, Beta}},
	{"arrsel", []Sort{Alpha, Beta, Gamma}},
	{"arrupd", []Sort{Alpha, Beta, Gamma, Alpha}},
	{"arrcst", []Sort{Alpha}},
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
func buildPolymorphicSymbols() map[string]*Const {
	m := make(map[string]*Const, len(polymorphicSymbolsDef))
	for _, def := range polymorphicSymbolsDef {
		var sort Sort
		if len(def.Sorts) > 1 {
			sort, _ = NewFunctionSort(def.Sorts...)
		} else {
			sort = def.Sorts[0]
		}
		m[def.Name] = NewConst(def.Name, sort)
	}
	return m
}

// IvyLogicConfig holds per-session ivylogic state.
type IvyLogicConfig struct {
	IuCfg              *IvyUtilsConfig // per-session version flags
	PolymorphicSymbols map[string]*Const
	DefaultSort        Sort
	AllowUnsorted      bool
	ReasonText         string
	Equals             *Const
}

// NewIvyLogicConfigOn creates a new IvyLogicConfig with the given per-session config.
func NewIvyLogicConfigOn(iuCfg *IvyUtilsConfig) *IvyLogicConfig {
	return &IvyLogicConfig{
		IuCfg:              iuCfg,
		PolymorphicSymbols: buildPolymorphicSymbols(),
		Equals:             NewConst("=", RelationSort([]Sort{TopS, TopS})),
	}
}

// NewIvyLogicConfig creates a new IvyLogicConfig with a fresh default config. For tests.
func NewIvyLogicConfig() *IvyLogicConfig {
	return NewIvyLogicConfigOn(NewIvyUtilsConfig())
}

// FindPolymorphicSymbolOn looks up a polymorphic symbol by name on this config.
func (cfg *IvyLogicConfig) FindPolymorphicSymbolOn(name string) (*Const, bool) {
	if cfg.IuCfg == nil || !cfg.IuCfg.HavePolymorphism {
		return nil, false
	}
	if c, ok := cfg.PolymorphicSymbols[name]; ok {
		return c, true
	}
	if strings.HasPrefix(name, "bfe[") {
		sort, _ := NewFunctionSort(Alpha, Beta)
		c := NewConst(name, sort)
		cfg.PolymorphicSymbols[name] = c
		return c, true
	}
	if len(name) > 0 && (name[0] >= '0' && name[0] <= '9' || name[0] == '"') {
		return NewConst(name, Alpha), true
	}
	return nil, false
}

// FindPolymorphicSymbol looks up a polymorphic symbol by name.
// For "bfe[...]" symbols, creates them on demand (not cached in global map).
// Returns false when IvyHavePolymorphism is disabled (language version <= 1.2).
// Matches Python ivy_logic.py find_polymorphic_symbol.
func FindPolymorphicSymbol(name string, iuCfg *IvyUtilsConfig) (*Const, bool) {
	if iuCfg == nil || !iuCfg.HavePolymorphism {
		return nil, false
	}
	if c, ok := polymorphicSymbols[name]; ok {
		return c, true
	}
	// Dynamic bfe[...] symbols — create fresh (not cached in global).
	// Per-session caching is done via IvyLogicConfig.FindPolymorphicSymbolOn.
	if strings.HasPrefix(name, "bfe[") {
		sort, _ := NewFunctionSort(Alpha, Beta)
		c := NewConst(name, sort)
		return c, true
	}
	// Numerals and string literals get polymorphic sort alpha.
	// Matches Python ivy_logic.py:358-359.
	if len(name) > 0 && (name[0] >= '0' && name[0] <= '9' || name[0] == '"') {
		return NewConst(name, Alpha), true
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
