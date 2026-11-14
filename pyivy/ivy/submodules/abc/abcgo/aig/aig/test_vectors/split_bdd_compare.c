#include "aig/aig/aigSplit.c"

/* Standalone console hooks; all BDD and AIG algorithms are original source. */
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

static unsigned aig_truth(Aig_Man_t *g, Aig_Obj_t *output)
{
    Aig_Obj_t *o;
    unsigned result = 0;
    int mask, i;
    unsigned char *values = ABC_CALLOC(unsigned char, Aig_ManObjNumMax(g));
    for (mask = 0; mask < 8; mask++)
    {
        values[0] = 1;
        Aig_ManForEachCi(g, o, i) values[o->Id] = (mask >> i)&1;
        Aig_ManForEachNode(g, o, i)
            values[o->Id] = (values[Aig_ObjFaninId0(o)] ^ Aig_ObjFaninC0(o)) &
                             (values[Aig_ObjFaninId1(o)] ^ Aig_ObjFaninC1(o));
        result |= (values[Aig_ObjFaninId0(output)] ^ Aig_ObjFaninC0(output)) << mask;
    }
    free(values);
    return result;
}

int main(void)
{
    Aig_Man_t *g, *q;
    Aig_Obj_t *inputs[3], *root, *term, *output;
    DdManager *dd;
    DdNode *function, *cofactor;
    Vec_Ptr_t *subset, *cofactors, *values, *random;
    unsigned table, selection, truth;
    int i, mask, assignment[3];
    values = Vec_PtrAlloc(12);
    for (i = 1; i <= 12; i++) Vec_PtrPush(values, (void *)(ABC_PTRUINT_T)(2*i));
    Aig_ManRandom(1);
    random = Aig_ManVecRandSubset(values, 4);
    printf("subset:");
    for (i = 0; i < Vec_PtrSize(random); i++) printf(" %u", (unsigned)(ABC_PTRUINT_T)Vec_PtrEntry(random, i));
    puts(""); Vec_PtrFree(random); Vec_PtrFree(values);
    while (scanf("%u%u", &table, &selection) == 2)
    {
        g = Aig_ManStart(100);
        for (i = 0; i < 3; i++) inputs[i] = Aig_ObjCreateCi(g);
        root = Aig_ManConst0(g);
        for (mask = 0; mask < 8; mask++)
        {
            if (!(table & (1u << mask))) continue;
            term = Aig_ManConst1(g);
            for (i = 0; i < 3; i++) term = Aig_And(g, term, Aig_NotCond(inputs[i], !((mask >> i)&1)));
            root = Aig_Or(g, root, term);
        }
        Aig_ObjCreateCo(g, root);
        if (selection & 8) Aig_ObjCreateCo(g, Aig_Exor(g, inputs[0], inputs[2]));
        Aig_ManSetRegNum(g, (selection & 8) != 0);
        if (selection & 16)
        {
            Aig_ManRandom(1);
            q = Aig_ManSplit(g, 2, 0);
            assert(q != NULL);
            printf("split %u %u %d: ", table, selection, Aig_ManCoNum(q));
            Aig_ManForEachCo(q, output, i) printf("%u ", aig_truth(q, output));
            puts("");
            Aig_ManStop(q); Aig_ManStop(g);
            continue;
        }
        dd = Aig_ManBuildPoBdd(g, &function);
        subset = Vec_PtrAlloc(3);
        for (i = 2; i >= 0; i--) if (selection & (1u << i)) Vec_PtrPush(subset, Cudd_bddIthVar(dd, i));
        cofactors = Aig_ManCofactorBdds(g, subset, dd, function);
        q = Aig_ManConvertBddsToAigs(g, dd, cofactors);
        printf("%u %u %d %d:", table, selection, Vec_PtrSize(cofactors), Aig_ManCoNum(q));
        Vec_PtrForEachEntry(DdNode *, cofactors, cofactor, i)
        {
            truth = 0;
            for (mask = 0; mask < 8; mask++)
            {
                assignment[0] = mask&1; assignment[1] = (mask>>1)&1; assignment[2] = (mask>>2)&1;
                if (Cudd_Eval(dd, cofactor, assignment) == Cudd_ReadOne(dd)) truth |= 1u << mask;
            }
            printf(" %u", truth);
        }
        printf(" | ");
        Aig_ManForEachCo(q, output, i) printf("%u ", aig_truth(q, output));
        puts("");
        Cudd_RecursiveDeref(dd, function);
        Vec_PtrForEachEntry(DdNode *, cofactors, cofactor, i) Cudd_RecursiveDeref(dd, cofactor);
        Vec_PtrFree(subset); Vec_PtrFree(cofactors);
        assert(Cudd_CheckZeroRef(dd) == 0);
        Cudd_Quit(dd); Aig_ManStop(q); Aig_ManStop(g);
    }
    return 0;
}
