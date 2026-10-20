package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/time"
)

// santiagoAt is the instant of a local hour in America/Santiago (UTC-3 in March 2026) on the given date.
func santiagoAt(t *testing.T, date string, hour, minute int) time.Unix {
	t.Helper()
	day, err := time.ParseDate(date)
	if err != nil {
		t.Fatal(err)
	}
	return time.LocalMinutesToUnixUTC(day, hour*60+minute, "America/Santiago")
}

func TestRoomOfOccupant(t *testing.T) {
	env := newTestEnv("tenantA")
	f, _ := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1"})
	box1, _ := env.mod.SaveRoom(roomlayout.Room{TenantId: "tenantA", FloorId: f.Id, Code: "B1", Name: "Box 1", IsActive: true})
	box2, _ := env.mod.SaveRoom(roomlayout.Room{TenantId: "tenantA", FloorId: f.Id, Code: "B2", Name: "Box 2", IsActive: true})
	for _, r := range []string{box1.Id, box2.Id} {
		if err := env.mod.SetRoomCategories("tenantA", r, []string{"cat1"}); err != nil {
			t.Fatal(err)
		}
	}
	// occ1: Mondays 09:00–12:00 in box 1; on 2026-03-17 (a Tuesday) only, 15:00–16:00 in box 2.
	if _, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{RoomId: box1.Id, CategoryId: "cat1", OccupantId: "occ1", DayOfWeek: roomlayout.WeekdayMonday, Start: "09:00", End: "12:00"}); err != nil {
		t.Fatalf("weekly shift: %v", err)
	}
	if _, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{RoomId: box2.Id, CategoryId: "cat1", OccupantId: "occ1", Date: "2026-03-17", Start: "15:00", End: "16:00"}); err != nil {
		t.Fatalf("dated shift: %v", err)
	}

	cases := []struct {
		name      string
		occupant  string
		at        time.Unix
		wantRoom  string
		wantFound bool
	}{
		{"weekly shift, inside", "occ1", santiagoAt(t, "2026-03-16", 10, 30), box1.Id, true},
		{"weekly shift, at the start", "occ1", santiagoAt(t, "2026-03-16", 9, 0), box1.Id, true},
		{"weekly shift, at the end (excluded)", "occ1", santiagoAt(t, "2026-03-16", 12, 0), "", false},
		{"dated shift", "occ1", santiagoAt(t, "2026-03-17", 15, 15), box2.Id, true},
		{"no shift that day", "occ1", santiagoAt(t, "2026-03-18", 10, 0), "", false},
		{"another occupant", "occ2", santiagoAt(t, "2026-03-16", 10, 30), "", false},
		{"no occupant", "", santiagoAt(t, "2026-03-16", 10, 30), "", false},
	}
	for _, c := range cases {
		room, found, err := env.mod.RoomOfOccupant("tenantA", c.occupant, c.at)
		if err != nil || room != c.wantRoom || found != c.wantFound {
			t.Errorf("%s: got (%q, %v, %v), want (%q, %v, nil)", c.name, room, found, err, c.wantRoom, c.wantFound)
		}
	}

	if _, _, err := env.mod.RoomOfOccupant("", "occ1", santiagoAt(t, "2026-03-16", 10, 0)); err != roomlayout.ErrTenantRequired {
		t.Errorf("no tenant: got %v, want ErrTenantRequired", err)
	}
	if room, found, _ := env.mod.RoomOfOccupant("tenantB", "occ1", santiagoAt(t, "2026-03-16", 10, 30)); found || room != "" {
		t.Errorf("another tenant must not see the shift, got %q", room)
	}
}
