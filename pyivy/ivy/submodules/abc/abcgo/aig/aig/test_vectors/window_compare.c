#include "aig/aig/aig.h"
unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f,int n,unsigned char *s) { abort(); }
char *vnsprintf(const char *f,va_list a) { abort(); }
int main(void)
{
    int cases,trial,i,a,b,rootId,size,limit,k,j,sizes[]={2,3,4,6},limits[]={0,1,3,100};
    assert(scanf("%d",&cases)==1);
    for (trial=0;trial<cases;trial++)
    {
        Aig_Man_t *p=Aig_ManStart(100);
        Aig_Obj_t *pool[15],*root,*o;
        Vec_Ptr_t *front=Vec_PtrAlloc(10),*visited=Vec_PtrAlloc(20);
        for (i=0;i<4;i++) pool[i]=Aig_ObjCreateCi(p);
        pool[4]=Aig_And(p,pool[0],pool[1]);
        for (i=0;i<10;i++) { assert(scanf("%d %d",&a,&b)==2); pool[5+i]=Aig_And(p,Aig_NotCond(pool[a/2],a&1),Aig_NotCond(pool[b/2],b&1)); }
        root=pool[4];
        for (i=0;i<15;i++) if (Aig_ObjIsAnd(Aig_Regular(pool[i]))) root=Aig_Regular(pool[i]);
        Aig_ObjCreateCo(p,root);
        for (i=0;i<4;i++) { assert(scanf("%d",&a)==1); Aig_ObjCreateCo(p,Aig_NotCond(pool[a/2],a&1)); }
        rootId=Aig_ObjId(root);
        for (k=0;k<4;k++) for (j=0;j<4;j++)
        {
            size=sizes[k]; limit=limits[j];
            Aig_ManFindCut(root,front,visited,size,limit);
            printf("case %d root %d size %d limit %d front:",trial,rootId,size,limit);
            Vec_PtrForEachEntry(Aig_Obj_t *,front,o,i) printf(" %d",Aig_ObjId(o));
            printf(" visited:");
            Vec_PtrForEachEntry(Aig_Obj_t *,visited,o,i) printf(" %d",Aig_ObjId(o));
            printf(" marks:");
            Aig_ManForEachObj(p,o,i) if (o->fMarkA) printf(" %d",i);
            puts("");
        }
        Vec_PtrFree(front); Vec_PtrFree(visited); Aig_ManStop(p);
    }
    return 0;
}
