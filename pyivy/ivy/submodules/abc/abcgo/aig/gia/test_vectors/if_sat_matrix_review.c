#include "map/if/ifSat.c"
#include <sys/resource.h>
int main(int argc,char **argv)
{
    struct rlimit core={0,0};FILE *file;sat_solver *xy[7]={0};Vec_Int_t *scratch=Vec_IntAlloc(2048);char op,kind;int k,n,outputs,i,index=0,ret;unsigned set;word truth[32],bound,free;
    setrlimit(RLIMIT_CORE,&core);setvbuf(stdout,NULL,_IONBF,0);if(argc!=2)return 2;file=fopen(argv[1],"r");assert(file);
    while(fscanf(file," %c",&op)==1){
        if(op=='B'){
            sat_solver *solver;assert(fscanf(file," %c %d",&kind,&k)==2);
            if(kind=='X'){if(!xy[k])xy[k]=If_ManSatBuildXY(k);solver=xy[k];}
            else solver=If_ManSatBuildXYZ(k);
            printf("build %c %d %d %d\n",kind,k,sat_solver_nvars(solver),(int)solver->stats.clauses);
            if(kind!='X')If_ManSatUnbuild(solver);
        }else if(op=='C'){
            assert(fscanf(file,"%d %d %x %d",&k,&n,&set,&outputs)==4);
            for(i=0;i<Abc_TtWordNum(n);i++){unsigned long long value;assert(fscanf(file,"%llx",&value)==1);truth[i]=value;}
            if(!xy[k])xy[k]=If_ManSatBuildXY(k);
            bound=ABC_CONST(0x0123456789abcdef);free=ABC_CONST(0xfedcba9876543210);
            ret=If_ManSatCheckXY(xy[k],k,truth,n,set,(outputs&1)?&bound:NULL,(outputs&2)?&free:NULL,scratch);
            printf("check %d %d %016llx %016llx assumptions",index++,ret,(unsigned long long)bound,(unsigned long long)free);
            for(i=0;i<Vec_IntSize(scratch);i++)printf(" %d",Vec_IntEntry(scratch,i));puts("");
        }else if(op=='S'){
            unsigned found;
            assert(fscanf(file,"%d %d",&k,&n)==2);
            for(i=0;i<Abc_TtWordNum(n);i++){unsigned long long value;assert(fscanf(file,"%llx",&value)==1);truth[i]=value;}
            if(!xy[k])xy[k]=If_ManSatBuildXY(k);
            found=If_ManSatCheckXYall(xy[k],k,truth,n,scratch);
            printf("search %d %x assumptions",index++,found);
            for(i=0;i<Vec_IntSize(scratch);i++)printf(" %d",Vec_IntEntry(scratch,i));puts("");
        }else return 3;
    }
    fclose(file);for(k=2;k<=6;k++)If_ManSatUnbuild(xy[k]);Vec_IntFree(scratch);return 0;
}
