#include "aig/gia/giaIf.c"
#include <sys/resource.h>

#define PTR_FIELDS(M) M(vCiArrs) M(vCoReqs) M(vCoArrs) M(vCoAttrs) M(vInArrs) M(vOutReqs) M(vNamesIn) M(vNamesOut) M(vConfigs) M(pManTime) M(pAigExtra) M(vRegClasses) M(vRegInits)
static Vec_Int_t *iv(int value){Vec_Int_t *v=Vec_IntAlloc(1);Vec_IntPush(v,value);return v;}
static Vec_Flt_t *fv(float value){Vec_Flt_t *v=Vec_FltAlloc(1);Vec_FltPush(v,value);return v;}
static Vec_Ptr_t *pv(const char *value){Vec_Ptr_t *v=Vec_PtrAlloc(1);Vec_PtrPush(v,Abc_UtilStrsav((char *)value));return v;}
static Gia_Man_t *make(int mask,int owner)
{
    Gia_Man_t *p=Gia_ManStart(1);int value=owner*100;
    if(mask&1)p->vCiArrs=iv(value+1);if(mask&2)p->vCoReqs=iv(value+2);if(mask&4)p->vCoArrs=iv(value+3);if(mask&8)p->vCoAttrs=iv(value+4);
    if(mask&16)p->vInArrs=fv(owner+0.25f);if(mask&32)p->vOutReqs=fv(owner+0.5f);
    if(mask&64)p->vNamesIn=pv(owner==1?"s-in":"d-in");if(mask&128)p->vNamesOut=pv(owner==1?"s-out":"d-out");
    if(mask&256)p->vConfigs=iv(value+9);if(mask&512)p->pCellStr=Abc_UtilStrsav(owner==1?"s":"d");
    if(mask&1024)p->pManTime=Tim_ManStart(owner,owner);
    if(mask&1){p->pAigExtra=Gia_ManStart(1);p->pAigExtra->pName=Abc_UtilStrsav(owner==1?"s-extra":"d-extra");}
    if(mask&2)p->vRegClasses=iv(value+11);if(mask&4)p->vRegInits=iv(value+12);
    p->And2Delay=owner*10+3;p->nAnd2Delay=owner*10+7;p->DefInArrs=owner+0.25f;p->DefOutReqs=owner+0.5f;return p;
}
static int intv(Vec_Int_t *v){return v?Vec_IntEntry(v,0):-1;}
static float fltv(Vec_Flt_t *v){return v?Vec_FltEntry(v,0):-1;}
static const char *ptrv(Vec_Ptr_t *v){return v?(const char *)Vec_PtrEntry(v,0):"-";}
static void print(const char *name,Gia_Man_t *p,Gia_Man_t *oldSrc,Gia_Man_t *oldDst)
{
    unsigned src=0,dst=0;int bit=0;
#define ORIGIN(f) if(p->f&&p->f==oldSrc->f)src|=1u<<bit;if(p->f&&p->f==oldDst->f)dst|=1u<<bit;bit++;
    PTR_FIELDS(ORIGIN)
#undef ORIGIN
    printf("%s %d %d %d %d %.2f %.2f %s %s %d %s %d %s %d %d %d %d %.2f %.2f origin %x %x\n",name,intv(p->vCiArrs),intv(p->vCoReqs),intv(p->vCoArrs),intv(p->vCoAttrs),fltv(p->vInArrs),fltv(p->vOutReqs),ptrv(p->vNamesIn),ptrv(p->vNamesOut),intv(p->vConfigs),p->pCellStr?p->pCellStr:"-",p->pManTime?Tim_ManCiNum(p->pManTime):-1,p->pAigExtra?p->pAigExtra->pName:"-",intv(p->vRegClasses),intv(p->vRegInits),p->And2Delay,p->nAnd2Delay,p->DefInArrs,p->DefOutReqs,src,dst);
}
static void release_metadata(Gia_Man_t *p)
{
    Vec_IntFreeP(&p->vCiArrs);Vec_IntFreeP(&p->vCoReqs);Vec_IntFreeP(&p->vCoArrs);Vec_IntFreeP(&p->vCoAttrs);
    Vec_FltFreeP(&p->vInArrs);Vec_FltFreeP(&p->vOutReqs);
    if(p->vNamesIn){free(Vec_PtrEntry(p->vNamesIn,0));Vec_PtrFree(p->vNamesIn);}if(p->vNamesOut){free(Vec_PtrEntry(p->vNamesOut,0));Vec_PtrFree(p->vNamesOut);}
    Vec_IntFreeP(&p->vConfigs);ABC_FREE(p->pCellStr);if(p->pManTime)Tim_ManStop(p->pManTime);if(p->pAigExtra)Gia_ManStop(p->pAigExtra);
    Vec_IntFreeP(&p->vRegClasses);Vec_IntFreeP(&p->vRegInits);
}
static void detach(Gia_Man_t *p)
{
#define CLEAR(f) p->f=NULL;
    PTR_FIELDS(CLEAR)
#undef CLEAR
    p->pCellStr=NULL;
}
int main(void)
{
    struct rlimit core={0,0};int mask,variant;setrlimit(RLIMIT_CORE,&core);
    for(mask=0;mask<2048;mask++)for(variant=0;variant<3;variant++){
        Gia_Man_t *src=make(mask,1),*dst=make(variant==1?2047:0,2),oldSrc=*src,oldDst=*dst,*target=variant==2?src:dst;
        Gia_ManTransferTiming(target,src);printf("case %d %d\n",mask,variant);print("dst",target,&oldSrc,&oldDst);print("src",src,&oldSrc,&oldDst);
        /* Own all fixture allocations independently of moves/overwrites, so
           even C's destructive self-transfer is tested without leaking. */
        detach(src);detach(dst);release_metadata(&oldSrc);release_metadata(&oldDst);Gia_ManStop(src);Gia_ManStop(dst);
    }
    return 0;
}
