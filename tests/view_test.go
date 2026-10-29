package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/router/loopback"
	"webtyp.com/view"
)

func TestPresenterViewsCoverage(t *testing.T) {
	env := newTestEnv("tenantA")
	caller := loopback.WithTenant("tenantA", env.mod)

	noopDone := func(err error) {}

	vFloor := roomlayout.NewFloorView(view.NewRecords().Caller(caller))
	if vFloor == nil {
		t.Fatalf("NewFloorView returned nil")
	}
	vFloor.Reload(noopDone)

	vRoom := roomlayout.NewRoomView(view.NewRecords().Caller(caller))
	if vRoom == nil {
		t.Fatalf("NewRoomView returned nil")
	}
	vRoom.Reload(noopDone)

	vEq := roomlayout.NewEquipmentView(view.NewRecords().Caller(caller))
	if vEq == nil {
		t.Fatalf("NewEquipmentView returned nil")
	}
	vEq.Reload(noopDone)

	}

func TestItemsDeclaredOnFields(t *testing.T) {
	f := &roomlayout.Floor{Id: "1", Name: "Floor 1"}
	if item := f.Item(); item.Label != "Floor 1" || item.ID != "1" {
		t.Fatalf("invalid floor Item: %+v", item)
	}

	r := &roomlayout.Room{Id: "2", Code: "B-101", Name: "Box 101"}
	if item := r.Item(); item.Label != "B-101" || item.Description != "Box 101" || item.ID != "2" {
		t.Fatalf("invalid room Item: %+v", item)
	}

	eq := &roomlayout.Equipment{Id: "3", Name: "Impresora"}
	if item := eq.Item(); item.Label != "Impresora" || item.ID != "3" {
		t.Fatalf("invalid equipment Item: %+v", item)
	}

	sForm1 := &roomlayout.RoomShiftForm{Id: "4", DayOfWeek: roomlayout.WeekdayMonday, Start: "09:00", End: "13:00", OccupantLabel: "Doc"}
	if item := sForm1.Item(); item.Label != "Doc" || item.Description != "1 · 09:00 · 13:00" {
		t.Fatalf("invalid shift form Item weekly: %+v", item)
	}

	sForm2 := &roomlayout.RoomShiftForm{Id: "5", Date: "2026-03-16", Start: "09:00", End: "13:00", OccupantLabel: "Doc"}
	if item := sForm2.Item(); item.Label != "Doc" || item.Description != "2026-03-16 · 09:00 · 13:00" {
		t.Fatalf("invalid shift form Item dated: %+v", item)
	}
}
