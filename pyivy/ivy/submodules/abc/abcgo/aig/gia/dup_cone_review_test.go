package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
	"slices"
	"strings"
	"testing"
)

func TestDfsSkipRejectsMarkedRegisterInputs(t *testing.T) {
	src := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a^1)
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	Gia_ManRi(src, 0).fMark1 = 1
	defer func() {
		if recover() == nil {
			t.Fatal("skipping an RI silently reclassified a property as next-state logic")
		}
	}()
	Gia_ManDupDfsSkip(src)
}

func coneReviewOutputRecords(t *testing.T, typed, layout, skip bool) string {
	t.Helper()
	dependencies := []uint{7, 56, 8, 0, 3, 0}
	if typed {
		dependencies = []uint{3, 255, 12, 28, 33}
	}
	var records strings.Builder
	for signs := int32(0); signs < 32; signs++ {
		for _, regs := range []int32{0, 2, 4} {
			properties := len(dependencies) - int(regs)
			count := properties
			if skip {
				count = 1 << uint(properties)
			}
			for choice := 0; choice < count; choice++ {
				src, _ := levelReviewGraph(signs, 0, typed)
				Gia_ManAppendCi(src)
				Gia_ManSetRegNum(src, regs)
				src.nConstrs = 1
				src.pCexSeq = util.AbcCexSentinel
				if layout {
					src = Gia_ManDupOrderDfsReverse(src)
					src.pCexSeq = util.AbcCexSentinel
				}
				var selected []int32
				if skip {
					for po := 0; po < properties; po++ {
						if choice&(1<<uint(po)) == 0 {
							selected = append(selected, int32(po))
						}
					}
				} else {
					selected = []int32{int32(choice)}
				}
				for i := int32(properties); i < Gia_ManCoNum(src); i++ {
					selected = append(selected, i)
				}
				for _, obj := range src.pObjs {
					obj.fMark0 = uint8(obj.id & 1)
					obj.fMark1 = uint8((obj.id >> 1) & 1)
				}
				if skip {
					for i := int32(0); i < Gia_ManCoNum(src); i++ {
						Gia_ManCo(src, i).fMark1 = 0
						if i < int32(properties) && choice&(1<<uint(i)) != 0 {
							Gia_ManCo(src, i).fMark1 = 1
						}
					}
				}
				marks := make([][2]uint8, src.nObjs)
				for id, obj := range src.pObjs {
					marks[id] = [2]uint8{obj.fMark0, obj.fMark1}
				}
				var out *Gia_Man_t
				if skip {
					out = Gia_ManDupDfsSkip(src)
				} else {
					out = Gia_ManDupDfsOnePo(src, int32(choice))
				}
				mask := uint(0)
				for _, index := range selected {
					mask |= dependencies[index]
				}
				name, spec := out.Identity()
				if Gia_ManCiNum(out) != 5 || Gia_ManCoNum(out) != int32(len(selected)) || out.nRegs != regs || Gia_ManPoNum(out) != int32(len(selected))-regs || out.nConstrs != 0 || Gia_ManAndNum(out) != int32(bits.OnesCount(mask)) || name != "levels" || spec != "levels.aig" || out.pCexSeq != nil {
					t.Fatal("cone copy changed retained interface/logic/metadata policy")
				}
				if !skip && Gia_ManIsNormalized(out) == 0 {
					t.Fatal("one-output copy not normalized")
				}
				for id, obj := range src.pObjs {
					if [2]uint8{obj.fMark0, obj.fMark1} != marks[id] {
						t.Fatal("cone copy changed source marks")
					}
				}
				for i := int32(0); i < 5; i++ {
					if Gia_ManCi(src, i).Value != uint32(Gia_ManCi(out, i).id<<1) {
						t.Fatal("input order changed")
					}
				}
				for i := int32(0); i < Gia_ManCoNum(src); i++ {
					position := slices.Index(selected, i)
					want := ^uint32(0)
					if position >= 0 {
						want = uint32(Gia_ManCo(out, int32(position)).id << 1)
					}
					if Gia_ManCo(src, i).Value != want {
						t.Fatal("wrong kept/dropped output mapping")
					}
				}
				fmt.Fprintf(&records, "%d %d %d %d %d %d %d", signs, regs, choice, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
				if !skip {
					for _, obj := range src.pObjs {
						fmt.Fprintf(&records, " %d", obj.Value)
					}
				}
				for assignment := 0; assignment < 32; assignment++ {
					inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0, assignment&16 != 0}
					outputs, _ := selectionReviewTruth(signs, typed, inputs)
					want := make([]bool, len(selected))
					for i, index := range selected {
						want[i] = outputs[index]
					}
					checkQuantificationOutputs(t, out, inputs, want)
					packed := 0
					for i, value := range want {
						if value {
							packed |= 1 << uint(i)
						}
					}
					fmt.Fprintf(&records, " %d", packed)
				}
				records.WriteByte('\n')
			}
		}
	}
	return records.String()
}

