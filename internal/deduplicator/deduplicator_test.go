package deduplicator

import (
	"testing"
)

func TestCreate(t *testing.T) {

	directory := "test-directory-name"
	hashingRoutines := 4

	deduplicator := New(directory, hashingRoutines)

	if deduplicator.directory != directory {
		t.Errorf("deduplicator.directory = %s; want %s", deduplicator.directory, directory)
	}

	if deduplicator.hashingRoutines != hashingRoutines {
		t.Errorf("deduplicator.hashingRoutines = %d; want %d", deduplicator.hashingRoutines, hashingRoutines)
	}

}
