package cudd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
	"unsafe"
)

const TESTCUDD_VERSION = "TestCudd Version #1.0, Release date 3/17/01"

var onames = []string{"C", "M"}

type testCuddOptions struct {
	cmu, autodyn, harwell, multiple, profile, keepperm, clearcache, blif, memoryRecord int
	approach, groupcheck, n, pr, nvars, slots, cache                                   int
	maxMemory                                                                          uint64
	dump, file                                                                         string
}

func parseTestCuddArgs(args []string, stderr io.Writer) (testCuddOptions, error) {
	o := testCuddOptions{approach: int(CUDD_REORDER_NONE), groupcheck: int(CUDD_GROUP_CHECK7), n: 4, nvars: 4, slots: CUDD_UNIQUE_SLOTS, cache: 127, file: "-"}
	fs := flag.NewFlagSet("testcudd", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { testCuddUsage(stderr) }
	fs.Var((*boolFlag)(&o.cmu), "C", "use CMU multiplication algorithm")
	fs.Var((*boolFlag)(&o.autodyn), "D", "enable automatic dynamic reordering")
	fs.Var((*boolFlag)(&o.harwell), "H", "read matrix in Harwell format")
	fs.Var((*boolFlag)(&o.memoryRecord), "M", "turn off memory allocation recording")
	fs.Var((*boolFlag)(&o.profile), "P", "print BDD heap profile")
	fs.Var((*boolFlag)(&o.multiple), "m", "read multiple matrices")
	fs.Var((*boolFlag)(&o.keepperm), "k", "print the variable permutation")
	fs.Var((*boolFlag)(&o.clearcache), "c", "clear the cache after each matrix")
	fs.Var((*boolFlag)(&o.blif), "b", "use blif as format for dumps")
	fs.IntVar(&o.slots, "S", o.slots, "number of slots for each subtable")
	fs.Uint64Var(&o.maxMemory, "X", 0, "target maximum memory in bytes")
	fs.IntVar(&o.approach, "a", o.approach, "reordering approach (0-17)")
	fs.IntVar(&o.groupcheck, "g", o.groupcheck, "aggregation criterion (0,5,7)")
	fs.IntVar(&o.n, "n", o.n, "number of variables for Walsh matrices")
	fs.IntVar(&o.pr, "p", 0, "verbosity")
	fs.IntVar(&o.nvars, "v", o.nvars, "initial variables in unique table")
	fs.IntVar(&o.cache, "x", o.cache, "initial cache size")
	fs.StringVar(&o.dump, "d", "", "dump DDs to file")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	if fs.NArg() > 1 {
		return o, fmt.Errorf("expected at most one input file")
	}
	if fs.NArg() == 1 {
		o.file = fs.Arg(0)
	}
	if o.approach < 0 || o.approach > 17 {
		return o, fmt.Errorf("invalid approach: %d", o.approach)
	}
	if o.slots < 0 || o.cache < 0 || o.n < 0 || o.nvars < 0 {
		return o, fmt.Errorf("sizes must be non-negative")
	}
	return o, nil
}

type boolFlag int

func (b *boolFlag) Set(value string) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	if v {
		*b = 1
	} else {
		*b = 0
	}
	return nil
}
func (b *boolFlag) String() string   { return strconv.Itoa(int(*b)) }
func (b *boolFlag) IsBoolFlag() bool { return true }

func main() { os.Exit(runTestCudd(os.Args[1:], os.Stdout, os.Stderr)) }

