package abc

import "testing"

func TestAbcExtractSfmMappedNetwork(t *testing.T) {
	n, _, _, pi, driver, consumer := abcSCLFixture(t)
	mapped, objects, err := Abc_NtkExtractSfm(n)
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 4 || objects[0] != pi || objects[1] != driver || objects[2] != consumer || objects[3] != n.Outputs()[0] {
		t.Fatalf("object order: %v", objects)
	}
	for _, input := range []bool{false, true} {
		out, err := mapped.Evaluate([]bool{input})
		if err != nil || len(out) != 1 || out[0] != input {
			t.Fatalf("input %v: %v %v", input, out, err)
		}
	}
	consumer.fanins[0].Complement = true
	if _, _, err := Abc_NtkExtractSfm(n); err == nil {
		t.Fatal("complemented mapped edge accepted")
	}
}

func TestAbcApplySfmMappedNetwork(t *testing.T) {
	n, _, _, pi, driver, consumer := abcSCLFixture(t)
	p, original, err := Abc_NtkExtractSfm(n)
	if err != nil {
		t.Fatal(err)
	}
	// Remove the first buffer and reconnect the second directly to the input.
	p.Deleted[1] = true
	p.Fanins[1] = nil
	p.Fanins[2] = []int{0}
	if err := Abc_NtkApplySfm(n, p, original); err != nil {
		t.Fatal(err)
	}
	if n.Object(driver.ID) != nil || n.Object(consumer.ID) != consumer || len(consumer.fanins) != 1 || consumer.fanins[0].Object != pi {
		t.Fatal("mapped network replacement did not preserve live identities")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestAbcApplySfmRejectsInvalidGateWithoutMutation(t *testing.T) {
	n, _, _, _, _, consumer := abcSCLFixture(t)
	p, original, err := Abc_NtkExtractSfm(n)
	if err != nil {
		t.Fatal(err)
	}
	old := consumer.fanins[0]
	p.Mapping[2] = nil
	if err := Abc_NtkApplySfm(n, p, original); err == nil {
		t.Fatal("invalid gate accepted")
	}
	if consumer.fanins[0] != old {
		t.Fatal("failed validation changed network")
	}
}
