package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
)

func TestDuplicateRoomCode(t *testing.T) {
	env := newTestEnv("tenantA")

	f, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1"})
	if err != nil {
		t.Fatalf("SaveFloor failed: %v", err)
	}

	r1, err := env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  f.Id,
		Code:     "R101",
		Name:     "Room 101",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("SaveRoom r1 failed: %v", err)
	}

	// Duplicate code in same tenant
	_, err = env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  f.Id,
		Code:     "R101",
		Name:     "Another Room 101",
		IsActive: true,
	})
	if err != roomlayout.ErrCodeAlreadyExists {
		t.Fatalf("expected ErrCodeAlreadyExists, got %v", err)
	}

	// Same code in tenantB should succeed
	fB, _ := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantB", Name: "Piso 1"})
	_, err = env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantB",
		FloorId:  fB.Id,
		Code:     "R101",
		Name:     "Room 101 Tenant B",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("SaveRoom in tenantB failed: %v", err)
	}

	// Updating r1 with same code should succeed
	r1.Name = "Room 101 Updated"
	_, err = env.mod.SaveRoom(r1)
	if err != nil {
		t.Fatalf("Update r1 failed: %v", err)
	}
}
