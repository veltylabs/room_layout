package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/time"
)

func TestListRoomDay(t *testing.T) {
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

	dateSec, _ := time.ParseDate("2026-03-16") // Monday
	dateSec = dateSec / 1_000_000_000

	_, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Doc Monday",
		DayOfWeek:     roomlayout.WeekdayMonday,
		Start:         "09:00",
		End:           "12:00",
	})
	if err != nil {
		t.Fatalf("SaveShift failed: %v", err)
	}

	dayShifts, err := env.mod.ListRoomDay("tenantA", r.Id, dateSec)
	if err != nil || len(dayShifts) != 1 {
		t.Fatalf("ListRoomDay expected 1 shift, got %d, err: %v", len(dayShifts), err)
	}
	if dayShifts[0].OccupantLabel != "Doc Monday" {
		t.Fatalf("expected 'Doc Monday', got '%s'", dayShifts[0].OccupantLabel)
	}
}
