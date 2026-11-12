package abcmain

// Run is the importable equivalent of main.c's thin process entry point.
// The calling executable decides whether to pass its result to os.Exit.
func Run(args []string) int { return int(Abc_RealMain(args)) }
