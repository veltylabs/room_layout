package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/time"
)

func TestListPublicAvailability(t *testing.T) {
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

	dateSec, _ := time.ParseDate("2026-03-16")
	dateSec = dateSec / 1_000_000_000

	// Bounds: 08:00 (480) - 12:00 (720)
	env.bounds.setBounds(dateSec, time.DayBounds{Open: true, OpenMin: 480, CloseMin: 720})

	// Shift from 08:00 to 09:00, and 09:10 to 10:00
	_, _ = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Secret Occupant",
		Date:          "2026-03-16",
		Start:         "08:00",
		End:           "09:00",
	})
	_, _ = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Secret Occupant",
		Date:          "2026-03-16",
		Start:         "09:10",
		End:           "10:00",
	})

	slots, err := env.mod.ListPublicAvailability(dateSec, "cat1")
	if err != nil {
		t.Fatalf("ListPublicAvailability failed: %v", err)
	}
	if len(slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(slots))
	}
	if slots[0].Start != "10:00" || slots[0].End != "12:00" {
		t.Fatalf("expected slot 10:00-12:00, got %s-%s", slots[0].Start, slots[0].End)
	}
}
