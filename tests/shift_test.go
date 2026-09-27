package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
)

func TestSaveShiftValidationAndOccupantSnapshot(t *testing.T) {
	env := newTestEnv("tenantA")

	f, _ := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1"})
	r, _ := env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  f.Id,
		Code:     "R101",
		Name:     "Room 101",
		IsActive: true,
	})

	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	// 1. Invalid range (start >= end)
	form := roomlayout.RoomShiftForm{
		RoomId:     r.Id,
		CategoryId: "cat1",
		DayOfWeek:  roomlayout.WeekdayMonday,
		Start:      "12:00",
		End:        "10:00",
	}
	_, err := env.mod.SaveShift("tenantA", form)
	if err != roomlayout.ErrInvalidRange {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}

	// 2. Unknown category
	form = roomlayout.RoomShiftForm{
		RoomId:     r.Id,
		CategoryId: "unknowncat",
		DayOfWeek:  roomlayout.WeekdayMonday,
		Start:      "10:00",
		End:        "12:00",
	}
	_, err = env.mod.SaveShift("tenantA", form)
	if err != roomlayout.ErrUnknownCategory {
		t.Fatalf("expected ErrUnknownCategory, got %v", err)
	}

	// 3. Category not allowed for room
	form = roomlayout.RoomShiftForm{
		RoomId:     r.Id,
		CategoryId: "cat2",
		DayOfWeek:  roomlayout.WeekdayMonday,
		Start:      "10:00",
		End:        "12:00",
	}
	_, err = env.mod.SaveShift("tenantA", form)
	if err != roomlayout.ErrCategoryNotAllowed {
		t.Fatalf("expected ErrCategoryNotAllowed, got %v", err)
	}

	// 4. Occupant snapshot test
	form = roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantId:    "occ1",
		OccupantLabel: "Custom Label That Should Be Overwritten",
		DayOfWeek:     roomlayout.WeekdayMonday,
		Start:         "10:00",
		End:           "12:00",
	}
	shift, err := env.mod.SaveShift("tenantA", form)
	if err != nil {
		t.Fatalf("SaveShift failed: %v", err)
	}
	if shift.OccupantLabel != "Dr. Juan Pérez" {
		t.Fatalf("expected occupant label 'Dr. Juan Pérez', got '%s'", shift.OccupantLabel)
	}
}
