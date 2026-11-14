// Signal range parsing from verCore.c.
package ver

import (
	"strconv"
	"strings"
)

// verAtoi follows the C parser's use of atoi: whitespace and a sign are
// accepted, and parsing stops at the first non-digit.
func verAtoi(s string) int32 {
	s = strings.TrimLeft(s, " \t\n\r\v\f")
	end := 0
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		end++
	}
	start := end
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == start {
		return 0
	}
	n, _ := strconv.ParseInt(s[:end], 10, 32)
	return int32(n)
}
func verRangeError(p *Ver_Man_t, message string) int32 {
	clear(p.sError[:])
	copy(p.sError[:], message)
	Ver_ParsePrintErrorMessage(p)
	return 0
}

func Ver_ParseSignalPrefix(p *Ver_Man_t, word *string, msb, lsb *int32) int32 {
	s := *word
	if len(s) == 0 || s[0] != '[' {
		panic("ver: range prefix must begin with '['")
	}
	high := verAtoi(s[1:])
	split := strings.IndexAny(s, ":]")
	if split < 0 {
		return verRangeError(p, "Cannot find closing bracket in this line.")
	}
	low := high
	if s[split] == ':' {
		low = verAtoi(s[split+1:])
		end := strings.IndexByte(s[split:], ']')
		if end < 0 {
			return verRangeError(p, "Cannot find closing bracket in this line.")
		}
		split += end + 1
		if split < len(s) && s[split] == '\\' {
			split++
			if end := strings.IndexByte(s[split:], ' '); end >= 0 {
				s = s[:split+end]
			}
		}
	}
	if high < 0 || low < 0 {
		panic("ver: negative signal range")
	}
	*word = s[split:]
	*msb, *lsb = high, low
	return 1
}

func Ver_ParseSignalSuffix(p *Ver_Man_t, word *string, msb, lsb *int32) int32 {
	s := *word
	if len(s) == 0 || s[len(s)-1] != ']' {
		panic("ver: range suffix must end with ']'")
	}
	split := strings.LastIndexAny(s[:len(s)-1], "[:")
	if split <= 0 {
		return verRangeError(p, "Cannot find opening bracket in signal name "+s+".")
	}
	if s[split] == '[' {
		*msb = verAtoi(s[split+1:])
		*lsb = *msb
		*word = s[:split]
		return 1
	}
	*lsb = verAtoi(s[split+1:])
	open := strings.LastIndexByte(s[:split], '[')
	if open <= 0 {
		return verRangeError(p, "Cannot find opening bracket in signal name "+s+".")
	}
	*msb = verAtoi(s[open+1:])
	*word = s[:open]
	return 1
}
