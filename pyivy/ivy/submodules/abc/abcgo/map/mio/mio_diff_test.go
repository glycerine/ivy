package mio

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

// Compile the actual original expression evaluator with a minimal vector API;
// the oracle algorithm is extracted directly from exp.h, not reimplemented.
func TestOriginalCExpressionTruth(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	original := filepath.Join("..", "..", "..", "src", "map", "mio", "exp.h")
	data, err := os.ReadFile(original)
	if err != nil {
		t.Skip("original exp.h unavailable")
	}
	function := func(name string) string {
		t.Helper()
		text := string(data)
		start := strings.Index(text, "static inline word "+name+"(")
		if start < 0 {
			t.Fatal("missing original", name)
		}
		body := strings.Index(text[start:], "{") + start
		depth := 1
		end := body + 1
		for depth > 0 && end < len(text) {
			if text[end] == '{' {
				depth++
			}
			if text[end] == '}' {
				depth--
			}
			end++
		}
		return text[start:end]
	}
	source := `#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
typedef uint64_t word;
typedef struct {int size;int *data;} Vec_Int_t;
#define EXP_CONST0 -1
#define EXP_CONST1 -2
#define ABC_CONST(x) UINT64_C(x)
#define ABC_CALLOC(t,n) ((t*)calloc((n),sizeof(t)))
#define ABC_FREE(x) free(x)
static inline int Vec_IntEntry(Vec_Int_t*p,int i){return p->data[i];}
static inline int Vec_IntEntryLast(Vec_Int_t*p){return p->data[p->size-1];}
static inline int Exp_NodeNum(Vec_Int_t*p){return p->size/2;}
` + function("Exp_Truth6Lit") + "\n" + function("Exp_Truth6") + `
int main(void){int n,count;while(scanf("%d%d",&n,&count)==2){Vec_Int_t p={count,calloc(count,sizeof(int))};for(int i=0;i<count;i++)if(scanf("%d",&p.data[i])!=1)return 2;printf("%016llx\n",(unsigned long long)Exp_Truth6(n,&p,NULL));free(p.data);}return 0;}
`
	dir := t.TempDir()
	cfile, binary := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle")
	if err = os.WriteFile(cfile, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(cc, "-std=c99", "-O2", cfile, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile original oracle: %v\n%s", err, output)
	}
	rng := rand.New(rand.NewSource(7823))
	var input bytes.Buffer
	var expected []uint64
	names := []string{"a", "b", "c", "d", "e", "f"}
	var formula func(int) string
	formula = func(depth int) string {
		if depth == 0 {
			return names[rng.Intn(len(names))]
		}
		left, right := formula(depth-1), formula(depth-1)
		operator := []string{"*", "+", "^"}[rng.Intn(3)]
		value := "(" + left + operator + right + ")"
		if rng.Intn(2) == 0 {
			value = "!" + value
		}
		return value
	}
	for test := 0; test < 100; test++ {
		expression, err := ParseFormula(formula(3), names)
		if err != nil {
			t.Fatal(err)
		}
		a := vec.Vec_IntArray(expression)
		fmt.Fprintf(&input, "6 %d", len(a))
		for _, lit := range a {
			fmt.Fprintf(&input, " %d", lit)
		}
		input.WriteByte('\n')
		expected = append(expected, Exp_Truth6(6, expression, nil))
	}
	command := exec.Command(binary)
	command.Stdin = &input
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	reader := bytes.NewReader(output)
	for i, want := range expected {
		var got uint64
		if _, err = fmt.Fscanf(reader, "%x\n", &got); err != nil || got != want {
			t.Fatalf("original C expression case %d: %016x != %016x (%v)", i, got, want, err)
		}
	}
}
