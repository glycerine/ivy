package gia

// Iterative randomized synthesis driver, adapted from giaDeep.c.
import (
	"fmt"
	"time"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
)

func giaDeepExecute(frame *abcmain.Frame, command string) int32 {
	if abcmain.Abc_FrameIsBatchMode() == 0 {
		abcmain.Abc_FrameSetBatchMode(1)
		defer abcmain.Abc_FrameSetBatchMode(0)
	}
	return abcmain.Cmd_CommandExecute(frame, command)
}

func Gia_ManDeepSynOne(noImprovement, timeout, goal, seed, useTwo, verbose int32) *Gia_Man_t {
	return giaManDeepSynOne(noImprovement, timeout, goal, seed, useTwo, verbose, time.Now)
}

func giaManDeepSynOne(noImprovement, timeout, goal, seed, useTwo, verbose int32, now func() time.Time) *Gia_Man_t {
	frame := abcmain.Abc_FrameGetGlobalFrame()
	initial, ok := abcmain.Abc_FrameReadGia(frame).(*Gia_Man_t)
	if !ok || initial == nil {
		fmt.Fprintln(frame.Err, "gia: no current graph for synthesis")
		return nil
	}
	start := now()
	best := Gia_ManDup(initial)
	last := int32(-1)
	util.Abc_Random(1)
	for s := int32(0); s < 10+seed; s++ {
		util.Abc_Random(0)
	}
	const limit int32 = 100000
	var i int32
	for i = 0; i < limit; i++ {
		random := util.Abc_Random(0)
		optionDch, optionFx := "", ""
		if random&1 != 0 {
			optionDch = " -f"
		}
		if random&4 != 0 {
			optionFx = "; &fx"
		}
		compress := "; &dc2"
		if random&2 != 0 {
			compress = "; &put; compress2rs; &get"
		}
		lut := int32(3) + i%4
		if useTwo != 0 {
			lut = 2 + i%5
		}
		command := fmt.Sprintf("&dch%s; &if -a -K %d; &mfs -e -W 20 -L 20%s%s", optionDch, lut, optionFx, compress)
		status := giaDeepExecute(frame, command)
		if status != 0 {
			fmt.Fprintf(frame.Err, "Something did not work out with the command %q.\n", command)
			Gia_ManStop(best)
			return nil
		}
		current, ok := abcmain.Abc_FrameReadGia(frame).(*Gia_Man_t)
		if !ok || current == nil {
			fmt.Fprintln(frame.Err, "gia: synthesis command removed the graph")
			Gia_ManStop(best)
			return nil
		}
		if Gia_ManAndNum(best) > Gia_ManAndNum(current) {
			Gia_ManStop(best)
			best = Gia_ManDup(current)
			last = i
			if verbose != 0 {
				fmt.Printf("Iter %6d : Time %8.2f sec : And = %6d  Lev = %3d  <== best : %s\n", i, now().Sub(start).Seconds(), Gia_ManAndNum(best), Gia_ManLevelNum(best), command)
			}
		}
		if goal > 0 && Gia_ManAndNum(best) <= goal {
			fmt.Printf("Quality goal (%d nodes <= %d nodes) is achieved after %d iterations and %.2f seconds.\n", Gia_ManAndNum(best), goal, i, now().Sub(start).Seconds())
			break
		}
		if timeout > 0 && now().Sub(start) >= time.Duration(timeout)*time.Second {
			fmt.Printf("Runtime limit (%d sec) is reached after %d iterations.\n", timeout, i)
			break
		}
		if i-last > noImprovement {
			fmt.Printf("Completed %d iterations without improvement in %.2f seconds.\n", noImprovement, now().Sub(start).Seconds())
			break
		}
	}
	if i == limit {
		fmt.Printf("Iteration limit (%d iters) is reached after %.2f seconds.\n", limit, now().Sub(start).Seconds())
	}
	return best
}
func Gia_ManDeepSyn(g *Gia_Man_t, iters, noImprovement, timeout, goal, seed, useTwo, verbose int32) *Gia_Man_t {
	if iters <= 0 {
		return Gia_ManDup(g)
	}
	initial, best := Gia_ManDup(g), Gia_ManDup(g)
	defer Gia_ManStop(initial)
	frame := abcmain.Abc_FrameGetGlobalFrame()
	for i := int32(0); i < iters; i++ {
		abcmain.Abc_FrameUpdateGia(frame, Gia_ManDup(initial))
		next := Gia_ManDeepSynOne(noImprovement, timeout, goal, seed+i, useTwo, verbose)
		if next == nil {
			Gia_ManStop(best)
			return nil
		}
		if Gia_ManAndNum(best) > Gia_ManAndNum(next) {
			Gia_ManStop(best)
			best = next
		} else {
			Gia_ManStop(next)
		}
	}
	return best
}
