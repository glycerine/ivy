#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManDupTopAnd_iter(Gia_Man_t *,int);
extern void Gia_ManPrintRepr(Gia_Man_t *);
extern int Gia_ManDupDfs2_rec(Gia_Man_t *,Gia_Man_t *,Gia_Obj_t *);

static Gia_Man_t *graph(int kind, int signs)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g), d = Gia_ManAppendCi(g);
    int u = Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    int v = Gia_ManAppendAnd(g,c^((signs>>2)&1),d^((signs>>3)&1));
    if (kind == 0 || kind == 1)
    {
        Gia_ManAppendCo(g,Gia_ManAppendAnd(g,u,v^1));
        if (kind == 1) Gia_ManAppendCo(g,a^1);
        else { Gia_ManAppendCo(g,v^1); Gia_ManAppendCo(g,1); }
    }
    else if (kind == 2) { Gia_ManAppendCo(g,u^1); Gia_ManAppendCo(g,v^1); }
    else if (kind == 3) { Gia_ManAppendCo(g,a^(signs&1)); Gia_ManAppendCo(g,b^((signs>>1)&1)); }
    else if (kind == 4)
    {
        int x = u, y = v, i;
        for (i = 0; i < 4; ++i) { int old = x; x = Gia_ManAppendAnd(g,x,y); y = Gia_ManAppendAnd(g,old,a); }
        Gia_ManAppendCo(g,x); Gia_ManAppendCo(g,x); Gia_ManAppendCo(g,1);
    }
    else { Gia_ManAppendCo(g,a); Gia_ManAppendCo(g,a^1); Gia_ManAppendCo(g,Gia_ManAppendAnd(g,a,b)); }
    Gia_ManAppendCi(g);
    g->pName = Abc_UtilStrsav("top"); g->pSpec = Abc_UtilStrsav("top.aig");
    return g;
}

static void values(Gia_Man_t *g, Gia_Man_t *out, int rows)
{
    Gia_Obj_t *obj;
    int i,mask,packed,*data = ABC_CALLOC(int,out->nObjs);
    Gia_ManForEachObj(g,obj,i) printf(" %u",obj->Value);
    for (mask = 0; mask < rows; ++mask)
    {
        Gia_ManForEachCi(out,obj,i) data[Gia_ObjId(out,obj)] = (mask>>i)&1;
        Gia_ManForEachAnd(out,obj,i)
            data[i] = (data[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj)) & (data[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
        packed = 0;
        Gia_ManForEachCo(out,obj,i) packed |= (data[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
        printf(" %d",packed);
    }
    puts(""); ABC_FREE(data);
}

static int sat_count(Gia_Man_t *g)
{
    Gia_Obj_t *obj;
    int i,mask,all,count=0,*data=ABC_CALLOC(int,g->nObjs);
    for (mask=0;mask<(1<<Gia_ManCiNum(g));++mask)
    {
        Gia_ManForEachCi(g,obj,i) data[Gia_ObjId(g,obj)]=(mask>>i)&1;
        Gia_ManForEachAnd(g,obj,i)
            data[i]=(data[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(data[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
        all=1;
        Gia_ManForEachCo(g,obj,i) all &= data[Gia_ObjFaninId0p(g,obj)]^Gia_ObjFaninC0(obj);
        count+=all;
    }
    ABC_FREE(data);return count;
}
static Gia_Man_t *representatives(int signs, int regs)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), x,y,z,u,zero,one,i;
    Gia_ManAppendCi(g); Gia_ManAppendCi(g);
    x = Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    y = Gia_ManAppendAnd(g,x^1,1); z = Gia_ManAppendAnd(g,y,1);
    u = Gia_ManAppendAnd(g,(a^(signs&1))^1,b^((signs>>1)&1));
    zero = Gia_ManAppendAnd(g,x,u); one = Gia_ManAppendAnd(g,zero^1,1);
    Gia_ManAppendCo(g,x^((signs>>2)&1)); Gia_ManAppendCo(g,y^((signs>>3)&1));
    Gia_ManAppendCo(g,z^((signs>>4)&1)); Gia_ManAppendCo(g,zero); Gia_ManAppendCo(g,one);
    Gia_ManSetRegNum(g,regs); Gia_ManSetPhase(g);
    g->pReprsOld = ABC_ALLOC(int,g->nObjs);
    for (i = 0; i < g->nObjs; ++i) g->pReprsOld[i] = -1;
    g->pReprsOld[y>>1] = x>>1; g->pReprsOld[z>>1] = y>>1;
    g->pReprsOld[zero>>1] = 0; g->pReprsOld[one>>1] = 0;
    return g;
}


int main(int argc, char **argv)
{
    int kind,signs,full,regs,i;
    Gia_Obj_t *obj;
    if (argc!=2) return 1;
    if (!strcmp(argv[1],"printer"))
    {
        Gia_Man_t *g=Gia_ManStart(1);
        for (i=0;i<3;++i) Gia_ManAppendCi(g);
        g->pReprsOld=ABC_FALLOC(int,g->nObjs);g->pReprsOld[3]=1;
        g->pReprs=ABC_CALLOC(Gia_Rpr_t,g->nObjs);
        for (i=0;i<g->nObjs;++i) Gia_ObjSetRepr(g,i,GIA_VOID);
        Gia_ObjSetRepr(g,3,2);Gia_ManPrintRepr(g);Gia_ManStop(g);return 0;
    }
    if (!strcmp(argv[1],"false"))
    {
        Gia_Man_t *g=Gia_ManStart(1),*out;
        int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g);
        Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,Gia_ManAppendAnd(g,a,b));
        out=Gia_ManDupTopAnd_iter(g,0);
        printf("%d %d\n",sat_count(g),sat_count(out));Gia_ManStop(out);Gia_ManStop(g);return 0;
    }
    if (!strcmp(argv[1],"top"))
    {
        for (kind=0;kind<6;++kind) for (signs=0;signs<16;++signs) for (full=0;full<2;++full)
        {
            Gia_Man_t *g=graph(kind,signs),*out=full?Gia_ManDupTopAnd(g,0):Gia_ManDupTopAnd_iter(g,0);
            printf("R %d %d %d %d %d %d",kind,signs,full,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out));
            values(g,out,32);Gia_ManStop(out);Gia_ManStop(g);
        }
        return 0;
    }
    if (!strcmp(argv[1],"classes"))
    {
        for (signs=0;signs<32;++signs) for (regs=0;regs<=4;regs+=2)
        {
            Gia_Man_t *g=representatives(signs,regs),*out=Gia_ManStart(g->nObjs),*temp;
            // giaDup.c's Classes entry passes a CO to an AND-only helper.
            // Compare its intended logic with the actual representative-aware
            // DFS helper and original cleanup, not that unusable entry point.
            Gia_ManFillValue(g);Gia_ManConst0(g)->Value=0;
            Gia_ManForEachCi(g,obj,i) obj->Value=Gia_ManAppendCi(out);
            Gia_ManHashAlloc(out);
            Gia_ManForEachCo(g,obj,i) Gia_ManDupDfs2_rec(out,g,obj);
            Gia_ManHashStop(out);Gia_ManSetRegNum(out,regs);
            out=Gia_ManCleanup(temp=out);Gia_ManStop(temp);
            printf("R %d %d %d",signs,regs,out->nObjs);values(g,out,16);
            Gia_ManStop(out);Gia_ManStop(g);
        }
        return 0;
    }
    return 1;
}
