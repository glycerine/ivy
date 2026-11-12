package if_

import (
	"fmt"
	"strings"
)

// dsdExpression removes the NUL terminator used by the original C buffer.
func dsdExpression(text string) string {
	if nul := strings.IndexByte(text, 0); nul >= 0 {
		return text[:nul]
	}
	return text
}

func dsdDelimiterMatches(text string) ([]int32, error) {
	matches := make([]int32, len(text))
	stack := make([]int, 0, 8)
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '(', '[', '<', '{':
			stack = append(stack, i)
		case ')', ']', '>', '}':
			if len(stack) == 0 {
				return nil, fmt.Errorf("unmatched closing delimiter at byte %d", i)
			}
			open := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if !dsdDelimitersMatch(text[open], text[i]) {
				return nil, fmt.Errorf("mismatched delimiter at byte %d", i)
			}
			matches[open] = int32(i)
		}
	}
	if len(stack) != 0 {
		return nil, fmt.Errorf("unclosed delimiter at byte %d", stack[len(stack)-1])
	}
	return matches, nil
}

func dsdDelimitersMatch(open, close byte) bool {
	switch open {
	case '(':
		return close == ')'
	case '[':
		return close == ']'
	case '<':
		return close == '>'
	case '{':
		return close == '}'
	default:
		return false
	}
}

func parseDecExistNode(text string, pos *int, matches []int32, support *int32, andOnly bool) (uint32, error) {
	if *pos >= len(text) {
		return 0, fmt.Errorf("unexpected end of DSD expression")
	}
	if text[*pos] == '!' {
		*pos++
		if *pos >= len(text) {
			return 0, fmt.Errorf("negation has no operand")
		}
	}
	for *pos < len(text) && ((text[*pos] >= 'A' && text[*pos] <= 'F') || (text[*pos] >= '0' && text[*pos] <= '9')) {
		*pos++
	}
	if *pos >= len(text) {
		return 0, fmt.Errorf("DSD literal has no operand")
	}
	ch := text[*pos]
	if ch >= 'a' && ch <= 'z' {
		*pos++
		*support++
		return 0, nil
	}

	switch ch {
	case '(', '[', '<', '{':
	default:
		return 0, fmt.Errorf("unexpected DSD token %q at byte %d", ch, *pos)
	}
	open := *pos
	close := int(matches[open])
	if close <= open || close >= len(text) {
		return 0, fmt.Errorf("invalid DSD match at byte %d", open)
	}
	if andOnly && ch == '[' {
		// XOR is a legal non-AND node but does not create an AND boundset.
	}
	*pos++
	var mask uint32
	var childSupports [8]int32
	nParts, nSimple := 0, 0
	for *pos < close {
		if nParts == len(childSupports) {
			return 0, fmt.Errorf("DSD operator at byte %d has more than eight children", open)
		}
		childSupport := int32(0)
		childMask, err := parseDecExistNode(text, pos, matches, &childSupport, andOnly)
		if err != nil {
			return 0, err
		}
		if *pos <= open || *pos > close {
			return 0, fmt.Errorf("DSD child crossed its parent delimiter at byte %d", open)
		}
		mask |= childMask
		childSupports[nParts] = childSupport
		if childSupport == 1 {
			nSimple++
		}
		*support += childSupport
		nParts++
	}
	if *pos != close {
		return 0, fmt.Errorf("DSD children do not end at byte %d", close)
	}
	if nParts < 2 {
		return 0, fmt.Errorf("DSD operator at byte %d has fewer than two children", open)
	}
	*pos++ // consume the close delimiter

	if andOnly {
		if ch != '(' {
			return mask, nil
		}
		if nSimple == 0 {
			return mask, nil
		}
		for subset := 1; subset < 1<<nParts; subset++ {
			sum := int32(0)
			for i := 0; i < nParts; i++ {
				if childSupports[i] > 1 && subset&(1<<i) != 0 {
					sum += childSupports[i]
				}
			}
			if sum > 8 {
				return 0, fmt.Errorf("AND decomposition support sum %d exceeds eight", sum)
			}
			if sum >= 2 {
				for i := 0; i < nSimple; i++ {
					bit := sum + int32(i)
					if bit < 32 {
						mask |= 1 << uint(bit)
					}
				}
			}
		}
		for i := 2; i < nSimple; i++ {
			mask |= 1 << uint(i)
		}
		return mask, nil
	}

	if ch == '<' || ch == '{' {
		if *support < 32 {
			mask |= 1 << uint(*support)
		}
		return mask, nil
	}
	for subset := 1; subset < 1<<nParts; subset++ {
		sum := int32(0)
		for i := 0; i < nParts; i++ {
			if subset&(1<<i) != 0 {
				sum += childSupports[i]
			}
		}
		if sum > 8 {
			return 0, fmt.Errorf("decomposition support sum %d exceeds eight", sum)
		}
		if sum >= 2 {
			mask |= 1 << uint(sum)
		}
	}
	return mask, nil
}

func checkDsdDecExist(text string, andOnly bool) (int32, error) {
	text = dsdExpression(text)
	if len(text) <= 1 {
		if andOnly {
			return 1, nil
		}
		return 0, nil
	}
	matches, err := dsdDelimiterMatches(text)
	if err != nil {
		return 0, err
	}
	pos := 0
	support := int32(0)
	mask, err := parseDecExistNode(text, &pos, matches, &support, andOnly)
	if err != nil {
		return 0, err
	}
	if pos != len(text) {
		return 0, fmt.Errorf("unconsumed DSD text at byte %d", pos)
	}
	return int32(mask), nil
}

// Dau_DsdCheckDecExist reports the support sizes for which the DSD tree
// admits a nontrivial boundset decomposition.
func Dau_DsdCheckDecExist(text string) int32 {
	mask, err := checkDsdDecExist(text, false)
	if err != nil {
		panic("invalid DSD expression: " + err.Error())
	}
	return mask
}

// Dau_DsdCheckDecAndExist reports support sizes for which the DSD tree admits
// an AND decomposition.
func Dau_DsdCheckDecAndExist(text string) int32 {
	mask, err := checkDsdDecExist(text, true)
	if err != nil {
		panic("invalid DSD expression: " + err.Error())
	}
	return mask
}
