package abc

import (
	"reflect"
	"testing"
)

func TestMappingTimingAdapterCopiesAndExpandsDefaults(t *testing.T) {
	n := NewNetwork("timed")
	Abc_NtkCreatePi(n)
	Abc_NtkCreatePi(n)
	Abc_NtkCreatePo(n)
	Abc_NtkCreatePo(n)
	arrivals, required := []float32{1.25}, []float32{8.5}
	Abc_NtkSetMappingTiming(n, &MappingTiming{ArrivalDefault: 2, RequiredDefault: 9, CIArrivals: arrivals, CORequired: required})
	arrivals[0], required[0] = 100, 100
	if got := Abc_NtkMappingArrivalTimes(n); !reflect.DeepEqual(got, []float32{1.25, 2}) {
		t.Fatalf("arrival times = %v", got)
	}
	if got := Abc_NtkMappingRequiredTimes(n); !reflect.DeepEqual(got, []float32{8.5, 9}) {
		t.Fatalf("required times = %v", got)
	}
}
