package abc

import (
	"testing"

	"github.com/glycerine/abcgo/misc/nm"
)

func TestObjectDeleteNameRemovesMapping(t *testing.T) {
	network := NewNetwork("names")
	object := network.NewObject(ObjectPI)
	if Abc_ObjAssignName(object, "input", "") != "input" {
		t.Fatal("failed to assign test name")
	}
	Abc_ObjDeleteName(object)
	if got := nm.Nm_ManFindIdByName(network.names, "input", int32(ObjectPI)); got != -1 {
		t.Fatalf("deleted name still maps to object %d", got)
	}
}

func TestObjectWriteDataDoesNotClearStructuralHash(t *testing.T) {
	network := NewNetwork("hash")
	object := network.NewObject(ObjectNode)
	key := andHashKey([2]Edge{{Object: object}, {Object: network.constant}})
	network.hashed[key] = object
	data := &struct{ value int }{value: 11}
	Abc_ObjWriteData(object, data)
	if Abc_ObjReadData(object) != data {
		t.Fatal("object data accessor changed the opaque payload")
	}
	if network.hashed[key] != object {
		t.Fatal("raw object data write cleared structural hash state")
	}
}
