package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/router/loopback"
)

func TestPresenterViewsCoverage(t *testing.T) {
	env := newTestEnv("tenantA")
	loop := loopback.WithTenant("tenantA", env.mod)

	noopDone := func(err error) {}

	vFloor := roomlayout.NewFloorView(loop)
	if vFloor == nil {
		t.Fatalf("NewFloorView returned nil")
	}
	vFloor.Reload(noopDone)

	vRoom := roomlayout.NewRoomView(loop)
	if vRoom == nil {
		t.Fatalf("NewRoomView returned nil")
	}
	vRoom.Reload(noopDone)

	vEq := roomlayout.NewEquipmentView(loop)
	if vEq == nil {
		t.Fatalf("NewEquipmentView returned nil")
	}
	vEq.Reload(noopDone)

	vShift := roomlayout.NewShiftView(loop, "tenantA", "r1")
	if vShift == nil {
		t.Fatalf("NewShiftView returned nil")
	}
	vShift.Reload(noopDone)

	// Test Itemizer methods
	f := &roomlayout.Floor{Id: "1", Name: "Floor 1"}
	if f.Item().ID != "1" || f.Item().Label != "Floor 1" {
		t.Fatalf("invalid floor Item")
	}

	r := &roomlayout.Room{Id: "2", Code: "B-101", Name: "Box 101"}
	if r.Item().Label != "B-101 · Box 101" {
		t.Fatalf("invalid room Item")
	}

	eq := &roomlayout.Equipment{Id: "3", Name: "Impresora"}
	if eq.Item().Label != "Impresora" {
		t.Fatalf("invalid equipment Item")
	}

	sForm1 := &roomlayout.RoomShiftForm{Id: "4", DayOfWeek: roomlayout.WeekdayMonday, Start: "09:00", End: "13:00", OccupantLabel: "Doc"}
	if sForm1.Item().Description != "Lunes 09:00–13:00" {
		t.Fatalf("invalid shift form Item weekly: %s", sForm1.Item().Description)
	}

	sForm2 := &roomlayout.RoomShiftForm{Id: "5", Date: "2026-03-16", Start: "09:00", End: "13:00", OccupantLabel: "Doc"}
	if sForm2.Item().Description != "2026-03-16 09:00–13:00" {
		t.Fatalf("invalid shift form Item dated: %s", sForm2.Item().Description)
	}
}
