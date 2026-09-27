package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/time"
)

func TestShiftsOnAndBoard(t *testing.T) {
	env := newTestEnv("tenantA")

	f, _ := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1", Position: 1})
	r, _ := env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  f.Id,
		Code:     "R101",
		Name:     "Room 101",
		IsActive: true,
	})
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	dateSec, _ := time.ParseDate("2026-03-16")
	dateSec = dateSec / 1_000_000_000

	// Local time 10:30 on March 16 (minute 630)
	env.clockSec = time.LocalMinutesToUnixUTC(dateSec+3*3600, 630, "America/Santiago")

	// Set bounds for this date: 08:00 (480) to 20:00 (1200)
	env.bounds.setBounds(dateSec, time.DayBounds{Open: true, OpenMin: 480, CloseMin: 1200})

	// Add busy shift: 10:00 - 12:00
	_, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Busy Doctor",
		Date:          "2026-03-16",
		Start:         "10:00",
		End:           "12:00",
	})
	if err != nil {
		t.Fatalf("Save busy shift failed: %v", err)
	}

	// Add next shift: 14:00 - 16:00
	_, err = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Next Doctor",
		Date:          "2026-03-16",
		Start:         "14:00",
		End:           "16:00",
	})
	if err != nil {
		t.Fatalf("Save next shift failed: %v", err)
	}

	board, err := env.mod.ListBoard("tenantA", "", "")
	if err != nil {
		t.Fatalf("ListBoard failed: %v", err)
	}
	if len(board) != 1 {
		t.Fatalf("expected 1 room on board, got %d", len(board))
	}
	b := board[0]
	if b.Status != roomlayout.BoardStatusBusy {
		t.Fatalf("expected status 'busy', got '%s'", b.Status)
	}
	if b.CurrentOccupantLabel != "Busy Doctor" {
		t.Fatalf("expected current occupant 'Busy Doctor', got '%s'", b.CurrentOccupantLabel)
	}
	if b.CurrentUntil != "12:00" {
		t.Fatalf("expected current_until '12:00', got '%s'", b.CurrentUntil)
	}
	if b.NextStart != "14:00" || b.NextOccupantLabel != "Next Doctor" {
		t.Fatalf("expected next shift 14:00 Next Doctor, got %s %s", b.NextStart, b.NextOccupantLabel)
	}
}

func TestMoveShiftOccurrence(t *testing.T) {
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

	weeklyShift, _ := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r1.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Doctor W",
		DayOfWeek:     roomlayout.WeekdayMonday,
		Start:         "10:00",
		End:           "12:00",
	})

	dateSec, _ := time.ParseDate("2026-03-16") // Monday
	dateSec = dateSec / 1_000_000_000

	// Move to r2 should fail because cat1 is not enabled for r2
	_, err := env.mod.MoveShiftOccurrence("tenantA", weeklyShift.Id, dateSec, r2.Id)
	if err != roomlayout.ErrCategoryNotAllowed {
		t.Fatalf("expected ErrCategoryNotAllowed, got %v", err)
	}

	// Enable cat1 for r2
	_ = env.mod.SetRoomCategories("tenantA", r2.Id, []string{"cat1"})

	// Move to r2 should succeed now
	movedShift, err := env.mod.MoveShiftOccurrence("tenantA", weeklyShift.Id, dateSec, r2.Id)
	if err != nil {
		t.Fatalf("MoveShiftOccurrence failed: %v", err)
	}
	if movedShift.RoomId != r2.Id || movedShift.SpecificDate != dateSec {
		t.Fatalf("unexpected moved shift: %+v", movedShift)
	}
}
