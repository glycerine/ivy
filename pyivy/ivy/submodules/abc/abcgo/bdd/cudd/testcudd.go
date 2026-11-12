package cudd

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile***********************************************************************

  FileName    [testcudd.c]

  PackageName [cudd]

  Synopsis    [Sanity check tests for some CUDD functions.]

  Description [testcudd reads a matrix with real coefficients and
  transforms it into an ADD. It then performs various operations on
  the ADD and on the BDD corresponding to the ADD pattern. Finally,
  testcudd tests functions relate to Walsh matrices and matrix
  multiplication.]

  SeeAlso     []

  Author      [Fabio Somenzi]

  Copyright   [Copyright (c) 1995-2004, Regents of the University of Colorado

  All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions
  are met:

  Redistributions of source code must retain the above copyright
  notice, this list of conditions and the following disclaimer.

  Redistributions in binary form must reproduce the above copyright
  notice, this list of conditions and the following disclaimer in the
  documentation and/or other materials provided with the distribution.

  Neither the name of the University of Colorado nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
  "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
  LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
  FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
  COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
  INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
  BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
  LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
  CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
  LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN
  ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
  POSSIBILITY OF SUCH DAMAGE.]

******************************************************************************/

//#include "util.h"
//#include "cuddInt.h"


/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

const TESTCUDD_VERSION =    "TestCudd Version #1.0, Release date 3/17/01"

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: testcudd.c,v 1.20 2009/03/08 02:49:02 fabio Exp $";
#endif

var onames []string = []string{ "C", "M" }; /* names of functions to be dumped */

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static void usage (char * prog);

// static FILE *open_file (char *filename, const char *mode);

// static int testIterators (DdManager *dd, DdNode *M, DdNode *C, int pr);

// static int testXor (DdManager *dd, DdNode *f, int pr, int nvars);

// static int testHamming (DdManager *dd, DdNode *f, int pr);

// static int testWalsh (DdManager *dd, int N, int cmu, int approach, int pr);


/**AutomaticEnd***************************************************************/


