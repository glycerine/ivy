#include "aig/gia/giaIf.c"
#include <sys/resource.h>

static Gia_Man_t *build(int variant)
{
    Gia_Man_t *p=Gia_ManStart(64);int a,b,c,d,roots[4],leaves[4][3],sizes[4]={3,3,2,2},i,k;
    a=Gia_ManAppendCi(p);b=Gia_ManAppendCi(p);c=Gia_ManAppendCi(p);d=Gia_ManAppendCi(p);
    roots[0]=Gia_ManAppendMux(p,a^(variant&1),b^((variant>>1)&1),c^((variant>>2)&1));
    roots[1]=Gia_ManAppendMux(p,d,roots[0],c);
    roots[2]=Gia_ManAppendAnd(p,roots[0],b);
    roots[3]=Gia_ManAppendAnd(p,roots[2],d);
    leaves[0][0]=a>>1;leaves[0][1]=b>>1;leaves[0][2]=c>>1;
    leaves[1][0]=d>>1;leaves[1][1]=roots[0]>>1;leaves[1][2]=c>>1;
    leaves[2][0]=roots[0]>>1;leaves[2][1]=b>>1;
    leaves[3][0]=roots[2]>>1;leaves[3][1]=d>>1;
    for(i=0;i<4;i++)Gia_ManAppendCo(p,roots[i]^(i&1));Gia_ManAppendCo(p,a);
    p->pName=Abc_UtilStrsav("mapping_A");p->vMapping=Vec_IntStart(Gia_ManObjNum(p));
    for(i=0;i<4;i++){
        Vec_IntWriteEntry(p->vMapping,roots[i]>>1,Vec_IntSize(p->vMapping));Vec_IntPush(p->vMapping,sizes[i]);
        for(k=0;k<sizes[i];k++)Vec_IntPush(p->vMapping,leaves[i][k]);Vec_IntPush(p->vMapping,i<3?-(roots[i]>>1):roots[i]>>1);
    }
    p->vPacking=Vec_IntAlloc(16);Vec_IntPush(p->vPacking,variant%4);
    for(i=1;i<=variant%4;i++){Vec_IntPush(p->vPacking,i);for(k=0;k<i;k++)Vec_IntPush(p->vPacking,roots[k]>>1);}
    return p;
}
int main(int argc,char **argv)
{
    struct rlimit core={0,0};Gia_Man_t *p;Gia_Obj_t *obj;int mode,variant,i,fs[3],counts[33];
    setrlimit(RLIMIT_CORE,&core);setvbuf(stdout,NULL,_IONBF,0);if(argc<3)return 2;mode=atoi(argv[1]);variant=atoi(argv[2]);
    if(mode==1){for(i=0;i<33;i++)counts[i]=(i*17+variant*13)%37;Gia_ManPrintNodeProfile(counts,variant%33);return 0;}
    p=build(variant);
    if(mode==2){
        assert(argc==4);Gia_ManPrintMappingStats(p,argv[3]);Gia_ManPrintMappingStats(p,argv[3]);
        ABC_FREE(p->pName);p->pName=Abc_UtilStrsav("mapping_B");Gia_ManPrintMappingStats(p,argv[3]);
        ABC_FREE(p->pName);p->pName=Abc_UtilStrsav("mapping_A");Gia_ManPrintMappingStats(p,argv[3]);
        Gia_ManStop(p);return 0;
    }
    if(mode==3){Vec_IntFreeP(&p->vMapping);p->vMapping=Vec_IntStart(Gia_ManObjNum(p));Gia_ManPrintLutStats(p);Gia_ManStop(p);return 0;}
    Gia_ManForEachLut(p,i)if(Gia_ObjLutSize(p,i)==3){Gia_ManPrintGetMuxFanins(p,Gia_ManObj(p,i),fs);printf("mux %d %d %d %d\n",i,fs[0],fs[1],fs[2]);}
    printf("duplicates %d\n",Gia_ManCountDupLut(p));Gia_ManPrintMappingStats(p,NULL);Gia_ManPrintPackingStats(p);Gia_ManPrintLutStats(p);
    Gia_ManForEachObj(p,obj,i)assert(!obj->fMark0&&!obj->fMark1);
    Gia_ManStop(p);return 0;
}
