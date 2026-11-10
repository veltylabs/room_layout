package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/view"
)

func TestScreenSaveAndDelete(t *testing.T) {
	env := newTestEnv("tenantA")
	caller := view.NewRecords().Caller(crud.LoopbackActingAs(nil, "u1", env.mod))

	// Create presenters
	vFloor := roomlayout.NewFloorView(caller)
	vRoom := roomlayout.NewRoomView(caller)
	vEq := roomlayout.NewEquipmentView(caller)

	var lastErr error
	done := func(e error) {
		lastErr = e
	}

	// Save a floor
	f := &roomlayout.Floor{
		Name: "Planta Principal",
	}
	vFloor.(view.Saver).Save([]model.Model{f}, done)
	if lastErr != nil {
		t.Fatalf("failed to save floor: %v", lastErr)
	}

	vFloor.Reload(done)
	if len(vFloor.Items()) != 1 {
		t.Fatalf("expected 1 floor, got %d", len(vFloor.Items()))
	}
	floorID := vFloor.Items()[0].ID

	// Save a room
	r := &roomlayout.Room{
		FloorId:  floorID,
		Code:     "B01",
		Name:     "Box 1",
		IsActive: true,
	}
	vRoom.(view.Saver).Save([]model.Model{r}, done)
	if lastErr != nil {
		t.Fatalf("failed to save room: %v", lastErr)
	}

	// A second room with the same code: the screen gets a conflict, not a 500.
	dup := &roomlayout.Room{FloorId: floorID, Code: "B01", Name: "Box 1 bis", IsActive: true}
	vRoom.(view.Saver).Save([]model.Model{dup}, done)
	if model.KindOf(lastErr) != model.ErrorConflict {
		t.Fatalf("duplicate room code: want a conflict on the client, got %v", lastErr)
	}
	lastErr = nil

	// Save equipment
	eq := &roomlayout.Equipment{
		Name: "Electrocardiógrafo",
	}
	vEq.(view.Saver).Save([]model.Model{eq}, done)
	if lastErr != nil {
		t.Fatalf("failed to save equipment: %v", lastErr)
	}

	vEq.Reload(done)
	if len(vEq.Items()) != 1 {
		t.Fatalf("expected 1 equipment, got %d", len(vEq.Items()))
	}

	// Delete equipment
	vEq.(view.Deleter).Delete([]string{vEq.Items()[0].ID}, done)
	if lastErr != nil {
		t.Fatalf("failed to delete equipment: %v", lastErr)
	}

	vFloor.Reload(done)
	vRoom.Reload(done)
	vEq.Reload(done)

	// The floor still has a room: deleting it is a conflict the screen can show.
	vFloor.(view.Deleter).Delete([]string{floorID}, done)
	if model.KindOf(lastErr) != model.ErrorConflict {
		t.Fatalf("deleting a floor with rooms: want a conflict, got %v", lastErr)
	}
}

// El tenant lo pone el servidor (D32): un registro que llega con tenant_id se rechaza como model.ErrorInvalid.
func TestScreenSaveIgnoresTenantFromRequest(t *testing.T) {
	env := newTestEnv("tenantA")
	caller := view.NewRecords().Caller(crud.LoopbackActingAs(nil, "u1", env.mod))
	vFloor := roomlayout.NewFloorView(caller)

	var lastErr error
	f := &roomlayout.Floor{TenantId: "other-tenant", Name: "Planta B"}
	vFloor.(view.Saver).Save([]model.Model{f}, func(e error) { lastErr = e })
	if model.KindOf(lastErr) != model.ErrorInvalid {
		t.Fatalf("save floor with tenant_id: want ErrorInvalid, got %v", lastErr)
	}
}
