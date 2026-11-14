package kit

import "testing"

func TestDsdManagerOwnsSourceCloudManager(t *testing.T) {
	p := Kit_DsdManAlloc(3, 1)
	cloud := Kit_DsdCloud(p)
	if cloud == nil || Cloud_One(cloud) == nil {
		t.Fatal("DSD manager did not initialize its Cloud manager")
	}
	for i := int32(0); i < 16; i++ {
		if Cloud_Var(cloud, i) == nil {
			t.Fatalf("Cloud variable %d missing", i)
		}
	}
	Kit_DsdManFree(p)
	if Kit_DsdCloud(p) != nil || cloud.one != nil {
		t.Fatal("DSD manager did not release its Cloud manager")
	}
}
