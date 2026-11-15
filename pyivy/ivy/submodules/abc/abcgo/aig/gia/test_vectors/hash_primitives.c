/* Expose the actual static hash key and MUX-recognition helpers. */
#include "aig/gia/giaHash.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static void static_helpers(void)
{
    int lits[12]={0,1,2,3,6,7,30,31,62,63,254,255},controls[7]={-1,0,1,2,63,126,127},sizes[4]={3,11,67,1009};
    int i,j,k,m,n=0,a,b,inputs[4],nodes[24];Gia_Man_t *g=Gia_ManStart(64);
    for(i=0;i<12;i++)for(j=0;j<12;j++)for(k=0;k<7;k++)for(m=0;m<4;m++)
        printf("key %d %d %d %d %d\n",lits[i],lits[j],controls[k],sizes[m],Gia_ManHashOne(lits[i],lits[j],controls[k],sizes[m]));
    for(i=0;i<4;i++)inputs[i]=Gia_ManAppendCi(g);
    for(i=0;i<4;i++)for(j=i+1;j<4;j++)for(k=0;k<4;k++)nodes[n++]=Gia_ManAppendAnd(g,inputs[i]^(k&1),inputs[j]^((k>>1)&1));
    for(i=0;i<n;i++)for(j=0;j<n;j++){
        Gia_Obj_t *x=Gia_ManObj(g,nodes[i]>>1),*y=Gia_ManObj(g,nodes[j]>>1),*yes,*no,*control;
        int opposing=0,xs[2]={Gia_ObjFaninLit0p(g,x),Gia_ObjFaninLit1p(g,x)},ys[2]={Gia_ObjFaninLit0p(g,y),Gia_ObjFaninLit1p(g,y)};
        for(a=0;a<2;a++)for(b=0;b<2;b++)if(xs[a]==(ys[b]^1))opposing=1;
        if(!opposing)continue;
        control=Gia_ObjRecognizeMuxTwo(x,y,&yes,&no);
        printf("recognize %d %d %d %d %d\n",i,j,Gia_ObjToLit(g,control),Gia_ObjToLit(g,yes),Gia_ObjToLit(g,no));
    }
    Gia_ManStop(g);
}
int main(void)
{
    int mode,i,j,k,a,b,c,x,y,z,t,r,lookup,pool[64],n,inputs[4],length,code,limit,position,value;
    for(mode=0;mode<4;mode++){
        Gia_Man_t *g=Gia_ManStart(64);Vec_Int_t *v=Vec_IntAlloc(5),*v2=Vec_IntAlloc(5),*v3=Vec_IntAlloc(5);
        for(i=0;i<4;i++)inputs[i]=Gia_ManAppendCi(g);
        if(mode==2)g->pMuxes=ABC_CALLOC(unsigned,g->nObjsAlloc);
        if(mode==3)g->fGiaSimple=1;else Gia_ManHashStart(g);
        pool[0]=0;pool[1]=1;n=2;
        for(i=0;i<4;i++){pool[n++]=inputs[i];pool[n++]=inputs[i]^1;}
        for(i=0;i<4;i++)for(j=i+1;j<4;j++)for(k=0;k<4;k++){
            x=Gia_ManHashAnd(g,inputs[i]^(k&1),inputs[j]^((k>>1)&1));pool[n++]=x;pool[n++]=x^1;
        }
        g->fAddStrash=mode==1;
        for(i=0;i<n;i++)for(j=0;j<n;j++){
            a=pool[i];b=pool[j];t=-2;
            if(mode!=3 || a<2 || b<2 || (a>>1)==(b>>1))t=Gia_ManHashAndTry(g,a,b);
            x=Gia_ManHashAnd(g,a,b);y=Gia_ManHashOr(g,a,b);z=Gia_ManHashXor(g,a,b);
            r=mode==2?Gia_ManHashXorReal(g,a,b):-2;lookup=-2;
            if(mode!=3 && a>=2 && b>=2 && (a>>1)!=(b>>1))lookup=Gia_ManHashLookup(g,Gia_ObjFromLit(g,a),Gia_ObjFromLit(g,b));
            printf("pair %d %d %d %d %d %d %d %d %d %d %llu %llu\n",mode,i,j,t,x,y,z,r,lookup,g->nObjs,(unsigned long long)g->nHashHit,(unsigned long long)g->nHashMiss);
        }
        for(i=0;i<10;i++)for(j=0;j<10;j++)for(k=0;k<10;k++){
            a=pool[i];b=pool[j];c=pool[k];x=Gia_ManHashMux(g,a,b,c);y=Gia_ManHashMaj(g,a,b,c);r=mode==2?Gia_ManHashMuxReal(g,a,b,c):-2;
            printf("triple %d %d %d %d %d %d %d %d %llu %llu\n",mode,i,j,k,x,y,r,g->nObjs,(unsigned long long)g->nHashHit,(unsigned long long)g->nHashMiss);
        }
        for(length=0,limit=1;length<=5;length++,limit*=6)for(code=0;code<limit;code++){
            Vec_IntClear(v);value=code;
            for(position=0;position<length;position++){Vec_IntPush(v,pool[value%6]);value/=6;}
            Vec_IntClear(v2);Vec_IntAppend(v2,v);Vec_IntClear(v3);Vec_IntAppend(v3,v);
            x=Gia_ManHashAndMulti(g,v);y=Gia_ManHashAndMulti2(g,v2);z=Gia_ManHashDualMiter(g,v3);
            printf("fold %d %d %d %d %d %d vector",mode,length,code,x,y,z);
            Vec_IntForEachEntry(v,a,i)printf(" %d",a);
            printf(" kept");Vec_IntForEachEntry(v2,a,i)printf(" %d",a);
            printf(" %d %llu %llu\n",g->nObjs,(unsigned long long)g->nHashHit,(unsigned long long)g->nHashMiss);
        }
        if(mode!=3){Gia_ManHashStop(g);Gia_ManHashStart(g);Gia_ManHashResize(g);}
        printf("profile %d\n",mode);Gia_ManHashProfile(g);
        Vec_IntFree(v);Vec_IntFree(v2);Vec_IntFree(v3);Gia_ManStop(g);
    }
    static_helpers();
    return 0;
}
