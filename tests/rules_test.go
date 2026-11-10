package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/crud"
	"webtyp.com/events"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/time"
)

// call runs one operation of the module as user through the same loopback the demo uses and answers the callback error.
func call(user string, pub events.Publisher, mod *roomlayout.Module, entity string, action model.Action, args model.Encodable, into model.Decodable) error {
	var callErr error
	crud.LoopbackActingAs(pub, user, mod).Call(
		router.Operation{Module: roomlayout.ModelName, Entity: entity, Action: action},
		args, into, func(err error) { callErr = err })
	return callErr
}

// failedField answers the field a refusal points at, or "".
func failedField(err error) string {
	if f, ok := err.(interface{ FailedField() string }); ok {
		return f.FailedField()
	}
	return ""
}

// roomOf creates a floor and a room of tenantA through the kept methods and answers the room.
func roomOf(t *testing.T, env *testEnv, code string) roomlayout.Room {
	t.Helper()
	f, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso " + code})
	if err != nil {
		t.Fatal(err)
	}
	r, err := env.mod.SaveRoom(roomlayout.Room{TenantId: "tenantA", FloorId: f.Id, Code: code, Name: "Sala " + code, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestFloor_CreateThroughCrud(t *testing.T) {
	env := newTestEnv("tenantA")
	err := call("u1", nil, env.mod, "room_floor", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.Floor{Name: "  Piso 1  ", Cols: 10, Rows: 10}},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	floors, err := env.mod.ListFloors("tenantA")
	if err != nil {
		t.Fatal(err)
	}
	if len(floors) != 1 {
		t.Fatalf("expected 1 floor, got %d", len(floors))
	}
	if floors[0].Name != "Piso 1" || floors[0].TenantId != "tenantA" {
		t.Fatalf("unexpected floor data: %+v", floors[0])
	}
}

func TestFloor_BlankNameRejected(t *testing.T) {
	env := newTestEnv("tenantA")
	err := call("u1", nil, env.mod, "room_floor", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.Floor{Name: "   "}},
	}, nil)
	if model.KindOf(err) != model.ErrorInvalid || failedField(err) != "name" {
		t.Fatalf("expected ErrorInvalid on name, got: %v (field %q)", err, failedField(err))
	}
}

func TestFloor_DeleteWithRoomsRefused(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	err := call("u1", nil, env.mod, "room_floor", model.Delete, &crud.DeleteArgs{IDs: []string{r.FloorId}}, nil)
	if model.KindOf(err) != model.ErrorConflict {
		t.Fatalf("expected ErrorConflict, got: %v", err)
	}

	f2, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso Vacío"})
	if err != nil {
		t.Fatal(err)
	}
	err = call("u1", nil, env.mod, "room_floor", model.Delete, &crud.DeleteArgs{IDs: []string{f2.Id}}, nil)
	if err != nil {
		t.Fatalf("expected no error deleting empty floor, got: %v", err)
	}
}

func TestRoom_CodeIsUniquePerTenant(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	err := call("u1", nil, env.mod, "room_space", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.Room{FloorId: r.FloorId, Code: "R1", Name: "xx", IsActive: true}},
	}, nil)
	if model.KindOf(err) != model.ErrorConflict {
		t.Fatalf("expected ErrorConflict on duplicate code, got: %v", err)
	}

	err = call("u1", nil, env.mod, "room_space", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.Room{FloorId: r.FloorId, Code: "R2", Name: "x2", IsActive: true}},
	}, nil)
	if err != nil {
		t.Fatalf("expected create R2 success, got: %v", err)
	}

	rooms, _ := env.mod.ListRooms("tenantA", false)
	var r2ID string
	for _, room := range rooms {
		if room.Code == "R2" {
			r2ID = room.Id
			break
		}
	}

	err = call("u1", nil, env.mod, "room_space", model.Update, &crud.UpdateArgs{
		IDs:    []string{r2ID},
		Fields: []string{"code"},
		Record: &roomlayout.Room{Code: "R1"},
	}, nil)
	if model.KindOf(err) != model.ErrorConflict {
		t.Fatalf("expected ErrorConflict on update code to duplicate, got: %v", err)
	}
}

func TestRoom_DeactivateIsAnUpdate(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	err := call("u1", nil, env.mod, "room_space", model.Update, &crud.UpdateArgs{
		IDs:    []string{r.Id},
		Fields: []string{"is_active"},
		Record: &roomlayout.Room{IsActive: false},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := env.mod.GetRoom("tenantA", r.Id)
	if err != nil {
		t.Fatal(err)
	}
	if got.IsActive || got.Code != "R1" {
		t.Fatalf("expected inactive room R1, got: %+v", got)
	}
}

func TestEquipment_DeleteRemovesItsLinks(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	e, err := env.mod.SaveEquipment(roomlayout.Equipment{TenantId: "tenantA", Name: "Camilla"})
	if err != nil {
		t.Fatal(err)
	}
	if err := env.mod.SetRoomEquipment("tenantA", r.Id, []string{e.Id}); err != nil {
		t.Fatal(err)
	}

	err = call("u1", nil, env.mod, "room_equipment", model.Delete, &crud.DeleteArgs{IDs: []string{e.Id}}, nil)
	if err != nil {
		t.Fatalf("unexpected error deleting equipment: %v", err)
	}

	eqs, err := env.mod.ListRoomEquipment("tenantA", r.Id)
	if err != nil {
		t.Fatal(err)
	}
	if len(eqs) != 0 {
		t.Fatalf("expected 0 room equipment, got %d", len(eqs))
	}

	var links roomlayout.RoomEquipmentList
	err = call("u1", nil, env.mod, "room_space_equipment", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.RoomEquipment{RoomId: r.Id},
	}, &links)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links, got %d", len(links))
	}
}

