package ivylogic

import (
	"strings"

	lg "github.com/glycerine/goivy/logic"
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
var polymorphicSymbols map[string]*lg.Const

func init() {
	polymorphicSymbols = make(map[string]*lg.Const, len(polymorphicSymbolsDef))
	for _, def := range polymorphicSymbolsDef {
		var sort lg.Sort
		if len(def.Sorts) > 1 {
			sort, _ = lg.NewFunctionSort(def.Sorts...)
		} else {
			sort = def.Sorts[0]
		}
		polymorphicSymbols[def.Name] = lg.NewConst(def.Name, sort)
	}
}

// FindPolymorphicSymbol looks up a polymorphic symbol by name.
// For "bfe[...]" symbols, creates them on demand.
func FindPolymorphicSymbol(name string) (*lg.Const, bool) {
	if c, ok := polymorphicSymbols[name]; ok {
		return c, true
	}
	// Dynamic bfe[...] symbols
	if strings.HasPrefix(name, "bfe[") {
		sort, _ := lg.NewFunctionSort(Alpha, Beta)
		c := lg.NewConst(name, sort)
		polymorphicSymbols[name] = c
		return c, true
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