/**Function********************************************************************

  Synopsis    [Main function for testcudd.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func main() {
    var fp *os.File;           /* pointer to input file */
    var file string  = (char *) "";    /* input file name */
    var dfp *os.File  = nil;    /* pointer to dump file */
    var dfile string;    /* file for DD dump */
    var dfunc [2]*DdNode;    /* addresses of the functions to be dumped */
    var dd *DdManager;    /* pointer to DD manager */
    var one *DdNode;    /* fast access to constant function */
    var M *DdNode;
    var x **DdNode;        /* pointers to variables */
    var y **DdNode;        /* pointers to variables */
    var xn **DdNode;    /* complements of row variables */
    var yn_ **DdNode;    /* complements of column variables */
    var xvars **DdNode;
    var yvars **DdNode;
    var C *DdNode;        /* result of converting from ADD to BDD */
    var ess *DdNode;    /* cube of essential variables */
    var shortP *DdNode;    /* BDD cube of shortest path */
    var largest *DdNode;    /* BDD of largest cube */
    var shortA *DdNode;    /* ADD cube of shortest path */
    var constN *DdNode;    /* value returned by evaluation of ADD */
    var ycube *DdNode;    /* cube of the negated y vars for c-proj */
    var CP *DdNode;        /* C-Projection of C */
    var CPr *DdNode;    /* C-Selection of C */
    var length int32;    /* length of the shortest path */
    var nx int32;            /* number of variables */
    var ny int32;
    var maxnx int32;
    var maxny int32;
    var m int32;
    var n int32;
    var N int32;
    var cmu int32;            /* use CMU multiplication */
    var pr int32;            /* verbose printout level */
    var harwell int32;
    var multiple int32;        /* read multiple matrices */
    var ok int32;
    var c int32;            /* variable to read in options */
    var approach int32;        /* reordering approach */
    var autodyn int32;        /* automatic reordering */
    var groupcheck int32;        /* option for group sifting */
    var profile int32;        /* print heap profile if != 0 */
    var keepperm int32;        /* keep track of permutation */
    var clearcache int32;        /* clear the cache after each matrix */
    var blifOrDot int32;        /* dump format: 0 -> dot, 1 -> blif, ... */
    var retval int32;        /* return value */
    var i int32;            /* loop index */
    var startTime long;        /* initial time */
    var lapTime long;
    var size int32;
    unsigned int cacheSize, maxMemory;
    unsigned int nvars,nslots;

    startTime = util_cpu_time();

    approach = CUDD_REORDER_NONE;
    autodyn = 0;
    pr = 0;
    harwell = 0;
    multiple = 0;
    profile = 0;
    keepperm = 0;
    cmu = 0;
    N = 4;
    nvars = 4;
    cacheSize = 127;
    maxMemory = 0;
    nslots = CUDD_UNIQUE_SLOTS;
    clearcache = 0;
    groupcheck = CUDD_GROUP_CHECK7;
    dfile = nil;
    blifOrDot = 0; /* dot format */

    /* Parse command line. */
    for (c = util_getopt(len(os.Args), os.Args, (char *) "CDHMPS:a:bcd:g:hkmn:p:v:x:X:"))
       != -1 {
    switch c {
    case 'C':
        cmu = 1;
        break;
    case 'D':
        autodyn = 1;
        break;
    case 'H':
        harwell = 1;
        break;
    case 'M':
#ifdef MNEMOSYNE
        (void) mnem_setrecording(0);
#endif
        break;
    case 'P':
        profile = 1;
        break;
    case 'S':
        nslots = strconv.Atoi(util_optarg);
        break;
    case 'X':
        maxMemory = strconv.Atoi(util_optarg);
        break;
    case 'a':
        approach = strconv.Atoi(util_optarg);
        break;
    case 'b':
        blifOrDot = 1; /* blif format */
        break;
    case 'c':
        clearcache = 1;
        break;
    case 'd':
        dfile = util_optarg;
        break;
    case 'g':
        groupcheck = strconv.Atoi(util_optarg);
        break;
    case 'k':
        keepperm = 1;
        break;
    case 'm':
        multiple = 1;
        break;
    case 'n':
        N = strconv.Atoi(util_optarg);
        break;
    case 'p':
        pr = strconv.Atoi(util_optarg);
        break;
    case 'v':
        nvars = strconv.Atoi(util_optarg);
        break;
    case 'x':
        cacheSize = strconv.Atoi(util_optarg);
        break;
    case 'h':
    fallthrough
default:
        usage(os.Args[0]);
        break;
    }
    }

    if len(os.Args) - util_optind == 0 {
    file = (char *) "-";
    } else if len(os.Args) - util_optind == 1 {
    file = os.Args[util_optind];
    } else {
    usage(os.Args[0]);
    }
    if (approach<0) || (approach>17) {
    (void) fmt.Fprintf(os.Stderr, "Invalid approach: %d \n", approach);
    usage(os.Args[0]);
    }

    if pr >= 0 {
    (void) fmt.Printf("# %s\n", TESTCUDD_VERSION);
    /* Echo command line and arguments. */
    (void) fmt.Printf("#");
    for i = 0; i < len(os.Args); i++ {
        (void) fmt.Printf(" %s", os.Args[i]);
    }
    (void) fmt.Printf("\n");
    (void) fflush(os.Stdout);
    }

    /* Initialize manager and provide easy reference to terminals. */
    dd = Cudd_Init(nvars,0,nslots,cacheSize,maxMemory);
    // inlining of macro DD_ONE
    one = ((dd).one);
    dd.groupcheck = (Cudd_AggregationType) groupcheck;
    if autodyn != 0 { Cudd_AutodynEnable(dd,CUDD_REORDER_SAME); }

    /* Open input file. */
    fp = open_file(file, "r");

    /* Open dump file if requested */
    if dfile != nil {
    dfp = open_file(dfile, "w");
    }

    x = y = xn = yn_ = nil;
    for {
    /* We want to start anew for every matrix. */
    maxnx = maxny = 0;
    nx = maxnx; ny = maxny;
    if pr>0 { lapTime = util_cpu_time(); }
    if harwell != 0 {
        if pr >= 0 { (void) fmt.Printf(":name: "); }
        ok = Cudd_addHarwell(fp, dd, &M, &x, &y, &xn, &yn_, &nx, &ny,
        &m, &n, 0, 2, 1, 2, pr);
    } else {
        ok = Cudd_addRead(fp, dd, &M, &x, &y, &xn, &yn_, &nx, &ny,
        &m, &n, 0, 2, 1, 2);
        if pr >= 0 {
        (void) fmt.Printf(":name: %s: %d rows %d columns\n", file, m, n); }
    }
    if !(ok != 0) {
        (void) fmt.Fprintf(os.Stderr, "Error reading matrix\n");
        os.Exit(int(1));
    }

    if nx > maxnx { maxnx = nx; }
    if ny > maxny { maxny = ny; }

    /* Build cube of negated y's. */
    // inlining of macro DD_ONE
    ycube = ((dd).one);
    Cudd_Ref(ycube);
    for i = maxny - 1; i >= 0; i-- {
        var tmpp *DdNode;
        tmpp = Cudd_bddAnd(dd,Cudd_Not(dd.vars[y[i].index]),ycube);
        if tmpp == nil { os.Exit(int(2)); }
        Cudd_Ref(tmpp);
        Cudd_RecursiveDeref(dd,ycube);
        ycube = tmpp;
    }
    /* Initialize vectors of BDD variables used by priority func. */
    xvars = ALLOC(DdNode *, nx);
    if xvars == nil { os.Exit(int(2)); }
    for i = 0; i < nx; i++ {
        xvars[i] = dd.vars[x[i].index];
    }
    yvars = ALLOC(DdNode *, ny);
    if yvars == nil { os.Exit(int(2)); }
    for i = 0; i < ny; i++ {
        yvars[i] = dd.vars[y[i].index];
    }

    /* Clean up */
    for i=0; i < maxnx; i++ {
        Cudd_RecursiveDeref(dd, x[i]);
        Cudd_RecursiveDeref(dd, xn[i]);
    }
    FREE(x);
    FREE(xn);
    for i=0; i < maxny; i++ {
        Cudd_RecursiveDeref(dd, y[i]);
        Cudd_RecursiveDeref(dd, yn_[i]);
    }
    FREE(y);
    FREE(yn_);

    if pr>0 {(void) fmt.Printf(":1: M"); Cudd_PrintDebug(dd,M,nx+ny,pr);}

    if pr>0 { (void) fmt.Printf(":2: time to read the matrix = %s\n", util_print_time(util_cpu_time() - lapTime)); }

    C = Cudd_addBddPattern(dd, M);
    if C == 0 { os.Exit(int(2)); }
    Cudd_Ref(C);
    if pr>0 {(void) fmt.Printf(":3: C"); Cudd_PrintDebug(dd,C,nx+ny,pr);}

    /* Test iterators. */
    retval = testIterators(dd,M,C,pr);
    if retval == 0 { os.Exit(int(2)); }

    cuddCacheProfile(dd,os.Stdout);

    /* Test XOR */
    retval = testXor(dd,C,pr,nx+ny);
    if retval == 0 { os.Exit(int(2)); }

    /* Test Hamming distance functions. */
    retval = testHamming(dd,C,pr);
    if retval == 0 { os.Exit(int(2)); }

    /* Test selection functions. */
    CP = Cudd_CProjection(dd,C,ycube);
    if CP == nil { os.Exit(int(2)); }
    Cudd_Ref(CP);
    if pr>0 {(void) fmt.Printf("ycube"); Cudd_PrintDebug(dd,ycube,nx+ny,pr);}
    if pr>0 {(void) fmt.Printf("CP"); Cudd_PrintDebug(dd,CP,nx+ny,pr);}

    if nx == ny {
        CPr = Cudd_PrioritySelect(dd,C,xvars,yvars,(DdNode **)nil,
        (DdNode *)nil,ny,Cudd_Xgty);
        if CPr == nil { os.Exit(int(2)); }
        Cudd_Ref(CPr);
        if pr>0 {(void) fmt.Printf(":4: CPr"); Cudd_PrintDebug(dd,CPr,nx+ny,pr);}
        if CP != CPr {
        (void) fmt.Printf("CP != CPr!\n");
        }
        Cudd_RecursiveDeref(dd, CPr);
    }

    /* Test inequality generator. */
    {
        var Nmin int32;
// inlining of macro ddMin

        if (ny) < (nx) {
            Nmin = (ny);
        } else {
            Nmin = (nx);
        }
        var q int32;
        var gen *DdGen;
        var cube *int32;
        var f *DdNode  = Cudd_Inequality(dd,Nmin,2,xvars,yvars);
        if f == nil { os.Exit(int(2)); }
        Cudd_Ref(f);
        if pr>0 {
        (void) fmt.Printf(":4: ineq");
        Cudd_PrintDebug(dd,f,nx+ny,pr);
        if pr>1 {
            Cudd_ForeachPrime(dd,Cudd_Not(f),Cudd_Not(f),gen,cube) {
            for q = 0; q < dd.size; q++ {
                switch cube[q] {
                case 0:
                (void) fmt.Printf("1");
                break;
                case 1:
                (void) fmt.Printf("0");
                break;
                case 2:
                (void) fmt.Printf("-");
                break;
                default:
                (void) fmt.Printf("?");
                }
            }
            (void) fmt.Printf(" 1\n");
            }
            (void) fmt.Printf("\n");
        }
        }
        Cudd_IterDerefBdd(dd, f);
    }
    FREE(xvars); FREE(yvars);

    Cudd_RecursiveDeref(dd, CP);
    Cudd_RecursiveDeref(dd, ycube);

    /* Test functions for essential variables. */
    ess = Cudd_FindEssential(dd,C);
    if ess == nil { os.Exit(int(2)); }
    Cudd_Ref(ess);
    if pr>0 {(void) fmt.Printf(":4: ess"); Cudd_PrintDebug(dd,ess,nx+ny,pr);}
    Cudd_RecursiveDeref(dd, ess);

    /* Test functions for shortest paths. */
    shortP = Cudd_ShortestPath(dd, M, nil, nil, &length);
    if shortP == nil { os.Exit(int(2)); }
    Cudd_Ref(shortP);
    if pr>0 {
        (void) fmt.Printf(":5: shortP"); Cudd_PrintDebug(dd,shortP,nx+ny,pr);
    }
    /* Test functions for largest cubes. */
    largest = Cudd_LargestCube(dd, Cudd_Not(C), &length);
    if largest == nil { os.Exit(int(2)); }
    Cudd_Ref(largest);
    if pr>0 {
        (void) fmt.Printf(":5b: largest");
        Cudd_PrintDebug(dd,largest,nx+ny,pr);
    }
    Cudd_RecursiveDeref(dd, largest);

    /* Test Cudd_addEvalConst and Cudd_addIteConstant. */
    shortA = Cudd_BddToAdd(dd,shortP);
    if shortA == nil { os.Exit(int(2)); }
    Cudd_Ref(shortA);
    Cudd_RecursiveDeref(dd, shortP);
    constN = Cudd_addEvalConst(dd,shortA,M);
    if constN == DD_NON_CONSTANT { os.Exit(int(2)); }
    if Cudd_addIteConstant(dd,shortA,M,constN) != constN { os.Exit(int(2)); }
    // inlining of macro cuddV
    if pr>0 {(void) fmt.Printf("The value of M along the chosen shortest path is %g\n", ((constN).type.value));}
    Cudd_RecursiveDeref(dd, shortA);

    shortP = Cudd_ShortestPath(dd, C, nil, nil, &length);
    if shortP == nil { os.Exit(int(2)); }
    Cudd_Ref(shortP);
    if pr>0 {
        (void) fmt.Printf(":6: shortP"); Cudd_PrintDebug(dd,shortP,nx+ny,pr);
    }

    /* Test Cudd_bddIteConstant and Cudd_bddLeq. */
    if !Cudd_bddLeq(dd,shortP,C) { os.Exit(int(2)); }
    if Cudd_bddIteConstant(dd,Cudd_Not(shortP),one,C) != one { os.Exit(int(2)); }
    Cudd_RecursiveDeref(dd, shortP);

    if profile != 0 {
        retval = cuddHeapProfile(dd);
    }

    size = dd.size;

    if pr>0 {
        (void) fmt.Printf("Average distance: %g\n", Cudd_AverageDistance(dd));
    }

    /* Reorder if so requested. */
    if approach != CUDD_REORDER_NONE {
#ifndef DD_STATS
        retval = Cudd_EnableReorderingReporting(dd);
        if (retval == 0) {
        (void) fprintf(stderr,"Error reported by Cudd_EnableReorderingReporting\n");
        exit(3);
        }
#endif
#ifdef DD_DEBUG
        retval = Cudd_DebugCheck(dd);
        if (retval != 0) {
        (void) fprintf(stderr,"Error reported by Cudd_DebugCheck\n");
        exit(3);
        }
        retval = Cudd_CheckKeys(dd);
        if (retval != 0) {
        (void) fprintf(stderr,"Error reported by Cudd_CheckKeys\n");
        exit(3);
        }
#endif
        retval = Cudd_ReduceHeap(dd,(Cudd_ReorderingType)approach,5);
        if retval == 0 {
        (void) fmt.Fprintf(os.Stderr, "Error reported by Cudd_ReduceHeap\n");
        os.Exit(int(3));
        }
#ifndef DD_STATS
        retval = Cudd_DisableReorderingReporting(dd);
        if (retval == 0) {
        (void) fprintf(stderr,"Error reported by Cudd_DisableReorderingReporting\n");
        exit(3);
        }
#endif
#ifdef DD_DEBUG
        retval = Cudd_DebugCheck(dd);
        if (retval != 0) {
        (void) fprintf(stderr,"Error reported by Cudd_DebugCheck\n");
        exit(3);
        }
        retval = Cudd_CheckKeys(dd);
        if (retval != 0) {
        (void) fprintf(stderr,"Error reported by Cudd_CheckKeys\n");
        exit(3);
        }
#endif
        if approach == CUDD_REORDER_SYMM_SIFT ||
        approach == CUDD_REORDER_SYMM_SIFT_CONV {
        Cudd_SymmProfile(dd,0,dd.size-1);
        }

        if pr>0 {
        (void) fmt.Printf("Average distance: %g\n", Cudd_AverageDistance(dd));
        }

        if keepperm != 0 {
        /* Print variable permutation. */
        (void) fmt.Printf("Variable Permutation:");
        for i=0; i<size; i++ {
            if i%20 == 0 { (void) fmt.Printf("\n"); }
            (void) fmt.Printf("%d ", dd.invperm[i]);
        }
        (void) fmt.Printf("\n");
        (void) fmt.Printf("Inverse Permutation:");
        for i=0; i<size; i++ {
            if i%20 == 0 { (void) fmt.Printf("\n"); }
            (void) fmt.Printf("%d ", dd.perm[i]);
        }
        (void) fmt.Printf("\n");
        }

        if pr>0 {(void) fmt.Printf("M"); Cudd_PrintDebug(dd,M,nx+ny,pr);}

        if profile != 0 {
        retval = cuddHeapProfile(dd);
        }

    }

    /* Dump DDs of C and M if so requested. */
    if dfile != nil {
        dfunc[0] = C;
        dfunc[1] = M;
        if blifOrDot == 1 {
        /* Only dump C because blif cannot handle ADDs */
        retval = Cudd_DumpBlif(dd,1,dfunc,nil,(char **)onames,
                       nil,dfp,0);
        } else {
        retval = Cudd_DumpDot(dd,2,dfunc,nil,(char **)onames,dfp);
        }
        if retval != 1 {
        (void) fmt.Fprintf(os.Stderr, "abnormal termination\n");
        os.Exit(int(2));
        }
    }

    Cudd_RecursiveDeref(dd, C);
    Cudd_RecursiveDeref(dd, M);

    if clearcache != 0 {
        if pr>0 {(void) fmt.Printf("Clearing the cache... ");}
        for i = dd.cacheSlots - 1; i>=0; i-- {
        dd.cache[i].data = NIL(DdNode);
        }
        if pr>0 {(void) fmt.Printf("done\n");}
    }
    if pr>0 {
        (void) fmt.Printf("Number of variables = %6d\t", dd.size);
        (void) fmt.Printf("Number of slots     = %6d\n", dd.slots);
        (void) fmt.Printf("Number of keys      = %6d\t", dd.keys);
        (void) fmt.Printf("Number of min dead  = %6d\n", dd.minDead);
    }

    
if !((multiple != 0) && !feof(fp)) { break }
} 

    _ = (fp).Close();
    if dfile != nil {
    _ = (dfp).Close();
    }

    /* Second phase: experiment with Walsh matrices. */
    if !testWalsh(dd,N,cmu,approach,pr) {
    os.Exit(int(2));
    }

    /* Check variable destruction. */
    assert(cuddDestroySubtables(dd,3), "cuddDestroySubtables(dd,3)");
    assert(Cudd_DebugCheck(dd) == 0, "Cudd_DebugCheck(dd) == 0");
    assert(Cudd_CheckKeys(dd) == 0, "Cudd_CheckKeys(dd) == 0");

    retval = Cudd_CheckZeroRef(dd);
    ok = retval != 0;  /* ok == 0 means O.K. */
    if retval != 0 {
    (void) fmt.Fprintf(os.Stderr, "%d non-zero DD reference counts after dereferencing\n", retval);
    }

    if pr >= 0 {
    (void) Cudd_PrintInfo(dd,os.Stdout);
    }

    Cudd_Quit(dd);

