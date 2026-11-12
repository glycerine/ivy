package kit

import "strings"

func kitPlaString(s string) string {
	if end := strings.IndexByte(s, 0); end >= 0 {
		return s[:end]
	}
	return s
}
func Kit_PlaIsConst0(s string) int32 {
	if strings.HasPrefix(s, " 0") {
		return 1
	}
	return 0
}
func Kit_PlaIsConst1(s string) int32 {
	if strings.HasPrefix(s, " 1") {
		return 1
	}
	return 0
}
func Kit_PlaIsBuf(s string) int32 {
	s = kitPlaString(s)
	if s == "1 1\n" || s == "0 0\n" {
		return 1
	}
	return 0
}
func Kit_PlaIsInv(s string) int32 {
	s = kitPlaString(s)
	if s == "0 1\n" || s == "1 0\n" {
		return 1
	}
	return 0
}
func Kit_PlaGetVarNum(s string) int32 {
	end := strings.IndexByte(kitPlaString(s), '\n')
	if end < 2 {
		return -1
	}
	return int32(end - 2)
}
func Kit_PlaGetCubeNum(s string) int32 { return int32(strings.Count(kitPlaString(s), "\n")) }
func Kit_PlaIsComplement(s string) int32 {
	end := strings.IndexByte(s, '\n')
	if end <= 0 {
		panic("kit: malformed PLA row")
	}
	if s[end-1] == '0' || s[end-1] == 'n' {
		return 1
	}
	return 0
}

