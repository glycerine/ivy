#include "aig/aig/aig.h"
#include "sat/bsat/satStore.h"
#include "sat/cnf/cnf.h"
unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f,int n,unsigned char *s) { abort(); }
char *vnsprintf(const char *f,va_list a) { abort(); }
extern int Aig_ManAddNewCnfToSolver(sat_solver *,Aig_Man_t *,Vec_Int_t *,Vec_Int_t *,Vec_Ptr_t *,int);
extern void Aig_ManPartResetNodePolarity(Aig_Man_t *);
int main(void)
{
    int align,stage,i,j;
    for (align=0;align<2;align++)
    {
        Aig_Man_t *g=Aig_ManStart(100);
        Aig_Obj_t *root=Aig_ManConst1(g),*input;
        Vec_Int_t *pio=Vec_IntAlloc(7),*vars;
        Vec_Ptr_t *outputs=Vec_PtrAlloc(0);
        sat_solver *s=sat_solver_new();
        Sto_Man_t *store;
        Sto_Cls_t *clause;
        for (i=0;i<6;i++)
        {
            input=Aig_ObjCreateCi(g); Vec_IntPush(pio,Aig_ObjId(input));
            root=Aig_And(g,root,input);
        }
        Aig_ObjCreateCo(g,root); Vec_IntPush(pio,Aig_ObjId(root));
        if (!align) Aig_ManPartResetNodePolarity(g);
        vars=Vec_IntStart(Aig_ManObjNumMax(g));
        sat_solver_store_alloc(s);
        for (stage=0;stage<2;stage++)
        {
            assert(Aig_ManAddNewCnfToSolver(s,g,vars,pio,outputs,align)==0);
            printf("align %d stage %d vars %d map:",align,stage,sat_solver_nvars(s));
            for (i=0;i<Vec_IntSize(vars);i++) printf(" %d",Vec_IntEntry(vars,i));
            puts("");
        }
        store=(Sto_Man_t *)s->pStore;
        Sto_ManForEachClause(store,clause)
        {
            printf("clause:");
            for (j=0;j<(int)clause->nLits;j++) printf(" %d",clause->pLits[j]);
            puts("");
        }
        sat_solver_delete(s); Vec_IntFree(vars); Vec_IntFree(pio); Vec_PtrFree(outputs); Aig_ManStop(g);
    }
    Cnf_ManFree();
    return 0;
}
