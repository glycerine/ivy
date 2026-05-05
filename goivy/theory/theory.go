// Package theory defines Ivy's built-in theories (int, nat, bv[n]).
package theory

import (
	"fmt"
	"strconv"
	"strings"

	lg "github.com/glycerine/ivy/goivy/logic"
)

// TheoryKind identifies the kind of theory.
type TheoryKind int

const (
	IntegerKind TheoryKind = iota
	NaturalKind
	BitVectorKind
)

// Theory represents a built-in Ivy theory.
type Theory struct {
	Name   string
	Kind   TheoryKind
	Args   []int
	Finite bool
}

func (t *Theory) String() string { return t.Name }

// Bits returns the bit-width for a bit-vector theory.
// It panics if called on a non-bit-vector theory.
func (t *Theory) Bits() int {
	if t.Kind != BitVectorKind || len(t.Args) == 0 {
		panic("Bits() called on non-bit-vector theory")
	}
	return t.Args[0]
}

// NewIntegerTheory returns a Theory for the int interpretation.
func NewIntegerTheory() *Theory {
	return &Theory{Name: "int", Kind: IntegerKind}
}

// NewNaturalTheory returns a Theory for the nat interpretation.
func NewNaturalTheory() *Theory {
	return &Theory{Name: "nat", Kind: NaturalKind}
}

// NewBitVectorTheory returns a Theory for the bv[n] interpretation.
func NewBitVectorTheory(bits int) *Theory {
	return &Theory{
		Name:   fmt.Sprintf("bv[%d]", bits),
		Kind:   BitVectorKind,
		Args:   []int{bits},
		Finite: true,
	}
}

// numParams maps theory base names to expected parameter counts.
var numParams = map[string]int{
	"int": 0,
	"nat": 0,
	"bv":  1,
}

// ParseTheory parses a theory name such as "int", "nat", or "bv[32]"
// and returns the corresponding Theory.
func ParseTheory(name string) (*Theory, error) {
	parts := strings.Split(name, "[")
	base := parts[0]
	paramStrs := parts[1:]

	for _, p := range paramStrs {
		if !strings.HasSuffix(p, "]") {
			return nil, fmt.Errorf("bad theory syntax: %s", name)
		}
	}

	params := make([]int, len(paramStrs))
	for i, p := range paramStrs {
		v, err := strconv.Atoi(p[:len(p)-1])
		if err != nil {
			return nil, fmt.Errorf("bad theory syntax: %s", name)
		}
		params[i] = v
	}

	expected, ok := numParams[base]
	if !ok {
		return nil, fmt.Errorf("unknown theory: %s", name)
	}
	if len(params) != expected {
		return nil, fmt.Errorf("wrong number of theory parameters: %s", name)
	}

	switch base {
	case "int":
		return NewIntegerTheory(), nil
	case "nat":
		return NewNaturalTheory(), nil
	case "bv":
		return NewBitVectorTheory(params[0]), nil
	}
	// unreachable
	return nil, fmt.Errorf("unknown theory: %s", name)
}

// --- Theory schema strings ---

// TheorySchemas16 holds the Ivy-language schema definitions for version < 1.7.
var TheorySchemas16 = map[string]string{
	"int": `#lang ivy
    schema rec[t] = {
	type q
	function base(X:t) : q
	function step(X:q,Y:t) : q
	function fun(X:t) : q
	#---------------------------------------------------------
	definition fun(X:t) = base(X) if X <= 0 else step(fun(X-1),X)
    }

    schema ind[t] = {
        relation p(X:t)
        {
            individual x:t
            property x <= 0 -> p(x)
        }
        {
            individual x:t
            property p(x) -> p(x+1)
        }
        #--------------------------
        property p(X)
    }

    schema lep[t] = {
        function n : t
        function p(X:t) : bool
        #---------------------------------------------------------
        property exists L. (L >= n & forall B. (B >= n & p(B)-> p(L) & L <= B))
    }
`,
}

// TheorySchemas17 holds the Ivy-language schema definitions for version >= 1.7.
var TheorySchemas17 = map[string]string{
	"int": `#lang ivy
    schema rec[t] = {
	type q
	function base(X:t) : q
	function step(X:q,Y:t) : q
	function fun(X:t) : q
	#---------------------------------------------------------
	definition fun(X:t) = base(X) if X <= 0 else step(fun(X-1),X)
    }

    schema ind[t] = {
        relation p(X:t)
        theorem [base] {
            individual x:t
            property x <= 0 -> p(x)
        }
        theorem [step] {
            individual x:t
            property p(x) -> p(x+1)
        }
        #--------------------------
        property p(X)
    }

    schema lep[t] = {
        function n : t
        function p(X:t) : bool
        #---------------------------------------------------------
        property exists L. (L >= n & forall B. (B >= n & p(B)-> p(L) & L <= B))
    }
`,
}

// Theories returns the appropriate schema map for the given version string.
// If version >= "1.7", it returns TheorySchemas17; otherwise TheorySchemas16.
func Theories(version string) map[string]string {
	if TheoryVersionLE("1.7", version) {
		return TheorySchemas17
	}
	return TheorySchemas16
}

// GetTheorySchemata returns the schema string for a given theory name,
// or the empty string if no schema is defined.
// The sort parameter allows passing a *logic.RangeSort directly.
// The version string determines which schema set to use.
func GetTheorySchemata(name string, sort lg.Sort, version string) string {
	if !TheoryVersionLE("1.6", version) {
		return ""
	}
	schemas := Theories(version)

	// RangeSort maps to int schemas.
	if _, ok := sort.(*lg.RangeSort); ok {
		return schemas["int"]
	}
	if strings.HasPrefix(name, "bv[") || name == "nat" {
		return schemas["int"]
	}
	if s, ok := schemas[name]; ok {
		return s
	}
	return ""
}

// GetSortTheory returns the theory associated with a first-order sort.
// If the sort has an interpretation in interp, the interpretation is
// parsed as a theory. If the interpretation is already a *Theory, it
// is returned directly. Otherwise, the sort itself is returned.
func GetSortTheory(sort lg.Sort, interp map[string]interface{}) interface{} {
	name := lg.SortName(sort)
	if v, ok := interp[name]; ok {
		switch val := v.(type) {
		case string:
			thy, err := ParseTheory(val)
			if err != nil {
				return sort
			}
			return thy
		case *Theory:
			return val
		default:
			return v
		}
	}
	return sort
}

// HasIntegerInterp checks whether a sort has an integer-like interpretation
// (int, nat, or range).
func HasIntegerInterp(sort lg.Sort, interp map[string]interface{}) bool {
	name := lg.SortName(sort)
	v, ok := interp[name]
	if !ok {
		return false
	}
	switch val := v.(type) {
	case string:
		return val == "int" || val == "nat"
	case *lg.RangeSort:
		_ = val
		return true
	}
	return false
}

// VersionLE returns true if version a <= version b, comparing
// dot-separated numeric components left to right.
func TheoryVersionLE(a, b string) bool {
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	maxLen := len(pa)
	if len(pb) > maxLen {
		maxLen = len(pb)
	}
	for i := 0; i < maxLen; i++ {
		va, vb := 0, 0
		if i < len(pa) {
			va, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			vb, _ = strconv.Atoi(pb[i])
		}
		if va < vb {
			return true
		}
		if va > vb {
			return false
		}
	}
	return true // equal
}
