package pla

/**CFile****************************************************************

  FileName    [plaCom.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: plaCom.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"path/filepath"
	"strconv"
	"strings"
)

func Pla_AbcGetMan(f *abcmain.Frame) *Pla_Man_t { p, _ := f.Resource("pAbcPla").(*Pla_Man_t); return p }
func Pla_AbcFreeMan(f *abcmain.Frame) {
	if p := Pla_AbcGetMan(f); p != nil {
		Pla_ManFree(p)
	}
	f.SetResource("pAbcPla", nil)
}
func Pla_AbcUpdateMan(f *abcmain.Frame, p *Pla_Man_t) {
	if Pla_AbcGetMan(f) == p {
		return
	}
	Pla_AbcFreeMan(f)
	f.SetResource("pAbcPla", p)
}
func Pla_SetMan(f *abcmain.Frame, p *Pla_Man_t) { Pla_AbcUpdateMan(f, p) }
func Pla_Init(f *abcmain.Frame) {
	commands := []struct {
		name    string
		handler func(*abcmain.Frame, int32, []string) int32
	}{{"|read", Abc_CommandReadPla}, {"|write", Abc_CommandWritePla}, {"|ps", Abc_CommandPs}, {"|gen", Abc_CommandGen}, {"|merge", Abc_CommandMerge}, {"|test", Abc_CommandTest}}
	for _, entry := range commands {
		handler := entry.handler
		if err := f.RegisterCommand(abcmain.Command{Name: entry.name, Group: "Two-level", Handler: func(frame *abcmain.Frame, args []string) int32 { return handler(frame, int32(len(args)), args) }}); err != nil {
			panic(err)
		}
	}
}
func Pla_End(f *abcmain.Frame) { Pla_AbcFreeMan(f) }

type commandOptions struct {
	toggles map[byte]bool
	numbers map[byte]int32
	args    []string
}

func parseCommand(argc int32, argv []string, toggles, numbers string) (commandOptions, error) {
	o := commandOptions{toggles: make(map[byte]bool), numbers: make(map[byte]int32)}
	if argc < 1 || int(argc) > len(argv) {
		return o, fmt.Errorf("invalid argument count")
	}
	for i := 1; i < int(argc); i++ {
		arg := argv[i]
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			o.args = append(o.args, arg)
			continue
		}
		for k := 1; k < len(arg); k++ {
			c := arg[k]
			if c == 'h' {
				return o, fmt.Errorf("help requested")
			}
			if strings.IndexByte(toggles, c) >= 0 {
				o.toggles[c] = !o.toggles[c]
				continue
			}
			if strings.IndexByte(numbers, c) >= 0 {
				var text string
				if k+1 < len(arg) {
					text = arg[k+1:]
					k = len(arg)
				} else {
					i++
					if i >= int(argc) {
						return o, fmt.Errorf("-%c requires an integer", c)
					}
					text = argv[i]
				}
				n, err := strconv.ParseInt(text, 10, 32)
				if err != nil || n < 0 {
					return o, fmt.Errorf("invalid -%c integer", c)
				}
				o.numbers[c] = int32(n)
				continue
			}
			return o, fmt.Errorf("unknown switch -%c", c)
		}
	}
	return o, nil
}
func commandUsage(f *abcmain.Frame, usage string, err error) int32 {
	if err != nil {
		fmt.Fprintln(f.Err, err)
	}
	fmt.Fprintln(f.Err, "usage:", usage)
	return 1
}
func Abc_CommandReadPla(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, err := parseCommand(argc, argv, "v", "")
	if err != nil {
		return commandUsage(f, "|read [-vh] <file_name>", err)
	}
	if len(o.args) != 1 {
		fmt.Fprintln(f.Out, "Input file name should be given on the command line.")
		return 0
	}
	if filepath.Ext(o.args[0]) != ".pla" {
		fmt.Fprintln(f.Out, "Unknown file extension.")
		return 0
	}
	p, err := ReadFile(o.args[0])
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 0
	}
	Pla_AbcUpdateMan(f, p)
	for _, warning := range p.Warnings {
		fmt.Fprintln(f.Out, "Warning:", warning)
	}
	return 0
}
func Abc_CommandWritePla(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, err := parseCommand(argc, argv, "v", "")
	if err != nil {
		return commandUsage(f, "|write [-vh] [file_name]", err)
	}
	p := Pla_AbcGetMan(f)
	if p == nil {
		fmt.Fprintln(f.Out, "There is no current design.")
		return 0
	}
	name := p.Name + "_out.v"
	if len(o.args) == 1 {
		name = o.args[0]
	} else if len(o.args) > 1 {
		fmt.Fprintln(f.Out, "Output file name should be given on the command line.")
		return 0
	}
	if err := Pla_WritePla(p, name); err != nil {
		fmt.Fprintln(f.Err, err)
	}
	return 0
}
func Abc_CommandPs(f *abcmain.Frame, argc int32, argv []string) int32 {
	_, err := parseCommand(argc, argv, "madv", "")
	if err != nil {
		return commandUsage(f, "|ps [-madvh]", err)
	}
	if p := Pla_AbcGetMan(f); p != nil {
		p.PrintStats(f.Out)
	} else {
		fmt.Fprintln(f.Out, "There is no current design.")
	}
	return 0
}
func Abc_CommandGen(f *abcmain.Frame, argc int32, argv []string) int32 {
	o, err := parseCommand(argc, argv, "spv", "IOPS")
	if err != nil {
		return commandUsage(f, "|gen [-IOPS num] [-spvh]", err)
	}
	ins, outs, cubes, seed := int32(8), int32(1), int32(20), int32(0)
	for k, v := range o.numbers {
		switch k {
		case 'I':
			ins = v
		case 'O':
			outs = v
		case 'P':
			cubes = v
		case 'S':
			seed = v
		}
	}
	if ins < 1 || ins > 30 || outs < 1 || outs > 30 || seed > 1000 || int64(cubes) > int64(1)<<uint(ins) || outs > 1 && int64(cubes) > int64(1)<<uint(outs)-1 {
		return commandUsage(f, "|gen [-IOPS num] [-spvh]", fmt.Errorf("invalid generation dimensions or seed"))
	}
	var p *Pla_Man_t
	if o.toggles['s'] {
		if err := Pla_GenSorter(ins); err != nil {
			fmt.Fprintln(f.Err, err)
			return 0
		}
	} else if o.toggles['p'] {
		p = Pla_ManPrimesDetector(ins)
	} else {
		Pla_SeedRandom(seed)
		p = Pla_ManGenerate(ins, outs, cubes, boolInt(o.toggles['v']))
	}
	Pla_AbcUpdateMan(f, p)
	return 0
}
func Abc_CommandMerge(f *abcmain.Frame, argc int32, argv []string) int32 {
	_, err := parseCommand(argc, argv, "mv", "")
	if err != nil {
		return commandUsage(f, "|merge [-mvh]", err)
	}
	if p := Pla_AbcGetMan(f); p != nil {
		Pla_ManDist1Merge(p)
	} else {
		fmt.Fprintln(f.Out, "There is no current design.")
	}
	return 0
}

// |test has no active transformation in the C source (plaCom.c:486-490).
func Abc_CommandTest(f *abcmain.Frame, argc int32, argv []string) int32 {
	_, err := parseCommand(argc, argv, "v", "N")
	if err != nil {
		return commandUsage(f, "|test [-N num] [-vh]", err)
	}
	return 0
}
