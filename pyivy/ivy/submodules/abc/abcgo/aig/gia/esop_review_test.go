package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func esopReviewTypedGraph(kind string, signs int32) *Gia_Man_t {
	g := Gia_ManStart(32)
	g.EnableMuxes()
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	a ^= signs & 1
	b ^= (signs >> 1) & 1
	c ^= (signs >> 2) & 1
	var lit int32
	switch kind {
	case "xor":
		lit = Gia_ManAppendXorReal(g, a, b)
	case "mux":
		lit = Gia_ManAppendMuxReal(g, c, a, b)
	case "buffer":
		lit = Gia_ManAppendBuf(g, a)
	default:
		panic("test gate")
	}
	Gia_ManAppendCo(g, lit)
	Gia_ManAppendCo(g, lit^1)
	return g
}

func esopReviewCheckCover(t *testing.T, src *Gia_Man_t, cover *Vec_Wec_t) {
	t.Helper()
	for bits := 0; bits < 1<<Gia_ManCiNum(src); bits++ {
		got := make([]bool, Gia_ManCoNum(src))
		for i := 0; i < vec.Vec_WecSize(cover); i++ {
			cube := vec.Vec_IntArray(vec.Vec_WecEntry(cover, i))
			if len(cube) == 0 {
				t.Fatal("missing output tag")
			}
			output := -cube[len(cube)-1] - 1
			if output < 0 || output >= Gia_ManCoNum(src) {
				t.Fatalf("invalid output tag: %v", cube)
			}
			product := true
			for _, lit := range cube[:len(cube)-1] {
				if lit < 0 || lit>>1 >= Gia_ManCiNum(src) {
					t.Fatalf("invalid variable: %v", cube)
				}
				product = product && ((bits&(1<<(lit>>1)) != 0) != (lit&1 != 0))
			}
			got[output] = got[output] != product
		}
		values := cofReviewValues(src, bits)
		for i := range got {
			if got[i] != values[Gia_ManCo(src, int32(i)).id] {
				t.Fatalf("input%d output%d cover%t graph%t", bits, i, got[i], values[Gia_ManCo(src, int32(i)).id])
			}
		}
	}
}

func TestEsopTypedGateTruthAndCover(t *testing.T) {
	for _, kind := range []string{"xor", "mux", "buffer"} {
		for signs := int32(0); signs < 8; signs++ {
			t.Run(fmt.Sprintf("%s/signs%d", kind, signs), func(t *testing.T) {
				src := esopReviewTypedGraph(kind, signs)
				defer Gia_ManStop(src)
				t.Run("graph", func(t *testing.T) {
					out := Eso_ManCompute(src, 0, nil)
					defer Gia_ManStop(out)
					cofReviewCheckRebuild(t, src, out)
				})
				t.Run("cover", func(t *testing.T) {
					var cover *Vec_Wec_t
					if Eso_ManCompute(src, 0, &cover) != nil {
						t.Fatal("extraction also returned graph")
					}
					esopReviewCheckCover(t, src, cover)
				})
			})
		}
	}
}

func TestEsopSequentialInterfaceMetadata(t *testing.T) {
	for _, layout := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse%t", layout), func(t *testing.T) {
			src := cofReviewGraph(5)
			Gia_ManHashStop(src)
			Gia_ManSetRegNum(src, 2)
			if layout {
				src = Gia_ManDupOrderDfsReverse(src)
			}
			if err := src.SetAigerProperties(0, 1); err != nil {
				t.Fatal(err)
			}
			src.SetIdentity("esop", "esop.aig")
			src.aigerLatchInits = []int32{1, 0}
			src.aigerComment = []byte("ESOP interface")
			out := Eso_ManCompute(src, 0, nil)
			if out == nil {
				t.Fatal("missing output graph")
			}
			defer Gia_ManStop(src)
			defer Gia_ManStop(out)
			cofReviewCheckRebuild(t, src, out)
			if out.nConstrs != 1 || !slices.Equal(out.aigerLatchInits, src.aigerLatchInits) || !slices.Equal(out.aigerComment, src.aigerComment) {
				t.Fatalf("lost metadata: constraints%d resets%v comment%q", out.nConstrs, out.aigerLatchInits, out.aigerComment)
			}
			if name, spec := out.Identity(); name != "esop" || spec != "esop.aig" {
				t.Fatal("lost identity")
			}
			out.aigerLatchInits[0] = 0
			out.aigerComment[0] = 'X'
			if src.aigerLatchInits[0] != 1 || string(src.aigerComment) != "ESOP interface" {
				t.Fatal("aliased interface metadata")
			}
		})
	}
}
