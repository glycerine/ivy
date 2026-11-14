#include "aig/aig/aig.h"
unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f,int n,unsigned char *s) { abort(); }
char *vnsprintf(const char *f,va_list a) { abort(); }

int main(void)
{
    Aig_Man_t *p=Aig_ManStart(100),*q;
    Aig_Obj_t *a=Aig_ObjCreateCi(p),*b=Aig_ObjCreateCi(p),*s=Aig_ObjCreateCi(p),*t=Aig_ObjCreateCi(p);
    Aig_Obj_t *x=Aig_And(p,a,s),*y=Aig_And(p,Aig_Not(a),t),*ns=Aig_Or(p,x,y),*nt=Aig_And(p,b,Aig_Not(s));
    Aig_Obj_t *po0=Aig_And(p,s,t),*po1=Aig_Or(p,a,Aig_Not(b)),*o,**map;
    int frames,flags,i,j,mask,*values,packed;
    p->pName=Abc_UtilStrsav("frame-matrix"); p->pSpec=Abc_UtilStrsav("source-spec");
    Aig_And(p,b,t); /* A dangling gate still receives a map before cleanup. */
    Aig_ObjCreateCo(p,po0); Aig_ObjCreateCo(p,po1); Aig_ObjCreateCo(p,ns); Aig_ObjCreateCo(p,nt);
    Aig_ManSetRegNum(p,2);
    for (frames=1;frames<=4;frames++) for (flags=0;flags<16;flags++)
    {
        q=Aig_ManFrames(p,frames,flags&1,(flags>>1)&1,(flags>>2)&1,(flags>>3)&1,&map);
        printf("case %d %d ci %d co %d regs %d true %d %d objects %d max %d name %s spec %s\n",frames,flags,Aig_ManCiNum(q),Aig_ManCoNum(q),Aig_ManRegNum(q),q->nTruePis,q->nTruePos,Aig_ManObjNum(q),Aig_ManObjNumMax(q),q->pName,q->pSpec);
        printf("map:");
        for (i=0;i<Aig_ManObjNumMax(p)*frames;i++)
        {
            if (!map[i]) printf(" -");
            else printf(" %c%d",Aig_ObjIsNone(Aig_Regular(map[i]))?'d':'l',2*Aig_ObjId(Aig_Regular(map[i]))+Aig_IsComplement(map[i]));
        }
        puts("");
        Aig_ManForEachObj(q,o,i)
        {
            printf("object %d refs %d level %d phase %d",i,Aig_ObjRefs(o),o->Level,o->fPhase);
            if (Aig_ObjIsCi(o)) printf(" ci %d",Aig_ObjCioId(o));
            else if (Aig_ObjIsCo(o)) printf(" co %d %d",Aig_ObjCioId(o),Aig_ObjFaninId0(o)*2+Aig_ObjFaninC0(o));
            else if (Aig_ObjIsAnd(o)) printf(" and %d %d",Aig_ObjFaninId0(o)*2+Aig_ObjFaninC0(o),Aig_ObjFaninId1(o)*2+Aig_ObjFaninC1(o));
            puts("");
        }
        values=ABC_CALLOC(int,Aig_ManObjNumMax(q));
        for (mask=0;mask<(1<<Aig_ManCiNum(q));mask++)
        {
            values[0]=1;
            Aig_ManForEachCi(q,o,i) values[Aig_ObjId(o)]=(mask>>i)&1;
            Aig_ManForEachNode(q,o,i) values[Aig_ObjId(o)]=(values[Aig_ObjFaninId0(o)]^Aig_ObjFaninC0(o))&(values[Aig_ObjFaninId1(o)]^Aig_ObjFaninC1(o));
            packed=0;
            Aig_ManForEachCo(q,o,j) packed|=(values[Aig_ObjFaninId0(o)]^Aig_ObjFaninC0(o))<<j;
            printf("truth %d %d\n",mask,packed);
        }
        ABC_FREE(values); ABC_FREE(map); Aig_ManStop(q);
    }
    Aig_ManStop(p);
    return 0;
}
