package bdc

import (
	"bytes"
	"regexp"
	"testing"
)

// Bdc_SpfdDecomposeTest prints "Total = %8d.  " then Abc_PrintTime( 1,
// "Time", ... ); the port printed "Total = 3. Time = 1.2µs".
func TestSPFDSubsetBenchmarkPrintsLikeABC(t *testing.T) {
	var out bytes.Buffer
	if err := SPFDSubsetBenchmark(&out, 100, 5, 1); err != nil {
		t.Fatal(err)
	}
	line := regexp.MustCompile(`^Total = [ 0-9]{8}\.  Time = +[0-9]+\.[0-9][0-9] sec\nTotal = [ 0-9]{8}\.  Time = +[0-9]+\.[0-9][0-9] sec\n$`)
	if !line.MatchString(out.String()) {
		t.Fatalf("got:\n%s", out.String())
	}
}
