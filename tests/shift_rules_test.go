package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/time"
	"webtyp.com/view"
)

func TestShift_CreateThroughCrud(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	if err := env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"}); err != nil {
		t.Fatal(err)
	}

	shiftRec := &roomlayout.RoomShift{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Dr. X",
		DayOfWeek:     1,
		StartMin:      540,
		EndMin:        600,
	}
	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{shiftRec},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error creating shift: %v", err)
	}

	shifts, err := env.mod.ListRoomShifts("tenantA", r.Id)
	if err != nil {
		t.Fatal(err)
	}
	if len(shifts) != 1 {
		t.Fatalf("expected 1 shift, got %d", len(shifts))
	}
	s := shifts[0]
	if !s.IsActive || s.DayOfWeek != 1 || s.TenantId != "tenantA" {
		t.Fatalf("unexpected shift data: %+v", s)
	}
}

func TestShift_RulesOfTheRange(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Dr. X", DayOfWeek: 1, StartMin: 600, EndMin: 540,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid on start >= end, got %v", err)
	}

	err = call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Dr. X", DayOfWeek: 1, StartMin: 540, EndMin: 1500,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid on end > 1440, got %v", err)
	}

	err = call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Dr. X", DayOfWeek: 9, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid on bad day_of_week, got %v", err)
	}

	err = call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: "nope", CategoryId: "cat1", OccupantLabel: "Dr. X", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorNotFound {
		t.Fatalf("expected ErrorNotFound on nonexistent room, got %v", err)
	}
}

func TestShift_CategoryRules(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "zzz", OccupantLabel: "Dr. X", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid on unknown category, got %v", err)
	}

	err = call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat2", OccupantLabel: "Dr. X", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid on not-enabled category, got %v", err)
	}
}

func TestShift_OccupantRules(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid without occupant, got %v", err)
	}

	err = call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantId: "occ1", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	shifts, _ := env.mod.ListRoomShifts("tenantA", r.Id)
	if len(shifts) != 1 || shifts[0].OccupantLabel != "Dr. Juan Pérez" {
		t.Fatalf("expected occupant label Dr. Juan Pérez, got %+v", shifts)
	}

	err = call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantId: "nope", DayOfWeek: 1, StartMin: 660, EndMin: 720,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid on unknown occupant ID, got %v", err)
	}
}

func TestShift_OverlapRejected(t *testing.T) {
	env := newTestEnv("tenantA")
	r1 := roomOf(t, env, "R1")
	r2 := roomOf(t, env, "R2")
	_ = env.mod.SetRoomCategories("tenantA", r1.Id, []string{"cat1"})
	_ = env.mod.SetRoomCategories("tenantA", r2.Id, []string{"cat1"})

	shiftRec := &roomlayout.RoomShift{
		RoomId: r1.Id, CategoryId: "cat1", OccupantId: "occ1", DayOfWeek: 1, StartMin: 540, EndMin: 600,
	}
	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{shiftRec},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Same shift again -> conflict
	err = call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r1.Id, CategoryId: "cat1", OccupantLabel: "Other Doc", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorConflict {
		t.Fatalf("expected ErrorConflict on room overlap, got %v", err)
	}

	// Another room with same occupant -> conflict
	err = call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r2.Id, CategoryId: "cat1", OccupantId: "occ1", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorConflict {
		t.Fatalf("expected ErrorConflict on occupant overlap, got %v", err)
	}

	// Different hours -> success
	err = call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r1.Id, CategoryId: "cat1", OccupantId: "occ1", DayOfWeek: 1, StartMin: 660, EndMin: 720,
		}},
	}, nil)
	if err != nil {
		t.Fatalf("expected success for non-overlapping shift, got %v", err)
	}
}

func TestShift_DatedMustFitOpeningHours(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	date, _ := time.ParseDateUTC("2026-03-16")
	env.hours.setBounds(date, time.DayLimits{Open: true, OpenMin: 480, CloseMin: 600}) // 08:00 - 10:00

	// 09:00 - 11:00 (540 - 660) -> outside bounds
	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", SpecificDate: date, StartMin: 540, EndMin: 660,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid outside bounds, got %v", err)
	}

	// 09:00 - 10:00 (540 - 600) -> success
	err = call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", SpecificDate: date, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if err != nil {
		t.Fatalf("expected success within bounds, got %v", err)
	}

	shifts, _ := env.mod.ListRoomShifts("tenantA", r.Id)
	if len(shifts) != 1 || shifts[0].DayOfWeek != 1 {
		t.Fatalf("expected DayOfWeek 1 computed, got %+v", shifts)
	}
}

