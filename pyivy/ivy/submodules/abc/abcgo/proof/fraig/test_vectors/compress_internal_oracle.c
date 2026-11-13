#include "proof/fraig/fraigInt.h"
#include <stdio.h>

int main(void)
{
    Fraig_Man_t manager = {0};
    Fraig_Node_t constant = {0}, a = {0}, b = {0}, gate = {0};
    Fraig_HashTable_t table = {0};
    unsigned a_storage[6] = {0x12345678, 0x80000000, 1};
    unsigned b_storage[6] = {0xa5a5a5a5, 0, 1};
    unsigned gate_storage[6] = {0};
    unsigned real[5] = {0x7fffffff}, temp[5] = {0};
    Fraig_Node_t *inputs[] = {&a, &b}, *nodes[] = {&constant, &a, &b, &gate};
    Fraig_NodeVec_t input_vec = {0}, node_vec = {0};

    input_vec.nSize = input_vec.nCap = 2;
    input_vec.pArray = inputs;
    node_vec.nSize = node_vec.nCap = 4;
    node_vec.pArray = nodes;
    constant.NumPi = -1;
    a.Num = 1; a.NumPi = 0; a.puSimR = a_storage; a.puSimD = a_storage + 1;
    b.Num = 2; b.NumPi = 1; b.puSimR = b_storage; b.puSimD = b_storage + 1;
    gate.Num = 3; gate.NumPi = -1;
    gate.p1 = &a; gate.p2 = Fraig_Not(&b);
    gate.puSimR = gate_storage; gate.puSimD = gate_storage + 1;
    Fraig_NodeSimulate(&gate, 0, 1, 1);
    manager.pConst1 = &constant;
    manager.vInputs = &input_vec;
    manager.vNodes = &node_vec;
    manager.vCones = Fraig_NodeVecAlloc(8);
    manager.mmSims = Fraig_MemFixedStart(24);
    manager.pTableF = &table;
    manager.nWordsRand = 1;
    manager.nWordsDyna = 5;
    manager.iWordStart = 3;
    manager.iWordPerm = 1;
    manager.iPatsPerm = 31;
    manager.pSimsReal = real;
    manager.pSimsTemp = temp;
    manager.vPatsReal = Msat_IntVecAlloc(4);
    Msat_IntVecPush(manager.vPatsReal, 32);

    Fraig_FeedBackCompress(&manager);
    printf("%d %d %d %d %08x %08x %08x %08x %08x %08x %08x\n",
           manager.iWordStart, manager.iWordPerm, manager.iPatsPerm,
           manager.nWordsDyna, a.puSimD[0], b.puSimD[0],
           gate.puSimR[0], gate.puSimD[0], gate.uHashR, gate.uHashD,
           manager.pSimsReal[0]);
    Msat_IntVecFree(manager.vPatsReal);
    Fraig_NodeVecFree(manager.vCones);
    Fraig_MemFixedStop(manager.mmSims, 0);
    return 0;
}
