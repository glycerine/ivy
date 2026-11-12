package glucose2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
	abcmain "github.com/glycerine/abcgo/base/main"
)

// AIGProvider allows an ABC manager to expose its graph without coupling this
// command package to a concrete manager implementation.
type AIGProvider interface{ Glucose2AIG() *aig.Graph }

func Glucose2_Init(frame *abcmain.Frame) {
	err := frame.RegisterCommand(abcmain.Command{
		Name:  "&glucose2",
		Group: "ABC9",
		Handler: func(f *abcmain.Frame, args []string) int32 {
			return Abc_CommandGlucose2(f, int32(len(args)), args)
		},
	})
	if err != nil {
		panic(err)
	}
}

// Glucose2_End mirrors the source's empty lifecycle hook; the frame owns its
// registered commands.
func Glucose2_End(frame *abcmain.Frame) {}

func Abc_CommandGlucose2(frame *abcmain.Frame, argc int32, argv []string) int32 {
	if argc < 1 || int(argc) > len(argv) {
		return 1
	}
	args := argv[1:argc]
	pars := Glucose_CreatePars(1, 0, 0, 0)
	var filename string
	usage := func() int32 {
		fmt.Fprintf(frame.Err, "usage: &glucose2 [-C num] [-pvh] <file.cnf>\n  -C num : conflict limit [default = %d]\n  -p : toggle preprocessing [default = %d]\n  -v : toggle verbosity [default = %d]\n", pars.ConflictLimit, boolInt(pars.Preprocess), boolInt(pars.Verbose))
		return 1
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			if filename != "" {
				return usage()
			}
			filename = arg
			continue
		}
		for j := 1; j < len(arg); j++ {
			switch arg[j] {
			case 'p':
				pars.Preprocess = !pars.Preprocess
			case 'v':
				pars.Verbose = !pars.Verbose
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
				pars.ConflictLimit = int32(n)
			case 'h':
				return usage()
			default:
				return usage()
			}
		}
	}

	var status int32
	var model []int32
	if filename != "" {
		var err error
		status, model, err = Glucose2_SolveCnf(filename, &pars)
		if err != nil {
			fmt.Fprintln(frame.Err, err)
			return 1
		}
	} else {
		value := frame.Resource("pGia")
		graph, ok := value.(*aig.Graph)
		if !ok {
			if provider, yes := value.(AIGProvider); yes {
				graph = provider.Glucose2AIG()
			}
		}
		if graph == nil {
			fmt.Fprintln(frame.Err, "Abc_CommandGlucose2(): There is no AIG.")
			return 1
		}
		status, model = Glucose2_SolveAig(graph, &pars)
	}

	switch status {
	case 10:
		fmt.Fprintln(frame.Out, "SATISFIABLE")
		if filename == "" {
			cex := &abcmain.Counterexample{PIs: int32(len(model)), Bits: make([]uint64, (len(model)+63)/64)}
			for i, value := range model {
				if value != 0 {
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
