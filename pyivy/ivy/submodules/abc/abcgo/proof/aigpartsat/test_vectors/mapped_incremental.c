#include "aig/aig/aig.h"
#include "sat/bsat/satStore.h"
#include "sat/cnf/cnf.h"
unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f,int n,unsigned char *s) { abort(); }
char *vnsprintf(const char *f,va_list a) { abort(); }
extern int Aig_ManAddNewCnfToSolver(sat_solver *,Aig_Man_t *,Vec_Int_t *,Vec_Int_t *,Vec_Ptr_t *,int);
extern void Aig_ManPartResetNodePolarity(Aig_Man_t *);
/* This fixture selects fSynthesize=0. Trap the unreachable optional rewrite
   callback rather than linking its unrelated sequential/timing subsystem. */
Aig_Man_t *__wrap_Dar_ManRwsat(Aig_Man_t *p,int balance,int verbose) { abort(); }
extern int Aig_ManPartitionedSat(Aig_Man_t *,int,int,int,int,int,int,int);
static void zero_budget(void)
{
    Aig_Man_t *g=Aig_ManStart(100);
    Aig_Obj_t *a=Aig_ObjCreateCi(g),*b=Aig_ObjCreateCi(g);
    Aig_Obj_t *x=Aig_And(g,a,b),*y=Aig_And(g,Aig_Not(a),b);
    int status,valid=0,*model;
    Aig_ObjCreateCo(g,x); Aig_ObjCreateCo(g,y);
    status=Aig_ManPartitionedSat(g,2,1,0,0,0,0,0);
    model=(int *)g->pData;
    if (status==0 && model) valid=(model[0]&&model[1])&&(!model[0]&&model[1]);
    printf("source status %d witness valid %d\n",status,valid);
    ABC_FREE(g->pData);
    Aig_ManStop(g); Cnf_ManFree();
}
extern Vec_Ptr_t *Aig_ManPartSplit(Aig_Man_t *,Vec_Int_t *,Vec_Ptr_t **,Vec_Ptr_t **);
static void sparse_labels(void)
{
    Aig_Man_t *g=Aig_ManStart(100);
    Aig_Obj_t *a=Aig_ObjCreateCi(g),*b=Aig_ObjCreateCi(g),*root=Aig_And(g,a,b);
    Vec_Int_t *labels;
    Vec_Ptr_t *parts,*pio,*outputs;
    int i;
    Aig_ObjCreateCo(g,root);
    labels=Vec_IntStart(Aig_ManObjNumMax(g)); Vec_IntWriteEntry(labels,Aig_ObjId(root),7);
    parts=Aig_ManPartSplit(g,labels,&pio,&outputs);
    printf("source parts %d output groups %d first outputs %d\n",Vec_PtrSize(parts),Vec_PtrSize(outputs),Vec_PtrSize((Vec_Ptr_t *)Vec_PtrEntry(outputs,0)));
    for (i=0;i<Vec_PtrSize(parts);i++) Aig_ManStop((Aig_Man_t *)Vec_PtrEntry(parts,i));
    Vec_PtrFree(parts); Vec_VecFree((Vec_Vec_t *)pio); Vec_VecFree((Vec_Vec_t *)outputs); Vec_IntFree(labels); Aig_ManStop(g);
}
int main(int argc,char **argv)
{
    if (argc==2 && !strcmp(argv[1],"zero-budget")) { zero_budget(); sparse_labels(); return 0; }
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
