package server

import "testing"

func TestHostedPersistenceSaveModelRevisionAndConflict(t *testing.T) {
	store := NewProjectStore()
	project := store.CreatePersonalProject("writer", "Model", "model")

	first, err := store.SaveModelRevision("writer", project.ID, "demo.ivy", "type t", 0)
	if err != nil {
		t.Fatal(err)
	}
	if first.Revision != 1 {
		t.Fatalf("revision = %d", first.Revision)
	}
	second, err := store.SaveModelRevision("writer", project.ID, "demo.ivy", "type u", 1)
	if err != nil {
		t.Fatal(err)
	}
	if second.Revision != 2 {
		t.Fatalf("revision = %d", second.Revision)
	}
	if _, err := store.SaveModelRevision("writer", project.ID, "demo.ivy", "stale", 1); err == nil {
		t.Fatal("expected stale base revision conflict")
	}
}

func TestHostedPersistenceAppendJobResultDoesNotOverwrite(t *testing.T) {
	store := NewProjectStore()
	project := store.CreatePersonalProject("reader", "Results", "results")

	first, err := store.AppendJobResult("reader", project.ID, "job-1", "first")
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.AppendJobResult("reader", project.ID, "job-1", "second")
	if err != nil {
		t.Fatal(err)
	}
	if second.Payload != first.Payload {
		t.Fatalf("append-only result overwritten: %+v", second)
	}
}

func TestHostedPersistenceGraphSnapshotsAreImmutable(t *testing.T) {
	store := NewProjectStore()
	project := store.CreatePersonalProject("reader", "Graphs", "graphs")

	if _, err := store.AppendGraphSnapshot("reader", project.ID, "snapshot-1", "first"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendGraphSnapshot("reader", project.ID, "snapshot-1", "second"); err == nil {
		t.Fatal("expected immutable snapshot conflict")
	}
}
