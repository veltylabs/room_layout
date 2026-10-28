package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/model"
	"webtyp.com/router/loopback"
	"webtyp.com/view"
)

func TestPresenterViewsCoverage(t *testing.T) {
	env := newTestEnv("tenantA")
	loop := view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod))

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
}

func TestRowsDeclaredOnFields(t *testing.T) {
	f := &roomlayout.Floor{Id: "1", Name: "Floor 1"}
	if row := model.RowOf(f); row.Title != "Floor 1" || row.ID != "1" {
		t.Fatalf("invalid floor RowOf: %+v", row)
	}

	r := &roomlayout.Room{Id: "2", Code: "B-101", Name: "Box 101"}
	if row := model.RowOf(r); row.Title != "B-101" || row.Detail != "Box 101" || row.ID != "2" {
		t.Fatalf("invalid room RowOf: %+v", row)
	}

	eq := &roomlayout.Equipment{Id: "3", Name: "Impresora"}
	if row := model.RowOf(eq); row.Title != "Impresora" || row.ID != "3" {
		t.Fatalf("invalid equipment RowOf: %+v", row)
	}

	sForm1 := &roomlayout.RoomShiftForm{Id: "4", DayOfWeek: roomlayout.WeekdayMonday, Start: "09:00", End: "13:00", OccupantLabel: "Doc"}
	if row := model.RowOf(sForm1); row.Title != "Doc" || row.Detail != "1 · 09:00 · 13:00" {
		t.Fatalf("invalid shift form RowOf weekly: %+v", row)
	}

	sForm2 := &roomlayout.RoomShiftForm{Id: "5", Date: "2026-03-16", Start: "09:00", End: "13:00", OccupantLabel: "Doc"}
	if row := model.RowOf(sForm2); row.Title != "Doc" || row.Detail != "2026-03-16 · 09:00 · 13:00" {
		t.Fatalf("invalid shift form RowOf dated: %+v", row)
	}
}