func testCuddUsage(out io.Writer) {
	fmt.Fprintln(out, "usage: testcudd [options] [file]")
	fmt.Fprintln(out, "   -C\t\tuse CMU multiplication algorithm")
	fmt.Fprintln(out, "   -D\t\tenable automatic dynamic reordering")
	fmt.Fprintln(out, "   -H\t\tread matrix in Harwell format")
	fmt.Fprintln(out, "   -M\t\tturns off memory allocation recording")
	fmt.Fprintln(out, "   -P\t\tprint BDD heap profile")
	fmt.Fprintln(out, "   -S n\t\tnumber of slots for each subtable")
	fmt.Fprintln(out, "   -X n\t\ttarget maximum memory in bytes")
	fmt.Fprintln(out, "   -a n\t\tchoose reordering approach (0-17)")
	fmt.Fprintln(out, "   \t\t\t0: same as autoMethod; 1: none; 2: random; 3: pivot")
	fmt.Fprintln(out, "   \t\t\t4: sifting; 5: sifting to convergence; 6-7: symmetric sifting")
	fmt.Fprintln(out, "   \t\t\t8-10: window of size 2-4; 11-13: window to convergence")
	fmt.Fprintln(out, "   \t\t\t14-15: group sifting; 16: simulated annealing; 17: genetic algorithm")
	fmt.Fprintln(out, "   -b\t\tuse blif as format for dumps")
	fmt.Fprintln(out, "   -c\t\tclear the cache after each matrix")
	fmt.Fprintln(out, "   -d file\tdump DDs to file")
	fmt.Fprintln(out, "   -g n\t\tselect aggregation criterion (0,5,7)")
	fmt.Fprintln(out, "   -h\t\tprints this message")
	fmt.Fprintln(out, "   -k\t\tprint the variable permutation")
	fmt.Fprintln(out, "   -m\t\tread multiple matrices (only with -H)")
	fmt.Fprintln(out, "   -n n\t\tnumber of variables")
	fmt.Fprintln(out, "   -p n\t\tcontrol verbosity")
	fmt.Fprintln(out, "   -v n\t\tinitial variables in the unique table")
	fmt.Fprintln(out, "   -x n\t\tinitial size of the cache")
}

func runTestCudd(args []string, stdout, stderr *os.File) int {
	started := time.Now()
	o, err := parseTestCuddArgs(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 2
		}
		fmt.Fprintln(stderr, err)
		return 2
	}
	if o.pr >= 0 {
		fmt.Fprintf(stdout, "# %s\n# testcudd", TESTCUDD_VERSION)
		for _, arg := range append([]string{"testcudd"}, args...) {
			fmt.Fprintf(stdout, " %s", arg)
		}
		fmt.Fprintln(stdout)
	}
	dd := Cudd_Init(uint32(o.nvars), 0, uint32(o.slots), uint32(o.cache), o.maxMemory)
	if dd == nil {
		fmt.Fprintln(stderr, "failed to initialize CUDD manager")
		return 1
	}
	Cudd_SetGroupcheck(dd, Cudd_AggregationType(o.groupcheck))
	if o.autodyn != 0 {
		Cudd_AutodynEnable(dd, CUDD_REORDER_SAME)
	}
	input, err := openTestFile(o.file, "r", os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(stderr, err)
		Cudd_Quit(dd)
		return 1
	}
	defer func() {
		if input != os.Stdin {
			_ = input.Close()
		}
	}()
	var dump *os.File
	if o.dump != "" {
		dump, err = os.Create(o.dump)
		if err != nil {
			fmt.Fprintln(stderr, err)
			Cudd_Quit(dd)
			return 1
		}
		defer dump.Close()
	}
	result := runMatrixTests(dd, input, dump, o, stdout, stderr)
	if result == 0 {
		Cudd_Quit(dd)
		return 2
	}
	if testWalsh(dd, int32(o.n), int32(o.cmu), Cudd_ReorderingType(o.approach), int32(o.pr), stdout, stderr) == 0 {
		Cudd_Quit(dd)
		return 2
	}
	if cuddDestroySubtables(dd, 3) == 0 || Cudd_DebugCheck(dd) != 0 || Cudd_CheckKeys(dd) != 0 {
		Cudd_Quit(dd)
		return 2
	}
	leaks := Cudd_CheckZeroRef(dd)
	if leaks != 0 {
		fmt.Fprintf(stderr, "%d non-zero DD reference counts after dereferencing\n", leaks)
	}
	if o.pr >= 0 {
		Cudd_PrintInfo(dd, stdout)
	}
	Cudd_Quit(dd)
	if o.pr > 0 {
		fmt.Fprintf(stdout, "total time = %s\n", time.Since(started))
	}
	if leaks != 0 {
		return 1
	}
	return 0
}

