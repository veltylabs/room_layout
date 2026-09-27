package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
)

func TestTenantIsolation(t *testing.T) {
	env := newTestEnv("tenantA")

	// Create Floor and Room in Tenant A
	fA, _ := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Floor A"})
	rA, _ := env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  fA.Id,
		Code:     "RA",
		Name:     "Room A",
		IsActive: true,
	})

	// Tenant B tries to read Room A
	_, err := env.mod.GetRoom("tenantB", rA.Id)
	if err != roomlayout.ErrNotFound {
		t.Fatalf("expected ErrNotFound for cross-tenant read, got %v", err)
	}

	// Tenant B tries to list rooms
	roomsB, err := env.mod.ListRooms("tenantB", false)
	if err != nil || len(roomsB) != 0 {
		t.Fatalf("expected empty room list for tenantB, got %d rooms, err: %v", len(roomsB), err)
	}

	// Tenant B tries to delete Floor A
	_ = env.mod.DeleteFloor("tenantB", fA.Id)

	// Verify Floor A still exists for Tenant A
	floorsA, err := env.mod.ListFloors("tenantA")
	if err != nil || len(floorsA) != 1 {
		t.Fatalf("expected Floor A to still exist, got %v", err)
	}
}
