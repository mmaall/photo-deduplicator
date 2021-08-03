package deduplicator

import (
	"fmt"
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

func TestAddToMap(t *testing.T) {

	pairCount := 10
	var inputPairs []pair

	for i := 0; i < pairCount; i++ {
		inputPairs = append(inputPairs, pair{
			key: fmt.Sprint(rune(i)),
			val: string(rune(int('a') + i)),
		})
	}

	fmt.Print("Created %d", pairCount, "pairs")

	// TODO: Finish up this test

}
