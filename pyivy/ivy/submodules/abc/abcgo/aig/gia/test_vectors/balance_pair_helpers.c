#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern void Gia_ManPrepareLastTwo(Gia_Man_t *,Vec_Int_t *);
extern int Gia_ManFindSharedNode(Gia_Man_t *,Vec_Int_t *,int);
extern void Gia_ManSuperCollect(Gia_Man_t *,Gia_Obj_t *,int);
static int level(int mode,int i){return mode==0?0:mode==1?i%3:mode==2?i/3:(12-i)%4;}
static void pair_helpers(void)
{
    int mode,i,j,k,n,rotation,phase,value,first,chosen,inputs[12],array[12];
    for(mode=0;mode<4;mode++){
        Gia_Man_t *g=Gia_ManStart(256);Vec_Int_t *v=Vec_IntAlloc(12),*copy=Vec_IntAlloc(12);
        for(i=0;i<12;i++)inputs[i]=Gia_ManAppendCi(g);
        g->vLevels=Vec_IntStart(g->nObjsAlloc);
        for(i=0;i<12;i++)Gia_ObjSetLevel(g,Gia_ManObj(g,inputs[i]>>1),level(mode,i));
        Gia_ManHashStart(g);
        for(i=0;i<12;i++)for(j=i+1;j<12;j++)for(phase=0;phase<4;phase++)if((i+j+phase)%3==0){
            value=Gia_ManHashAnd(g,inputs[i]^(phase&1),inputs[j]^((phase>>1)&1));Gia_ObjSetGateLevel(g,Gia_ManObj(g,value>>1));
        }
        for(n=2;n<=12;n++)for(rotation=0;rotation<12;rotation++){
            for(i=0;i<n;i++)array[i]=inputs[(rotation+i)%12]^((rotation+i)&1);
            for(i=1;i<n;i++){value=array[i];j=i;while(j>0 && Gia_ObjLevelId(g,array[j-1]>>1)<Gia_ObjLevelId(g,value>>1)){array[j]=array[j-1];j--;}array[j]=value;}
            Vec_IntClear(v);for(i=0;i<n;i++)Vec_IntPush(v,array[i]);Gia_ManPrepareLastTwo(g,v);
            printf("prepare %d %d %d",mode,n,rotation);Vec_IntForEachEntry(v,value,k)printf(" %d",value);puts("");
            for(phase=0;phase<2;phase++){
                Vec_IntClear(copy);for(i=0;i<n;i++)Vec_IntPush(copy,array[i]);first=inputs[(rotation+3)%12]^phase;
                chosen=Gia_ManFindSharedNode(g,copy,first);printf("find %d %d %d %d %d",mode,n,rotation,phase,chosen);
                Vec_IntForEachEntry(copy,value,k)printf(" %d",value);puts("");
            }
        }
        Vec_IntFree(v);Vec_IntFree(copy);Gia_ManStop(g);
    }
}
static void collector_helpers(void)
{
    int mode,size,extra,leafRefs,strict,i,k,inputs[64],base,root,value;
    for(mode=0;mode<2;mode++)for(size=8;size<=64;size+=56)for(extra=0;extra<3;extra++)for(leafRefs=0;leafRefs<2;leafRefs++)for(strict=0;strict<2;strict++){
        Gia_Man_t *g=Gia_ManStart(256);for(i=0;i<size;i++)inputs[i]=Gia_ManAppendCi(g);
        root=base=mode?Gia_ManAppendXorReal(g,inputs[0],inputs[1]):Gia_ManAppendAnd(g,inputs[0],inputs[1]);
        for(i=2;i<size;i++)root=mode?Gia_ManAppendXorReal(g,root,inputs[i]):Gia_ManAppendAnd(g,root,inputs[i]);
        Gia_ManAppendCo(g,root);for(i=0;i<extra;i++)Gia_ManAppendCo(g,base);
        if(leafRefs){Gia_ManAppendCo(g,inputs[0]);Gia_ManAppendCo(g,inputs[1]);}
        Gia_ManCreateRefs(g);Gia_ManSuperCollect(g,Gia_ManObj(g,root>>1),strict);
        printf("collect %d %d %d %d %d",mode,size,extra,leafRefs,strict);Vec_IntForEachEntry(g->vSuper,value,k)printf(" %d",value);puts("");
        Gia_ManStop(g);
    }
}
int main(void){pair_helpers();collector_helpers();return 0;}