func TestRoomEquipment_CompositeKey(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	e, err := env.mod.SaveEquipment(roomlayout.Equipment{TenantId: "tenantA", Name: "Camilla"})
	if err != nil {
		t.Fatal(err)
	}

	err = call("u1", nil, env.mod, "room_space_equipment", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomEquipment{RoomId: r.Id, EquipmentId: e.Id}},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = call("u1", nil, env.mod, "room_space_equipment", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomEquipment{RoomId: r.Id, EquipmentId: e.Id}},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error on duplicate link (idempotent create): %v", err)
	}

	err = call("u1", nil, env.mod, "room_space_equipment", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomEquipment{RoomId: "nope", EquipmentId: e.Id}},
	}, nil)
	if failedField(err) != "room_id" {
		t.Fatalf("expected failedField room_id, got: %v (%q)", err, failedField(err))
	}

	err = call("u1", nil, env.mod, "room_space_equipment", model.Delete, &crud.DeleteArgs{
		IDs: []string{model.JoinKey(r.Id, e.Id)},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	eqs, _ := env.mod.ListRoomEquipment("tenantA", r.Id)
	if len(eqs) != 0 {
		t.Fatalf("expected 0 links after delete, got %d", len(eqs))
	}
}

func TestRoomCategory_UnknownCategoryRejected(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	err := call("u1", nil, env.mod, "room_category", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomCategory{RoomId: r.Id, CategoryId: "nope"}},
	}, nil)
	if failedField(err) != "category_id" {
		t.Fatalf("expected failedField category_id, got: %v (%q)", err, failedField(err))
	}

	err = call("u1", nil, env.mod, "room_category", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.RoomCategory{RoomId: r.Id, CategoryId: "cat1"}},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cats, err := env.mod.ListRoomCategories("tenantA", r.Id)
	if err != nil || len(cats) != 1 {
		t.Fatalf("expected 1 category, got %d (err %v)", len(cats), err)
	}
}

func TestRoomCategory_DeleteInUseRefused(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	if err := env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"}); err != nil {
		t.Fatal(err)
	}
	shift, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId:        r.Id,
		CategoryId:    "cat1",
		OccupantLabel: "Doc",
		DayOfWeek:     roomlayout.WeekdayMonday,
		Start:         "09:00",
		End:           "10:00",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = call("u1", nil, env.mod, "room_category", model.Delete, &crud.DeleteArgs{
		IDs: []string{model.JoinKey(r.Id, "cat1")},
	}, nil)
	if model.KindOf(err) != model.ErrorConflict {
		t.Fatalf("expected ErrorConflict deleting in-use category, got: %v", err)
	}

	if err := env.mod.DeleteShift("tenantA", shift.Id); err != nil {
		t.Fatal(err)
	}

	err = call("u1", nil, env.mod, "room_category", model.Delete, &crud.DeleteArgs{
		IDs: []string{model.JoinKey(r.Id, "cat1")},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error deleting free category: %v", err)
	}
}

func TestRoomShift_ReadByRoom(t *testing.T) {
	env := newTestEnv("tenantA")
	r1 := roomOf(t, env, "R1")
	r2 := roomOf(t, env, "R2")
	_ = env.mod.SetRoomCategories("tenantA", r1.Id, []string{"cat1"})
	_ = env.mod.SetRoomCategories("tenantA", r2.Id, []string{"cat1"})

	_, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId: r1.Id, CategoryId: "cat1", OccupantLabel: "Doc 1", DayOfWeek: roomlayout.WeekdayMonday, Start: "09:00", End: "10:00",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId: r2.Id, CategoryId: "cat1", OccupantLabel: "Doc 2", DayOfWeek: roomlayout.WeekdayMonday, Start: "09:00", End: "10:00",
	})
	if err != nil {
		t.Fatal(err)
	}

	var list roomlayout.RoomShiftList
	err = call("u1", nil, env.mod, "room_shift", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.RoomShift{RoomId: r1.Id},
	}, &list)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].RoomId != r1.Id {
		t.Fatalf("expected 1 shift for room r1, got %+v", list)
	}
}

