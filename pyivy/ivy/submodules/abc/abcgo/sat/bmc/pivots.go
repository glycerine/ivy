package bmc

import (
	"strconv"
	"strings"
)

// parsePivotIDs matches fscanf("%d") over a whitespace-separated pivot file:
// parsing stops at the first token that is not a base-10 integer.
func parsePivotIDs(data []byte) []int32 {
	fields := strings.Fields(string(data))
	ids := make([]int32, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.ParseInt(field, 10, 32)
		if err != nil {
			break
		}
		ids = append(ids, int32(value))
	}
	return ids
}
