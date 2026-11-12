package bsat2

import "fmt"

import "os"

/*****************************************************************************************[Main.cc]
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
Copyright (c) 2007,      Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

//#include <errno.h>

//#include <signal.h>
//#include "misc/zlib/zlib.h"

#ifndef _WIN32
#include <sys/resource.h>
#endif

//#include "System.h"
//#include "ParseUtils.h"
//#include "Options.h"
//#include "Dimacs.h"
//#include "SimpSolver.h"

//using namespace Minisat;


//=================================================================================================

// extern void printStats(Solver& solver);


static Solver* solver;
// Terminate by notifying the solver and back out gracefully. This is mainly to have a test-case
// for this feature of the Solver as it may take longer than an immediate call to '_exit()'.
func SIGINT_interrupt(signum int32) { solver.interrupt(); }

// Note that '_exit()' rather than 'exit()' has to be used. The reason is that 'exit()' calls
// destructors and may cause deadlocks if a malloc/free function happens to be running (these
// functions are guarded by locks for multithreaded use).
func SIGINT_exit(signum int32) {
    fmt.Printf("\n"); fmt.Printf("*** INTERRUPTED ***\n");
    if solver.verbosity > 0{
        printStats(*solver);
        fmt.Printf("\n"); fmt.Printf("*** INTERRUPTED ***\n"); }
    _exit(1); }


//=================================================================================================
// Main:

func MainSimp(argc int32, argv *string) int32 {
    try {
        setUsageHelp("USAGE: %s [options] <input-file> <result-output-file>\n\n  where input may be either in plain or gzipped DIMACS.\n");
        // printf("This is MiniSat 2.0 beta\n");
        
#if defined(__linux__)
        fpu_control_t oldcw, newcw;
        _FPU_GETCW(oldcw); newcw = (oldcw & ~_FPU_EXTENDED) | _FPU_DOUBLE; _FPU_SETCW(newcw);
        printf("WARNING: for repeatability, setting FPU to use double precision\n");
#endif
        // Extra options:
        //
        IntOption    verb   ("MAIN", "verb",   "Verbosity level (0=silent, 1=some, 2=more).", 1, IntRange(0, 2));
        BoolOption   pre    ("MAIN", "pre",    "Completely turn on/off any preprocessing.", true);
        StringOption dimacs ("MAIN", "dimacs", "If given, stop after preprocessing and write the result to this file.");
        IntOption    cpu_lim("MAIN", "cpu-lim","Limit on CPU time allowed in seconds.\n", INT32_MAX, IntRange(0, INT32_MAX));
        IntOption    mem_lim("MAIN", "mem-lim","Limit on memory usage in megabytes.\n", INT32_MAX, IntRange(0, INT32_MAX));

        if !parseOptions(argc, argv, true) {
            return 1; }
        
        var S SimpSolver;
        var initial_time float64  = cpuTime();

        if !pre { S.eliminate(true); }

        S.verbosity = verb;
        
        solver = &S;
/*
        // Use signal handlers that forcibly quit until the solver will be able to respond to
        // interrupts:
        signal(SIGINT, SIGINT_exit);
        signal(SIGXCPU,SIGINT_exit);

        // Set limit on CPU-time:
        if (cpu_lim != INT32_MAX){
            rlimit rl;
            getrlimit(RLIMIT_CPU, &rl);
            if (rl.rlim_max == RLIM_INFINITY || (rlim_t)cpu_lim < rl.rlim_max){
                rl.rlim_cur = cpu_lim;
                if (setrlimit(RLIMIT_CPU, &rl) == -1)
                    printf("WARNING! Could not set resource limit: CPU-time.\n");
            } }

        // Set limit on virtual memory:
        if (mem_lim != INT32_MAX){
            rlim_t new_mem_lim = (rlim_t)mem_lim * 1024*1024;
            rlimit rl;
            getrlimit(RLIMIT_AS, &rl);
            if (rl.rlim_max == RLIM_INFINITY || new_mem_lim < rl.rlim_max){
                rl.rlim_cur = new_mem_lim;
                if (setrlimit(RLIMIT_AS, &rl) == -1)
                    printf("WARNING! Could not set resource limit: Virtual memory.\n");
            } }
*/        
        if argc == 1 {
            fmt.Printf("Reading from standard input... Use '--help' for help.\n");
            return 1;
        }

        var in gzFile;
        if argc == 1 {
            in = gzdopen(0, "rb");
        } else {
            in = gzopen(argv[1], "rb");
        }
        if in == nil {
            fmt.Printf("ERROR! Could not open file: %s\n", argc == 1 ? "<stdin>" : argv[1]);
os.Exit(int(1)); }
        
        if S.verbosity > 0{
            fmt.Printf("============================[ Problem Statistics ]=============================\n");
            fmt.Printf("|                                                                             |\n"); }
        
        parse_DIMACS(in, S);
        gzclose(in);
        var res *os.File  = (argc >= 3) ? os.OpenFile(argv[2], os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666) : nil;

        if S.verbosity > 0{
            fmt.Printf("|  Number of variables:  %12d                                         |\n", S.nVars());
            fmt.Printf("|  Number of clauses:    %12d                                         |\n", S.nClauses()); }
        
        var parsed_time float64  = cpuTime();
        if S.verbosity > 0 {
            fmt.Printf("|  Parse time:           %12.2f s                                       |\n", parsed_time - initial_time); }

        // Change to signal-handlers that will only notify the solver and allow it to terminate
        // voluntarily:
