#include "proof/fraig/fraigInt.h"
#include <stdio.h>

int main(void)
{
    Fraig_Man_t manager = {0};
    Fraig_Node_t constant = {0}, a = {0}, b = {0};
    Fraig_HashTable_t table = {0};
    unsigned sim_a[8] = {0x80000000, 1, 0};
    unsigned sim_b[8] = {0, 1, 0};
    unsigned real[8] = {0}, temp[8] = {0};
    Fraig_Node_t *inputs[] = {&a, &b};
    Fraig_Node_t *nodes[] = {&constant, &a, &b};
    Fraig_NodeVec_t input_vec = {0}, node_vec = {0};
    Fraig_Node_t *bins[] = {&a};

    input_vec.nSize = input_vec.nCap = 2;
    input_vec.pArray = inputs;
    node_vec.nSize = node_vec.nCap = 3;
    node_vec.pArray = nodes;
    constant.NumPi = -1;
    constant.Num = 0;
    a.NumPi = 0;
    a.Num = 1;
    a.puSimD = sim_a;
    b.NumPi = 1;
    b.Num = 2;
    b.puSimD = sim_b;
    manager.vInputs = &input_vec;
    manager.vNodes = &node_vec;
    manager.vCones = Fraig_NodeVecAlloc(8);
    manager.mmSims = Fraig_MemFixedStart(48);
    table.nBins = 1;
    table.pBins = bins;
    a.pNextD = &b;
    manager.pTableF = &table;
    manager.nBTLimit = 0;
    manager.nWordsDyna = 8;
    manager.iWordStart = 3;
    manager.iWordPerm = 1;
    manager.iPatsPerm = 31;
    real[0] = 0x7fffffff;
    manager.pSimsReal = real;
    manager.pSimsTemp = temp;
    manager.vPatsReal = Msat_IntVecAlloc(4);
    Msat_IntVecPush(manager.vPatsReal, 32);
    Msat_IntVecPush(manager.vPatsReal, 31);

    Fraig_FeedBackCompress(&manager);
    printf("%d %d %d %08x %08x %08x %08x %08x %08x\n",
           manager.iWordStart, manager.iWordPerm, manager.iPatsPerm,
           sim_a[0], sim_a[1], sim_b[0], sim_b[1], real[0], real[1]);
    Msat_IntVecFree(manager.vPatsReal);
    Fraig_NodeVecFree(manager.vCones);
    Fraig_MemFixedStop(manager.mmSims, 0);
    return 0;
}
