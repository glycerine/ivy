#include "opt/dau/dau.h"
#include <stdio.h>
static char *pool[]={"a","!a","(ab)","[ab]","<abc>","!(a[bc])","(a!b)","[a(bc)]","CA{abc}","!E8{abc}"};
static void emit(int a,int b,int ca,int cb){
    int pa[3]={0,1,2},pb[3]={2,1,3};
    char *result=Dau_DsdMerge(pool[a],pa,pool[b],pb,ca,cb,4);
    printf("%s\t%s\t%d\t%d\t%s\n",pool[a],pool[b],ca,cb,result?result:"ERROR");
}
int main(int argc,char **argv){
    for(int a=0;a<10;a++)for(int b=0;b<10;b++)
        for(int ca=0;ca<2;ca++)for(int cb=0;cb<2;cb++)emit(a,b,ca,cb);
    unsigned state=861;
    for(int i=0;i<500;i++){
        state=state*1664525+1013904223;int a=state%10;
        state=state*1664525+1013904223;int b=state%10;
        state=state*1664525+1013904223;int ca=state>>31;
        state=state*1664525+1013904223;int cb=state>>31;
        emit(a,b,ca,cb);
    }
    return 0;
}
