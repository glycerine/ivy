#include "aig/gia/giaIf.c"
#include <sys/resource.h>

static unsigned randomState;
static unsigned next_random(void){return randomState=randomState*1664525u+1013904223u;}
static void print_array(const char *name,int *data,int size){int i;printf("%s",name);for(i=0;i<size;i++)printf(" %d",data[i]);puts("");}
static Gia_Man_t *build(unsigned seed,int mode)
{
    Gia_Man_t *p=Gia_ManStart(64);int i,k,lits[16],roots[12],nRoots=12;
    randomState=seed;
    for(i=0;i<4;i++)lits[i]=Gia_ManAppendCi(p);
    for(i=0;i<6;i++){
        int a=(next_random()%(4+2*i)),b=(next_random()%(3+2*i));if(b>=a)b++;
        a=lits[a]^(next_random()&1);b=lits[b]^(next_random()&1);
        roots[2*i]=lits[4+2*i]=Gia_ManAppendAnd(p,a,b);
        roots[2*i+1]=lits[5+2*i]=Gia_ManAppendAnd(p,a,b);
    }
    for(i=0;i<nRoots;i++)Gia_ManAppendCo(p,roots[i]^(i&1));
    p->pSibls=ABC_CALLOC(int,Gia_ManObjNum(p));
    for(i=0;i<nRoots;i+=2)p->pSibls[roots[i+1]>>1]=roots[i]>>1;
    p->vMapping=Vec_IntStart(Gia_ManObjNum(p));
    for(i=0;i<nRoots;i++){
        Gia_Obj_t *obj=Gia_ManObj(p,roots[i]>>1);int leaves[4],nLeaves=4;
        for(k=0;k<4;k++)leaves[k]=k+1;
        if(mode%3==1||(mode%3==2&&(i&1))){leaves[0]=Gia_ObjFaninId0p(p,obj);leaves[1]=Gia_ObjFaninId1p(p,obj);nLeaves=2;}
        Vec_IntWriteEntry(p->vMapping,roots[i]>>1,Vec_IntSize(p->vMapping));Vec_IntPush(p->vMapping,nLeaves);
        for(k=0;k<nLeaves;k++)Vec_IntPush(p->vMapping,leaves[k]);Vec_IntPush(p->vMapping,roots[i]>>1);
    }
    p->pName=Abc_UtilStrsav("mapping_review");
    return p;
}
int main(int argc,char **argv)
{
    struct rlimit core={0,0};Gia_Man_t *p;int i,luts,edges,levels,*objectLevels=NULL;Vec_Str_t *label;Vec_Int_t *visited;
    setrlimit(RLIMIT_CORE,&core);setvbuf(stdout,NULL,_IONBF,0);if(argc<3)return 2;
    p=build((unsigned)strtoul(argv[1],NULL,10),atoi(argv[2]));
    if(argc>3&&atoi(argv[3])==1){
        Gia_Man_t *q=Gia_ManStart(256);int a=Gia_ManAppendCi(q),b=Gia_ManAppendCi(q),c=Gia_ManAppendCi(q),x=Gia_ManAppendAnd(q,a,b),y=Gia_ManAppendAnd(q,x,c);
        Gia_ManStop(p);p=q;for(i=0;i<200;i++)Gia_ManAppendCo(p,i<13?y:x);
        p->vMapping=Vec_IntStart(Gia_ManObjNum(p));
        Vec_IntWriteEntry(p->vMapping,x>>1,Vec_IntSize(p->vMapping));Vec_IntPush(p->vMapping,2);Vec_IntPush(p->vMapping,a>>1);Vec_IntPush(p->vMapping,b>>1);Vec_IntPush(p->vMapping,x>>1);
        Vec_IntWriteEntry(p->vMapping,y>>1,Vec_IntSize(p->vMapping));Vec_IntPush(p->vMapping,2);Vec_IntPush(p->vMapping,x>>1);Vec_IntPush(p->vMapping,c>>1);Vec_IntPush(p->vMapping,y>>1);
        Gia_ManPrintMappingStats(p,NULL);Gia_ManStop(p);return 0;
    }
    if(argc>3&&atoi(argv[3])==3){
        Vec_IntFreeP(&p->vMapping);printf("missing mapping %d\n",atoi(argv[2])?Gia_ManComputeOverlap2(p):Gia_ManComputeOverlap(p));Gia_ManStop(p);return 0;
    }
    if(argc>3&&atoi(argv[3])==2){
        p->vPacking=Vec_IntAlloc(1);Vec_IntPush(p->vPacking,0);Gia_ManPrintPackingStats(p);Gia_ManStop(p);return 0;
    }
    printf("metrics %d %d %d %d\n",Gia_ManLutNum(p),Gia_ManLutFaninCount(p),Gia_ManLutSizeMax(p),Gia_ManLutLevel(p,&objectLevels));
    print_array("levels",objectLevels,Gia_ManObjNum(p));ABC_FREE(objectLevels);
    Gia_ManLutParams(p,&luts,&edges,&levels);printf("params %d %d %d\n",luts,edges,levels);
    Gia_ManSetRefsMapped(p);print_array("refs",p->pRefs,Gia_ManObjNum(p));
    Gia_ManSetLutRefs(p);print_array("lutrefs",p->pLutRefs,Gia_ManObjNum(p));
    label=Vec_StrStart(Gia_ManObjNum(p));visited=Vec_IntAlloc(32);
    Gia_ManForEachLut(p,i){int count=Gia_ManComputeOverlap2One(p,i,label,visited);printf("cone %d %d %d",i,Gia_ManComputeOverlapOne(p,i),count);print_array(" visited",Vec_IntArray(visited),Vec_IntSize(visited));}
    printf("overlap %d %d\n",Gia_ManComputeOverlap(p),Gia_ManComputeOverlap2(p));
    printf("choice %d\n",Gia_ManChoiceLevel(p));print_array("choicelevels",Vec_IntArray(p->vLevels),Gia_ManObjNum(p));
    Gia_ManPrintMappingStats(p,argc>3?argv[3]:NULL);Gia_ManPrintLutStats(p);
    Vec_StrFree(label);Vec_IntFree(visited);Gia_ManStop(p);return 0;
}
