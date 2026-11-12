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
			g.vPacking = nil
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
