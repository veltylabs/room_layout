package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/time"
)

func TestFindFreeRooms(t *testing.T) {
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

	dateSec, _ := time.ParseDate("2026-03-16")
	dateSec = dateSec / 1_000_000_000

	// Shift in r1 from 10:00 to 12:00
	_, _ = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r1.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Doc 1",
		Date:          "2026-03-16",
		Start:         "10:00",
		End:           "12:00",
	})

	// Search free rooms between 11:00 and 13:00 -> r1 is busy, r2 is free
	freeRooms, err := env.mod.FindFreeRooms("tenantA", dateSec, 660, 780, "cat1") // 11:00 (660) to 13:00 (780)
	if err != nil {
		t.Fatalf("FindFreeRooms failed: %v", err)
	}
	if len(freeRooms) != 1 || freeRooms[0].RoomId != r2.Id {
		t.Fatalf("expected r2 as free room, got %+v", freeRooms)
	}
}