#ifdef MNEMOSYNE
    mnem_writestats();
#endif

    if pr>0 { (void) fmt.Printf("total time = %s\n", util_print_time(util_cpu_time() - startTime)); }

    if pr >= 0 { util_print_cpu_stats(os.Stdout); }
    os.Exit(int(ok));
    /* NOTREACHED */

} /* end of main */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Prints usage info for testcudd.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func usage(prog string) {
    (void) fmt.Fprintf(os.Stderr, "usage: %s [options] [file]\n", prog);
    (void) fmt.Fprintf(os.Stderr, "   -C\t\tuse CMU multiplication algorithm\n");
    (void) fmt.Fprintf(os.Stderr, "   -D\t\tenable automatic dynamic reordering\n");
    (void) fmt.Fprintf(os.Stderr, "   -H\t\tread matrix in Harwell format\n");
    (void) fmt.Fprintf(os.Stderr, "   -M\t\tturns off memory allocation recording\n");
    (void) fmt.Fprintf(os.Stderr, "   -P\t\tprint BDD heap profile\n");
    (void) fmt.Fprintf(os.Stderr, "   -S n\t\tnumber of slots for each subtable\n");
    (void) fmt.Fprintf(os.Stderr, "   -X n\t\ttarget maximum memory in bytes\n");
    (void) fmt.Fprintf(os.Stderr, "   -a n\t\tchoose reordering approach (0-13)\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t0: same as autoMethod\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t1: no reordering (default)\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t2: random\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t3: pivot\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t4: sifting\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t5: sifting to convergence\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t6: symmetric sifting\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t7: symmetric sifting to convergence\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t8-10: window of size 2-4\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t11-13: window of size 2-4 to conv.\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t14: group sifting\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t15: group sifting to convergence\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t16: simulated annealing\n");
    (void) fmt.Fprintf(os.Stderr, "   \t\t\t17: genetic algorithm\n");
    (void) fmt.Fprintf(os.Stderr, "   -b\t\tuse blif as format for dumps\n");
    (void) fmt.Fprintf(os.Stderr, "   -c\t\tclear the cache after each matrix\n");
    (void) fmt.Fprintf(os.Stderr, "   -d file\tdump DDs to file\n");
    (void) fmt.Fprintf(os.Stderr, "   -g\t\tselect aggregation criterion (0,5,7)\n");
    (void) fmt.Fprintf(os.Stderr, "   -h\t\tprints this message\n");
    (void) fmt.Fprintf(os.Stderr, "   -k\t\tprint the variable permutation\n");
    (void) fmt.Fprintf(os.Stderr, "   -m\t\tread multiple matrices (only with -H)\n");
    (void) fmt.Fprintf(os.Stderr, "   -n n\t\tnumber of variables\n");
    (void) fmt.Fprintf(os.Stderr, "   -p n\t\tcontrol verbosity\n");
    (void) fmt.Fprintf(os.Stderr, "   -v n\t\tinitial variables in the unique table\n");
    (void) fmt.Fprintf(os.Stderr, "   -x n\t\tinitial size of the cache\n");
    os.Exit(int(2));
} /* end of usage */


