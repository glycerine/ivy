package startingpoint

import "strings"

func (p *libraryPass) convertMutex(c libraryCall) (string, string, bool) {
	method := ""
	switch c.name {
	case "pthread_mutex_lock":
		method = "Lock"
	case "pthread_mutex_unlock":
		method = "Unlock"
	default:
		return "", "", false
	}
	if len(c.args) != 1 {
		return "", arity(c, 1), false
	}
	receiver := unwrapMutexOperand(c.args[0])
	if strings.HasPrefix(receiver, "&") {
		receiver = unwrapMutexOperand(strings.TrimSpace(receiver[1:]))
	}
	if !identifier.MatchString(receiver) {
		receiver = "(" + receiver + ")"
	}
	return receiver + "." + method + "()", "Go mutex method returns no status; repair C integer-result uses downstream. No closure is generated.", true
}

func unwrapMutexOperand(s string) string {
	s = strings.TrimSpace(s)
	for strings.HasPrefix(s, "(") {
		all, err := Lex(s)
		if err != nil {
			break
		}
		var ts []Token
		for _, t := range all {
			if t.Kind != "space" && t.Kind != "comment" {
				ts = append(ts, t)
			}
		}
		pairs, err := MatchDelimiters(ts)
		if err != nil || len(ts) < 2 || pairs[0] != len(ts)-1 {
			break
		}
		s = strings.TrimSpace(s[ts[0].End:ts[len(ts)-1].Start])
	}
	return s
}