func openTestFile(name, mode string, stdin, stdout *os.File) (*os.File, error) {
	if name == "-" {
		if mode[0] == 'r' {
			return stdin, nil
		}
		return stdout, nil
	}
	if mode[0] == 'r' {
		return os.Open(name)
	}
	return os.Create(name)
}

func runMatrixTests(dd *DdManager, input io.Reader, dump *os.File, o testCuddOptions, stdout, stderr io.Writer) int {
	var M *DdNode
	var x, y, xn, yn []*DdNode
	var nx, ny, rows, cols int32
	if o.harwell != 0 {
		err := forEachHarwellMatrix(input, dd, o, stdout, func(hb *HarwellMatrix) bool {
			M, x, y, xn, yn = hb.Function, hb.X, hb.Y, hb.XNeg, hb.YNeg
			nx, ny, rows, cols = int32(len(x)), int32(len(y)), int32(hb.Rows), int32(hb.Cols)
			return M != nil && runMatrix(dd, M, x, y, xn, yn, nx, ny, o, dump, stdout, stderr) != 0
		})
		if err != nil {
			fmt.Fprintf(stderr, "Error reading matrix: %v\n", err)
			return 0
		}
		return 1
	} else {
		if Cudd_addRead(input, dd, &M, &x, &y, &xn, &yn, &nx, &ny, &rows, &cols, 0, 2, 1, 2) == 0 {
			fmt.Fprintln(stderr, "Error reading matrix")
			return 0
		}
		if o.pr >= 0 {
			fmt.Fprintf(stdout, ":name: %s: %d rows %d columns\n", o.file, rows, cols)
		}
	}
	if M == nil {
		fmt.Fprintln(stderr, "matrix reader returned a nil ADD")
		return 0
	}
	return runMatrix(dd, M, x, y, xn, yn, nx, ny, o, dump, stdout, stderr)
}

// forEachHarwellMatrix implements testcudd's do/while read loop. Each read
// starts with empty variable vectors; the callback owns and releases the
// matrix function and variable references before the next read begins.
func forEachHarwellMatrix(input io.Reader, dd *DdManager, o testCuddOptions, stdout io.Writer, visit func(*HarwellMatrix) bool) error {
	var verbose io.Writer
	if o.pr > 0 {
		verbose = stdout
	}
	count := 0
	for {
		if o.pr >= 0 {
			fmt.Fprint(stdout, ":name: ")
		}
		matrix, more, err := Cudd_addHarwellNext(input, dd, 0, 2, 1, 2, verbose)
		if err != nil {
			return err
		}
		if !more {
			if count == 0 {
				return fmt.Errorf("empty Harwell input")
			}
			return nil
		}
		if !visit(matrix) {
			return fmt.Errorf("matrix processing failed")
		}
		count++
		if o.multiple == 0 {
			return nil
		}
	}
}

