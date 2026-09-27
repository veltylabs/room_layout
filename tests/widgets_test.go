package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/form"
)

func TestFormWidgetsRegression(t *testing.T) {
	ids := &simpleIDs{}

	// 1. RoomShiftForm
	fShift, err := form.New("p1", &roomlayout.RoomShiftForm{}, ids)
	if err != nil {
		t.Fatalf("form.New for RoomShiftForm failed: %v", err)
	}
	if fShift == nil {
		t.Fatalf("expected non-nil form for RoomShiftForm")
	}

	// 2. Room
	fRoom, err := form.New("p2", &roomlayout.Room{}, ids)
	if err != nil {
		t.Fatalf("form.New for Room failed: %v", err)
	}
	if fRoom == nil {
		t.Fatalf("expected non-nil form for Room")
	}

	// 3. Floor
	fFloor, err := form.New("p3", &roomlayout.Floor{}, ids)
	if err != nil {
		t.Fatalf("form.New for Floor failed: %v", err)
	}
	if fFloor == nil {
		t.Fatalf("expected non-nil form for Floor")
	}

	// 4. Equipment
	fEq, err := form.New("p4", &roomlayout.Equipment{}, ids)
	if err != nil {
		t.Fatalf("form.New for Equipment failed: %v", err)
	}
	if fEq == nil {
		t.Fatalf("expected non-nil form for Equipment")
	}
}
