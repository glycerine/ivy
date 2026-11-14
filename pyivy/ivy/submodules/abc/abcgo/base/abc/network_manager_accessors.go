package abc

import "github.com/glycerine/abcgo/misc/nm"

// Abc_NtkReadFunctionManager returns the network's borrowed function manager.
func Abc_NtkReadFunctionManager(n *Network) any { return n.FunctionManager }

// Abc_NtkSetFunctionManager stores the function manager without changing ownership.
func Abc_NtkSetFunctionManager(n *Network, manager any) { n.FunctionManager = manager }

// Abc_ObjReadData returns the object's opaque source-compatible pData payload.
func Abc_ObjReadData(o *Object) any { return Abc_ObjData(o) }

// Abc_ObjWriteData stores the object's opaque pData payload.
func Abc_ObjWriteData(o *Object, data any) { o.Data = data }

// Abc_NtkReadNameManager returns the network's borrowed name manager.
func Abc_NtkReadNameManager(n *Network) *nm.Manager { return n.names }
