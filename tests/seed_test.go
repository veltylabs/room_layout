package tests

import (
	"testing"

	"github.com/veltylabs/room_layout/seed"
)

func TestSeedLoad(t *testing.T) {
	env := newTestEnv("tenantA")

	data, err := seed.Load(env.mod, "tenantA", seed.LoadOptions{
		CategoryIDs: []string{"cat1", "cat2"},
		OccupantIDs: []string{"occ1", "occ2"},
	})
	if err != nil {
		t.Fatalf("seed.Load failed: %v", err)
	}

	if len(data.Floors) != 2 {
		t.Fatalf("expected 2 floors, got %d", len(data.Floors))
	}
	if len(data.Rooms) != 5 {
		t.Fatalf("expected 5 rooms, got %d", len(data.Rooms))
	}
	if len(data.Equipment) != 3 {
		t.Fatalf("expected 3 equipment, got %d", len(data.Equipment))
	}
	if len(data.Shifts) != 6 {
		t.Fatalf("expected 6 shifts, got %d", len(data.Shifts))
	}
}
