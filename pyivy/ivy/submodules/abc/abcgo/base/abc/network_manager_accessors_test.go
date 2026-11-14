package abc

import "testing"

func TestFunctionManagerAndObjectDataAccessors(t *testing.T) {
	network := NewNetwork("accessor-test")
	manager := &struct{ value int }{value: 7}
	Abc_NtkSetFunctionManager(network, manager)
	if Abc_NtkReadNameManager(network) == nil {
		t.Fatal("network name manager accessor returned nil")
	}
	if Abc_NtkReadFunctionManager(network) != manager {
		t.Fatal("function manager accessor changed the opaque payload")
	}

	object := network.NewObject(ObjectNode)
	data := &struct{ value int }{value: 11}
	Abc_ObjWriteData(object, data)
	if Abc_ObjReadData(object) != data {
		t.Fatal("object data accessor changed the opaque payload")
	}
}
