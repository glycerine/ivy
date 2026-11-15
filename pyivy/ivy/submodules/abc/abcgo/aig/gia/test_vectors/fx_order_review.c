#include "aig/gia/giaFx.c"
static void run(int test,int inputs,int start,int rows[][5],int sizes[],int n)
{
    Vec_Wec_t *cubes=Vec_WecAlloc(n);Vec_Int_t *first=NULL,*count=NULL,*order;int i,j;
    for(i=0;i<n;i++){Vec_Int_t *row=Vec_WecPushLevel(cubes);for(j=0;j<sizes[i];j++)Vec_IntPush(row,rows[i][j]);}
    order=Gia_ManFxTopoOrder(cubes,inputs,start,&first,&count);
    printf("case%d",test);
    if(!order) printf(" unchanged first%d count%d",first==NULL,count==NULL);
    else {
        printf(" order");for(i=0;i<Vec_IntSize(order);i++)printf(" %d",Vec_IntEntry(order,i));
        printf(" first");for(i=0;i<Vec_IntSize(first);i++)printf(" %d",Vec_IntEntry(first,i));
        printf(" count");for(i=0;i<Vec_IntSize(count);i++)printf(" %d",Vec_IntEntry(count,i));
    }
    puts("");Vec_IntFreeP(&order);Vec_IntFreeP(&first);Vec_IntFreeP(&count);Vec_WecFree(cubes);
}
int main(void)
{
    int a[][5]={{2,0},{3,2}},as[]={2,2};
    int b[][5]={{2,8},{3,0,2},{4,2},{4,1}},bs[]={2,3,2,2};
    int c[][5]={{3,14,4},{6,8},{7,0,2}},cs[]={3,2,3};
    int d[][5]={{0},{1,0}},ds[]={1,2};
    run(0,2,4,a,as,2);run(1,2,4,b,bs,4);run(2,3,5,c,cs,3);run(3,0,1,d,ds,2);run(4,0,0,d,ds,0);
    return 0;
}