func runMatrix(dd *DdManager, M *DdNode, x, y, xn, yn []*DdNode, nx, ny int32, o testCuddOptions, dump *os.File, stdout, stderr io.Writer) int {
	defer Cudd_RecursiveDeref(dd, M)
	start := time.Now()
	maxny := ny
	ycube := Cudd_ReadOne(dd)
	Cudd_Ref(ycube)
	for i := maxny - 1; i >= 0; i-- {
		tmp := Cudd_bddAnd(dd, Cudd_Not(Cudd_ReadVars(dd, int32(Cudd_NodeReadIndex(y[i])))), ycube)
		if tmp == nil {
			return 0
		}
		Cudd_Ref(tmp)
		Cudd_RecursiveDeref(dd, ycube)
		ycube = tmp
	}
	defer Cudd_RecursiveDeref(dd, ycube)
	xvars, yvars := make([]*DdNode, nx), make([]*DdNode, ny)
	for i := range xvars {
		xvars[i] = Cudd_ReadVars(dd, int32(Cudd_NodeReadIndex(x[i])))
	}
	for i := range yvars {
		yvars[i] = Cudd_ReadVars(dd, int32(Cudd_NodeReadIndex(y[i])))
	}
	for _, n := range append(append([]*DdNode(nil), x...), xn...) {
		Cudd_RecursiveDeref(dd, n)
	}
	for _, n := range append(append([]*DdNode(nil), y...), yn...) {
		Cudd_RecursiveDeref(dd, n)
	}
	if o.pr > 0 {
		fmt.Fprint(stdout, ":1: M")
		Cudd_PrintDebug(dd, M, nx+ny, int32(o.pr))
	}
	if o.pr > 0 {
		fmt.Fprintf(stdout, ":2: time to read the matrix = %s\n", time.Since(start))
	}
	C := Cudd_addBddPattern(dd, M)
	if C == nil {
		return 0
	}
	Cudd_Ref(C)
	defer Cudd_RecursiveDeref(dd, C)
	if o.pr > 0 {
		fmt.Fprint(stdout, ":3: C")
		Cudd_PrintDebug(dd, C, nx+ny, int32(o.pr))
	}
	if testIterators(dd, M, C, int32(o.pr), stdout) == 0 {
		return 0
	}
	cuddCacheProfile(dd, os.Stdout)
	if testXor(dd, C, int32(o.pr), nx+ny, stdout) == 0 || testHamming(dd, C, int32(o.pr), stdout) == 0 {
		return 0
	}
	CP := Cudd_CProjection(dd, C, ycube)
	if CP == nil {
		return 0
	}
	Cudd_Ref(CP)
	defer Cudd_RecursiveDeref(dd, CP)
	if o.pr > 0 {
		fmt.Fprint(stdout, "ycube")
		Cudd_PrintDebug(dd, ycube, nx+ny, int32(o.pr))
		fmt.Fprint(stdout, "CP")
		Cudd_PrintDebug(dd, CP, nx+ny, int32(o.pr))
	}
	if nx == ny {
		cpr := Cudd_PrioritySelect(dd, C, xvars, yvars, nil, nil, int(ny), Cudd_Xgty)
		if cpr == nil {
			return 0
		}
		Cudd_Ref(cpr)
		if o.pr > 0 {
			fmt.Fprint(stdout, ":4: CPr")
			Cudd_PrintDebug(dd, cpr, nx+ny, int32(o.pr))
		}
		if CP != cpr {
			fmt.Fprintln(stdout, "CP != CPr!")
		}
		Cudd_RecursiveDeref(dd, cpr)
	}
	f := Cudd_Inequality(dd, int(min32(nx, ny)), 2, xvars, yvars)
	if f == nil {
		return 0
	}
	Cudd_Ref(f)
	if o.pr > 0 {
		fmt.Fprint(stdout, ":4: ineq")
		Cudd_PrintDebug(dd, f, nx+ny, int32(o.pr))
		if o.pr > 1 {
			printPrimeCubes(dd, Cudd_Not(f), stdout)
		}
	}
	Cudd_IterDerefBdd(dd, f)
	ess := Cudd_FindEssential(dd, C)
	if ess == nil {
		return 0
	}
	Cudd_Ref(ess)
	if o.pr > 0 {
		fmt.Fprint(stdout, ":4: ess")
		Cudd_PrintDebug(dd, ess, nx+ny, int32(o.pr))
	}
	Cudd_RecursiveDeref(dd, ess)
	var length int32
	short := Cudd_ShortestPath(dd, M, nil, nil, &length)
	if short == nil {
		return 0
	}
	Cudd_Ref(short)
	if o.pr > 0 {
		fmt.Fprint(stdout, ":5: shortP")
		Cudd_PrintDebug(dd, short, nx+ny, int32(o.pr))
	}
	largest := Cudd_LargestCube(dd, Cudd_Not(C), &length)
	if largest == nil {
		return 0
	}
	Cudd_Ref(largest)
	if o.pr > 0 {
		fmt.Fprint(stdout, ":5b: largest")
		Cudd_PrintDebug(dd, largest, nx+ny, int32(o.pr))
	}
	Cudd_RecursiveDeref(dd, largest)
	shortA := Cudd_BddToAdd(dd, short)
	if shortA == nil {
		Cudd_RecursiveDeref(dd, short)
		return 0
	}
	Cudd_Ref(shortA)
	Cudd_RecursiveDeref(dd, short)
	constN := Cudd_addEvalConst(dd, shortA, M)
	if constN == DD_NON_CONSTANT || Cudd_addIteConstant(dd, shortA, M, constN) != constN {
		Cudd_RecursiveDeref(dd, shortA)
		return 0
	}
	if o.pr > 0 {
		fmt.Fprintf(stdout, "The value of M along the chosen shortest path is %g\n", Cudd_V(constN))
	}
	Cudd_RecursiveDeref(dd, shortA)
	short = Cudd_ShortestPath(dd, C, nil, nil, &length)
	if short == nil {
		return 0
	}
	Cudd_Ref(short)
	if o.pr > 0 {
		fmt.Fprint(stdout, ":6: shortP")
		Cudd_PrintDebug(dd, short, nx+ny, int32(o.pr))
	}
	if !Cudd_bddLeq(dd, short, C) || Cudd_bddIteConstant(dd, Cudd_Not(short), Cudd_ReadOne(dd), C) != Cudd_ReadOne(dd) {
		Cudd_RecursiveDeref(dd, short)
		return 0
	}
	Cudd_RecursiveDeref(dd, short)
	if o.profile != 0 {
		cuddHeapProfile(dd)
	}
	size := Cudd_ReadSize(dd)
	if o.pr > 0 {
		fmt.Fprintf(stdout, "Average distance: %g\n", Cudd_AverageDistance(dd))
	}
	if Cudd_ReorderingType(o.approach) != CUDD_REORDER_NONE {
		if Cudd_EnableReorderingReporting(dd) == 0 {
			fmt.Fprintln(stderr, "Error reported by Cudd_EnableReorderingReporting")
			return 0
		}
		if Cudd_ReduceHeap(dd, Cudd_ReorderingType(o.approach), 5) == 0 {
			fmt.Fprintln(stderr, "Error reported by Cudd_ReduceHeap")
			return 0
		}
		if Cudd_DisableReorderingReporting(dd) == 0 {
			fmt.Fprintln(stderr, "Error reported by Cudd_DisableReorderingReporting")
			return 0
		}
		if Cudd_ReorderingType(o.approach) == CUDD_REORDER_SYMM_SIFT || Cudd_ReorderingType(o.approach) == CUDD_REORDER_SYMM_SIFT_CONV {
			Cudd_SymmProfile(dd, 0, Cudd_ReadSize(dd)-1)
		}
		if o.pr > 0 {
			fmt.Fprintf(stdout, "Average distance: %g\n", Cudd_AverageDistance(dd))
		}
		if o.keepperm != 0 {
			printPermutation(dd, size, stdout)
		}
		if o.pr > 0 {
			fmt.Fprint(stdout, "M")
			Cudd_PrintDebug(dd, M, nx+ny, int32(o.pr))
		}
		if o.profile != 0 {
			cuddHeapProfile(dd)
		}
	}
	if dump != nil {
		var status int32
		if o.blif != 0 {
			status = Cudd_DumpBlif(dd, []*DdNode{C}, nil, nil, "testcudd", dump, 0)
		} else {
			status = Cudd_DumpDot(dd, []*DdNode{C, M}, nil, onames, dump)
		}
		if status != 1 {
			fmt.Fprintln(stderr, "abnormal termination")
			return 0
		}
	}
	if o.clearcache != 0 {
		if o.pr > 0 {
			fmt.Fprint(stdout, "Clearing the cache... ")
		}
		cuddCacheFlush(dd)
		if o.pr > 0 {
			fmt.Fprintln(stdout, "done")
		}
	}
	if o.pr > 0 {
		fmt.Fprintf(stdout, "Number of variables = %6d\tNumber of slots = %6d\nNumber of keys = %6d\tNumber of min dead = %6d\n", Cudd_ReadSize(dd), Cudd_ReadSlots(dd), Cudd_ReadKeys(dd), Cudd_ReadMinDead(dd))
	}
	return 1
}

