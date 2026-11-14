#include "aig/aig/aig.h"
extern void Aig_MuxTest(void);

/* Console environment only; the experiment and all graph operations below
   are linked from the original implementation. */
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

/* Observe the final graph at the source's BLIF export boundary. The Go test
   separately checks that its public entry point really writes the BLIF file. */
void __wrap_Aig_ManDumpBlif(Aig_Man_t *p, char *name, Vec_Ptr_t *ins, Vec_Ptr_t *outs)
{
    Aig_Obj_t *o;
    int i;
    assert(!strcmp(name, "test1.blif") && !ins && !outs);
    Aig_ManForEachObj(p, o, i)
    {
        if (Aig_ObjIsAnd(o))
            printf("and %d %d %d\n", i,
                Aig_ObjFaninId0(o)*2 + Aig_ObjFaninC0(o),
                Aig_ObjFaninId1(o)*2 + Aig_ObjFaninC1(o));
        else if (Aig_ObjIsCo(o))
            printf("out %d %d\n", i, Aig_ObjFaninId0(o)*2 + Aig_ObjFaninC0(o));
    }
}

int main(void)
{
    Aig_MuxTest();
    return 0;
}
