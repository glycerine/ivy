package startingpoint

import (
	"regexp"
	"strings"
)

var callbackParameter = regexp.MustCompile(`^\s*([A-Za-z_]\w*(?:::[A-Za-z_]\w*)*\s*\**)\s*\(\s*\*\s*([A-Za-z_]\w*)\s*\)\s*\(([\s\S]*)\)\s*$`)
var unnamedParameterType = regexp.MustCompile(`^\s*[A-Za-z_]\w*(?:::[A-Za-z_]\w*)*\s*\**\s*&?\s*$`)

func goParameter(arg string) (string, bool) {
	if m := namedType.FindStringSubmatch(arg); m != nil {
		return m[3] + " " + arrayType(goType(m[1]+m[2]), m[4], true), true
	}
	m := callbackParameter.FindStringSubmatch(arg)
	if m == nil {
		return "", false
	}
	var params []string
	if strings.TrimSpace(m[3]) != "" && strings.TrimSpace(m[3]) != "void" {
		args, _ := splitExpressions(m[3])
		for _, a := range args {
			if unnamedParameterType.MatchString(a) {
				params = append(params, goType(a))
				continue
			}
			// Parameter names are optional within the callback's signature.
			if named := namedType.FindStringSubmatch(a); named != nil {
				params = append(params, arrayType(goType(named[1]+named[2]), named[4], true))
				continue
			}
			return "", false
		}
	}
	result := goType(m[1])
	if result != "" {
		result = " " + result
	}
	return m[2] + " func(" + strings.Join(params, ", ") + ")" + result, true
}