func printPermutation(dd *DdManager, size int32, out io.Writer) {
	fmt.Fprint(out, "Variable Permutation:")
	for i := int32(0); i < size; i++ {
		if i%20 == 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "%d ", Cudd_ReadInvPerm(dd, i))
	}
	fmt.Fprint(out, "\nInverse Permutation:")
	for i := int32(0); i < size; i++ {
		if i%20 == 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "%d ", Cudd_ReadPerm(dd, i))
	}
	fmt.Fprintln(out)
}
func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
func printPrimeCubes(dd *DdManager, f *DdNode, out io.Writer) {
	var cube *int32
	gen := Cudd_FirstPrime(dd, f, f, &cube)
	if gen == nil {
		return
	}
	defer Cudd_GenFree(gen)
	for Cudd_IsGenEmpty(gen) == 0 {
		for i := int32(0); i < Cudd_ReadSize(dd); i++ {
			switch unsafe.Slice(cube, int(Cudd_ReadSize(dd)))[i] {
			case 0:
				fmt.Fprint(out, "1")
			case 1:
				fmt.Fprint(out, "0")
			case 2:
				fmt.Fprint(out, "-")
			default:
				fmt.Fprint(out, "?")
			}
		}
		fmt.Fprintln(out, " 1")
		if Cudd_NextPrime(gen, &cube) == 0 {
			break
		}
	}
	fmt.Fprintln(out)
}

