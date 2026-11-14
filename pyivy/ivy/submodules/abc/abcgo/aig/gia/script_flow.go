package gia

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func Gia_ManPerformMap(ands, lutSize, cutNum, minAve, useMfs, verbose int32) {
	frame := abcmain.Abc_FrameGetGlobalFrame()
	run := func(s string) { frame.Execute(s) }
	opt := ""
	if minAve != 0 {
		opt = "-t"
	}
	run(fmt.Sprintf("&unmap; &lf -K %d -C %d -k %s; &save", lutSize, cutNum, opt))
	if verbose != 0 {
		fmt.Println("MAPPING:\nMapping with &lf -k:")
		run("&ps")
	}
	run(fmt.Sprintf("&unmap; &lf -K %d -C %d %s; &save", lutSize, cutNum, opt))
	if verbose != 0 {
		fmt.Println("Mapping with &lf:")
		run("&ps")
	}
	if (lutSize == 4 && ands < 100000) || (lutSize == 6 && ands < 2000) {
		command := fmt.Sprintf("&unmap; &if -sz -S %d%d -K %d -C %d %s", lutSize, lutSize, 2*lutSize-1, 2*cutNum, opt)
		run(command)
		if g, ok := abcmain.Abc_FrameReadGia(frame).(*Gia_Man_t); ok && g != nil {
			g.VPacking = nil
		}
		run("&save")
		if verbose != 0 {
			fmt.Printf("Mapping with %s:\n", command)
			run("&ps")
		}
	}
	run("&load")
	if useMfs != 0 {
		run("&put; mfs2 -W 4 -M 500 -C 7000; &get -m")
	}
	if verbose != 0 {
		fmt.Println("Mapping final:")
		run("&ps")
	}
}
func Gia_ManPerformRound(mapped, ands, levels, lutSize, cutNum, minAve, useMfs, verbose int32) {
	_, _ = mapped, levels
	frame := abcmain.Abc_FrameGetGlobalFrame()
	run := func(s string) { frame.Execute(s) }
	opt := ""
	if minAve != 0 {
		opt = "-t"
	}
	for _, limit := range []int32{50000, 20000} {
		if ands < limit {
			run("")
			run(fmt.Sprintf("&dsdb; &dch -C 500; &if -K %d -C %d %s; &save", lutSize, cutNum, opt))
			if verbose != 0 {
				fmt.Printf("Mapping with &dch -C 500; &if -K %d -C %d %s:\n", lutSize, cutNum, opt)
				run("&ps")
			}
			run("&st")
		}
	}
	Gia_ManPerformMap(ands, lutSize, cutNum, minAve, useMfs, verbose)
	run("&st")
	run("&dsdb")
	Gia_ManPerformMap(ands, lutSize, cutNum, minAve, useMfs, verbose)
	run("&st")
	run("&syn2 -m -R 10; &dsdb")
	run(fmt.Sprintf("&blut -a -K %d", lutSize))
	Gia_ManPerformMap(ands, lutSize, cutNum, minAve, useMfs, verbose)
}
func Gia_ManPerformFlow(mapped, ands, levels, lutSize, cutNum, minAve, useMfs, verbose int32) {
	frame := abcmain.Abc_FrameGetGlobalFrame()
	if mapped != 0 {
		frame.Execute("&st")
	}
	Gia_ManPerformRound(mapped, ands, levels, lutSize, cutNum, minAve, useMfs, verbose)
	frame.Execute("&st; &sopb")
	Gia_ManPerformRound(mapped, ands, levels, lutSize, cutNum, minAve, useMfs, verbose)
}