// Go strings are immutable; return the cover with its output polarity toggled.
func Kit_PlaComplement(s string) string {
	out := []byte(kitPlaString(s))
	for i, c := range out {
		if c != '\n' {
			continue
		}
		if i == 0 {
			panic("kit: malformed PLA row")
		}
		switch out[i-1] {
		case '0':
			out[i-1] = '1'
		case '1':
			out[i-1] = '0'
		case 'x':
			out[i-1] = 'n'
		case 'n':
			out[i-1] = 'x'
		default:
			panic("kit: invalid PLA output")
		}
	}
	return string(out)
}
func Kit_PlaStart(pool any, nCubes, nVars int32) string {
	if nCubes < 0 || nVars < 0 {
		panic("kit: negative PLA dimensions")
	}
	return strings.Repeat(strings.Repeat("-", int(nVars))+" 1\n", int(nCubes))
}
func Kit_PlaCreateFromIsop(pool any, nVars int32, cover []uint32) string {
	if nVars < 0 || nVars > 16 {
		panic("kit: unsupported PLA dimension")
	}
	var out strings.Builder
	for _, cube := range cover {
		for v := int32(0); v < nVars; v++ {
			switch cube >> uint(2*v) & 3 {
			case 0:
				out.WriteByte('-')
			case 1:
				out.WriteByte('0')
			case 2:
				out.WriteByte('1')
			default:
				panic("kit: contradictory ISOP cube")
			}
		}
		out.WriteString(" 1\n")
	}
	return out.String()
}
func Kit_PlaToIsop(s string, cover *[]uint32) {
	s = kitPlaString(s)
	nVars := Kit_PlaGetVarNum(s)
	if nVars < 0 || nVars > 16 {
		panic("kit: malformed PLA cover")
	}
	*cover = (*cover)[:0]
	width := int(nVars) + 3
	if len(s)%width != 0 {
		panic("kit: truncated PLA cover")
	}
	for start := 0; start < len(s); start += width {
		var cube uint32
		for v := int32(0); v < nVars; v++ {
			switch s[start+int(v)] {
			case '-':
			case '0':
				cube |= 1 << uint(2*v)
			case '1':
				cube |= 2 << uint(2*v)
			default:
				panic("kit: invalid PLA literal")
			}
		}
		if s[start+int(nVars)] != ' ' || s[start+width-1] != '\n' {
			panic("kit: malformed PLA delimiter")
		}
		*cover = append(*cover, cube)
	}
}
func Kit_PlaStoreSop(pool any, s string) string { return strings.Clone(kitPlaString(s)) }
func Kit_PlaFromTruth(pool any, truth []uint32, nVars int32, cover *[]uint32) string {
	if Kit_TruthIsConst0(truth, nVars) != 0 {
		return " 0\n"
	}
	if Kit_TruthIsConst1(truth, nVars) != 0 {
		return " 1\n"
	}
	phase := Kit_TruthIsop(truth, nVars, cover, 0)
	if phase < 0 {
		panic("kit: ISOP memory limit exceeded")
	}
	s := Kit_PlaCreateFromIsop(pool, nVars, *cover)
	if phase != 0 {
		s = Kit_PlaComplement(s)
	}
	return s
}
func Kit_PlaFromIsop(buffer *[]byte, nVars int32, cover []uint32) string {
	s := Kit_PlaCreateFromIsop(nil, nVars, cover)
	if buffer != nil {
		*buffer = append((*buffer)[:0], s...)
	}
	return s
}
func Kit_PlaFromTruthNew(truth []uint32, nVars int32, cover *[]uint32, buffer *[]byte) string {
	phase := Kit_TruthIsop(truth, nVars, cover, 1)
	if phase < 0 {
		panic("kit: ISOP memory limit exceeded")
	}
	s := ""
	if len(*cover) == 0 {
		s = " 0\n"
	} else if len(*cover) == 1 && (*cover)[0] == 0 {
		s = " 1\n"
	} else {
		s = Kit_PlaCreateFromIsop(nil, nVars, *cover)
	}
	if phase != 0 {
		s = Kit_PlaComplement(s)
	}
	if buffer != nil {
		*buffer = append((*buffer)[:0], s...)
	}
	return s
}
func Kit_PlaToTruth6(s string, nVars int32) uint64 {
	if nVars < 0 || nVars > 6 {
		panic("kit: single-word PLA supports at most six variables")
	}
	vars := make([][]uint32, nVars)
	n := max(nVars, 6)
	for i := range vars {
		vars[i] = make([]uint32, 2)
		Kit_TruthIthVar(vars[i], n, int32(i))
	}
	out, temp := make([]uint32, 2), make([]uint32, 2)
	Kit_PlaToTruth(s, nVars, vars, temp, out)
	if nVars < 6 {
		return uint64(out[0]) | uint64(out[0])<<32
	}
	return uint64(out[0]) | uint64(out[1])<<32
}
func Kit_PlaToTruth(s string, nVars int32, vars [][]uint32, temp, out []uint32) {
	s = kitPlaString(s)
	if nVars < 0 || nVars > 16 {
		panic("kit: unsupported PLA variable count")
	}
	if Kit_PlaIsConst0(s) != 0 {
		Kit_TruthClear(out, nVars)
		return
	}
	if Kit_PlaIsConst1(s) != 0 {
		Kit_TruthFill(out, nVars)
		return
	}
	width := int(nVars) + 3
	if len(s) == 0 || len(s)%width != 0 {
		panic("kit: malformed PLA representation")
	}
	if len(vars) < int(nVars) {
		panic("kit: missing elementary PLA truth tables")
	}
	Kit_TruthClear(out, nVars)
	phase := Kit_PlaIsComplement(s)
	for start := 0; start < len(s); start += width {
		if s[start+int(nVars)] != ' ' || s[start+width-1] != '\n' {
			panic("kit: malformed PLA row")
		}
		Kit_TruthFill(temp, nVars)
		for v := int32(0); v < nVars; v++ {
			switch s[start+int(v)] {
			case '1':
				Kit_TruthAnd(temp, temp, vars[v], nVars)
			case '0':
				Kit_TruthSharp(temp, temp, vars[v], nVars)
			case '-':
			default:
				panic("kit: invalid PLA literal")
			}
		}
		Kit_TruthOr(out, out, temp, nVars)
	}
	if phase != 0 {
		Kit_TruthNot(out, out, nVars)
	}
}
