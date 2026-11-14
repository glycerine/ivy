package acb

// acbLexBuffer follows Acb_VerilogSimpleLex's strtok traversal on a mutable
// byte buffer. intern must use the source's preassigned keyword IDs.
func acbLexBuffer(source []byte, intern func(string) int32) []int32 {
	buffer := append([]byte(nil), source...)
	limit := len(buffer)
	for i, c := range buffer {
		if c == 0 {
			limit = i
			break
		}
	}
	acbRemoveComments(buffer[:limit])
	delimiter := func(c byte) bool {
		switch c {
		case ' ', '\n', '\r', '\t', '(', ')', ',', ';', '=':
			return true
		}
		return false
	}
	var out []int32
	cursor := 0
	for cursor < limit {
		for cursor < limit && delimiter(buffer[cursor]) {
			cursor++
		}
		if cursor >= limit || buffer[cursor] == 0 {
			break
		}
		start := cursor
		for cursor < limit && buffer[cursor] != 0 && !delimiter(buffer[cursor]) {
			cursor++
		}
		token := string(buffer[start:cursor])
		if cursor < limit && buffer[cursor] != 0 {
			buffer[cursor] = 0
			cursor++
		}
		id := intern(token)
		if token == "assign" {
			out = append(out, 6)
		} else {
			out = append(out, id)
		}
		if id >= 6 && id < 16 {
			// strtok overwrote the first delimiter; C scans past that NUL to
			// locate the '(' on the same line, skipping a gate instance name.
			for pos := start; pos < limit && buffer[pos] != '\n'; pos++ {
				if buffer[pos] == '(' {
					cursor = pos
					break
				}
			}
		}
	}
	return out
}

// acbRemoveComments blanks C++-style line comments in place, retaining newlines.
func acbRemoveComments(buffer []byte) {
	for i := 0; i+1 < len(buffer) && buffer[i] != 0; i++ {
		if buffer[i] == '/' && buffer[i+1] == '/' {
			for i < len(buffer) && buffer[i] != 0 && buffer[i] != '\n' {
				buffer[i] = ' '
				i++
			}
		}
	}
}