/**Function********************************************************************

  Synopsis    [Opens a file.]

  Description [Opens a file, or fails with an error message and exits.
  Allows '-' as a synonym for standard input.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func open_file(filename string, mode string) *os.File {
    var fp *os.File;

    if int32(strings.Compare(filename, "-")) == 0 {
    if mode[0] == 'r' {
        return os.Stdin;
    } else {
        return os.Stdout;
    }
    } else if (fp = fopen(filename, mode)) == nil {
    perror(filename);
    os.Exit(int(1));
    }
    return fp;

} /* end of open_file */


/**Function********************************************************************

  Synopsis    [Tests Walsh matrix multiplication.]

  Description [Tests Walsh matrix multiplication.  Return 1 if successful;
  0 otherwise.]

  SideEffects [May create new variables in the manager.]

  SeeAlso     []

******************************************************************************/
static int
testWalsh(
  DdManager *dd /* manager */,
  int N /* number of variables */,
  int cmu /* use CMU approach to matrix multiplication */,
  int approach /* reordering approach */,
  int pr /* verbosity level */)
{
    var walsh1, walsh2, wtw *DdNode;
    var x, v, z **DdNode;
    var i, retval int32;
    // inlining of macro DD_ONE
    var one *DdNode  = ((dd).one);
    // inlining of macro DD_ZERO
    var zero *DdNode  = ((dd).zero);

    if N > 3 {
    x = ALLOC(DdNode *,N);
    v = ALLOC(DdNode *,N);
    z = ALLOC(DdNode *,N);

    for i = N-1; i >= 0; i-- {
        Cudd_Ref(x[i]=cuddUniqueInter(dd,3*i,one,zero));
        Cudd_Ref(v[i]=cuddUniqueInter(dd,3*i+1,one,zero));
        Cudd_Ref(z[i]=cuddUniqueInter(dd,3*i+2,one,zero));
    }
    Cudd_Ref(walsh1 = Cudd_addWalsh(dd,v,z,N));
    if pr>0 {(void) fmt.Printf("walsh1"); Cudd_PrintDebug(dd,walsh1,2*N,pr);}
    Cudd_Ref(walsh2 = Cudd_addWalsh(dd,x,v,N));
    if cmu != 0 {
        Cudd_Ref(wtw = Cudd_addTimesPlus(dd,walsh2,walsh1,v,N));
    } else {
        Cudd_Ref(wtw = Cudd_addMatrixMultiply(dd,walsh2,walsh1,v,N));
    }
    if pr>0 {(void) fmt.Printf("wtw"); Cudd_PrintDebug(dd,wtw,2*N,pr);}

    if approach != CUDD_REORDER_NONE {
#ifdef DD_DEBUG
        retval = Cudd_DebugCheck(dd);
        if (retval != 0) {
        (void) fprintf(stderr,"Error reported by Cudd_DebugCheck\n");
        return(0);
        }
#endif
        retval = Cudd_ReduceHeap(dd,(Cudd_ReorderingType)approach,5);
        if retval == 0 {
        (void) fmt.Fprintf(os.Stderr, "Error reported by Cudd_ReduceHeap\n");
        return(0);
        }
#ifdef DD_DEBUG
        retval = Cudd_DebugCheck(dd);
        if (retval != 0) {
        (void) fprintf(stderr,"Error reported by Cudd_DebugCheck\n");
        return(0);
        }
#endif
        if approach == CUDD_REORDER_SYMM_SIFT ||
        approach == CUDD_REORDER_SYMM_SIFT_CONV {
        Cudd_SymmProfile(dd,0,dd.size-1);
        }
    }
    /* Clean up. */
    Cudd_RecursiveDeref(dd, wtw);
    Cudd_RecursiveDeref(dd, walsh1);
    Cudd_RecursiveDeref(dd, walsh2);
    for i=0; i < N; i++ {
        Cudd_RecursiveDeref(dd, x[i]);
        Cudd_RecursiveDeref(dd, v[i]);
        Cudd_RecursiveDeref(dd, z[i]);
    }
    FREE(x);
    FREE(v);
    FREE(z);
    }
    return(1);

} /* end of testWalsh */

