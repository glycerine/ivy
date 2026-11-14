#include "aig/aig/aig.h"
unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f,int n,unsigned char *s) { abort(); }
char *vnsprintf(const char *f,va_list a) { abort(); }
static int literal(Aig_Obj_t *o) { return o?2*Aig_ObjId(Aig_Regular(o))+Aig_IsComplement(o):-1; }
int main(void)
{
    Aig_Man_t *g=Aig_ManStart(100);
    Aig_Obj_t *a=Aig_ObjCreateCi(g),*b=Aig_ObjCreateCi(g),*ghost,*raw;
    Aig_Type_t kinds[]={AIG_OBJ_CI,AIG_OBJ_CO,AIG_OBJ_BUF,AIG_OBJ_AND,AIG_OBJ_EXOR};
    Aig_Obj_t *left[]={NULL,Aig_ManConst1(g),Aig_Not(a),b,Aig_Not(b)};
    Aig_Obj_t *right[]={NULL,NULL,NULL,Aig_Not(a),a};
    int i;
    printf("initial ghost %d children %d %d\n",Aig_ObjType(Aig_ManGhost(g)),literal(Aig_ObjChild0(Aig_ManGhost(g))),literal(Aig_ObjChild1(Aig_ManGhost(g))));
    for (i=0;i<5;i++)
    {
        ghost=Aig_ObjCreateGhost(g,left[i],right[i],kinds[i]);
        printf("ghost %d kind %d children %d %d\n",i,Aig_ObjType(ghost),literal(Aig_ObjChild0(ghost)),literal(Aig_ObjChild1(ghost)));
    }
    printf("terminals: %d %d %d %d\n",Aig_ObjChild0(Aig_ManConst1(g))==NULL,Aig_ObjChild1(Aig_ManConst1(g))==NULL,Aig_ObjChild0(a)==NULL,Aig_ObjChild1(a)==NULL);
    raw=Aig_ManFetchMemory(g);
    printf("reserved type %d none %d count %d slots %d children %d %d\n",Aig_ObjType(raw),Aig_ObjIsNone(raw),Aig_ManObjNum(g),Aig_ManObjNumMax(g),literal(Aig_ObjChild0(raw)),literal(Aig_ObjChild1(raw)));
    raw->Type=AIG_OBJ_BUF; raw->pFanin0=Aig_ManConst1(g);
    Aig_ObjClean(raw);
    printf("cleared type %d none %d count %d children %d %d\n",Aig_ObjType(raw),Aig_ObjIsNone(raw),Aig_ManObjNum(g),literal(Aig_ObjChild0(raw)),literal(Aig_ObjChild1(raw)));
    Aig_ManRecycleMemory(g,raw);
    printf("recycled type %d none %d count %d slots %d\n",Aig_ObjType(raw),Aig_ObjIsNone(raw),Aig_ManObjNum(g),Aig_ManObjNumMax(g));
    Aig_ManStop(g);
    return 0;
}
