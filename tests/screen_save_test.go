package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/model"
	"webtyp.com/router/loopback"
	"webtyp.com/view"
)

func TestScreenSaveAndDelete(t *testing.T) {
	env := newTestEnv("tenantA")
	baseCaller := loopback.WithTenant("tenantA", env.mod)
	caller := view.NewRecords().Caller(baseCaller)

	// Create presenters
	vFloor := roomlayout.NewFloorView(caller)
	vRoom := roomlayout.NewRoomView(caller)
	vEq := roomlayout.NewEquipmentView(caller)

	var lastErr error
	done := func(e error) {
		lastErr = e
	}

	// Save a floor
	f := &roomlayout.Floor{Id: "fl-1",
		Name: "Planta Principal",
	}
	vFloor.(view.Saver).Save([]model.Model{f}, done)
	if lastErr != nil {
		t.Fatalf("failed to save floor: %v", lastErr)
	}


	// Save a room
	r := &roomlayout.Room{Id: "rm-1",
		FloorId: f.Id,
		Code: "B01",
		Name: "Box 1",
		IsActive: true,
	}
	vRoom.(view.Saver).Save([]model.Model{r}, done)
	if lastErr != nil {
		t.Fatalf("failed to save room: %v", lastErr)
	}




	// Save equipment
	eq := &roomlayout.Equipment{Id: "eq-1",
		Name: "Electrocardiógrafo",
	}
	vEq.(view.Saver).Save([]model.Model{eq}, done)
	if lastErr != nil {
		t.Fatalf("failed to save equipment: %v", lastErr)
	}


	// Verify lists
	vFloor.Reload(func(err error) {
		if err != nil {
			t.Fatalf("failed to reload floor view: %v", err)
		}
		if len(vFloor.Items()) != 1 {
			t.Fatalf("expected 1 floor, got %d", len(vFloor.Items()))
		}
	})

	// Delete
	vEq.(view.Deleter).Delete([]string{vEq.Items()[0].ID}, done)
	if lastErr != nil {
		t.Fatalf("failed to delete equipment: %v", lastErr)
	}

	vFloor.Reload(done); vRoom.Reload(done); vEq.Reload(done)
	// Deactivate room (Room doesn't use generic Delete list op)


	vFloor.(view.Deleter).Delete([]string{vFloor.Items()[0].ID}, done)
	if lastErr != nil {
		t.Fatalf("failed to delete floor: %v", lastErr)
	}
}
