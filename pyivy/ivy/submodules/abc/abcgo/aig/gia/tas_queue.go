package gia

// Tas_Que_t supports queue snapshots and null-terminated clause segments.
type Tas_Que_t struct {
	iHead, iTail int32
	pData        []*Gia_Obj_t
}

func Tas_QueIsEmpty(q *Tas_Que_t) int32 { return giaBool(q.iHead == q.iTail) }

func Tas_QuePush(q *Tas_Que_t, node *Gia_Obj_t) {
	if q.iTail < 0 || int(q.iTail) > len(q.pData) {
		panic("gia: invalid SAT queue tail")
	}
	if int(q.iTail) == len(q.pData) {
		q.pData = append(q.pData, node)
	} else {
		q.pData[q.iTail] = node
	}
	q.iTail++
}

func Tas_QueHasNode(q *Tas_Que_t, node *Gia_Obj_t) int32 {
	for i := q.iHead; i < q.iTail && q.pData[i] != nil; i++ {
		if q.pData[i] == node {
			return 1
		}
	}
	return 0
}

func Tas_QueStore(q *Tas_Que_t, head, tail *int32) {
	*head, *tail = q.iHead, q.iTail
	for i := *head; i < *tail; i++ {
		Tas_QuePush(q, q.pData[i])
	}
	q.iHead = *tail
}

func Tas_QueRestore(q *Tas_Que_t, head, tail int32) {
	if head < 0 || tail < head || int(tail) > len(q.pData) {
		panic("gia: invalid SAT queue snapshot")
	}
	q.iHead, q.iTail = head, tail
}

func Tas_QueFinish(q *Tas_Que_t) int32 {
	if q.iHead >= q.iTail {
		panic("gia: cannot finish empty clause")
	}
	head := q.iHead
	Tas_QuePush(q, nil)
	q.iHead = q.iTail
	return head
}
