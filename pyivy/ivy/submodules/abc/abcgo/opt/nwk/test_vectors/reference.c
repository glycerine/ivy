/* Original active NWK graph, timing, matching and retiming algorithms. */
#include "opt/nwk/nwk.h"
#include "opt/nwk/nwkMerge.h"
static Nwk_Man_t * manager(void){Nwk_Man_t *p=calloc(1,sizeof(*p));p->vCis=Vec_PtrAlloc(0);p->vCos=Vec_PtrAlloc(0);p->vObjs=Vec_PtrAlloc(0);p->vTemp=Vec_PtrAlloc(0);p->nFanioPlus=2;p->pMemObjs=Aig_MmFlexStart();return p;}
int main(int argc,char **argv){int kind=atoi(argv[1]),seed=atoi(argv[2]),i;unsigned state=seed+1;Nwk_Man_t *p;Nwk_Obj_t *o,*a,*b;Vec_Ptr_t *cut;Nwk_Grf_t *graph;
 if(kind==2){graph=Nwk_ManGraphAlloc(16);for(i=0;i<36;i++){state=1664525*state+1013904223;int x=state%16;state=1664525*state+1013904223;int y=state%16;Nwk_ManGraphHashEdge(graph,x,y);}Nwk_ManGraphSolve(graph);printf("RESULT");for(i=0;i<Vec_IntSize(graph->vPairs);i++)printf(" %d",Vec_IntEntry(graph->vPairs,i));puts("");return 0;}
 p=manager();for(i=0;i<6;i++)Nwk_ManCreateCi(p,0);
 for(i=0;i<12;i++){state=1664525*state+1013904223;a=Nwk_ManObj(p,state%Vec_PtrSize(p->vObjs));state=1664525*state+1013904223;b=Nwk_ManObj(p,state%Vec_PtrSize(p->vObjs));o=Nwk_ManCreateNode(p,2,0);Nwk_ObjAddFanin(o,a);Nwk_ObjAddFanin(o,b);}
 for(i=0;i<4;i++){o=Nwk_ManCreateCo(p);Nwk_ObjAddFanin(o,Nwk_ManObj(p,17-i));}
 if(kind==3){printf("RESULT %.6g %d\n",Nwk_ManDelayTraceLut(p),Nwk_ManLevel(p));return 0;}
 cut=kind==0?Nwk_ManRetimeCutForward(p,4,0):Nwk_ManRetimeCutBackward(p,4,0);printf("RESULT");for(i=0;i<Vec_PtrSize(cut);i++)printf(" %d",Nwk_ObjId((Nwk_Obj_t*)Vec_PtrEntry(cut,i)));puts("");return 0;
}
unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void){return 0;}
int Gia_ManToBridgeText(FILE *f,int n,unsigned char *b){return fwrite(b,1,n,f);}
char *vnsprintf(const char *format,va_list args){va_list copy;int n;char *b;va_copy(copy,args);n=vsnprintf(NULL,0,format,copy);va_end(copy);b=malloc(n+1);vsnprintf(b,n+1,format,args);return b;}
