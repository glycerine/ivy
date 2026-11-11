package startingpoint

import (
	"regexp"
	"strings"
)

var arrayDimension = regexp.MustCompile(`\[([^\[\]]*)\]`)

// Array parameters decay in C; a Go slice represents the outer dimension.
// Fixed object arrays retain their bounds and dimension order.
func arrayType(element, suffix string, parameter bool) string {
	var prefix strings.Builder
	for i, dim := range arrayDimension.FindAllStringSubmatch(suffix, -1) {
		bound := strings.TrimSpace(dim[1])
		if parameter && i == 0 {
			bound = ""
		}
		prefix.WriteString("[" + bound + "]")
	}
	return prefix.String() + element
}
