package ivyutils

import (
	"strconv"
	"strings"
)

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
	return &LocationTuple{Filename: filename, Line: line}
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

// LinenoStr extracts location string from an object that has a GetLineno-like method.
// Corresponds to Python's lineno_str(ast).
// Accepts anything with a Lineno() or GetLineno() method returning a location.
func LinenoStr(node interface{}) string {
	// Try *LocationTuple directly
	if lt, ok := node.(*LocationTuple); ok {
		r := lt.String()
		r = strings.TrimSuffix(r, ": ")
		return r
	}
	// Try interface with GetLineno returning LocationTuple
	type hasLinenoLT interface {
		GetLinenoLT() *LocationTuple
	}
	if n, ok := node.(hasLinenoLT); ok {
		lt := n.GetLinenoLT()
		if lt == nil {
			return ""
		}
		r := lt.String()
		r = strings.TrimSuffix(r, ": ")
		return r
	}
	return ""
}