/**Function********************************************************************

  Synopsis    [Tests iterators.]

  Description [Tests iterators on cubes and nodes.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func testIterators(dd *DdManager, M *DdNode, C *DdNode, pr int32) int32 {
    var cube *int32;
    CUDD_VALUE_TYPE value;
    var gen *DdGen;
    var q int32;

    /* Test iterator for cubes. */
    if pr>1 {
    (void) fmt.Printf("Testing iterator on cubes:\n");
    Cudd_ForeachCube(dd,M,gen,cube,value) {
        for q = 0; q < dd.size; q++ {
        switch cube[q] {
        case 0:
            (void) fmt.Printf("0");
            break;
        case 1:
            (void) fmt.Printf("1");
            break;
        case 2:
            (void) fmt.Printf("-");
            break;
        default:
            (void) fmt.Printf("?");
        }
        }
        (void) fmt.Printf(" %g\n", value);
    }
    (void) fmt.Printf("\n");
    }

    if pr>1 {
    (void) fmt.Printf("Testing prime expansion of cubes:\n");
    if !Cudd_bddPrintCover(dd,C,C) { return(0); }
    }

    if pr>1 {
    (void) fmt.Printf("Testing iterator on primes (CNF):\n");
    Cudd_ForeachPrime(dd,Cudd_Not(C),Cudd_Not(C),gen,cube) {
        for q = 0; q < dd.size; q++ {
        switch cube[q] {
        case 0:
            (void) fmt.Printf("1");
            break;
        case 1:
            (void) fmt.Printf("0");
            break;
        case 2:
            (void) fmt.Printf("-");
            break;
        default:
            (void) fmt.Printf("?");
        }
        }
        (void) fmt.Printf(" 1\n");
    }
    (void) fmt.Printf("\n");
    }

    /* Test iterator on nodes. */
    if pr>2 {
    var node *DdNode;
    (void) fmt.Printf("Testing iterator on nodes:\n");
    Cudd_ForeachNode(dd,M,gen,node) {
        if Cudd_IsConstant(node) {
#if SIZEOF_VOID_P == 8
        (void) printf("ID = 0x%lx\tvalue = %-9g\n",
                  (ptruint) node /
                  (ptruint) sizeof(DdNode),
                  Cudd_V(node));
#else
        (void) printf("ID = 0x%x\tvalue = %-9g\n",
                  (ptruint) node /
                  (ptruint) sizeof(DdNode),
                  Cudd_V(node));
#endif
        } else {
#if SIZEOF_VOID_P == 8
        (void) printf("ID = 0x%lx\tindex = %u\tr = %u\n",
                  (ptruint) node /
                  (ptruint) sizeof(DdNode),
                  node->index, node->ref);
#else
        (void) printf("ID = 0x%x\tindex = %u\tr = %u\n",
                  (ptruint) node /
                  (ptruint) sizeof(DdNode),
                  node->index, node->ref);
#endif
        }
    }
    (void) fmt.Printf("\n");
    }
    return(1);

} /* end of testIterators */


