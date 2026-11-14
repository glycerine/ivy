package bac

import "strings"

// bacRangeWidth follows the source's bracket-and-equals range syntax.
func bacRangeWidth(text string) int32 {
	if len(text) == 0 || text[0] != '[' {
		panic("bac: range does not start with '['")
	}
	left := bacAtoi(text[1:])
	right := left
	if i := strings.IndexByte(text, '='); i >= 0 {
		right = bacAtoi(text[i+1:])
	}
	if left > right {
		return 1 + left - right
	}
	return 1 + right - left
}