func TestReadModel_Options(t *testing.T) {
	env := newTestEnv("tenantA")

	var catList roomlayout.CategoryOptionList
	err := call("u1", nil, env.mod, "category_option", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.CategoryOption{},
	}, &catList)
	if err != nil {
		t.Fatal(err)
	}
	if len(catList) != 3 {
		t.Fatalf("expected 3 category options, got %d", len(catList))
	}

	var occList roomlayout.OccupantOptionList
	err = call("u1", nil, env.mod, "occupant_option", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.OccupantOption{},
	}, &occList)
	if err != nil {
		t.Fatal(err)
	}
	if len(occList) != 2 {
		t.Fatalf("expected 2 occupant options, got %d", len(occList))
	}
}

func TestReadModel_DayShift(t *testing.T) {
	env := newTestEnv("tenantA")
	r := roomOf(t, env, "R1")
	_ = env.mod.SetRoomCategories("tenantA", r.Id, []string{"cat1"})

	created, err := env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{
		RoomId: r.Id, CategoryId: "cat1", OccupantLabel: "Doc", DayOfWeek: roomlayout.WeekdayMonday, Start: "09:00", End: "10:00",
	})
	if err != nil {
		t.Fatal(err)
	}

	dateSec, _ := time.ParseDateUTC("2026-03-16") // Monday
	expectedShifts, err := env.mod.ListRoomDay("tenantA", r.Id, dateSec)
	if err != nil {
		t.Fatal(err)
	}

	var list roomlayout.DayShiftList
	err = call("u1", nil, env.mod, "day_shift", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.DayShift{RoomId: r.Id, Date: "2026-03-16"},
	}, &list)
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != len(expectedShifts) {
		t.Fatalf("expected %d shifts, got %d", len(expectedShifts), len(list))
	}
	if len(list) > 0 && list[0].ShiftId != created.Id {
		t.Fatalf("expected shift ID %q, got %q", created.Id, list[0].ShiftId)
	}
}

func TestReadModel_FreeRoomNeedsItsFilter(t *testing.T) {
	env := newTestEnv("tenantA")
	_ = roomOf(t, env, "R1")

	var list roomlayout.FreeRoomList
	err := call("u1", nil, env.mod, "free_room", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.FreeRoom{},
	}, &list)
	if failedField(err) != "date" {
		t.Fatalf("expected failedField date, got: %v", err)
	}

	err = call("u1", nil, env.mod, "free_room", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.FreeRoom{Date: "2026-03-16", Start: "xx", End: "10:00"},
	}, &list)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Fatalf("expected ErrorInvalid on bad start, got: %v", err)
	}

	dateSec, _ := time.ParseDateUTC("2026-03-16")
	expectedFree, err := env.mod.FindFreeRooms("tenantA", dateSec, 540, 600, "")
	if err != nil {
		t.Fatal(err)
	}

	err = call("u1", nil, env.mod, "free_room", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.FreeRoom{Date: "2026-03-16", Start: "09:00", End: "10:00"},
	}, &list)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != len(expectedFree) {
		t.Fatalf("expected %d free rooms, got %d", len(expectedFree), len(list))
	}
}

func TestReadModel_PublicSlotNeedsDate(t *testing.T) {
	env := newTestEnv("tenantA")

	var list roomlayout.PublicSlotList
	err := call("u1", nil, env.mod, "public_slot", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.PublicSlot{},
	}, &list)
	if failedField(err) != "date" {
		t.Fatalf("expected failedField date, got: %v", err)
	}

	err = call("u1", nil, env.mod, "public_slot", model.Read, &crud.ReadArgs{
		Limit:  crud.MaxPage,
		Filter: &roomlayout.PublicSlot{Date: "2026-03-16"},
	}, &list)
	if err != nil {
		t.Fatalf("unexpected error with valid date: %v", err)
	}
}

func TestTenant_CrudSeesOnlyItsTenant(t *testing.T) {
	env := newTestEnv("tenantA")
	_, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantB", Name: "Otro"})
	if err != nil {
		t.Fatal(err)
	}

	var list roomlayout.FloorList
	err = call("u1", nil, env.mod, "room_floor", model.Read, &crud.ReadArgs{Limit: crud.MaxPage}, &list)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 floors for tenantA, got %d", len(list))
	}
}

func TestMount_MatchesDefinition(t *testing.T) {
	env := newTestEnv("tenantA")
	if err := crud.Mount(nil, nil, nil, env.mod); err != nil {
		t.Fatalf("crud.Mount failed: %v", err)
	}
}

func TestCreate_PublishesChange(t *testing.T) {
	env := newTestEnv("tenantA")
	pub := &mockPublisher{}
	err := call("u1", pub, env.mod, "room_floor", model.Create, &crud.CreateArgs{
		Records: []model.Model{&roomlayout.Floor{Name: "Piso Pub"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(pub.events))
	}
	if _, ok := pub.events[0].Payload.(*crud.Change); !ok {
		t.Fatalf("expected payload to be *crud.Change, got %T", pub.events[0].Payload)
	}
}
