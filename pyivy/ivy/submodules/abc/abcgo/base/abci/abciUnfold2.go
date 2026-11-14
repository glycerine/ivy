package abci

// parseUnfold2Options mirrors the active switches in abciUnfold2.c. The
// source's -F option is commented out, so nFrames remains fixed at one.
func parseUnfold2Options(argc int32, argv []string) (nConfs, nProps, verbose int32, ok bool) {
	nConfs, nProps = 1000, 1000
	Extra_UtilGetoptReset()
	for c := Extra_UtilGetopt(argc, argv, "CPvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "CPvh") {
		switch c {
		case 'C':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-C\" should be followed by an integer.\n")
				return nConfs, nProps, verbose, false
			}
			nConfs = abcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if nConfs < 0 {
				return nConfs, nProps, verbose, false
			}
		case 'P':
			if globalUtilOptind >= argc {
				Abc_Print(-1, "Command line switch \"-P\" should be followed by an integer.\n")
				return nConfs, nProps, verbose, false
			}
			nProps = abcAtoi(argv[globalUtilOptind])
			globalUtilOptind++
			if nProps < 0 {
				return nConfs, nProps, verbose, false
			}
		case 'v':
			verbose ^= 1
		case 'h':
			return nConfs, nProps, verbose, false
		default:
			return nConfs, nProps, verbose, false
		}
	}
	return nConfs, nProps, verbose, true
}

func printUnfold2Usage(nConfs, nProps, verbose int32) {
	Abc_Print(-2, "usage: unfold2 [-FCP num] [-savh]\n")
	Abc_Print(-2, "\t         unfold hidden constraints as separate outputs\n")
	Abc_Print(-2, "\t-C num : the max number of conflicts in SAT solving [default = %d]\n", nConfs)
	Abc_Print(-2, "\t-P num : the max number of constraint propagations [default = %d]\n", nProps)
	label := "no"
	if verbose != 0 {
		label = "yes"
	}
	Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", label)
	Abc_Print(-2, "\t-h     : print the command usage\n")
}

// Abc_CommandUnfold2 extracts hidden constraints from a sequential strashed
// network and replaces the frame's current network on success.
func Abc_CommandUnfold2(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	nConfs, nProps, verbose, ok := parseUnfold2Options(argc, argv)
	if !ok {
		printUnfold2Usage(nConfs, nProps, verbose)
		return 1
	}
	pNtk := Abc_FrameReadNtk(pAbc)
	if pNtk == nil {
		Abc_Print(-1, "Empty network.\n")
		return 1
	}
	if Abc_NtkIsComb(pNtk) != 0 {
		Abc_Print(-1, "The network is combinational.\n")
		return 0
	}
	if Abc_NtkIsStrash(pNtk) == 0 {
		Abc_Print(-1, "Currently only works for structurally hashed circuits.\n")
		return 0
	}
	if Abc_NtkConstrNum(pNtk) > 0 {
		Abc_Print(-1, "Constraints are already extracted.\n")
		return 0
	}
	if Abc_NtkPoNum(pNtk) > 1 {
		Abc_Print(-1, "Functional constraint extraction works for single-output miters (use \"orpos\").\n")
		return 0
	}
	pNtkRes := Abc_NtkDarUnfold2(pNtk, 1, nConfs, nProps, 0, 0, verbose)
	if pNtkRes == nil {
		Abc_Print(1, "Transformation has failed.\n")
		return 0
	}
	Abc_FrameReplaceCurrentNetwork(pAbc, pNtkRes)
	return 0
}

func parseFold2Options(argc int32, argv []string) (verbose int32, ok bool) {
	Extra_UtilGetoptReset()
	for c := Extra_UtilGetopt(argc, argv, "cvh"); c != -1; c = Extra_UtilGetopt(argc, argv, "cvh") {
		switch c {
		case 'v':
			verbose ^= 1
		case 'h':
			return verbose, false
		default:
			return verbose, false
		}
	}
	return verbose, true
}

func Abc_CommandFold2(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	verbose, ok := parseFold2Options(argc, argv)
	if !ok {
		Abc_Print(-2, "usage: fold [-cvh]\n")
		Abc_Print(-2, "\t         folds constraints represented as separate outputs\n")
		Abc_Print(-2, "\t-v     : toggle printing verbose information [default = %s]\n", map[bool]string{true: "yes", false: "no"}[verbose != 0])
		Abc_Print(-2, "\t-h     : print the command usage\n")
		return 1
	}
	pNtk := Abc_FrameReadNtk(pAbc)
	if pNtk == nil {
		Abc_Print(-1, "Empty network.\n")
		return 1
	}
	if Abc_NtkIsStrash(pNtk) == 0 {
		Abc_Print(-1, "Currently only works for structurally hashed circuits.\n")
		return 0
	}
	if Abc_NtkConstrNum(pNtk) == 0 {
		Abc_Print(0, "The network has no constraints.\n")
		return 0
	}
	if Abc_NtkIsComb(pNtk) != 0 {
		Abc_Print(0, "The network is combinational.\n")
	}
	pNtkRes := Abc_NtkDarFold2(pNtk, 0, verbose, 0)
	if pNtkRes == nil {
		Abc_Print(1, "Transformation has failed.\n")
		return 0
	}
	Abc_FrameReplaceCurrentNetwork(pAbc, pNtkRes)
	return 0
}
