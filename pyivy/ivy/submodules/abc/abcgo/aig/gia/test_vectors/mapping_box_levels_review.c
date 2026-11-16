#include "aig/gia/giaIf.c"
#include "aig/gia/giaTim.c"
#include <sys/resource.h>

static unsigned state;
static unsigned next(void){return state=state*1664525u+1013904223u;}
static Gia_Man_t *build(unsigned seed,int fault)
{
    Gia_Man_t *p=Gia_ManStart(256);Tim_Man_t *timing;Vec_Ptr_t *tables;Vec_Int_t *roots=Vec_IntAlloc(100);
    int pis=2+seed%3,boxes=2+seed%3,ins[4],outs[4],totalIns=0,totalOuts=0,ci[16],available[16],nAvailable=pis;
    int i,j,k,ciOffset,coOffset,root,a,b,depth;state=seed;
    for(i=0;i<boxes;i++){ins[i]=1+(seed+i)%3;outs[i]=1+(seed+2*i)%2;totalIns+=ins[i];totalOuts+=outs[i];}
    for(i=0;i<pis+totalOuts;i++)ci[i]=Gia_ManAppendCi(p);
    for(i=0;i<pis;i++)available[i]=ci[i];ciOffset=pis;
    for(i=0;i<boxes;i++){
        for(j=0;j<ins[i];j++){
            a=available[next()%nAvailable];if(fault&&i==0&&j==0)a=ci[pis+outs[0]];
            root=a;depth=1+(seed+i+j)%3;
            for(k=0;k<depth;k++){
                b=available[next()%nAvailable];if((b>>1)==(root>>1))b=available[((b>>1)==(available[0]>>1))?1:0];
                {int pa=next()&1,pb=next()&1;root=Gia_ManAppendAnd(p,root^pa,b^pb);}Vec_IntPush(roots,root>>1);
            }
            Gia_ManAppendCo(p,root);
        }
        for(j=0;j<outs[i];j++)available[nAvailable++]=ci[ciOffset++];
    }
    for(j=0;j<2;j++){
        root=ci[pis+totalOuts-1-j%totalOuts];
        for(k=0;k<2+j;k++){root=Gia_ManAppendAnd(p,root,ci[k%pis]);Vec_IntPush(roots,root>>1);}
        Gia_ManAppendCo(p,root^(j&1));
    }
    p->vMapping=Vec_IntStart(Gia_ManObjNum(p));
    Vec_IntForEachEntry(roots,root,i){Gia_Obj_t *obj=Gia_ManObj(p,root);Vec_IntWriteEntry(p->vMapping,root,Vec_IntSize(p->vMapping));Vec_IntPush(p->vMapping,2);Vec_IntPush(p->vMapping,Gia_ObjFaninId0p(p,obj));Vec_IntPush(p->vMapping,Gia_ObjFaninId1p(p,obj));Vec_IntPush(p->vMapping,root);}
    timing=Tim_ManStart(pis+totalOuts,totalIns+2);tables=Vec_PtrAlloc(boxes);ciOffset=pis;coOffset=0;
    for(i=0;i<boxes;i++){
        float *table=ABC_ALLOC(float,3+ins[i]*outs[i]);table[0]=i;table[1]=ins[i];table[2]=outs[i];
        for(j=0;j<ins[i]*outs[i];j++)table[3+j]=((seed+i+j)%4==0)?-(float)ABC_INFINITY:(float)((seed+j)%7)/2;
        Vec_PtrPush(tables,table);Tim_ManCreateBox(timing,coOffset,ins[i],ciOffset,outs[i],i,0);ciOffset+=outs[i];coOffset+=ins[i];
    }
    Tim_ManSetDelayTables(timing,tables);p->pManTime=timing;Vec_IntFree(roots);return p;
}
static void levels(Gia_Man_t *p,const char *name){int i;printf("%s",name);for(i=0;i<Gia_ManObjNum(p);i++)printf(" %d",Gia_ObjLevelId(p,i));puts("");}
int main(int argc,char **argv)
{
    struct rlimit core={0,0};Gia_Man_t *p;int repeat,luts,edges,depth,fault;
    setrlimit(RLIMIT_CORE,&core);setvbuf(stdout,NULL,_IONBF,0);if(argc!=3)return 2;fault=atoi(argv[2]);p=build((unsigned)strtoul(argv[1],NULL,10),fault);
    for(repeat=0;repeat<2;repeat++){
        Gia_ManLutParams(p,&luts,&edges,&depth);printf("params %d %d %d\n",luts,edges,depth);levels(p,"lut-levels");
        if(!fault){printf("choice %d\n",Gia_ManChoiceLevel(p));levels(p,"choice-levels");}
    }
    Gia_ManStop(p);return 0;
}