//        signal(SIGINT, SIGINT_interrupt);
//        signal(SIGXCPU,SIGINT_interrupt);

        S.eliminate(true);
        var simplified_time float64  = cpuTime();
        if S.verbosity > 0{
            fmt.Printf("|  Simplification time:  %12.2f s                                       |\n", simplified_time - parsed_time);
            fmt.Printf("|                                                                             |\n"); }

        if !S.okay(){
            if res != nil { fmt.Fprintf(res, "UNSAT\n");
_ = (res).Close(); }
            if S.verbosity > 0{
                fmt.Printf("===============================================================================\n");
                fmt.Printf("Solved by simplification\n");
                printStats(S);
                fmt.Printf("\n"); }
            fmt.Printf("UNSATISFIABLE\n");
            os.Exit(int(20));
        }

        if dimacs{
            if S.verbosity > 0 {
                fmt.Printf("==============================[ Writing DIMACS ]===============================\n"); }
            S.toDimacs((char*)dimacs);
            if S.verbosity > 0 {
                printStats(S); }
            os.Exit(int(0));
        }

        vec<Lit> dummy;
        var ret lbool  = S.solveLimited(dummy);
        
        if S.verbosity > 0{
            printStats(S);
            fmt.Printf("\n"); }
        fmt.Printf(ret == l_True ? "SATISFIABLE\n" : ret == l_False ? "UNSATISFIABLE\n" : "INDETERMINATE\n");
        if res != nil{
            if ret == l_True{
                fmt.Fprintf(res, "SAT\n");
                for var i int32  = 0; i < S.nVars(); i++ {
                    if S.model[i] != l_Undef {
                        fmt.Fprintf(res, "%s%s%d", (i==0)?"":" ", (S.model[i]==l_True)?"":"-", i+1); }
                }
                fmt.Fprintf(res, " 0\n");
            }else if ret == l_False {
                fmt.Fprintf(res, "UNSAT\n");
            } else {
                fmt.Fprintf(res, "INDET\n"); }
            _ = (res).Close();
        }

//#ifdef NDEBUG
//        exit(ret == l_True ? 10 : ret == l_False ? 20 : 0);     // (faster than "return", which will invoke the destructor for 'Solver')
//#else
        if ret == l_True {
            return 10;
        } else {
            if ret == l_False {
                return 20;
            } else {
                return 0;
            }
        }
//#endif
    }
 catch (OutOfMemoryException&){
        fmt.Printf("===============================================================================\n");
        fmt.Printf("INDETERMINATE\n");
        os.Exit(int(0));
    }
}
