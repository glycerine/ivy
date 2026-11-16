#include "map/if/ifSat.c"
static word nextWord(word *state) { return *state = *state * ABC_CONST(6364136223846793005) + ABC_CONST(1442695040888963407); }
int main(void)
{
    word state=7654321,tables[3];int k,trial,j,m,l,f,p,v,count,result,flipped;lit assumptions[193];
    for(k=2;k<=6;k++) {
        sat_solver *solver=If_ManSatBuildXYZ(k);l=1<<k;f=1<<(3*k-2);
        for(trial=0;trial<12;trial++) {
            for(j=0;j<3;j++)tables[j]=nextWord(&state);
            for(j=0;j<8;j++) {
                m=j==0?0:j==1?f-1:nextWord(&state)%f;
                for(flipped=0;flipped<2;flipped++) {
                    count=0;for(p=0;p<3;p++)for(v=0;v<l;v++)assumptions[count++]=Abc_Var2Lit(p*l+v,!((tables[p]>>v)&1));
                    v=4*(m>>(2*k))+((tables[0]>>(m%l))&1)+2*((tables[1]>>((m>>k)%l))&1);
                    result=((tables[2]>>v)&1)^flipped;
                    assumptions[count++]=Abc_Var2Lit(3*l+m,!result);
                    result=sat_solver_solve(solver,assumptions,assumptions+count,0,0,0,0);
                    printf("xyz %d %d %d %d %d\n",k,trial,m,flipped,result);
                }
            }
        }
        If_ManSatUnbuild(solver);
    }
    return 0;
}
