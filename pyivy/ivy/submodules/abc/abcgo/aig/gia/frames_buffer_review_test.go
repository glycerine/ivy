package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func TestFramesGeneratedNativeAbstraction(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	native, err := testoracle.Command(t, exe, "--generated-abs").CombinedOutput()
	if err != nil {
		t.Fatalf("C generated abstraction: %v\n%s", err, native)
	}
	var got strings.Builder
	for seed := uint32(0); seed < 32; seed++ {
		for frames := int32(1); frames <= 5; frames++ {
			g := framesReviewGenerated(seed)
			fmt.Fprintf(&got, "abstraction %d f%d\n", seed, frames)
			for row, data := range Vec_PtrArray(Gia_ManUnrollAbs(g, frames)) {
				fmt.Fprintf(&got, "row%d ids", row)
				for _, id := range Vec_IntArray(data.(*Vec_Int_t)) {
					fmt.Fprintf(&got, " %d", id)
				}
				got.WriteByte('\n')
			}
			Gia_ManStop(g)
		}
	}
	esopReviewCompare(t, native, got.String())
}

func TestFramesBufferDrivers(t *testing.T) {
	rows := 0
	for seed := uint32(0); seed < 8; seed++ {
		for _, frames := range []int32{1, 3} {
			for init := int32(0); init < 2; init++ {
				for disable := int32(0); disable < 2; disable++ {
					for or := int32(0); or < 2; or++ {
						g, _ := falseReviewTypedGraph(seed)
						p := Gia_ParFra_t{NFrames: frames, FInit: init, FDisableSt: disable, FOrPos: or}
						q := Gia_ManFrames(g, &p)
						rows += framesReviewCheckSequence(t, g, q, p)
						Gia_ManStop(q)
						Gia_ManStop(g)
					}
				}
			}
		}
	}
	t.Logf("128 buffer/mixed-gate configurations and %d independent sequential histories", rows)
}

func TestFramesBufferIncremental(t *testing.T) {
	rows := 0
	for seed := uint32(0); seed < 8; seed++ {
		for save := int32(0); save < 2; save++ {
			for init := int32(0); init < 2; init++ {
				g, _ := falseReviewTypedGraph(seed)
				p := Gia_ParFra_t{NFrames: 3, fSaveLastLit: save, FInit: init}
				raw := Gia_ManUnrollStart(g, &p)
				var q *Gia_Man_t
				for horizon := int32(1); horizon <= 3; horizon++ {
					q = Gia_ManUnrollAdd(raw, horizon).(*Gia_Man_t)
					for assignment := 0; assignment < 1<<Gia_ManCiNum(q); assignment++ {
						state := false
						if init == 0 {
							stateOffset := int32(0)
							if save == 0 {
								stateOffset = 9
							}
							state = assignment>>stateOffset&1 != 0
						}
						var expected []bool
						last := false
						for f := int32(0); f < horizon; f++ {
							offset := f * 3
							if save != 0 && init == 0 {
								offset++
							}
							input := assignment >> offset & 7
							if state {
								input |= 8
							}
							values := cofReviewValues(g, input)
							if save == 0 {
								for po := int32(0); po < 2; po++ {
									expected = append(expected, values[Gia_ManPo(g, po).id])
								}
							}
							last, state = values[Gia_ManPo(g, 1).id], values[Gia_ManRi(g, 0).id]
						}
						actual := cofReviewValues(q, assignment)
						if len(expected) != int(Gia_ManCoNum(q)) {
							t.Fatal("wrong incremental output count")
						}
						for co, want := range expected {
							if actual[Gia_ManCo(q, int32(co)).id] != want {
								t.Fatalf("seed%d save%d init%d horizon%d assignment%d output%d", seed, save, init, horizon, assignment, co)
							}
						}
						lit := Gia_ManUnrollLastLit(raw)
						if save != 0 {
							if got := actual[lit>>1] != (lit&1 != 0); got != last {
								t.Fatal("saved mixed/buffer property differs")
							}
						} else if lit != 0 {
							t.Fatal("fixed mode saved a property")
						}
						rows++
					}
				}
				Gia_ManUnrollStop(raw)
				Gia_ManStop(q)
				Gia_ManStop(g)
			}
		}
	}
	t.Logf("96 incremental buffer/mixed configurations and %d independent histories", rows)
}
