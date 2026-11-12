package epd

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The C reference is only a test oracle; package epd has no C runtime dependency.
func TestOriginalCDecimalDifferential(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable for reference test")
	}
	root := "../../../src/bdd/epd"
	header, err := os.ReadFile(filepath.Join(root, "epd.h"))
	if err != nil {
		t.Fatal(err)
	}
	impl, err := os.ReadFile(filepath.Join(root, "epd.c"))
	if err != nil {
		t.Fatal(err)
	}
	var code strings.Builder
	code.WriteString(`#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <assert.h>
#define ABC_NAMESPACE_HEADER_START
#define ABC_NAMESPACE_HEADER_END
#define ABC_NAMESPACE_IMPL_START
#define ABC_NAMESPACE_IMPL_END
#define ABC_ALLOC(T,n) ((T*)malloc(sizeof(T)*(n)))
#define ABC_FREE(p) free(p)
`)
	code.Write(header)
	for _, line := range strings.Split(string(impl), "\n") {
		if !strings.HasPrefix(line, "#include") {
			code.WriteString(line + "\n")
		}
	}
	code.WriteString(`int main(int argc,char **argv) {
 EpDouble x; char text[100]; int n=atoi(argv[1]);
 EpdPow2(n,&x); EpdGetString(&x,text); printf("%s\n",text);
 EpdPow2Decimal(n,&x); printf("%.17g %d\n",x.type.value,x.exponent);
 return 0;
}
`)
	dir := t.TempDir()
	src := filepath.Join(dir, "reference.c")
	binary := filepath.Join(dir, "reference")
	if err := os.WriteFile(src, []byte(code.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-DNDEBUG", src, "-lm", "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile reference: %v: %s", err, out)
	}
	for _, n := range []int{-1022, -1000, -100, -1, 0, 1, 10, 100, 1023, 1024, 10000, 100000} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			output, err := exec.Command(binary, strconv.Itoa(n)).CombinedOutput()
			if err != nil {
				t.Fatal(err, string(output))
			}
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			var p EpDouble
			EpdPow2(n, &p)
			if got := EpdGetString(&p); got != lines[0] {
				t.Fatalf("format Go %s C %s", got, lines[0])
			}
			var value float64
			var exponent int
			if _, err := fmt.Sscan(lines[1], &value, &exponent); err != nil {
				t.Fatal(err)
			}
			EpdPow2Decimal(n, &p)
			if p.Exponent != exponent || math.Abs(p.Value-value) > math.Abs(value)*2e-14 {
				t.Fatalf("decimal Go %+v C %g e%d", p, value, exponent)
			}
		})
	}
}
