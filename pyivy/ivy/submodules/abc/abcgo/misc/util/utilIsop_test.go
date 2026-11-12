package util

import (
	"fmt"
	"math/bits"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func isopTestStretch(t word, n int32) word {
	if n >= 6 {
		return t
	}
	width := uint(1 << uint(n))
	mask := word(1)<<width - 1
	t &= mask
	for shift := width; shift < 64; shift *= 2 {
		t |= t << shift
	}
	return t
}
func isopTestCoverage(t *testing.T, on, dc, result []word, n int32, cost word, cover []int32, xor bool) {
	t.Helper()
	count := Abc_CostCubes(cost)
	if count < 0 || int(count) > len(cover) {
		t.Fatal("bad cube count", cost)
	}
	cover = cover[:count]
	rebuilt := make([]word, isopWords(n))
	mode := int32(0)
	if xor {
		mode = 1
	}
	Abc_IsopBuildTruth(cover, n, rebuilt, mode, 0)
	for i := range rebuilt {
		if rebuilt[i]&on[i] != on[i] || rebuilt[i]&^dc[i] != 0 {
			t.Fatalf("coverage mismatch n=%d word=%d got=%x on=%x dc=%x cubes=%v", n, i, rebuilt[i], on[i], dc[i], cover)
		}
		if result != nil && rebuilt[i] != result[i] {
			t.Fatal("returned truth differs from cube truth")
		}
	}
	if Abc_IsopCountLits(cover, n) != Abc_CostLits(cost) {
		t.Fatal("cost literal count", cost, cover)
	}
}
func TestIsopExhaustiveThreeVariables(t *testing.T) {
	limit := Abc_Cube2Cost(1000)
	for table := word(0); table < 256; table++ {
		on := []word{isopTestStretch(table, 3)}
		result := make([]word, 1)
		cover := make([]int32, 100)
		cost := Abc_IsopCheck(on, on, result, 3, limit, cover)
		isopTestCoverage(t, on, on, result, 3, cost, cover, false)
		count := Abc_CostCubes(cost)
		for i := int32(0); i < count; i++ {
			without := append([]int32(nil), cover[:i]...)
			without = append(without, cover[i+1:count]...)
			truth := make([]word, 1)
			Abc_IsopBuildTruth(without, 3, truth, 0, 0)
			if truth[0] == on[0] {
				t.Fatalf("redundant cube for truth %x", on[0])
			}
		}
		esop := make([]int32, 100)
		cost = Abc_EsopCheck(on, 3, limit, esop)
		isopTestCoverage(t, on, on, nil, 3, cost, esop, true)
	}
}
func TestIsopDontCareCoverage(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for n := int32(1); n <= 6; n++ {
		for k := 0; k < 150; k++ {
			dc := isopTestStretch(r.Uint64(), n)
			on := isopTestStretch(r.Uint64(), n) & dc
			cover := make([]int32, 500)
			result := make([]word, 1)
			cost := Abc_Isop6Cover(on, dc, result, n, Abc_Cube2Cost(1000), cover)
			isopTestCoverage(t, []word{on}, []word{dc}, result, n, cost, cover, false)
			greedy := make([]int32, 500)
			cost = Abc_IsopNew([]word{on}, []word{dc}, result, n, Abc_Cube2Cost(1000), greedy)
			isopTestCoverage(t, []word{on}, []word{dc}, result, n, cost, greedy, false)
		}
	}
}
func TestIsopLargeSupport(t *testing.T) {
	wrappers := []FUNC_ISOP{Abc_Isop7Cover, Abc_Isop8Cover, Abc_Isop9Cover, Abc_Isop10Cover, Abc_Isop11Cover, Abc_Isop12Cover, Abc_Isop13Cover, Abc_Isop14Cover, Abc_Isop15Cover, Abc_Isop16Cover}
	for n := int32(7); n <= 16; n++ {
		on := make([]word, isopWords(n))
		for i := range on {
			on[i] = isopMask(0) & isopMask(1)
			if i&(1<<uint(n-7)) != 0 {
				on[i] = ^on[i]
			}
		}
		copyOn := append([]word(nil), on...)
		result := make([]word, len(on))
		cover := make([]int32, 100)
		cost := wrappers[n-7](on, on, result, Abc_Cube2Cost(1000), cover)
		isopTestCoverage(t, on, on, result, n, cost, cover, false)
		cost = Abc_EsopCheck(on, n, Abc_Cube2Cost(1000), cover)
		isopTestCoverage(t, on, on, nil, n, cost, cover, true)
		cost = Abc_IsopNew(on, on, result, n, Abc_Cube2Cost(1000), cover)
		isopTestCoverage(t, on, on, result, n, cost, cover, false)
		if !reflect.DeepEqual(on, copyOn) {
			t.Fatal("input mutated")
		}
	}
}
func TestIsopLimitsPolarityAndConstants(t *testing.T) {
	for _, constant := range []word{0, ^word(0)} {
		var cover []int32
		flag := Abc_Isop([]word{constant}, 0, 10, &cover, 1)
		result := make([]word, 1)
		Abc_IsopBuildTruth(cover, 0, result, 0, flag)
		if result[0] != constant {
			t.Fatal("constant result", flag, cover)
		}
	}
	on := []word{isopMask(0) | isopMask(1) | isopMask(2)}
	before := append([]word(nil), on...)
	var cover []int32
	flag := Abc_Isop(on, 3, 100, &cover, 1)
	if flag != 1 || len(cover) != 1 {
		t.Fatal("wrong preferred polarity", flag, cover)
	}
	result := make([]word, 1)
	Abc_IsopBuildTruth(cover, 3, result, 0, flag)
	if result[0] != on[0] || !reflect.DeepEqual(on, before) {
		t.Fatal("polarity/input mismatch")
	}
	if Abc_Isop(on, 3, 1, &cover, 0) != -1 {
		t.Fatal("cube limit ignored")
	}
	limit := Abc_Cube2Cost(1)
	if Abc_IsopCheck(on, on, result, 3, limit, nil) != limit || Abc_EsopCheck(on, 3, limit, nil) != limit {
		t.Fatal("packed cutoff not saturated")
	}
}
func TestIsopCnfAndCubeCost(t *testing.T) {
	on := []word{isopMask(0) & isopMask(1)}
	cover := make([]int32, 100)
	count := Abc_IsopCnf(on, 2, 100, cover)
	if count <= 0 {
		t.Fatal(count)
	}
	relation := make([]word, 1)
	Abc_IsopBuildTruth(cover[:count], 3, relation, 0, 0)
	expected := (on[0] &^ isopMask(2)) | (^on[0] & isopMask(2))
	if relation[0] != expected {
		t.Fatal("CNF polarity encoding")
	}
	for _, cube := range cover[:count] {
		if bits.OnesCount32(uint32(cube)) == 0 {
			t.Fatal("missing output literal")
		}
	}
}

// Compare recursion costs and literal order against the original C routines.
// The production package has no C dependency; this differential test is optional
// on systems without a C compiler or the reference checkout.
func TestIsopScalarCReference(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable for reference comparison")
	}
	source, err := os.ReadFile("../../../src/misc/util/utilIsop.c")
	if err != nil {
		t.Fatal(err)
	}
	extract := func(name string) string {
		pattern := regexp.MustCompile(`(?m)^(?:static inline )?(?:word|int|void) ` + name + `\([^\n]*\)\n\{`)
		loc := pattern.FindIndex(source)
		if loc == nil {
			t.Fatalf("missing C reference %s", name)
		}
		open := loc[1] - 1
		depth := 1
		end := open + 1
		for ; end < len(source) && depth > 0; end++ {
			if source[end] == '{' {
				depth++
			}
			if source[end] == '}' {
				depth--
			}
		}
		return string(source[loc[0]:end]) + "\n"
	}
	code := `#include <stdint.h>
#include <stdio.h>
#include <assert.h>
typedef uint64_t word;
static word s_Truths6[6]={0xAAAAAAAAAAAAAAAAULL,0xCCCCCCCCCCCCCCCCULL,0xF0F0F0F0F0F0F0F0ULL,0xFF00FF00FF00FF00ULL,0xFFFF0000FFFF0000ULL,0xFFFFFFFF00000000ULL};
static word s_Truths6Neg[6];
#define Abc_Cube2Cost(n) ((word)(n)<<32)
#define Abc_CostCubes(c) ((int)((c)>>32))
#define Abc_Var2Lit(v,c) ((v)*2+(c))
#define Abc_MaxWord(a,b) ((a)>(b)?(a):(b))
word Abc_Tt6Cofactor0(word t,int v){word x=t&~s_Truths6[v];return x|(x<<(1<<v));}
word Abc_Tt6Cofactor1(word t,int v){word x=t&s_Truths6[v];return x|(x>>(1<<v));}
int Abc_Tt6HasVar(word t,int v){return Abc_Tt6Cofactor0(t,v)!=Abc_Tt6Cofactor1(t,v);}
`
	for _, name := range []string{"Abc_IsopAddLits", "Abc_Isop6Cover", "Abc_EsopAddLits", "Abc_Esop6Cover"} {
		code += extract(name)
	}
	code += `int main(void){for(int v=0;v<6;v++)s_Truths6Neg[v]=~s_Truths6[v];for(int x=0;x<256;x++){word t=x;for(int s=8;s<64;s*=2)t|=t<<s;word r=0;int cubes[100];word cost=Abc_Isop6Cover(t,t,&r,3,Abc_Cube2Cost(1000),cubes);printf("%llu ",(unsigned long long)cost);for(int i=0;i<Abc_CostCubes(cost);i++)printf("%u ",(unsigned)cubes[i]);cost=Abc_Esop6Cover(t,3,Abc_Cube2Cost(1000),cubes);printf("%llu ",(unsigned long long)cost);for(int i=0;i<Abc_CostCubes(cost);i++)printf("%u ",(unsigned)cubes[i]);puts("");}return 0;}`
	dir := t.TempDir()
	file := filepath.Join(dir, "reference.c")
	binary := filepath.Join(dir, "reference")
	if err := os.WriteFile(file, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c99", "-O2", file, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("reference compile: %v\n%s", err, out)
	}
	out, err := exec.Command(binary).Output()
	if err != nil {
		t.Fatal(err)
	}
	reader := strings.NewReader(string(out))
	for x := word(0); x < 256; x++ {
		truth := isopTestStretch(x, 3)
		for mode := 0; mode < 2; mode++ {
			var expected word
			if _, err := fmt.Fscan(reader, &expected); err != nil {
				t.Fatal(err)
			}
			cover := make([]int32, 100)
			var cost word
			if mode == 0 {
				cost = Abc_Isop6Cover(truth, truth, make([]word, 1), 3, Abc_Cube2Cost(1000), cover)
			} else {
				cost = Abc_Esop6Cover(truth, 3, Abc_Cube2Cost(1000), cover)
			}
			if cost != expected {
				t.Fatalf("truth %x mode %d cost %x != C %x", truth, mode, cost, expected)
			}
			for i := int32(0); i < Abc_CostCubes(cost); i++ {
				var cube uint32
				if _, err := fmt.Fscan(reader, &cube); err != nil {
					t.Fatal(err)
				}
				if uint32(cover[i]) != cube {
					t.Fatalf("truth %x mode %d cube %d = %x != C %x", truth, mode, i, cover[i], cube)
				}
			}
		}
	}
}