func TestShift_UpdateMergesFields(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	shifts, _ := env.mod.ListRoomShifts("tenantA", r.Id)
	shiftID := shifts[0].Id

	err = call("u1", nil, env.mod, "room_shift", model.Update, &crud.UpdateArgs{
		IDs:    []string{shiftID},
		Fields: []string{"end_min"},
		Record: &roomlayout.RoomShift{EndMin: 660},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	updated, _ := env.mod.ListRoomShifts("tenantA", r.Id)
	if updated[0].EndMin != 660 || updated[0].StartMin != 540 {
		t.Fatalf("expected end_min 660 and start_min 540, got %+v", updated[0])
	}

	err = call("u1", nil, env.mod, "room_shift", model.Update, &crud.UpdateArgs{
		IDs:    []string{shiftID},
		Fields: []string{"tenant_id"},
		Record: &roomlayout.RoomShift{},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid updating uneditable field tenant_id, got %v", err)
	}
}

func TestShift_UpdateRoomChecksTheTarget(t *testing.T) {
	env := newTestEnv("tenantA")
	r1 := roomOf(t, env, "R1")
	r2 := roomOf(t, env, "R2")
	_ = env.mod.SetRoomCategories("tenantA", r1.Id, []string{"cat1"})

	date, _ := time.ParseDateUTC("2026-03-16")
	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r1.Id, CategoryId: "cat1", OccupantLabel: "Doc", SpecificDate: date, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	shifts, _ := env.mod.ListRoomShifts("tenantA", r1.Id)
	shiftID := shifts[0].Id

	// Update room_id to R2 which lacks cat1 -> ErrorInvalid
	err = call("u1", nil, env.mod, "room_shift", model.Update, &crud.UpdateArgs{
		IDs:    []string{shiftID},
		Fields: []string{"room_id"},
		Record: &roomlayout.RoomShift{RoomId: r2.Id},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid when target room lacks category, got %v", err)
	}

	// Enable cat1 in R2
	if err := env.mod.SetRoomCategories("tenantA", r2.Id, []string{"cat1"}); err != nil {
		t.Fatal(err)
	}

	err = call("u1", nil, env.mod, "room_shift", model.Update, &crud.UpdateArgs{
		IDs:    []string{shiftID},
		Fields: []string{"room_id"},
		Record: &roomlayout.RoomShift{RoomId: r2.Id},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error updating room_id: %v", err)
	}

	r2Shifts, _ := env.mod.ListRoomShifts("tenantA", r2.Id)
	if len(r2Shifts) != 1 || r2Shifts[0].Id != shiftID {
		t.Fatalf("expected shift to be in R2, got %+v", r2Shifts)
	}
}

func TestShift_DeleteRemovesItsCancellations(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	shifts, _ := env.mod.ListRoomShifts("tenantA", r.Id)
	s := shifts[0]

	date, _ := time.ParseDateUTC("2026-03-16")
	err = call("u1", nil, env.mod, "room_shift_cancellation", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShiftCancellation{
			ShiftId: s.Id, SpecificDate: date,
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = call("u1", nil, env.mod, "room_shift", model.Delete, &crud.DeleteArgs{
		IDs: []string{s.Id},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var cancs roomlayout.RoomShiftCancellationList
	err = call("u1", nil, env.mod, "room_shift_cancellation", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.RoomShiftCancellation{},
	}, &cancs)
	if err != nil {
		t.Fatal(err)
	}
	if len(cancs) != 0 {
		t.Fatalf("expected 0 cancellations after deleting shift, got %d", len(cancs))
	}
}

func TestCancellation_WeeklyOnly(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	date, _ := time.ParseDateUTC("2026-03-16")
	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", SpecificDate: date, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	shifts, _ := env.mod.ListRoomShifts("tenantA", r.Id)
	s := shifts[0]

	err = call("u1", nil, env.mod, "room_shift_cancellation", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShiftCancellation{
			ShiftId: s.Id, SpecificDate: date,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid cancelling non-weekly shift, got %v", err)
	}
}

func TestCancellation_WrongWeekday(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	shifts, _ := env.mod.ListRoomShifts("tenantA", r.Id)
	s := shifts[0]

	tuesdayDate, _ := time.ParseDateUTC("2026-03-17")
	err = call("u1", nil, env.mod, "room_shift_cancellation", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShiftCancellation{
			ShiftId: s.Id, SpecificDate: tuesdayDate,
		}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid cancelling occurrence on wrong weekday, got %v", err)
	}
}

func TestCancellation_Idempotent(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	shifts, _ := env.mod.ListRoomShifts("tenantA", r.Id)
	s := shifts[0]

	mondayDate, _ := time.ParseDateUTC("2026-03-16")
	err = call("u1", nil, env.mod, "room_shift_cancellation", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShiftCancellation{
			ShiftId: s.Id, SpecificDate: mondayDate,
		}},
	}, nil)
	if err != nil {
		t.Fatalf("first cancellation failed: %v", err)
	}

	err = call("u1", nil, env.mod, "room_shift_cancellation", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShiftCancellation{
			ShiftId: s.Id, SpecificDate: mondayDate,
		}},
	}, nil)
	if err != nil {
		t.Fatalf("second cancellation failed: %v", err)
	}

	dayShifts, _ := env.mod.ListRoomDay("tenantA", r.Id, mondayDate)
	if len(dayShifts) != 0 {
		t.Fatalf("expected 0 day shifts after cancellation, got %d", len(dayShifts))
	}
}

func TestCancellation_MoveCreatesDatedShift(t *testing.T) {
	env := newTestEnv("tenantA")
	r1 := roomOf(t, env, "R1")
	r2 := roomOf(t, env, "R2")
	_ = env.mod.SetRoomCategories("tenantA", r1.Id, []string{"cat1"})
	_ = env.mod.SetRoomCategories("tenantA", r2.Id, []string{"cat1"})

	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r1.Id, CategoryId: "cat1", OccupantLabel: "Doc", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	shifts, _ := env.mod.ListRoomShifts("tenantA", r1.Id)
	s := shifts[0]

	mondayDate, _ := time.ParseDateUTC("2026-03-16")
	err = call("u1", nil, env.mod, "room_shift_cancellation", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShiftCancellation{
			ShiftId: s.Id, SpecificDate: mondayDate, MovedToRoomId: r2.Id,
		}},
	}, nil)
	if err != nil {
		t.Fatalf("move cancellation failed: %v", err)
	}

	r2Shifts, _ := env.mod.ListRoomShifts("tenantA", r2.Id)
	if len(r2Shifts) != 1 || r2Shifts[0].SpecificDate != mondayDate {
		t.Fatalf("expected 1 dated shift in R2, got %+v", r2Shifts)
	}

	var cancs roomlayout.RoomShiftCancellationList
	_ = call("u1", nil, env.mod, "room_shift_cancellation", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.RoomShiftCancellation{},
	}, &cancs)
	if len(cancs) != 1 || cancs[0].MovedToRoomId != r2.Id {
		t.Fatalf("expected cancellation with MovedToRoomId == R2.Id, got %+v", cancs)
	}
}

func TestCancellation_MoveToARoomWithoutTheCategory(t *testing.T) {
	env := newTestEnv("tenantA")
	r1 := roomOf(t, env, "R1")
	r3 := roomOf(t, env, "R3")
	_ = env.mod.SetRoomCategories("tenantA", r1.Id, []string{"cat1"})

	err := call("u1", nil, env.mod, "room_shift", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShift{
			RoomId: r1.Id, CategoryId: "cat1", OccupantLabel: "Doc", DayOfWeek: 1, StartMin: 540, EndMin: 600,
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	shifts, _ := env.mod.ListRoomShifts("tenantA", r1.Id)
	s := shifts[0]

	mondayDate, _ := time.ParseDateUTC("2026-03-16")
	err = call("u1", nil, env.mod, "room_shift_cancellation", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomShiftCancellation{
			ShiftId: s.Id, SpecificDate: mondayDate, MovedToRoomId: r3.Id,
		}},
	}, nil)
	if failedField(err) != "moved_to_room_id" {
		t.Fatalf("expected failedField moved_to_room_id, got %v (%q)", err, failedField(err))
	}

	r3Shifts, _ := env.mod.ListRoomShifts("tenantA", r3.Id)
	if len(r3Shifts) != 0 {
		t.Fatalf("expected no shift created in R3, got %d", len(r3Shifts))
	}
}

func TestShiftScreen_SavesAndDeletes(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	caller := view.NewRecords().Caller(crud.LoopbackActingAs(nil, "u1", env.mod))
	p := roomlayout.NewShiftView(caller, r.Id)

	var lastErr error
	done := func(e error) { lastErr = e }

	p.(view.Saver).Save([]model.Model{&roomlayout.RoomShiftForm{
		CategoryId:    "cat1",
		OccupantLabel: "Dr. X",
		DayOfWeek:     roomlayout.WeekdayMonday,
		Start:         "09:00",
		End:           "10:00",
	}}, done)
	if lastErr != nil {
		t.Fatalf("failed to save shift: %v", lastErr)
	}

	shifts, err := env.mod.ListRoomShifts("tenantA", r.Id)
	if err != nil || len(shifts) != 1 {
		t.Fatalf("expected 1 shift, got %d (err %v)", len(shifts), err)
	}

	p.(view.Deleter).Delete([]string{shifts[0].Id}, done)
	if lastErr != nil {
		t.Fatalf("failed to delete shift: %v", lastErr)
	}

	shiftsAfter, _ := env.mod.ListRoomShifts("tenantA", r.Id)
	if len(shiftsAfter) != 0 {
		t.Fatalf("expected 0 shifts after delete, got %d", len(shiftsAfter))
	}
}

func TestFormToShift_Conversions(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	// Bad hour range
	_, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", DayOfWeek: "1", Start: "xx", End: "10:00",
	})
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid on bad start hour, got %v", err)
	}

	// Bad weekday
	_, err = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", DayOfWeek: "9", Start: "09:00", End: "10:00",
	})
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid on bad day of week, got %v", err)
	}

	// Successful weekly
	s1, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", DayOfWeek: "1", Start: "09:30", End: "10:00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s1.StartMin != 570 || s1.EndMin != 600 || s1.DayOfWeek != 1 {
		t.Fatalf("unexpected weekly conversion: %+v", s1)
	}

	// Successful dated
	s2, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", Date: "2026-03-16", Start: "10:30", End: "11:00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dateSec, _ := time.ParseDateUTC("2026-03-16")
	if s2.SpecificDate != dateSec || s2.StartMin != 630 || s2.EndMin != 660 {
		t.Fatalf("unexpected dated conversion: %+v", s2)
	}
}
