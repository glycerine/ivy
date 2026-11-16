#include "aig/gia/giaIf.c"
#include <sys/resource.h>
extern void Abc_Start(void);
extern void Abc_Stop(void);
int main(int argc,char **argv)
{
    struct rlimit core={0,0};Gia_Man_t *p,*q;Vec_Int_t *configs;Gia_Obj_t *obj;int a,b,c,x,root,i,mask;
    int configData[2]={0,1},arrivals[3]={7,11,13},required[1]={30},output[1]={91};
    setrlimit(RLIMIT_CORE,&core);setvbuf(stdout,NULL,_IONBF,0);if(argc!=2)return 2;Abc_Start();
    p=Gia_ManStart(16);p->pName=Abc_UtilStrsav("balance-review");p->pSpec=Abc_UtilStrsav("source-spec");
    a=Gia_ManAppendCi(p);b=Gia_ManAppendCi(p);c=Gia_ManAppendCi(p);x=Gia_ManAppendAnd(p,a,b^1);root=Gia_ManAppendAnd(p,x,c);Gia_ManAppendCo(p,root^1);
    configs=p->vConfigs=Vec_IntAllocArrayCopy(configData,2);p->pCellStr=Abc_UtilStrsav("retained-cell");
    p->vCiArrs=Vec_IntAllocArrayCopy(arrivals,3);p->vCoReqs=Vec_IntAllocArrayCopy(required,1);p->vCoArrs=Vec_IntAllocArrayCopy(output,1);
    if(atoi(argv[1])==3){
        If_Par_t pars;If_Man_t *manager;Gia_ManSetIfParsDefault(&pars);pars.nLutSize=4;manager=Gia_ManToIf(p,&pars);
        printf("if-name %s\n",manager->pName);ABC_FREE(p->pName);p->pName=Abc_UtilStrsav("changed");printf("after-source-rename %s\n",manager->pName);
        If_ManStop(manager);Gia_ManStop(p);Abc_Stop();return 0;
    }
    if(atoi(argv[1])==2){If_Par_t pars;Gia_ManSetIfParsDefault(&pars);pars.nLutSize=4;q=Gia_ManPerformMapping(p,&pars);}
    else q=atoi(argv[1])?Gia_ManPerformDsdBalance(p,6,8,0,0):Gia_ManPerformSopBalance(p,8,0,0);assert(q);
    printf("config-owner %d %s source-cleared %d\n",atoi(argv[1])==2?0:q->vConfigs==configs,q->pCellStr?q->pCellStr:"-",!p->vConfigs&&!p->pCellStr);
    printf("identity %s %s\n",q->pName,q->pSpec);
    printf("ci-arrivals");for(i=0;i<Vec_IntSize(q->vCiArrs);i++)printf(" %d",Vec_IntEntry(q->vCiArrs,i));puts("");
    printf("co-required %d co-arrival %d\n",Vec_IntEntry(q->vCoReqs,0),Vec_IntEntry(q->vCoArrs,0));
    printf("truth ");for(mask=0;mask<8;mask++){
        Gia_ManConst0(q)->fMark0=0;Gia_ManForEachCi(q,obj,i)obj->fMark0=(mask>>i)&1;
        Gia_ManForEachAnd(q,obj,i)obj->fMark0=(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj))&(Gia_ObjFanin1(obj)->fMark0^Gia_ObjFaninC1(obj));
        obj=Gia_ManCo(q,0);putchar('0'+(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj)));
    }puts("");Gia_ManStop(q);Gia_ManStop(p);Abc_Stop();return 0;
}
