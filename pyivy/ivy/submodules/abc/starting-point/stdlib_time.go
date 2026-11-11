package startingpoint

func (p *libraryPass) convertTime(c libraryCall) (string, string, bool) {
	switch c.name {
	case "time":
		if len(c.args) != 1 {
			return "", arity(c, 1), false
		}
		now := p.imports.use("time") + "Now().Unix()"
		switch c.args[0] {
		case "nil", "NULL", "0":
			return now, "Wall-clock Unix seconds; time_t uses int64 in this draft.", true
		}
		return iife("spResult *int64", "int64", "spNow := "+now+"\nif spResult != nil { *spResult = spNow }; return spNow", c.args[0]), "Unix seconds returned and written through the optional pointer once; time_t uses int64.", true
	case "clock":
		if len(c.args) != 0 {
			return "", arity(c, 0), false
		}
		if p.shadowed["CLOCKS_PER_SEC"] {
			return "", "Source overrides CLOCKS_PER_SEC; adapt that unit policy before replacing clock.", false
		}
		return p.imports.use("time") + "Now().UnixNano()", "Draft policy changes CPU time to wall-clock nanoseconds. CLOCKS_PER_SEC becomes time.Second; clock adjustments can affect differences. Use time.Now/time.Since for monotonic elapsed timing when repairing the surrounding code.", true
	case "difftime":
		if len(c.args) != 2 {
			return "", arity(c, 2), false
		}
		// Avoid time.Duration's ~290-year range and signed subtraction overflow.
		body := "if spEnd >= spStart { return float64(uint64(spEnd) - uint64(spStart)) }\nreturn -float64(uint64(spStart) - uint64(spEnd))"
		return iife("spEnd, spStart int64", "float64", body, "int64("+c.args[0]+")", "int64("+c.args[1]+")"), "Difference of int64 Unix seconds, with unsigned magnitude subtraction to avoid signed overflow and Duration saturation.", true
	}
	return "", "", false
}
