package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
)

func TestCategoryOperations(t *testing.T) {
	env := newTestEnv("tenantA")

	f, _ := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1"})
	r, _ := env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  f.Id,
		Code:     "R101",
		Name:     "Room 101",
		IsActive: true,
	})

	err := env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1", "cat2"})
	if err != nil {
		t.Fatalf("SetRoomCategories failed: %v", err)
	}

	cats, err := env.mod.ListRoomCategories("tenantA", r.Id)
	if err != nil || len(cats) != 2 {
		t.Fatalf("ListRoomCategories expected 2, got %d, err: %v", len(cats), err)
	}

	// Create shift in cat1
	_, err = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Doc A",
		DayOfWeek:     roomlayout.WeekdayMonday,
		Start:         "09:00",
		End:           "12:00",
	})
	if err != nil {
		t.Fatalf("SaveShift failed: %v", err)
	}

	// Attempting to remove cat1 should return ErrCategoryInUse
	err = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat2"})
	if err != roomlayout.ErrCategoryInUse {
		t.Fatalf("expected ErrCategoryInUse, got %v", err)
	}
}
