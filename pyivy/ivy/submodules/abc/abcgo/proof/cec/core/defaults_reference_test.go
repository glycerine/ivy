package core

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestDefaultsMatchOriginalCECSource(t *testing.T) {
	data, err := os.ReadFile("../../../../src/proof/cec/cecCore.c")
	if os.IsNotExist(err) {
		t.Skip("original C source unavailable for reference comparison")
	}
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	read := func(function, field string) int {
		t.Helper()
		start := strings.Index(source, "void "+function+"(")
		if start < 0 {
			t.Fatal("source function missing", function)
		}
		end := strings.Index(source[start:], "\n}")
		if end < 0 {
			t.Fatal("source body missing")
		}
		pattern := regexp.MustCompile(`p->` + field + `\s*=\s*(-?[0-9]+)\s*;`)
		match := pattern.FindStringSubmatch(source[start : start+end])
		if match == nil {
			t.Fatal("source default missing", field)
		}
		number, err := strconv.Atoi(match[1])
		if err != nil {
			t.Fatal(err)
		}
		return number
	}
	sat, cec := DefaultSATOptions(), DefaultCECOptions()
	if sat.ConflictLimit != int64(read("Cec_ManSatSetDefaultParams", "nBTLimit")) || sat.MaxSATVariables != read("Cec_ManSatSetDefaultParams", "nSatVarMax") || sat.RecycleQueries != read("Cec_ManSatSetDefaultParams", "nCallsRecycle") || sat.PolarFlip != (read("Cec_ManSatSetDefaultParams", "fPolarFlip") != 0) || cec.ConflictLimit != int64(read("Cec_ManCecSetDefaultParams", "nBTLimit")) {
		t.Fatal("native options diverged from original defaults", sat, cec)
	}
}
