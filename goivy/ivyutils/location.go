package ivyutils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// bad: duplicate of ast.Location, which parser produces.
// we should scap this and replace it with that.

// LocationTuple represents a source location with an optional reference chain.
// Corresponds to Python's LocationTuple(tuple) which extends tuple with
// filename, line, and optional reference properties.
type LocationTuple struct {
	Filename  string
	Line      int
	Reference *LocationTuple
}

// Location creates a new LocationTuple with the given filename and line.
// Corresponds to Python's Location(filename=None, line=None).
func Location(filename string, line int) *LocationTuple {
	return &LocationTuple{
		//Filename: filename,
		Filename: xtracer.NormalizeLine(filename),
		Line:     line,
	}
}

// Nowhere returns a sentinel LocationTuple representing "no location".
// Corresponds to Python's nowhere().
func Nowhere() *LocationTuple {
	return Location("nowhere", 0)
}

// IsNowhere returns true if the location is the "nowhere" sentinel.
// Corresponds to Python's is_nowhere(lineno).
func IsNowhere(loc *LocationTuple) bool {
	return loc != nil && loc.Filename == "nowhere"
}

// String formats the location for error messages.
// Matches Python's LocationTuple.__str__ (non-Windows path).
func (lt *LocationTuple) String() string {
	if lt == nil {
		return ""
	}
	if lt.Reference != nil {
		return lt.Reference.String()
	}
	res := ""
	if lt.Filename != "" {
		res += lt.Filename + ": "
	}
	if lt.Line > 0 {
		res += "line " + strconv.Itoa(lt.Line) + ": "
	}
	return res
}

// LinenoStr extracts location string from an object.
// Corresponds to Python's lineno_str(ast):
//
//	if not hasattr(ast,'lineno'): return ''
//	r = str(ast.lineno)
//	if r.endswith(': '): r = r[:-2]
//	return r
//
// Accepts *LocationTuple directly, or any object implementing Locatable
// (has GetLinenoLT() *LocationTuple method).
func LinenoStr(node interface{}) string {
	if node == nil {
		return ""
	}
	// Direct LocationTuple
	if lt, ok := node.(*LocationTuple); ok {
		r := lt.String()
		return strings.TrimSuffix(r, ": ")
	}
	// Object with GetLinenoLT() — matches Locatable interface from error.go
	type hasLinenoLT interface {
		GetLinenoLT() *LocationTuple
	}
	if n, ok := node.(hasLinenoLT); ok {
		lt := n.GetLinenoLT()
		if lt == nil {
			return ""
		}
		r := lt.String()
		return strings.TrimSuffix(r, ": ")
	}
	// Object with a Stringer .lineno — generic fallback
	type hasLineno interface {
		Lineno() fmt.Stringer
	}
	if n, ok := node.(hasLineno); ok {
		s := n.Lineno()
		if s == nil {
			return ""
		}
		r := s.String()
		return strings.TrimSuffix(r, ": ")
	}
	return ""
}
