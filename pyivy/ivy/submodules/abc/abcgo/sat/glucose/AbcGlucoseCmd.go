package glucose

/**CFile****************************************************************

  FileName    [AbcGlucoseCmd.cpp]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver Glucose 3.0 by Gilles Audemard and Laurent Simon.]

  Synopsis    [Interface to Glucose.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 6, 2017.]

  Revision    [$Id: AbcGlucoseCmd.cpp,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	abcmain "github.com/glycerine/abcgo/base/main"
	"strconv"
	"strings"
)

// AIGProvider adapts another manager without importing unfinished GIA packages.
type AIGProvider interface{ GlucoseAIG() *aig.Graph }

func Glucose_Init(frame *abcmain.Frame) {
	if err := frame.RegisterCommand(abcmain.Command{Name: "&glucose", Group: "ABC9", Handler: func(f *abcmain.Frame, args []string) int32 { return Abc_CommandGlucose(f, int32(len(args)), args) }}); err != nil {
		panic(err)
	}
}

// Glucose_End matches the source's empty lifecycle hook; the frame owns commands.
func Glucose_End(frame *abcmain.Frame) {}
func Abc_CommandGlucose(frame *abcmain.Frame, argc int32, argv []string) int32 {
	if argc < 1 || int(argc) > len(argv) {
		return 1
	}
	args := argv[1:argc]
	pars := Glucose_CreatePars(1, 0, 0, 0)
	var file string
	usage := func() int32 {
		fmt.Fprintf(frame.Err, "usage: &glucose [-C num] [-pvh] <file.cnf>\n  -C num : conflict limit [default = %d]\n  -p : toggle preprocessing [default = %d]\n  -v : toggle verbosity [default = %d]\n", pars.NConfls, pars.Pre, pars.Verb)
		return 1
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			if file != "" {
				return usage()
			}
			file = arg
			continue
		}
		for j := 1; j < len(arg); j++ {
			switch arg[j] {
			case 'p':
				pars.Pre ^= 1
			case 'v':
				pars.Verb ^= 1
			case 'C':
				var value string
				if j+1 < len(arg) {
					value = arg[j+1:]
					j = len(arg)
				} else {
					i++
					if i >= len(args) {
						return usage()
					}
					value = args[i]
				}
				n, err := strconv.ParseInt(value, 10, 32)
				if err != nil || n < 0 {
					return usage()
				}
				pars.NConfls = int32(n)
			case 'h':
				return usage()
			default:
				return usage()
			}
		}
	}
	var status int32
	var model []int32
	if file != "" {
		var err error
		status, model, err = Glucose_SolveCnf(file, &pars)
		if err != nil {
			fmt.Fprintln(frame.Err, err)
			return 1
		}
	} else {
		value := frame.Resource("pGia")
		p, ok := value.(*aig.Graph)
		if !ok {
			if adapter, yes := value.(AIGProvider); yes {
				p = adapter.GlucoseAIG()
			}
		}
		if p == nil {
			fmt.Fprintln(frame.Err, "Abc_CommandGlucose(): There is no AIG.")
			return 1
		}
		status, model = Glucose_SolveAig(p, &pars)
	}
	switch status {
	case 10:
		fmt.Fprintln(frame.Out, "SATISFIABLE")
		if file == "" {
			cex := &abcmain.Counterexample{PIs: int32(len(model)), Bits: make([]uint64, (len(model)+63)/64)}
			for i, v := range model {
				if v != 0 {
					cex.Bits[i/64] |= uint64(1) << uint(i&63)
				}
			}
			abcmain.Abc_FrameReplaceCex(frame, &cex)
		}
	case 20:
		fmt.Fprintln(frame.Out, "UNSATISFIABLE")
	default:
		fmt.Fprintln(frame.Out, "INDETERMINATE")
	}
	return 0
}
