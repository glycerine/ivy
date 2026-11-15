#include "aig/gia/giaFx.c"
#include "base/main/main.h"
#include <sys/resource.h>
static unsigned next(unsigned *state) {return *state=*state*1664525u+1013904223u;}
int main(void)
{
    int n,seed,phase,row,i,pos,lit,input;unsigned state,mask;char sop[64];Gia_Obj_t *obj;
    struct rlimit core={0,0};setrlimit(RLIMIT_CORE,&core);
    Abc_Start();
    for(n=3;n<=8;n++)for(seed=1;seed<=16;seed++)for(phase=0;phase<2;phase++) {
        Gia_Man_t *g=Gia_ManStart(128);Vec_Int_t *leaves=Vec_IntAlloc(n);
        state=seed;mask=seed&((1u<<n)-1);pos=0;
        for(row=0;row<3;row++) {
            sop[pos++]=row==0?'1':'0';sop[pos++]=row==0?'-':row==1?'1':'0';
            for(i=2;i<n;i++)sop[pos++]="-01"[(next(&state)>>16)%3];
            sop[pos++]=' ';sop[pos++]='0'+phase;sop[pos++]='\n';
        }
        sop[pos]=0;
        for(i=0;i<n;i++)Vec_IntPush(leaves,Gia_ManAppendCi(g)^((mask>>i)&1));
        Gia_ManHashAlloc(g);lit=Gia_ManFactorNode(g,sop,leaves);
        printf("n%d seed%d phase%d lit%d objs%d ",n,seed,phase,lit,Gia_ManObjNum(g));
        for(input=0;input<(1<<n);input++) {
            Gia_ManConst0(g)->fMark0=0;
            Gia_ManForEachCi(g,obj,i)obj->fMark0=(input>>i)&1;
            Gia_ManForEachAnd(g,obj,i)obj->fMark0=(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj))&(Gia_ObjFanin1(obj)->fMark0^Gia_ObjFaninC1(obj));
            putchar('0'+(Gia_ManObj(g,lit>>1)->fMark0^(lit&1)));
        }
        puts("");
        {
            Vec_Wec_t *cubes=Vec_WecAlloc(4);Vec_Str_t *phases=Vec_StrStart(n+1);Gia_Man_t *out;
            Gia_ManAppendCo(g,lit);Gia_ManObj(g,lit>>1)->Value=n;
            Vec_StrWriteEntry(phases,n,(1-phase)^(lit&1));
            for(row=0;row<3;row++) {
                Vec_Int_t *cube=Vec_WecPushLevel(cubes);Vec_IntPush(cube,n);
                for(i=0;i<n;i++) if(sop[row*(n+3)+i]!='-') {
                    int var=i==0?n+1:i;
                    int sign=(sop[row*(n+3)+i]=='0')^((mask>>i)&1);
                    Vec_IntPush(cube,2*var+sign);
                }
                Vec_IntSelectSort(Vec_IntArray(cube)+1,Vec_IntSize(cube)-1);
            }
            {Vec_Int_t *cube=Vec_WecPushLevel(cubes);Vec_IntPush(cube,n+1);Vec_IntPush(cube,0);}
            out=Gia_ManFxInsert(g,cubes,phases);
            printf("insert n%d seed%d phase%d and%d ",n,seed,phase,Gia_ManAndNum(out));
            for(input=0;input<(1<<n);input++) {
                Gia_ManConst0(out)->fMark0=0;
                Gia_ManForEachCi(out,obj,i)obj->fMark0=(input>>i)&1;
                Gia_ManForEachAnd(out,obj,i)obj->fMark0=(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj))&(Gia_ObjFanin1(obj)->fMark0^Gia_ObjFaninC1(obj));
                obj=Gia_ManCo(out,0);putchar('0'+(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj)));
            }
            puts("");Vec_WecFree(cubes);Vec_StrFree(phases);Gia_ManStop(out);
        }
        Vec_IntFree(leaves);Gia_ManStop(g);
    }
    Abc_Stop();
    return 0;
}
