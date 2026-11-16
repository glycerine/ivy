#include "aig/gia/giaIf.c"
#include <sys/resource.h>
extern void Abc_Start(void);
extern void Abc_Stop(void);
static unsigned state;
static unsigned next(void){return state=state*1664525u+1013904223u;}
static Gia_Man_t *build(unsigned seed)
{
    Gia_Man_t *p=Gia_ManStart(64);int pool[10],roots[12],i,a,b,pa,pb,choices=seed&1,config[2]={0,1};state=seed;
    for(i=0;i<4;i++)pool[i]=Gia_ManAppendCi(p);
    for(i=0;i<6;i++){
        a=next()%(4+i);b=next()%(3+i);if(b>=a)b++;pa=(next()>>12)&1;pb=(next()>>12)&1;
        roots[2*i]=Gia_ManAppendAnd(p,pool[a]^pa,pool[b]^pb);
        roots[2*i+1]=pool[4+i]=Gia_ManAppendAnd(p,pool[a]^pa,pool[b]^pb);
    }
    if(choices){p->pSibls=ABC_CALLOC(int,p->nObjs+6);for(i=0;i<6;i++)p->pSibls[roots[2*i+1]>>1]=roots[2*i]>>1;}
    for(i=0;i<12;i++)if(!choices||(i&1))Gia_ManAppendCo(p,roots[i]^(i&1));
    Gia_ManSetRegNum(p,(seed>>2)%2);p->pName=Abc_UtilStrsav("driver-review");p->pSpec=Abc_UtilStrsav("driver-spec");
    p->vConfigs=Vec_IntAllocArrayCopy(config,2);p->pCellStr=Abc_UtilStrsav("driver-cell");
    p->vCoArrs=Vec_IntAlloc(Gia_ManCoNum(p));for(i=0;i<Gia_ManCoNum(p);i++)Vec_IntPush(p->vCoArrs,100+i);
    if(seed%3==0){p->vCiArrs=Vec_IntAlloc(4);for(i=0;i<4;i++)Vec_IntPush(p->vCiArrs,1+2*i);p->vCoReqs=Vec_IntAlloc(Gia_ManCoNum(p));for(i=0;i<Gia_ManCoNum(p);i++)Vec_IntPush(p->vCoReqs,50+i);}
    return p;
}
static void vector(const char *name,Vec_Int_t *v){int i;printf("%s",name);if(v)for(i=0;i<Vec_IntSize(v);i++)printf(" %d",Vec_IntEntry(v,i));puts("");}
static void report(Gia_Man_t *p)
{
    int i,k,mask,result;Gia_Obj_t *obj;
    printf("result %d %d %d %d %d %d %d\n",p->nObjs,Gia_ManCiNum(p),Gia_ManCoNum(p),p->nRegs,p->vMapping?Gia_ManLutNum(p):-1,p->MappedDelay,p->MappedArea);
    printf("identity %s %s\ncell %s\n",p->pName,p->pSpec,p->pCellStr?p->pCellStr:"-");
    vector("configs",p->vConfigs);vector("ci-arrivals",p->vCiArrs);vector("co-required",p->vCoReqs);vector("co-arrivals",p->vCoArrs);
    Gia_ManForEachObj(p,obj,i){
        if(i==0)puts("const0");else if(Gia_ObjIsCi(obj))printf("ci %d\n",Gia_ObjCioId(obj));
        else if(Gia_ObjIsCo(obj))printf("co %d\n",Gia_ObjFaninLit0p(p,obj));else printf("and %d %d\n",Gia_ObjFaninLit0p(p,obj),Gia_ObjFaninLit1p(p,obj));
    }
    /* Canonicalize only index offsets and unreachable duplicate records: Go
       already compacts those. Preserve every active root, leaf order and marker. */
    printf("mapping %s\n",p->vMapping?"present":"none");
    if(p->vMapping)for(i=1;i<Gia_ManObjNum(p);i++){
        int offset=Vec_IntEntry(p->vMapping,i),size,j;if(!offset)continue;size=Vec_IntEntry(p->vMapping,offset);
        printf("lut %d",i);for(j=0;j<size+2;j++)printf(" %d",Vec_IntEntry(p->vMapping,offset+j));puts("");
    }
    vector("packing",p->vPacking);
    printf("truth");for(mask=0;mask<16;mask++){
        Gia_ManConst0(p)->fMark0=0;Gia_ManForEachCi(p,obj,k)obj->fMark0=(mask>>k)&1;
        Gia_ManForEachAnd(p,obj,i)obj->fMark0=(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj))&(Gia_ObjFanin1(obj)->fMark0^Gia_ObjFaninC1(obj));
        result=0;Gia_ManForEachCo(p,obj,k)result|=(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj))<<k;printf(" %d",result);
    }puts("");
}
int main(int argc,char **argv)
{
    struct rlimit core={0,0};Gia_Man_t *p,*q;int mode;If_Par_t pars;
    setrlimit(RLIMIT_CORE,&core);setvbuf(stdout,NULL,_IONBF,0);if(argc!=3&&argc!=4)return 2;Abc_Start();mode=atoi(argv[2]);p=build((unsigned)strtoul(argv[1],NULL,10));
    if(mode<4){Gia_ManSetIfParsDefault(&pars);pars.nLutSize=3+((strtoul(argv[1],NULL,10)>>1)&1);pars.fHashMapping=mode&1;if(mode>=2){pars.fTruth=1;pars.fDeriveLuts=1;}q=Gia_ManPerformMapping(p,&pars);}
    else if(mode==4)q=Gia_ManPerformSopBalance(p,8,0,0);else q=Gia_ManPerformDsdBalance(p,6,8,0,0);
    assert(q);
    if(argc==4){int offset=q->vMapping?Vec_IntEntry(q->vMapping,0):0;printf("reserved-constant-record %d active-luts %d\n",offset?Vec_IntEntry(q->vMapping,offset):-1,q->vMapping?Gia_ManLutNum(q):-1);}
    report(q);Gia_ManStop(q);Gia_ManStop(p);Abc_Stop();return 0;
}