func Gia_ManPerformFlow2(mapped, ands, levels, lutSize, cutNum, balance, minAve, useMfs, verbose int32) {
	_ = levels
	frame := abcmain.Abc_FrameGetGlobalFrame()
	run := func(s string) { frame.Execute(s) }
	suffix := ""
	if minAve != 0 {
		suffix = "t"
	}
	mfs := ""
	if useMfs != 0 {
		mfs = "&put; mfs2 -W 4 -M 500 -C 7000; &get -m;"
	}
	syncSize := lutSize
	if syncSize == 0 {
		syncSize = 6
	}
	kopt := ""
	if lutSize != 0 {
		kopt = fmt.Sprintf("-K %d ", lutSize)
	}
	c1 := fmt.Sprintf("&synch2 -K %d -C 500;  &if -m%s %s-C %d; %s &save", syncSize, suffix, kopt, cutNum, mfs)
	c2 := fmt.Sprintf("&dch -C 500;          &if -m%s %s-C %d; %s &save", suffix, kopt, cutNum+4, mfs)
	c3 := fmt.Sprintf("&synch2 -K %d -C 500;  &lf -m%s -E 5 %s-C %d; %s &save", syncSize, suffix, kopt, cutNum, mfs)
	if verbose != 0 {
		fmt.Println("Trying synthesis...")
	}
	if mapped != 0 {
		run("&st")
	}
	run(c1)
	if verbose != 0 {
		run("&ps")
	}
	run("&st")
	run(c2)
	if verbose != 0 {
		run("&ps")
	}
	run("&load")
	if verbose != 0 {
		run("&ps")
	}
	stages := []struct {
		limit   int32
		command string
	}{{0, "&st; &sopb -R 10 -C 4"}, {100000, "&st; &sopb -R 10"}, {50000, "&st; &sopb -R 10"}}
	for _, stage := range stages {
		if stage.limit != 0 && ands > stage.limit {
			break
		}
		if balance != 0 {
			if verbose != 0 {
				fmt.Println("Trying SOP balancing...")
			}
			run(stage.command)
		}
		run(c3)
		if verbose != 0 {
			run("&ps")
		}
		run("&st")
		run(c2)
		if verbose != 0 {
			run("&ps")
		}
		run("&load")
		if verbose != 0 {
			run("&ps")
		}
	}
}

// giaFlow3Commands builds the three command sequences used by the flow.
func giaFlow3Commands(lutSize, cutNum int32, useMfs, useLutLib, recRunning bool) [3]string {
	mode := "g"
	if recRunning {
		mode = "y"
	}
	mfs := ""
	if useMfs {
		mfs = "&mfs; "
	}
	if useLutLib {
		return [3]string{
			fmt.Sprintf("&st; &if -C %d;       &save; &st; &syn2; &if -C %d;       &save; &load", cutNum, cutNum),
			fmt.Sprintf("&st; &if -%s -K 6; &dch -f; &if -C %d;       %s&save; &load", mode, cutNum, mfs),
			fmt.Sprintf("&st; &if -%s -K 6; &synch2; &if -C %d;       %s&save; &load", mode, cutNum, mfs),
		}
	}
	return [3]string{
		fmt.Sprintf("&st; &if -C %d -K %d; &save; &st; &syn2; &if -C %d -K %d; &save; &load", cutNum, lutSize, cutNum, lutSize),
		fmt.Sprintf("&st; &if -%s -K 6; &dch -f; &if -C %d -K %d; %s&save; &load", mode, cutNum, lutSize, mfs),
		fmt.Sprintf("&st; &if -%s -K 6; &synch2; &if -C %d -K %d; %s&save; &load", mode, cutNum, lutSize, mfs),
	}
}

// giaPerformFlow3 executes each stage against the current frame graph, which
// may change after any command. The recorder state comes from the LMS owner.
func giaPerformFlow3(frame *abcmain.Frame, lutSize, cutNum int32, useMfs, useLutLib, recRunning, verbose bool) {
	commands := giaFlow3Commands(lutSize, cutNum, useMfs, useLutLib, recRunning)
	if verbose {
		mode := "SOP balancing"
		if recRunning {
			mode = "LMS"
		}
		fmt.Printf("Trying simple synthesis with %s...\n", mode)
	}
	frame.Execute(commands[0])
	if verbose {
		frame.Execute("&ps")
	}
	for i, stage := range []struct {
		limit int32
		name  string
	}{{200000, "medium"}, {10000, "harder"}} {
		g, ok := abcmain.Abc_FrameReadGia(frame).(*Gia_Man_t)
		if !ok || g == nil {
			panic("gia: synthesis flow requires a current GIA")
		}
		if Gia_ManAndNum(g) < stage.limit {
			if verbose {
				fmt.Printf("Trying %s synthesis...\n", stage.name)
			}
			frame.Execute(commands[i+1])
			if verbose {
				frame.Execute("&ps")
			}
		}
	}
	if verbose {
		fmt.Println("Final result...")
		frame.Execute("&ps")
	}
}