func TestDfsOutputConesPreservePropertiesAndState(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			for _, skip := range []bool{false, true} {
				t.Run(fmt.Sprintf("typed%v_layout%v_skip%v", typed, layout, skip), func(t *testing.T) { coneReviewOutputRecords(t, typed, layout, skip) })
			}
		}
	}
}

func TestDfsOneOutputMatchesOriginalSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "dfs_one_output.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C one-output: %v: %s", err, native)
	}
	got, want := strings.Split(strings.TrimSpace(coneReviewOutputRecords(t, false, false, false)), "\n"), strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases %d, C cases %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case %d:\nGo %s\nC %s", i, got[i], want[i])
		}
	}
}

func TestDfsOutputConeBoundaries(t *testing.T) {
	for _, kind := range []string{"constants", "wire", "sequential"} {
		t.Run(kind, func(t *testing.T) {
			src := Gia_ManStart(1)
			var inputRows, wants [][]bool
			switch kind {
			case "constants":
				Gia_ManAppendCo(src, 1)
				inputRows = [][]bool{nil}
				wants = [][]bool{{true}}
			case "wire":
				a := Gia_ManAppendCi(src)
				Gia_ManAppendCi(src)
				Gia_ManAppendCo(src, a^1)
				inputRows = [][]bool{{false, false}, {true, true}}
				wants = [][]bool{{true}, {false}}
			case "sequential":
				Gia_ManAppendCi(src)
				r := Gia_ManAppendCi(src)
				Gia_ManAppendCo(src, 1)
				Gia_ManAppendCo(src, r^1)
				Gia_ManSetRegNum(src, 1)
				inputRows = [][]bool{{false, false}, {true, true}}
				wants = [][]bool{{true, true}, {true, false}}
			}
			one := Gia_ManDupDfsOnePo(src, 0)
			for i, inputs := range inputRows {
				checkQuantificationOutputs(t, one, inputs, wants[i])
			}
			Gia_ManPo(src, 0).fMark1 = 1
			skipped := Gia_ManDupDfsSkip(src)
			if Gia_ManPoNum(skipped) != 0 || skipped.nRegs != src.nRegs || Gia_ManCiNum(skipped) != Gia_ManCiNum(src) {
				t.Fatal("skipping all properties changed the state interface")
			}
			for i, inputs := range inputRows {
				checkQuantificationOutputs(t, skipped, inputs, wants[i][1:])
			}
		})
	}
	empty := Gia_ManStart(1)
	out := Gia_ManDupDfsSkip(empty)
	if out.nObjs != 1 || out.nRegs != 0 {
		t.Fatal("empty skip copy changed graph")
	}
	for _, choice := range []int32{-1, 0, 1<<31 - 1} {
		t.Run(fmt.Sprint(choice), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid property index accepted")
				}
			}()
			Gia_ManDupDfsOnePo(empty, choice)
		})
	}
	stateOnly := Gia_ManStart(1)
	r := Gia_ManAppendCi(stateOnly)
	Gia_ManAppendCo(stateOnly, r)
	Gia_ManSetRegNum(stateOnly, 1)
	out = Gia_ManDupDfsSkip(stateOnly)
	checkQuantificationOutputs(t, out, []bool{true}, []bool{true})
	if out.nRegs != 1 || Gia_ManPoNum(out) != 0 {
		t.Fatal("register-only graph lost state")
	}
}

func TestDfsSkipRejectsEachRegisterInput(t *testing.T) {
	for marked := int32(0); marked < 3; marked++ {
		t.Run(fmt.Sprint(marked), func(t *testing.T) {
			src := Gia_ManStart(1)
			a := Gia_ManAppendCi(src)
			regs := []int32{Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)}
			Gia_ManAppendCo(src, a)
			Gia_ManAppendCo(src, a^1)
			for _, reg := range regs {
				Gia_ManAppendCo(src, reg)
			}
			Gia_ManSetRegNum(src, 3)
			Gia_ManRi(src, marked).fMark1 = 1
			for _, obj := range src.pObjs {
				obj.Value = 99
			}
			defer func() {
				if recover() == nil {
					t.Error("marked register input accepted")
				}
				for _, obj := range src.pObjs {
					if obj.Value != 99 {
						t.Error("invalid skip changed source mappings before rejecting")
					}
				}
			}()
			Gia_ManDupDfsSkip(src)
		})
	}
}
