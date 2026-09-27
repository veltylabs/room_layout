package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
)

func TestDeleteFloorInUse(t *testing.T) {
	env := newTestEnv("tenantA")

	f, _ := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1"})
	_, _ = env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  f.Id,
		Code:     "R101",
		Name:     "Room 101",
		IsActive: true,
	})

	err := env.mod.DeleteFloor("tenantA", f.Id)
	if err != roomlayout.ErrFloorInUse {
		t.Fatalf("expected ErrFloorInUse, got %v", err)
	}
}