func testWalsh(dd *DdManager, n, cmu int32, approach Cudd_ReorderingType, pr int32, out, errOut io.Writer) int32 {
	if n <= 3 {
		return 1
	}
	x, v, z := make([]*DdNode, n), make([]*DdNode, n), make([]*DdNode, n)
	one, zero := Cudd_ReadOne(dd), Cudd_ReadZero(dd)
	for i := n - 1; i >= 0; i-- {
		x[i] = cuddUniqueInter(dd, 3*i, one, zero)
		v[i] = cuddUniqueInter(dd, 3*i+1, one, zero)
		z[i] = cuddUniqueInter(dd, 3*i+2, one, zero)
		Cudd_Ref(x[i])
		Cudd_Ref(v[i])
		Cudd_Ref(z[i])
	}
	walsh1 := Cudd_addWalsh(dd, v, z, n)
	if walsh1 == nil {
		return 0
	}
	Cudd_Ref(walsh1)
	if pr > 0 {
		fmt.Fprint(out, "walsh1")
		Cudd_PrintDebug(dd, walsh1, 2*n, pr)
	}
	walsh2 := Cudd_addWalsh(dd, x, v, n)
	if walsh2 == nil {
		Cudd_RecursiveDeref(dd, walsh1)
		return 0
	}
	Cudd_Ref(walsh2)
	var product *DdNode
	if cmu != 0 {
		product = Cudd_addTimesPlus(dd, walsh2, walsh1, v, n)
	} else {
		product = Cudd_addMatrixMultiply(dd, walsh2, walsh1, v, n)
	}
	if product == nil {
		Cudd_RecursiveDeref(dd, walsh1)
		Cudd_RecursiveDeref(dd, walsh2)
		return 0
	}
	Cudd_Ref(product)
	if pr > 0 {
		fmt.Fprint(out, "wtw")
		Cudd_PrintDebug(dd, product, 2*n, pr)
	}
	if approach != CUDD_REORDER_NONE && Cudd_ReduceHeap(dd, approach, 5) == 0 {
		fmt.Fprintln(errOut, "Error reported by Cudd_ReduceHeap")
		return 0
	}
	if approach == CUDD_REORDER_SYMM_SIFT || approach == CUDD_REORDER_SYMM_SIFT_CONV {
		Cudd_SymmProfile(dd, 0, Cudd_ReadSize(dd)-1)
	}
	Cudd_RecursiveDeref(dd, product)
	Cudd_RecursiveDeref(dd, walsh1)
	Cudd_RecursiveDeref(dd, walsh2)
	for i := int32(0); i < n; i++ {
		Cudd_RecursiveDeref(dd, x[i])
		Cudd_RecursiveDeref(dd, v[i])
		Cudd_RecursiveDeref(dd, z[i])
	}
	return 1
}

