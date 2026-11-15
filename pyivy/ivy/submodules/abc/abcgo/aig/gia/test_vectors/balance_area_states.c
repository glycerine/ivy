/* Include the source to inspect its private DAM state, not a reimplementation. */
#include "aig/gia/giaBalAig.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static Gia_Man_t *area_graph(int family,int signs)
{
    int subsets[4][3]={{0,1,2},{1,3,2},{0,3,2},{0,1,3}},ci[6],roots[8],nr=family==0 || family==2?2:4,i,j,a,b,c,x,mux,buf,unused;
    Gia_Man_t *g=Gia_ManStart(256);
    for(i=0;i<6;i++)ci[i]=Gia_ManAppendCi(g);
    g->pMuxes=ABC_CALLOC(unsigned,g->nObjsAlloc);Gia_ManHashStart(g);
    for(i=0;i<nr;i++){
        int xorGate=family==2 || (family==3 && (i&1));
        a=ci[subsets[i][0]]^((signs>>((subsets[i][0]+i)%3))&1);
        b=ci[subsets[i][1]]^((signs>>((subsets[i][1]+i)%3))&1);
        c=ci[subsets[i][2]]^((signs>>((subsets[i][2]+i)%3))&1);
        if(family==5){
            x=ci[i%6]^((signs>>((i%6)%3))&1);
            for(j=1;j<6;j++)x=Gia_ManHashAnd(g,x,ci[(i+j)%6]^((signs>>(((i+j)%6)%3))&1));
            roots[i]=x;
        }else if(xorGate){x=Gia_ManHashXorReal(g,a,b);roots[i]=Gia_ManHashXorReal(g,x,c);}
        else {x=Gia_ManHashAnd(g,a,b);roots[i]=Gia_ManHashAnd(g,x,c);}
    }
    if(family==4){
        buf=Gia_ManAppendBuf(g,roots[1]);mux=Gia_ManHashMuxReal(g,ci[4],roots[0],buf);
        unused=Gia_ManHashMuxReal(g,ci[5],ci[0],ci[3]);Gia_ManObj(g,unused>>1)->fMark0=1;
        roots[nr++]=mux;roots[nr++]=mux^1;roots[nr++]=buf;
    }
    for(i=0;i<nr;i++)Gia_ManAppendCo(g,roots[i]^((signs>>(i%3))&1));
    Gia_ManAppendCo(g,ci[4]^(signs&1));Gia_ManAppendCo(g,roots[0]^1);
    Gia_ManSetRegNum(g,2);Gia_ManHashStop(g);return g;
}
static unsigned float_bits(float v){unsigned bits;memcpy(&bits,&v,sizeof(bits));return bits;}
static void snapshot(Dam_Man_t *p,int step)
{
    int i,j,*set,hand;Gia_Obj_t *obj;
    printf("stage %d %d %d %d %d %d %d %08x sets",step,p->pGia->nObjs,p->nLevelMax,p->nAnds,p->nDivs,p->nGain,p->nGainX,float_bits(Vec_QueTopPriority(p->vQue)));
    Gia_ManForEachObj(p->pGia,obj,i){hand=Dam_ObjHand(p,i);if(!hand)continue;set=Dam_ObjSet(p,i);printf(" %d:%d",i,hand);for(j=0;j<=set[0];j++)printf(":%d",set[j]);}
    printf(" divs");for(i=1;i<=Hash_IntManEntryNum(p->vHash);i++){
        hand=Dam_DivHand(p,i);
        printf(" %d:%d:%d:%d:%08x:%d:%d:%d",i,Hash_IntObjData0(p->vHash,i),Hash_IntObjData1(p->vHash,i),Hash_IntObjData2(p->vHash,i),float_bits(Vec_FltEntry(p->vCounts,i)),Vec_QueIsMember(p->vQue,i),Vec_IntEntry(p->vDivLevR,i),hand);
        if(hand>0){set=Dam_DivSet(p,i);for(j=0;j<=set[0];j++)printf(":%d",set[j]);}
    }
    printf(" queue %d %d marks",Vec_QueSize(p->vQue),Vec_QueTop(p->vQue));
    Gia_ManForEachObj(p->pGia,obj,i)printf(" %d",obj->fMark0|(obj->fMark1<<1));
    printf(" visit");Vec_IntForEachEntry(p->vVisit,i,j)printf(" %d",i);
    printf(" levels");Gia_ManForEachObj(p->pGia,obj,i)printf(" %d",Gia_ObjLevel(p->pGia,obj));
    puts("");
}
static void result(Gia_Man_t *g,Gia_Man_t *out)
{
    int i,mask,packed,*values;Gia_Obj_t *obj;
    printf("result %d %d %d %d values",out->nObjs,out->nRegs,out->nXors,out->nMuxes);
    Gia_ManForEachObj(g,obj,i)printf(" %u",obj->Value);
    printf(" nodes");Gia_ManForEachObj(out,obj,i){
        if(Gia_ObjIsConst0(obj))printf(" 0");else if(Gia_ObjIsCi(obj))printf(" 1");
        else if(Gia_ObjIsCo(obj))printf(" 6:%d",Gia_ObjFaninLit0p(out,obj));
        else if(Gia_ObjIsBuf(obj))printf(" 5:%d",Gia_ObjFaninLit0p(out,obj));
        else if(Gia_ObjIsMux(out,obj))printf(" 4:%d:%d:%d",Gia_ObjFaninLit2p(out,obj),Gia_ObjFaninLit1p(out,obj),Gia_ObjFaninLit0p(out,obj));
        else printf(" %d:%d:%d",Gia_ObjIsXor(obj)?3:2,Gia_ObjFaninLit0p(out,obj),Gia_ObjFaninLit1p(out,obj));
    }
    printf(" truth");values=ABC_CALLOC(int,out->nObjs);
    for(mask=0;mask<64;mask++){
        Gia_ManForEachCi(out,obj,i)values[Gia_ObjId(out,obj)]=(mask>>i)&1;
        Gia_ManForEachAnd(out,obj,i){
            int a=values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj),b=values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj);
            if(Gia_ObjIsBuf(obj))values[i]=a;
            else if(Gia_ObjIsMux(out,obj))values[i]=(values[Gia_ObjFaninId2(out,i)]^Gia_ObjFaninC2(out,obj))?b:a;
            else if(Gia_ObjIsXor(obj))values[i]=a^b;else values[i]=a&b;
        }
        packed=0;Gia_ManForEachCo(out,obj,i)packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
        printf(" %d",packed);
    }
    puts("");ABC_FREE(values);
}
int main(void)
{
    int family,signs,arrival,li,limits[5]={0,1,2,4,8},levels[6]={0,2,1,3,0,1},i,step,div;Gia_Obj_t *obj;
    for(family=0;family<6;family++)for(signs=0;signs<8;signs++)for(arrival=0;arrival<2;arrival++)for(li=0;li<5;li++){
        Gia_Man_t *g=area_graph(family,signs),*out;Dam_Man_t *p=Dam_ManAlloc(g);Vec_Int_t *inputs=arrival?Vec_IntAllocArrayCopy(levels,6):NULL;
        p->nLevelMax=Gia_ManSetLevels(g,inputs);p->vNodLevR=Gia_ManReverseLevel(g);Vec_IntFillExtra(g->vLevels,3*g->nObjs/2,0);
        Gia_ManForEachObj(g,obj,i){if(!Gia_ObjIsMux(g,obj))obj->fMark0=1;obj->fMark1=i&1;}
        printf("case %d %d %d %d\n",family,signs,arrival,limits[li]);Dam_ManCreatePairs(p,0);snapshot(p,0);
        for(step=0;step<limits[li] && Vec_QueTopPriority(p->vQue)>=2;step++){
            div=Vec_QuePop(p->vQue);printf("extract %d\n",div);Dam_ManUpdate(p,div);snapshot(p,step+1);
        }
        out=Dam_ManMultiAig(p);result(g,out);if(inputs)Vec_IntFree(inputs);Dam_ManFree(p);Gia_ManStop(out);Gia_ManStop(g);
    }
    return 0;
}
