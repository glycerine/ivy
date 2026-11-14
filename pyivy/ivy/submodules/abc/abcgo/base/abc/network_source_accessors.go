package abc

import "github.com/glycerine/abcgo/misc/util"

// Abc_NtkReadSeqModel and Abc_NtkSetSeqModel expose the sequence-model slot
// represented by Network.SequenceModel.
func Abc_NtkReadSeqModel(n *Network) *util.Abc_Cex_t { return n.SequenceModel }
func Abc_NtkSetSeqModel(n *Network, model *util.Abc_Cex_t) {
	n.SequenceModel = model
}

// Abc_NtkReadExdc and Abc_NtkSetExdc expose Network.Exdc directly.
func Abc_NtkReadExdc(n *Network) *Network { return n.Exdc }
func Abc_NtkSetExdc(n, exdc *Network)     { n.Exdc = exdc }

// Abc_ObjReadCopy and Abc_ObjSetCopy expose the regular object pointer stored
// in Object.Copy. Complemented copy bindings remain available through the
// separate CopyEdge accessors.
func Abc_ObjReadCopy(o *Object) *Object { return o.Copy }
func Abc_ObjSetCopy(o, copy *Object)    { o.Copy = copy }
