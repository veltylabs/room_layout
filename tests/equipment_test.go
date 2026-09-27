package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
)

func TestEquipmentCRUD(t *testing.T) {
	env := newTestEnv("tenantA")

	eq, err := env.mod.SaveEquipment(roomlayout.Equipment{
		TenantId: "tenantA",
		Name:     "Electrocardiógrafo",
	})
	if err != nil {
		t.Fatalf("SaveEquipment failed: %v", err)
	}

	list, err := env.mod.ListEquipment("tenantA")
	if err != nil || len(list) != 1 {
		t.Fatalf("ListEquipment expected 1, got %d, err: %v", len(list), err)
	}

	f, _ := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1"})
	r, _ := env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  f.Id,
		Code:     "R101",
		Name:     "Room 101",
		IsActive: true,
	})

	err = env.mod.SetRoomEquipment("tenantA", r.Id, []string{eq.Id})
	if err != nil {
		t.Fatalf("SetRoomEquipment failed: %v", err)
	}

	reqList, err := env.mod.ListRoomEquipment("tenantA", r.Id)
	if err != nil || len(reqList) != 1 {
		t.Fatalf("ListRoomEquipment expected 1, got %d, err: %v", len(reqList), err)
	}

	err = env.mod.DeleteEquipment("tenantA", eq.Id)
	if err != nil {
		t.Fatalf("DeleteEquipment failed: %v", err)
	}
}
