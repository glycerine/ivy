package abc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"math/rand"
	"strings"
	"testing"
)

func TestSopConstructorsAndTruth(t *testing.T) {
	for n := int32(1); n <= 6; n++ {
		flags := make([]int32, n)
		for i := range flags {
			flags[i] = int32(i & 1)
		}
		constructors := []string{Abc_SopCreateAnd(nil, n, flags), Abc_SopCreateOr(nil, n, flags), Abc_SopCreateOrMultiCube(nil, n, flags), Abc_SopCreateNor(nil, n), Abc_SopCreateNand(nil, n)}
		for mask := 0; mask < 1<<uint(n); mask++ {
			and, or := true, false
			plainAnd, plainOr := true, false
			for v := int32(0); v < n; v++ {
				x := mask&(1<<uint(v)) != 0
				plainAnd = plainAnd && x
				plainOr = plainOr || x
				if flags[v] != 0 {
					x = !x
				}
				and = and && x
				or = or || x
			}
			want := []bool{and, or, or, !plainOr, !plainAnd}
			for i, s := range constructors {
				if Abc_SopCheck(s, n) == 0 || Abc_SopToTruth(s, n)&(1<<uint(mask)) != 0 != want[i] {
					t.Fatalf("n%d mask%d constructor%d", n, mask, i)
				}
			}
		}
	}
	if Abc_SopToTruth(Abc_SopCreateMux(nil), 3) != 0xd8d8d8d8d8d8d8d8 {
		t.Fatal("MUX input order")
	}
	if Abc_SopIsInv(Abc_SopCreateInv(nil)) != 1 || Abc_SopIsBuf(Abc_SopCreateBuf(nil)) != 1 {
		t.Fatal("buffer/inverter predicates")
	}
}
func TestSopTruthAndIsopRoundTrips(t *testing.T) {
	r := rand.New(rand.NewSource(612))
	cover := vec.Vec_IntAlloc(0)
	for trial := 0; trial < 300; trial++ {
		n := int32(1 + trial%7)
		words := 1
		if n > 6 {
			words = 2
		}
		truth := make([]uint64, words)
		for i := range truth {
			truth[i] = r.Uint64()
		}
		if n < 6 {
			period := uint(1) << uint(n)
			mask := (uint64(1) << period) - 1
			x := truth[0] & mask
			for shift := period; shift < 64; shift *= 2 {
				x |= x << shift
			}
			truth[0] = x
		}
		s := Abc_SopCreateFromTruthIsop(nil, n, truth, cover)
		res := make([]uint64, words)
		cube := make([]uint64, words)
		if Abc_SopIsConst0(s) != 0 || Abc_SopIsConst1(s) != 0 {
			continue
		}
		Abc_SopToTruthBig(s, n, sopVariables(n, words), cube, res)
		for i := range res {
			if res[i] != truth[i] {
				t.Fatalf("ISOP round trip trial%d", trial)
			}
		}
		Abc_SopToIsop(s, cover)
		rebuilt := Abc_SopCreateFromIsop(nil, n, cover)
		if Abc_SopIsComplement(s) != 0 {
			rebuilt = Abc_SopComplement(rebuilt)
		}
		if s != rebuilt {
			t.Fatal("cube encoding round trip")
		}
	}
}
func TestSopComplementAndTextOrder(t *testing.T) {
	s := "1-0 1\n-11 1\n"
	truth := Abc_SopToTruth(s, 3)
	if Abc_SopToTruth(Abc_SopComplement(s), 3) != ^truth || Abc_SopComplement(Abc_SopComplement(s)) != s {
		t.Fatal("complement")
	}
	for v := int32(0); v < 3; v++ {
		changed := Abc_SopToTruth(Abc_SopComplementVar(s, v), 3)
		for mask := 0; mask < 64; mask++ {
			if (changed >> uint(mask) & 1) != (truth >> uint(mask^(1<<uint(v))) & 1) {
				t.Fatal("variable complement")
			}
		}
	}
	if Abc_SopFromTruthBin("1000") != "11 1\n" || Abc_SopFromTruthHex("8") != "11 1\n" || Abc_SopFromTruthBin("0100") != "10 1\n" || Abc_SopFromTruthHex("4") != "01 1\n" {
		t.Fatal("source text variable ordering")
	}
	if Abc_SopFromTruthBin("0000") != "" || Abc_SopFromTruthHex("z") != "" {
		t.Fatal("invalid text convention")
	}
	x := Abc_SopCreateXorSpecial(nil, 4)
	if Abc_SopIsExorType(x) != 1 || Abc_SopGetPhase(Abc_SopComplement(x)) != 0 {
		t.Fatal("x/n tags")
	}
}
func TestSopMVCodecsAndValidation(t *testing.T) {
	if Abc_SopEncoderPos(nil, 2, 4) != "d0\n2 1\n" || Abc_SopEncoderLog(nil, 1, 6) != "d0\n(2,3) 1\n" || Abc_SopDecoderLog(nil, 3) != "0 0 0\n1 0 1\n0 1 2\n" {
		t.Fatal("MV codec")
	}
	if strings.Count(Abc_SopDecoderPos(nil, 3), "\n") != 3 {
		t.Fatal("positional decoder")
	}
	for _, s := range []string{"1? 1\n", "11 1", "11 0\n11 1\n", "11 z\n"} {
		if Abc_SopCheck(s, 2) != 0 {
			t.Fatal("invalid SOP accepted")
		}
	}
}
func TestSopLargeTruthAndMinterms(t *testing.T) {
	s := "1------1 1\n-0----1- 1\n"
	n := int32(8)
	vars := sopVariables(n, 4)
	res, cube := make([]uint64, 4), make([]uint64, 4)
	Abc_SopToTruthBig(s, n, vars, cube, res)
	for mint := 0; mint < 256; mint++ {
		want := mint&1 != 0 && mint&128 != 0 || mint&2 == 0 && mint&64 != 0
		if (res[mint/64]&(1<<uint(mint%64)) != 0) != want {
			t.Fatal("large truth")
		}
	}
	truth := []uint32{0x80000001}
	cover := Abc_SopCreateFromTruth(nil, 5, truth)
	if Abc_SopGetCubeNum(cover) != 2 || bits.OnesCount64(Abc_SopToTruth(cover, 5)) != 4 {
		t.Fatal("minterm constructor")
	}
}

func TestSopAllTwoVariableIsops(t *testing.T) {
	cover := vec.Vec_IntAlloc(0)
	for truth := uint64(0); truth < 16; truth++ {
		word := truth
		for shift := uint(4); shift < 64; shift *= 2 {
			word |= word << shift
		}
		s := Abc_SopCreateFromTruthIsop(nil, 2, []uint64{word}, cover)
		var actual uint64
		if Abc_SopGetVarNum(s) == 0 {
			actual = Abc_SopToTruth(s, 0)
		} else {
			actual = Abc_SopToTruth(s, 2)
		}
		if actual != word {
			t.Fatalf("truth%x ISOP %q", truth, s)
		}
	}
}
