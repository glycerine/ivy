package bac

import "strconv"

// bacAtoi implements the decimal-prefix behavior used by C atoi.
func bacAtoi(s string) int32 {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r' || s[start] == '\n' || s[start] == '\v' || s[start] == '\f') {
		start++
	}
	end := start
	if end < len(s) && (s[end] == '+' || s[end] == '-') {
		end++
	}
	first := end
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == first {
		return 0
	}
	n, _ := strconv.ParseInt(s[start:end], 10, 32)
	return int32(n)
}

// bacReadHeaderLine follows BacManReadBacLine's 1000-byte header buffer.
// The returned position is the first byte after the newline.
func bacReadHeaderLine(data []byte, pos int) (string, int, bool) {
	if pos < 0 || pos >= len(data) {
		return "", pos, false
	}
	start := pos
	for pos < len(data) && data[pos] != '\n' && pos-start < 1000 {
		pos++
	}
	if pos >= len(data) || data[pos] != '\n' || pos-start >= 1000 {
		return "", pos, false
	}
	return string(data[start:pos]), pos + 1, true
}

// bacParseHeaderNumbers follows BacManReadBacNameAndNums: fields are separated
// by literal spaces, at least two numeric positions must be present, and
// omitted third and fourth values remain -1. Like atoi, a numeric field may
// begin with whitespace or a sign and is zero when no digits are present.
func bacParseHeaderNumbers(line string) (string, [4]int32, bool) {
	nums := [4]int32{-1, -1, -1, -1}
	i := 0
	for i < len(line) && line[i] != ' ' {
		i++
	}
	if i == len(line) {
		return "", nums, false
	}
	name := line[:i]
	for field := 0; field < 4; field++ {
		i++
		nums[field] = bacAtoi(line[i:])
		if field == 3 {
			return name, nums, true
		}
		for i < len(line) && line[i] != ' ' {
			i++
		}
		if i == len(line) {
			return name, nums, field != 0
		}
	}
	return name, nums, true
}
