#include "aig/gia/giaFx.c"

static Gia_Man_t *make_graph(int mode,int signs,int reorder)
{
    Gia_Man_t *g=Gia_ManStart(64);
    int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g),d=Gia_ManAppendCi(g);
    int x=Gia_ManAppendAnd(g,a^(signs&1),b);
    int y=Gia_ManAppendAnd(g,x,c^((signs>>1)&1));
    int z=Gia_ManAppendAnd(g,x,d^((signs>>2)&1));
    int w=Gia_ManAppendAnd(g,y^((signs>>3)&1),z);
    int roots[4]={x>>1,y>>1,z>>1,w>>1};
    int cuts[4][4]={{a>>1,b>>1},{x>>1,c>>1},{x>>1,d>>1},{y>>1,z>>1}};
    int sizes[4]={2,2,2,2},i,j,tmp;
    if(mode==1) {cuts[1][0]=a>>1;cuts[1][1]=b>>1;cuts[1][2]=c>>1;sizes[1]=3;cuts[2][0]=a>>1;cuts[2][1]=b>>1;cuts[2][2]=d>>1;sizes[2]=3;}
    if(mode==2) {for(i=0;i<4;i++)cuts[3][i]=i+1;sizes[3]=4;}
    Gia_ManAppendCo(g,w^(signs&1));
    if(mode<2) {Gia_ManAppendCo(g,y);Gia_ManAppendCo(g,z^1);}
    if(reorder) {tmp=Vec_IntEntry(g->vCis,0);Vec_IntWriteEntry(g->vCis,0,Vec_IntEntry(g->vCis,3));Vec_IntWriteEntry(g->vCis,3,tmp);Gia_ObjSetCioId(Gia_ManCi(g,0),0);Gia_ObjSetCioId(Gia_ManCi(g,3),3);}
    g->vMapping=Vec_IntStart(Gia_ManObjNum(g));
    for(i=mode==2?3:mode;i<4;i++) {
        Vec_IntWriteEntry(g->vMapping,roots[i],Vec_IntSize(g->vMapping));
        Vec_IntPush(g->vMapping,sizes[i]);
        for(j=sizes[i]-1;j>=0;j--)Vec_IntPush(g->vMapping,cuts[i][j]);
        Vec_IntPush(g->vMapping,roots[i]);
    }
    return g;
}
static void print_truth(Gia_Man_t *g)
{
    int input,i,k; Gia_Obj_t *obj; unsigned mask[3]={0};
    for(input=0;input<16;input++) {
        Gia_ManConst0(g)->fMark0=0;
        Gia_ManForEachCi(g,obj,i) obj->fMark0=(input>>i)&1;
        Gia_ManForEachAnd(g,obj,i) obj->fMark0=(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj))&(Gia_ObjFanin1(obj)->fMark0^Gia_ObjFaninC1(obj));
        Gia_ManForEachCo(g,obj,k)mask[k]|=(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj))<<input;
    }
    for(i=0;i<Gia_ManCoNum(g);i++)printf(" %04x",mask[i]);
}
int main(int argc, char **argv)
{
    int mode,signs,reverse,reorder,i,j,limit; Gia_Man_t *p,*q;Vec_Wec_t *cubes;Vec_Str_t *phases;Vec_Int_t *row;
    for(mode=0;mode<3;mode++)for(signs=0;signs<16;signs++)for(reverse=0;reverse<2;reverse++)for(reorder=0;reorder<2;reorder++) {
        p=make_graph(mode,signs,reorder);
        cubes=Gia_ManFxRetrieve(p,&phases,reverse);
        printf("retrieve m%d s%d r%d o%d phases",mode,signs,reverse,reorder);
        for(i=0;i<Vec_StrSize(phases);i++)printf("%d",Vec_StrEntry(phases,i));
        for(i=0;i<Gia_ManObjNum(p);i++)printf(" v%u",Gia_ManObj(p,i)->Value);
        for(i=0;i<Vec_WecSize(cubes);i++){row=Vec_WecEntry(cubes,i);printf(" [");for(j=0;j<Vec_IntSize(row);j++)printf("%s%d",j?",":"",Vec_IntEntry(row,j));printf("]");}
        puts("");
        Vec_WecFree(cubes);Vec_StrFree(phases);Gia_ManStop(p);
        for(limit=0;limit<=4;limit+=4) {
            p=make_graph(mode,signs,reorder);q=Gia_ManPerformFx(p,limit,0,reverse,0,0);
            /* Explicit corrected mode for the known Gia_ManDup port-order
               defect. Dup writes CO Value scratch; rebuilding leaves it ~0. */
            if(argc>1 && reorder && Gia_ManCo(p,0)->Value!=~0u) {
                Gia_ManDupRemapCis(q,p);Gia_ManDupRemapCos(q,p);
            }
            printf("driver m%d s%d r%d o%d limit%d ci%d co%d and%d",mode,signs,reverse,reorder,limit,Gia_ManCiNum(q),Gia_ManCoNum(q),Gia_ManAndNum(q));
            print_truth(q);puts("");Gia_ManStop(q);Gia_ManStop(p);
        }
    }
    return 0;
}
