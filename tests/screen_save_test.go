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
	r := &roomlayout.Room{
		FloorId:  f.Id,
		Code:     "B01",
		Name:     "Box 1",
		IsActive: true,
	}
	vRoom.(view.Saver).Save([]model.Model{r}, done)
	if lastErr != nil {
		t.Fatalf("failed to save room: %v", lastErr)
	}

	// A second room with the same code: the screen gets a conflict, not a 500.
	dup := &roomlayout.Room{FloorId: f.Id, Code: "B01", Name: "Box 1 bis", IsActive: true}
	vRoom.(view.Saver).Save([]model.Model{dup}, done)
	if model.KindOf(lastErr) != model.ErrorConflict {
		t.Fatalf("duplicate room code: want a conflict on the client, got %v", lastErr)
	}
	lastErr = nil

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

	vFloor.Reload(done)
	vRoom.Reload(done)
	vEq.Reload(done)
	// Deactivate room (Room doesn't use generic Delete list op)

	// The floor still has a room: deleting it is a conflict the screen can show.
	vFloor.(view.Deleter).Delete([]string{vFloor.Items()[0].ID}, done)
	if model.KindOf(lastErr) != model.ErrorConflict {
		t.Fatalf("deleting a floor with rooms: want a conflict, got %v", lastErr)
	}
}

// El tenant lo pone el servidor (D32): un registro que llega con otro tenant se
// guarda en el de esta instalación, y una lista nunca devuelve otro tenant.
func TestScreenSaveIgnoresTenantFromRequest(t *testing.T) {
	env := newTestEnv("tenantA")
	caller := view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod))
	vFloor := roomlayout.NewFloorView(caller)

	var lastErr error
	f := &roomlayout.Floor{TenantId: "other-tenant", Name: "Planta B"}
	vFloor.(view.Saver).Save([]model.Model{f}, func(e error) { lastErr = e })
	if lastErr != nil {
		t.Fatalf("save floor: %v", lastErr)
	}

	mine, err := env.mod.ListFloors("tenantA")
	if err != nil || len(mine) != 1 {
		t.Fatalf("server's tenant: got %d floors (err %v), want 1", len(mine), err)
	}
	other, err := env.mod.ListFloors("other-tenant")
	if err != nil || len(other) != 0 {
		t.Fatalf("tenant sent by the request: got %d floors (err %v), want 0", len(other), err)
	}
}
