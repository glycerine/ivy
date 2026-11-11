package startingpoint

import (
	"regexp"
	"strconv"
	"strings"
)

var cFormatVerb = regexp.MustCompile(`%([-+# 0]*)([0-9]*|\*)?(\.(?:[0-9]*|\*))?([hljztL]*)([diouxXfFeEgGaAcspn%])`)

func convertFormatLiteral(arg string) (string, bool) {
	value, err := strconv.Unquote(arg)
	if err != nil {
		return arg, true
	}
	safe := true
	value = cFormatVerb.ReplaceAllStringFunc(value, func(spec string) string {
		m := cFormatVerb.FindStringSubmatch(spec)
		verb := m[5]
		if verb == "n" || verb == "a" || verb == "A" || ((verb == "s" || verb == "c") && m[4] != "") {
			safe = false
			return spec
		}
		if verb == "i" || verb == "u" {
			verb = "d"
		}
		return "%" + m[1] + m[2] + m[3] + verb
	})
	return strconv.Quote(value), safe
}
func (p *libraryPass) convertFormatted(c libraryCall) (string, string, bool) {
	if c.name != "printf" && c.name != "fprintf" && c.name != "sprintf" {
		return "", "", false
	}
	a := append([]string(nil), c.args...)
	format := 0
	if c.name != "printf" {
		format = 1
	}
	if len(a) <= format {
		return "", "Formatting call has no format argument.", false
	}
	converted, safe := convertFormatLiteral(a[format])
	if !safe {
		return "", "Format uses %n, hex-float or wide-character semantics requiring manual adaptation.", false
	}
	a[format] = converted
	// Known NUL-terminated byte buffers used with %s should be explicit strings.
	for i := format + 1; i < len(a); i++ {
		if p.kind(c.rawArgs[i]) == "bytes" {
			if v, ok := p.text(c.rawArgs[i]); ok {
				a[i] = v
			}
		}
	}
	fm := p.imports.use("fmt")
	note := "C length modifiers and unsigned/integer verbs normalized for literal formats. Review dynamic formats, argument types and byte/string representation."
	if c.name == "sprintf" {
		if !c.standalone || !identifier.MatchString(c.rawArgs[0]) {
			return "", "sprintf count-result or complex destination requires structural repair.", false
		}
		value := fm + "Sprintf(" + strings.Join(a[1:], ", ") + ")"
		if buffer, ok := p.bytes(c.rawArgs[0]); ok {
			return iife("spDst []byte, spText string", "", "if len(spText) >= len(spDst) { panic(\"sprintf exceeds known buffer\") }\ncopy(spDst, spText); spDst[len(spText)] = 0", buffer, value), note, true
		}
		return a[0] + " = " + value, note + " Destination assumed to be a Go string.", true
	}
	name := "Printf"
	if c.name == "fprintf" {
		name = "Fprintf"
	}
	call := fm + name + "(" + strings.Join(a, ", ") + ")"
	return call, note + " Go count/error results are left for downstream repair.", true
}
