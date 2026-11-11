package startingpoint

func (p *libraryPass) convertText(c libraryCall) (string, string, bool) {
	a := c.args
	switch c.name {
	case "strlen":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		if buffer, ok := p.bytes(c.rawArgs[0]); ok {
			body := "if spEnd := " + p.imports.use("bytes") + "IndexByte(spBytes, 0); spEnd >= 0 { return spEnd }; return len(spBytes)"
			return iife("spBytes []byte", "int", body, buffer), "Byte-buffer length ends at first NUL, not array capacity.", true
		}
		value := a[0]
		if v, ok := p.text(value); ok {
			value = v
		}
		return "len(" + value + ")", "Byte length; known byte arrays/slices are truncated at NUL. String defaults assume no embedded terminator.", true
	case "strcmp", "strncmp":
		n := 2
		if c.name == "strncmp" {
			n = 3
		}
		if len(a) != n {
			return "", arity(c, n), false
		}
		left, right := a[0], a[1]
		if v, ok := p.text(left); ok {
			left = v
		}
		if v, ok := p.text(right); ok {
			right = v
		}
		st := p.imports.use("strings")
		if n == 2 {
			return "int32(" + st + "Compare(" + left + ", " + right + "))", "Comparison sign retained; strings use the selected C-string representation.", true
		}
		body := "if spN < 0 { panic(\"negative C string bound\") }\nif len(spA) > spN { spA = spA[:spN] }; if len(spB) > spN { spB = spB[:spN] }\nreturn int32(" + st + "Compare(spA, spB))"
		return iife("spA, spB string, spN int", "int32", body, left, right, "int("+a[2]+")"), "Byte-bounded comparison; no Unicode or rune-count substitution.", true
	case "strcpy", "strcat":
		if len(a) != 2 {
			return "", arity(c, 2), false
		}
		if !c.standalone || p.kind(c.rawArgs[0]) != "string" || !identifier.MatchString(c.rawArgs[0]) {
			return "", "String copy/append needs a standalone call and a known string destination; buffer aliasing/capacity and pointer-result uses remain for review.", false
		}
		text := a[1]
		if v, ok := p.text(text); ok {
			text = v
		}
		op := " = "
		if c.name == "strcat" {
			op = " += "
		}
		return a[0] + op + text, "String-value representation replaces mutable C storage; verify aliasing, capacity and NULL assumptions.", true
	case "strdup":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		text, ok := p.text(a[0])
		if !ok {
			return "", "strdup requires a known string/byte buffer representation.", false
		}
		return "(" + text + ")", "Go strings are values; POSIX allocation failure and pointer identity are not retained.", true
	case "memcpy", "memmove", "memset", "memcmp":
		if len(a) != 3 {
			return "", arity(c, 3), false
		}
		left, ok := p.bytes(c.rawArgs[0])
		if !ok {
			return "", "Raw memory operation retained: destination must be a known byte array/slice.", false
		}
		if c.name == "memset" {
			if !c.standalone {
				return "", "memset pointer result needs representation repair.", false
			}
			return iife("spDst []byte, spValue int32, spN int", "", "for spI := range spDst[:spN] { spDst[spI] = byte(spValue) }", left, "int32("+a[1]+")", "int("+a[2]+")"), "Explicit byte fill; byte count and bounds retained.", true
		}
		right, ok := p.bytes(c.rawArgs[1])
		if !ok {
			return "", "Raw memory operation retained: source must be a known byte array/slice.", false
		}
		if c.name == "memcmp" {
			return iife("spA, spB []byte, spN int", "int32", "return int32("+p.imports.use("bytes")+"Compare(spA[:spN], spB[:spN]))", left, right, "int("+a[2]+")"), "Byte comparison sign retained; no struct-layout or padding guesses.", true
		}
		if !c.standalone {
			return "", "Memory-copy pointer result needs representation repair.", false
		}
		return iife("spDst, spSrc []byte, spN int", "", "copy(spDst[:spN], spSrc[:spN])", left, right, "int("+a[2]+")"), "Go copy handles overlap; byte extents are preserved, not guessed from element counts.", true
	}
	return "", "", false
}

// Explicitly keep pointer-returning searches out of the direct-call map: an
// Index result (-1 or byte offset) is not a C pointer (NULL or address).