/**Function********************************************************************

  Synopsis    [Tests the functions related to the exclusive OR.]

  Description [Tests the functions related to the exclusive OR. It
  builds the boolean difference of the given function in three
  different ways and checks that the results is the same. Returns 1 if
  successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func testXor(dd *DdManager, f *DdNode, pr int32, nvars int32) int32 {
    var f1, f0, res1, res2 *DdNode;
    var x int32;

    /* Extract cofactors w.r.t. mid variable. */
    x = nvars / 2;
    f1 = Cudd_Cofactor(dd,f,dd.vars[x]);
    if f1 == nil { return(0); }
    Cudd_Ref(f1);

    f0 = Cudd_Cofactor(dd,f,Cudd_Not(dd.vars[x]));
    if f0 == nil {
    Cudd_RecursiveDeref(dd,f1);
    return(0);
    }
    Cudd_Ref(f0);

    /* Compute XOR of cofactors with ITE. */
    res1 = Cudd_bddIte(dd,f1,Cudd_Not(f0),f0);
    if res1 == nil { return(0); }
    Cudd_Ref(res1);

    if pr>0 {(void) fmt.Printf("xor1"); Cudd_PrintDebug(dd,res1,nvars,pr);}

    /* Compute XOR of cofactors with XOR. */
    res2 = Cudd_bddXor(dd,f1,f0);
    if res2 == nil {
    Cudd_RecursiveDeref(dd,res1);
    return(0);
    }
    Cudd_Ref(res2);

    if res1 != res2 {
    if pr>0 {(void) fmt.Printf("xor2"); Cudd_PrintDebug(dd,res2,nvars,pr);}
    Cudd_RecursiveDeref(dd,res1);
    Cudd_RecursiveDeref(dd,res2);
    return(0);
    }
    Cudd_RecursiveDeref(dd,res1);
    Cudd_RecursiveDeref(dd,f1);
    Cudd_RecursiveDeref(dd,f0);

    /* Compute boolean difference directly. */
    res1 = Cudd_bddBooleanDiff(dd,f,x);
    if res1 == nil {
    Cudd_RecursiveDeref(dd,res2);
    return(0);
    }
    Cudd_Ref(res1);

    if res1 != res2 {
    if pr>0 {(void) fmt.Printf("xor3"); Cudd_PrintDebug(dd,res1,nvars,pr);}
    Cudd_RecursiveDeref(dd,res1);
    Cudd_RecursiveDeref(dd,res2);
    return(0);
    }
    Cudd_RecursiveDeref(dd,res1);
    Cudd_RecursiveDeref(dd,res2);
    return(1);

} /* end of testXor */