func testIterators(dd *DdManager, M, C *DdNode, pr int32, out io.Writer) int32 {
	var cube *int32
	var value CUDD_VALUE_TYPE
	gen := Cudd_FirstCube(dd, M, &cube, &value)
	if gen == nil {
		return 0
	}
	if pr > 1 {
		fmt.Fprintln(out, "Testing iterator on cubes:")
		for Cudd_IsGenEmpty(gen) == 0 {
			printCube(dd, cube, false, out)
			fmt.Fprintf(out, " %g\n", value)
			if Cudd_NextCube(gen, &cube, &value) == 0 {
				break
			}
		}
		fmt.Fprintln(out)
	}
	Cudd_GenFree(gen)
	if pr > 1 {
		fmt.Fprintln(out, "Testing prime expansion of cubes:")
		if Cudd_bddPrintCover(dd, C, C) == 0 {
			return 0
		}
	}
	if pr > 1 {
		fmt.Fprintln(out, "Testing iterator on primes (CNF):")
		printPrimeCubes(dd, Cudd_Not(C), out)
	}
	if pr > 2 {
		fmt.Fprintln(out, "Testing iterator on nodes:")
		var node *DdNode
		ngen := Cudd_FirstNode(dd, M, &node)
		if ngen == nil {
			return 0
		}
		for Cudd_IsGenEmpty(ngen) == 0 {
			r := Cudd_Regular(node)
			if Cudd_IsConstant(node) {
				fmt.Fprintf(out, "ID = 0x%x\tvalue = %-9g\n", r.ID, Cudd_V(node))
			} else {
				fmt.Fprintf(out, "ID = 0x%x\tindex = %d\tr = %d\n", r.ID, Cudd_NodeReadIndex(node), r.Ref)
			}
			if Cudd_NextNode(ngen, &node) == 0 {
				break
			}
		}
		Cudd_GenFree(ngen)
		fmt.Fprintln(out)
	}
	return 1
}
func printCube(dd *DdManager, cube *int32, negate bool, out io.Writer) {
	values := unsafe.Slice(cube, int(Cudd_ReadSize(dd)))
	for _, v := range values {
		switch v {
		case 0:
			if negate {
				fmt.Fprint(out, "1")
			} else {
				fmt.Fprint(out, "0")
			}
		case 1:
			if negate {
				fmt.Fprint(out, "0")
			} else {
				fmt.Fprint(out, "1")
			}
		case 2:
			fmt.Fprint(out, "-")
		default:
			fmt.Fprint(out, "?")
		}
	}
}
func testXor(dd *DdManager, f *DdNode, pr, nvars int32, out io.Writer) int32 {
	x := nvars / 2
	f1 := Cudd_Cofactor(dd, f, Cudd_ReadVars(dd, x))
	if f1 == nil {
		return 0
	}
	Cudd_Ref(f1)
	f0 := Cudd_Cofactor(dd, f, Cudd_Not(Cudd_ReadVars(dd, x)))
	if f0 == nil {
		Cudd_RecursiveDeref(dd, f1)
		return 0
	}
	Cudd_Ref(f0)
	res1 := Cudd_bddIte(dd, f1, Cudd_Not(f0), f0)
	if res1 == nil {
		Cudd_RecursiveDeref(dd, f1)
		Cudd_RecursiveDeref(dd, f0)
		return 0
	}
	Cudd_Ref(res1)
	if pr > 0 {
		fmt.Fprint(out, "xor1")
		Cudd_PrintDebug(dd, res1, nvars, pr)
	}
	res2 := Cudd_bddXor(dd, f1, f0)
	if res2 == nil {
		Cudd_RecursiveDeref(dd, res1)
		Cudd_RecursiveDeref(dd, f1)
		Cudd_RecursiveDeref(dd, f0)
		return 0
	}
	Cudd_Ref(res2)
	if res1 != res2 {
		if pr > 0 {
			fmt.Fprint(out, "xor2")
			Cudd_PrintDebug(dd, res2, nvars, pr)
		}
		Cudd_RecursiveDeref(dd, res1)
		Cudd_RecursiveDeref(dd, res2)
		Cudd_RecursiveDeref(dd, f1)
		Cudd_RecursiveDeref(dd, f0)
		return 0
	}
	Cudd_RecursiveDeref(dd, res1)
	Cudd_RecursiveDeref(dd, f1)
	Cudd_RecursiveDeref(dd, f0)
	res1 = Cudd_bddBooleanDiff(dd, f, int(x))
	if res1 == nil {
		Cudd_RecursiveDeref(dd, res2)
		return 0
	}
	Cudd_Ref(res1)
	if res1 != res2 {
		if pr > 0 {
			fmt.Fprint(out, "xor3")
			Cudd_PrintDebug(dd, res1, nvars, pr)
		}
		Cudd_RecursiveDeref(dd, res1)
		Cudd_RecursiveDeref(dd, res2)
		return 0
	}
	Cudd_RecursiveDeref(dd, res1)
	Cudd_RecursiveDeref(dd, res2)
	return 1
}
func testHamming(dd *DdManager, f *DdNode, pr int32, out io.Writer) int32 {
	size := Cudd_ReadSize(dd)
	vars := make([]*DdNode, size)
	for i := range vars {
		vars[i] = Cudd_bddIthVar(dd, int32(i))
	}
	minBdd := Cudd_bddPickOneMinterm(dd, Cudd_Not(f), vars, size)
	if minBdd == nil {
		return 0
	}
	Cudd_Ref(minBdd)
	if pr > 0 {
		fmt.Fprint(out, "Chosen minterm for Hamming distance test: ")
		Cudd_PrintDebug(dd, minBdd, size, pr)
	}
	minterm := make([]int32, size)
	scan := minBdd
	zero := Cudd_Not(Cudd_ReadOne(dd))
	for !Cudd_IsConstant(scan) {
		r := Cudd_Regular(scan)
		t, e := Cudd_T(r), Cudd_E(r)
		if r != scan {
			t, e = Cudd_Not(t), Cudd_Not(e)
		}
		if t == zero {
			minterm[Cudd_NodeReadIndex(r)] = 0
			scan = e
		} else {
			minterm[Cudd_NodeReadIndex(r)] = 1
			scan = t
		}
	}
	Cudd_RecursiveDeref(dd, minBdd)
	d := Cudd_MinHammingDist(dd, f, minterm, size)
	fmt.Fprintf(out, "Minimum Hamming distance = %d\n", d)
	return 1
}
