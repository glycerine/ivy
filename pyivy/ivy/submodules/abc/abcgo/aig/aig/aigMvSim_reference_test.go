package aig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Compile the actual original AND/hash/create routines, rather than a second Go
// implementation of the same rewrite. The Go port itself has no C dependency.
func TestMultivalueAndMatchesOriginalC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable for original-source reference test")
	}
	original, err := os.ReadFile("../../../src/aig/saig/saigSimMv.c")
	if os.IsNotExist(err) {
		t.Skip("original C source unavailable for reference comparison")
	}
	if err != nil {
		t.Fatal(err)
	}
	extract := func(name string) string {
		t.Helper()
		text := string(original)
		at := strings.Index(text, name+"(")
		if at < 0 {
			t.Fatal("reference routine missing", name)
		}
		line := strings.LastIndex(text[:at], "\n") + 1
		brace := strings.Index(text[at:], "{") + at
		if brace < at {
			t.Fatal("reference body missing")
		}
		depth := 0
		for i := brace; i < len(text); i++ {
			switch text[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return text[line:i+1] + "\n"
				}
			}
		}
		t.Fatal("reference braces missing")
		return ""
	}
	header := `#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#define ABC_REALLOC(type,p,n) ((type *)realloc((p),sizeof(type)*(n)))
#define Abc_MaxInt(a,b) ((a)>(b)?(a):(b))
typedef struct { int iFan0,iFan1,iNext; } Saig_MvAnd_t;
typedef struct { Saig_MvAnd_t *pAigNew; unsigned char *pLevels; int nObjs,nObjsAlloc,nPis;int *pTNodes;int nTNodesSize;} Saig_MvMan_t;
#define SAIG_UNDEF_VALUE 0x1ffffffe
static int Saig_MvConst0(void){return 1;}
static int Saig_MvIsConst1(int a){return a==0;}
static int Saig_MvIsConst(int a){return a<2;}
static int Saig_MvIsUndef(int a){return a==SAIG_UNDEF_VALUE;}
static int Saig_MvUndef(void){return SAIG_UNDEF_VALUE;}
static int Saig_MvNot(int a){return a^1;}
static int Saig_MvLit2Var(int a){return a>>1;}
static int Saig_MvVar2Lit(int a){return a<<1;}
static int Saig_MvIsComplement(int a){return a&1;}
static int Saig_MvLev(Saig_MvMan_t*p,int a){return p->pLevels[a>>1];}
`
	var source strings.Builder
	source.WriteString(header)
	for _, name := range []string{"Saig_MvCreateObj", "Saig_MvHash", "Saig_MvTableFind", "Saig_MvAnd"} {
		source.WriteString(extract(name))
	}
	cases := [][3]int32{{2, 2, 1}, {2, 3, 1}, {0, 2, 1}, {1, 2, 1}, {SAIG_UNDEF_VALUE, 1, 1}, {SAIG_UNDEF_VALUE, 0, 1}, {SAIG_UNDEF_VALUE, SAIG_UNDEF_VALUE, 1}, {2, 4, 0}, {2, 4, 1}, {4, 2, 1}, {3, 4, 1}, {3, 5, 1}}
	source.WriteString(`int main(void){Saig_MvMan_t p={0};p.nObjsAlloc=64;p.nTNodesSize=17;p.pAigNew=calloc(64,sizeof(*p.pAigNew));p.pLevels=calloc(64,1);p.pTNodes=calloc(17,sizeof(int));Saig_MvCreateObj(&p,0,0);Saig_MvCreateObj(&p,0,0);Saig_MvCreateObj(&p,0,0);`)
	for _, c := range cases {
		fmt.Fprintf(&source, "printf(\"%%d\\n\",Saig_MvAnd(&p,%d,%d,%d));\n", c[0], c[1], c[2])
	}
	source.WriteString(`int last=2;for(int i=0;i<300;i++){last=Saig_MvAnd(&p,last,2+2*(i%2),1);printf("%d\n",last);} return 0;}` + "\n")
	dir := t.TempDir()
	input, binary := filepath.Join(dir, "mv.c"), filepath.Join(dir, "mv")
	if err := os.WriteFile(input, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c99", "-O0", input, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("reference compile: %v\n%s", err, out)
	}
	out, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("reference execution: %v\n%s", err, out)
	}
	g := NewGraph()
	lo := g.NewInput()
	g.AddOutput(lo)
	g.nRegs = 1
	p := Aig_MvManStart(g, 2)
	p.newLiteral()
	p.newLiteral()
	var want strings.Builder
	for _, c := range cases {
		fmt.Fprintln(&want, p.and(c[0], c[1], c[2] != 0))
	}
	last := int32(2)
	for i := 0; i < 300; i++ {
		last = p.and(last, 2+2*int32(i%2), true)
		fmt.Fprintln(&want, last)
	}
	if string(out) != want.String() {
		t.Fatalf("original C:\n%sGo:\n%s", out, want.String())
	}
}