/**Function********************************************************************

  Synopsis    [Tests the Hamming distance functions.]

  Description [Tests the Hammming distance functions. Returns
  1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func testHamming(dd *DdManager, f *DdNode, pr int32) int32 {
    var vars **DdNode;
    var minBdd *DdNode;
    var zero *DdNode;
    var scan *DdNode;
    var i int32;
    var d int32;
    var minterm *int32;
    var size int32  = Cudd_ReadSize(dd);

    vars = ALLOC(DdNode *, size);
    if vars == nil { return(0); }
    for i = 0; i < size; i++ {
    vars[i] = Cudd_bddIthVar(dd,i);
    }

    minBdd = Cudd_bddPickOneMinterm(dd,Cudd_Not(f),vars,size);
    Cudd_Ref(minBdd);
    if pr > 0 {
    (void) fmt.Printf("Chosen minterm for Hamming distance test: ");
    Cudd_PrintDebug(dd,minBdd,size,pr);
    }

    minterm = ALLOC(int,size);
    if minterm == nil {
    FREE(vars);
    Cudd_RecursiveDeref(dd,minBdd);
    return(0);
    }
    scan = minBdd;
    zero = Cudd_Not(DD_ONE(dd));
    for !Cudd_IsConstant(scan) {
    var R *DdNode  = Cudd_Regular(scan);
    var T *DdNode  = Cudd_T(R);
    var E *DdNode  = Cudd_E(R);
    if R != scan {
        T = Cudd_Not(T);
        E = Cudd_Not(E);
    }
    if T == zero {
        minterm[R.index] = 0;
        scan = E;
    } else {
        minterm[R.index] = 1;
        scan = T;
    }
    }
    Cudd_RecursiveDeref(dd,minBdd);

    d = Cudd_MinHammingDist(dd,f,minterm,size);

    (void) fmt.Printf("Minimum Hamming distance = %d\n", d);

    FREE(vars);
    FREE(minterm);
    return(1);

} /* end of testHamming */
