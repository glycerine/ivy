package cba

import "strconv"

func Cba_CharIsDigit(c byte) int32 {
	if c >= '0' && c <= '9' {
		return 1
	}
	return 0
}

// Cba_StrCmpInt compares from the first differing byte, treating digit runs as
// decimal integers exactly as the C comparator does. A byte past the end of
// the string represents the terminating C NUL.
func Cba_StrCmpInt(p1, p2 string, i int32) int32 {
	at := func(s string, j int32) byte {
		if j >= int32(len(s)) {
			return 0
		}
		return s[j]
	}
	digit := func(c byte) bool { return c >= '0' && c <= '9' }
	a, b := at(p1, i), at(p2, i)
	if digit(a) || digit(b) {
		if i > 0 && digit(at(p1, i-1)) || digit(a) && digit(b) {
			for i--; i >= 0 && digit(at(p1, i)); i-- {
			}
			i++
			if !digit(at(p1, i)) || !digit(at(p2, i)) || i > 0 && (digit(at(p1, i-1)) || digit(at(p2, i-1))) {
				panic("cba: invalid numeric comparison")
			}
			number := func(s string) int64 {
				end := i
				for digit(at(s, end)) {
					end++
				}
				n, err := strconv.ParseInt(s[i:end], 10, 32)
				if err != nil {
					panic("cba: numeric name outside C int range")
				}
				return n
			}
			n1, n2 := number(p1), number(p2)
			if n1 < n2 {
				return -1
			}
			if n1 > n2 {
				return 1
			}
			panic("cba: equal numeric name segments")
		}
	}
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	panic("cba: equal names")
}

func Cba_StrCmp(p1, p2 *string) int32 {
	a, b := *p1, *p2
	var i int32
	for int(i) < len(a) && int(i) < len(b) && a[i] != 0 && b[i] != 0 {
		if a[i] != b[i] {
			return Cba_StrCmpInt(a, b, i)
		}
		i++
	}
	return Cba_StrCmpInt(a, b, i)
}
