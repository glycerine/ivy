package gia

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"time"
)

func Gia_ManHasChoices_very_old(p *Gia_Man_t) int32 {
	if p.PReprs == nil || p.PNexts == nil {
		return 0
	}
	var reprs, nexts int
	for _, obj := range p.pObjs {
		if Gia_ObjReprObj(p, obj.id) != nil {
			reprs++
		}
		if Gia_ObjNext(p, obj.id) != 0 {
			nexts++
		}
	}
	if reprs == 0 {
		fmt.Println("Warning: AIG has repr data-strucure but not reprs.")
		return 0
	}
	fmt.Printf("%d nodes have reprs.\n%d nodes have nexts.\n", reprs, nexts)
	p.Refs = nil
	Gia_ManCreateRefs(p)
	var missing, referenced, choices int
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		repr := Gia_ObjReprObj(p, obj.id)
		if Gia_ObjRefNum(p, obj) == 0 {
			if repr == nil {
				missing++
			} else {
				choices++
			}
		} else if repr != nil {
			referenced++
		}
		if repr != nil && Gia_ObjRepr(p, obj.id) >= obj.id {
			panic("gia: invalid choice representative")
		}
	}
	if choices == 0 {
		return 0
	}
	if missing != 0 {
		fmt.Printf("Gia_ManHasChoices_very_old(): Error: %d internal nodes have no fanout and no repr.\n", missing)
	}
	if referenced != 0 {
		fmt.Printf("Gia_ManHasChoices_very_old(): Error: %d internal nodes have both fanout and repr.\n", referenced)
	}
	return 1
}

// Gia_ManGroupProve executes the supplied ABC command on successive output
// groups and preserves each output's status and available counterexample.
func Gia_ManGroupProve(initial *Gia_Man_t, command string, groupSize, verbose int32) *Vec_Int_t {
	if groupSize <= 0 || command == "" || initial.nConstrs != 0 {
		panic("gia: invalid grouped proof parameters")
	}
	frame := abcmain.Abc_FrameGetGlobalFrame()
	p := Gia_ManDup(initial)
	defer Gia_ManStop(p)
	fmt.Printf("RUNNING MultiProve: Group size = %d. Command line = %q.\n", groupSize, command)
	start := time.Now()
	var commandTime time.Duration
	outputs := Vec_IntArray(Vec_IntStartNatural(int(Gia_ManPoNum(p))))
	statuses := Vec_IntAlloc(len(outputs))
	counterexamples := make([]any, 0, len(outputs))
	for first := 0; first < len(outputs); first += int(groupSize) {
		last := min(first+int(groupSize), len(outputs))
		group := Gia_ManDupCones(p, outputs[first:last], int32(last-first), 0)
		fmt.Printf("GROUP %4d : %4d <= PoId < %4d : ", first/int(groupSize), first, last)
		abcmain.Abc_FrameUpdateGia(frame, group)
		when := time.Now()
		status := abcmain.Cmd_CommandExecute(frame, command)
		commandTime += time.Since(when)
		if status != 0 {
			fmt.Println("Command failed.")
			return nil
		}
		if groupSize == 1 {
			Vec_IntPush(statuses, abcmain.Abc_FrameReadProbStatus(frame))
			counterexamples = append(counterexamples, abcmain.Abc_FrameReadCex(frame))
			continue
		}
		values := abcmain.Abc_FrameReadPoStatuses(frame)
		cexes := abcmain.Abc_FrameReadCexVec(frame)
		if len(values) < last-first || cexes != nil && len(cexes) < last-first {
			panic("gia: incomplete grouped proof results")
		}
		for i := 0; i < last-first; i++ {
			Vec_IntPush(statuses, values[i])
			if cexes != nil {
				counterexamples = append(counterexamples, cexes[i])
			} else {
				counterexamples = append(counterexamples, nil)
			}
		}
	}
	for _, cex := range counterexamples {
		if cex != nil {
			frame.SetListValues("vCexVec", counterexamples)
			break
		}
	}
	fmt.Printf("SUMMARY:  Properties = %6d. UNSAT = %6d. SAT = %6d. UNDEC = %6d.\n", len(outputs), Vec_IntCountEntry(statuses, 1), Vec_IntCountEntry(statuses, 0), Vec_IntCountEntry(statuses, -1))
	fmt.Printf("Command time = %.2f sec  Total time = %.2f sec\n", commandTime.Seconds(), time.Since(start).Seconds())
	return statuses
}
