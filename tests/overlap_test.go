package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/time"
)

func TestShiftOverlapCases(t *testing.T) {
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

	// Existing weekly shift: Mondays 09:00 - 13:00
	_, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Doctor A",
		DayOfWeek:     roomlayout.WeekdayMonday,
		Start:         "09:00",
		End:           "13:00",
	})
	if err != nil {
		t.Fatalf("Save initial weekly shift failed: %v", err)
	}

	// Edge case: 13:00 - 15:00 on Monday should NOT overlap
	_, err = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Doctor B",
		DayOfWeek:     roomlayout.WeekdayMonday,
		Start:         "13:00",
		End:           "15:00",
	})
	if err != nil {
		t.Fatalf("Adjacent shift 13:00-15:00 failed: %v", err)
	}

	// Weekly candidate 10:00 - 12:00 on Monday -> ErrRoomOverlap
	_, err = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Doctor C",
		DayOfWeek:     roomlayout.WeekdayMonday,
		Start:         "10:00",
		End:           "12:00",
	})
	if err != roomlayout.ErrRoomOverlap {
		t.Fatalf("expected ErrRoomOverlap, got %v", err)
	}

	// Cancel Monday occurrence of shift 1 on date 2026-03-16 (a Monday)
	mondayDateSec, _ := time.ParseDate("2026-03-16")
	mondayDateSec = mondayDateSec / 1_000_000_000

	// Get shift 1 ID
	shifts, _ := env.mod.ListRoomShifts("tenantA", r.Id)
	weeklyShiftID := shifts[0].Id

	err = env.mod.CancelShiftOccurrence("tenantA", weeklyShiftID, mondayDateSec)
	if err != nil {
		t.Fatalf("CancelShiftOccurrence failed: %v", err)
	}

	// Dated candidate on 2026-03-16 10:00-12:00 -> should NOT overlap because weekly occurrence is cancelled
	_, err = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Doctor D",
		Date:          "2026-03-16",
		Start:         "10:00",
		End:           "12:00",
	})
	if err != nil {
		t.Fatalf("Dated shift on cancelled date failed: %v", err)
	}
}

func TestOccupantOverlap(t *testing.T) {
	env := newTestEnv("tenantA")

	f, _ := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1"})
	r1, _ := env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  f.Id,
		Code:     "R101",
		Name:     "Room 101",
		IsActive: true,
	})
	r2, _ := env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  f.Id,
		Code:     "R102",
		Name:     "Room 102",
		IsActive: true,
	})

	_ = env.mod.SetRoomCategories("tenantA", r1.Id, []string{"cat1"})
	_ = env.mod.SetRoomCategories("tenantA", r2.Id, []string{"cat1"})

	// occ1 in r1 on Monday 10:00 - 12:00
	_, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:     r1.Id,
		CategoryId: "cat1",
		OccupantId: "occ1",
		DayOfWeek:  roomlayout.WeekdayMonday,
		Start:      "10:00",
		End:        "12:00",
	})
	if err != nil {
		t.Fatalf("Save shift in r1 failed: %v", err)
	}

	// occ1 in r2 on Monday 11:00 - 13:00 -> ErrOccupantOverlap
	_, err = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:     r2.Id,
		CategoryId: "cat1",
		OccupantId: "occ1",
		DayOfWeek:  roomlayout.WeekdayMonday,
		Start:      "11:00",
		End:        "13:00",
	})
	if err != roomlayout.ErrOccupantOverlap {
		t.Fatalf("expected ErrOccupantOverlap, got %v", err)
	}
}
