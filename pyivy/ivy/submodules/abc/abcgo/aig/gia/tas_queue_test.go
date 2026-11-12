package gia

import "testing"

func TestTasQueueSnapshotRestoreAndClause(t *testing.T) {
	q := &Tas_Que_t{}
	a, b, c := &Gia_Obj_t{}, &Gia_Obj_t{}, &Gia_Obj_t{}
	Tas_QuePush(q, a)
	Tas_QuePush(q, b)
	var head, tail int32
	Tas_QueStore(q, &head, &tail)
	if head != 0 || tail != 2 || Tas_QueHasNode(q, a) != 1 || q.iHead != 2 {
		t.Fatal("snapshot did not duplicate live queue")
	}
	Tas_QuePush(q, c)
	Tas_QueRestore(q, head, tail)
	if Tas_QueHasNode(q, c) != 0 {
		t.Fatal("restore retained speculative element")
	}
	Tas_QuePush(q, c)
	start := Tas_QueFinish(q)
	if start != 0 || Tas_QueIsEmpty(q) != 1 || q.pData[3] != nil || q.pData[2] != c {
		t.Fatal("clause was not terminated")
	}
}
