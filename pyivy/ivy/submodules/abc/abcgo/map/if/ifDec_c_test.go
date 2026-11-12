package if_

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const decOraclePrelude = `#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
typedef uint64_t word;
typedef struct If_Man_t If_Man_t;
#define ABC_CONST(x) UINT64_C(x)
#define ABC_NAMESPACE_IMPL_START
#define ABC_NAMESPACE_IMPL_END
static int Abc_Tt6FirstBit(word t){return t?__builtin_ctzll(t):-1;}
static word Abc_Tt6Stretch(word t,int n){int width=1<<n;if(width<64)t&=(UINT64_C(1)<<width)-1;for(int size=width;size<64;size*=2)t|=t<<size;return t;}
static int Abc_TtHasVar(word*t,int n,int v){for(int m=0;m<(1<<n);m++)if(!(m&(1<<v))&&((t[m/64]>>(m%64))&1)!=((t[(m|(1<<v))/64]>>((m|(1<<v))%64))&1))return 1;return 0;}
static void Abc_TtSwapVars(word*t,int n,int a,int b){word copy[16]={0};int words=n<=6?1:1<<(n-6);for(int m=0;m<words*64;m++){int source=m;if(((m>>a)&1)!=((m>>b)&1))source^=(1<<a)|(1<<b);copy[m/64]|=((t[source/64]>>(source%64))&1)<<(m%64);}for(int w=0;w<words;w++)t[w]=copy[w];}
/* Diagnostic formatting has no effect on decomposition decisions. */
void Extra_PrintBinary(FILE*f,unsigned*t,int n){for(int i=n-1;i>=0;i--)fprintf(f,"%u",(t[i/32]>>(i%32))&1);}
void Extra_PrintHex(FILE*f,unsigned*t,int n){int words=n<=5?1:1<<(n-5);for(int i=words-1;i>=0;i--)fprintf(f,"%08x",t[i]);}
void Kit_DsdPrintFromTruth(unsigned*t,int n){Extra_PrintHex(stderr,t,n);}
`

func TestDecOriginalCComparison(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	directory := t.TempDir()
	var objects []string
	for _, family := range []string{"07", "08", "10"} {
		sourcePath := filepath.Join("..", "..", "..", "src", "map", "if", "ifDec"+family+".c")
		data, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Skip("original decomposition source unavailable")
		}
		var source strings.Builder
		prelude := decOraclePrelude
		if family != "07" {
			prelude = prelude[:strings.Index(prelude, "/* Diagnostic")]
		}
		source.WriteString(prelude)
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "#include") {
				continue
			}
			source.WriteString(line)
			source.WriteByte('\n')
		}
		file, object := filepath.Join(directory, "dec"+family+".c"), filepath.Join(directory, "dec"+family+".o")
		if err = os.WriteFile(file, []byte(source.String()), 0600); err != nil {
			t.Fatal(err)
		}
		if output, err := exec.Command(cc, "-std=c99", "-O2", "-DNDEBUG", "-c", file, "-o", object).CombinedOutput(); err != nil {
			t.Fatalf("compile original Dec%s: %v\n%s", family, err, output)
		}
		objects = append(objects, object)
	}
	main := `#include <stdint.h>
#include <stdio.h>
typedef uint64_t word;
word If_Dec5Perform(word,int);word If_Dec6Perform(word,int);word If_Dec7Perform(word*,int);int If_Dec08Perform(word*,int,int);int If_Dec10Perform(word*,int,int);
int main(void){int mode,n;while(scanf("%d%d",&mode,&n)==2){word t[16];int words=n<=6?1:1<<(n-6);for(int i=0;i<words;i++){unsigned long long x;if(scanf("%llx",&x)!=1)return 2;t[i]=x;}int result;if(mode==7){result=(n==5?If_Dec5Perform(t[0],0):n==6?If_Dec6Perform(t[0],0):If_Dec7Perform(t,0))!=0;}else if(mode==8)result=If_Dec08Perform(t,n,0);else result=If_Dec10Perform(t,n,0);printf("%d\n",result);}return 0;}`
	mainPath, binary := filepath.Join(directory, "main.c"), filepath.Join(directory, "oracle")
	if err = os.WriteFile(mainPath, []byte(main), 0600); err != nil {
		t.Fatal(err)
	}
	arguments := append([]string{"-std=c99", "-O2", mainPath}, objects...)
	arguments = append(arguments, "-o", binary)
	if output, err := exec.Command(cc, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("link original oracle: %v\n%s", err, output)
	}
	rng := rand.New(rand.NewSource(710810))
	type testCase struct {
		mode  int
		n     int32
		truth []uint64
		want  int32
	}
	var cases []testCase
	add := func(n int32, truth []uint64) {
		if decSupport(truth, n) != (1<<uint(n))-1 {
			return
		}
		if n <= 7 {
			var got uint64
			switch n {
			case 5:
				got = If_Dec5Perform(truth[0], 0)
			case 6:
				got = If_Dec6Perform(truth[0], 0)
			case 7:
				got = If_Dec7Perform(truth, 0)
			}
			want := int32(0)
			if got != 0 {
				want = 1
			}
			cases = append(cases, testCase{7, n, truth, want})
		}
		if n >= 6 && n <= 8 {
			cases = append(cases, testCase{8, n, truth, If_Dec08Perform(truth, n, 0)})
		}
		if n >= 6 {
			cases = append(cases, testCase{10, n, truth, If_Dec10Perform(truth, n, 0)})
		}
	}
	for n := int32(5); n <= 10; n++ {
		for trial := 0; trial < 50; trial++ {
			truth := make([]uint64, decWords(n))
			for i := range truth {
				truth[i] = rng.Uint64()
			}
			if n == 5 {
				truth[0] = uint64(uint32(truth[0])) * 0x100000001
			}
			add(n, truth)
		}
		add(n, decTestTruth(n, func(m int) bool {
			parity := false
			for m != 0 {
				parity = !parity
				m &= m - 1
			}
			return parity
		}))
	}
	for n := 5; n <= 7; n++ {
		for trial := 0; trial < 60; trial++ {
			order := rng.Perm(n)
			bound := order[:4]
			free := append([]int(nil), order[4:]...)
			free = append(free, bound[:7-n]...)
			config := uint64(uint16(rng.Uint32())) | uint64(uint16(rng.Uint32()))<<32
			for i, v := range bound {
				config |= uint64(v) << uint(16+4*i)
			}
			for i, v := range free {
				config |= uint64(v) << uint(48+4*i)
			}
			config |= uint64(7) << 60
			add(int32(n), decConfigTruth(config, int32(n)))
		}
	}
	var input bytes.Buffer
	for _, test := range cases {
		fmt.Fprintf(&input, "%d %d", test.mode, test.n)
		for _, word := range test.truth {
			fmt.Fprintf(&input, " %016x", word)
		}
		input.WriteByte('\n')
	}
	command := exec.Command(binary)
	command.Stdin = &input
	output, err := command.Output()
	if err != nil {
		t.Fatal("original oracle", err)
	}
	reader := bytes.NewReader(output)
	for i, test := range cases {
		var original int32
		if _, err = fmt.Fscanf(reader, "%d\n", &original); err != nil || original != test.want {
			t.Fatalf("original comparison %d mode=%d n=%d truth=%x: C=%d Go=%d (%v)", i, test.mode, test.n, test.truth, original, test.want, err)
		}
	}
	t.Logf("compared %d original-C decomposition decisions", len(cases))
}
