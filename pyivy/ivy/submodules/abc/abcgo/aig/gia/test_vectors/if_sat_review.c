#include "map/if/ifSat.c"
#include <sys/resource.h>
int main(void)
{
    struct rlimit core={0,0};sat_solver *p;Vec_Int_t *scratch=Vec_IntAlloc(32);word truth,bound,free;int result,i;
    setrlimit(RLIMIT_CORE,&core);p=If_ManSatBuildXY(2);printf("build XY %d %d\n",sat_solver_nvars(p),(int)p->stats.clauses);
    truth=ABC_CONST(0x8080808080808080);bound=free=0;result=If_ManSatCheckXY(p,2,&truth,3,5,&bound,&free,scratch);
    printf("and-result %d assumptions",result);for(i=0;i<Vec_IntSize(scratch);i++)printf(" %d",Vec_IntEntry(scratch,i));puts("");
    truth=0;result=If_ManSatCheckXY(p,2,&truth,0,0,&bound,&free,scratch);printf("empty-partition %d\n",result);
    truth=ABC_CONST(0xaaaaaaaaaaaaaaaa);for(i=0;i<4;i++)if(i!=2){result=If_ManSatCheckXY(p,2,&truth,1,i,&bound,&free,scratch);printf("one-var-set%d %d\n",i,result);}
    truth=ABC_CONST(0x8080808080808080);bound=ABC_CONST(0x123456789abcdef0);result=If_ManSatCheckXY(p,2,&truth,3,5,&bound,NULL,scratch);printf("single-output %d %016llx\n",result,(unsigned long long)bound);
    truth=ABC_CONST(0xacacacacacacacac);bound=0x1234;free=0x5678;result=If_ManSatCheckXY(p,2,&truth,3,5,&bound,&free,scratch);printf("unsat-outputs %d %llx %llx\n",result,(unsigned long long)bound,(unsigned long long)free);
    If_ManSatUnbuild(p);p=If_ManSatBuildXYZ(2);printf("build XYZ %d %d\n",sat_solver_nvars(p),(int)p->stats.clauses);If_ManSatUnbuild(p);Vec_IntFree(scratch);return 0;
}
