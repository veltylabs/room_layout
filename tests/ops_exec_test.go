package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/router"
)

type testExecRegistry struct {
	handlers map[string]router.HandlerFunc
}

func (r *testExecRegistry) Operation(name string, h router.HandlerFunc) router.Route {
	if r.handlers == nil {
		r.handlers = make(map[string]router.HandlerFunc)
	}
	r.handlers[name] = h
	return &mockRoute{name: name}
}

type testOpContext struct {
	args      model.Encodable
	status    int
	resp      model.Encodable
	decodeErr error
}

func (c *testOpContext) Method() string                     { return "POST" }
func (c *testOpContext) Path() string                       { return "" }
func (c *testOpContext) Body() []byte                       { return nil }
func (c *testOpContext) GetHeader(k string) string          { return "" }
func (c *testOpContext) SetHeader(k, v string)              {}
func (c *testOpContext) WriteStatus(code int)               { c.status = code }
func (c *testOpContext) Write(b []byte) (int, error)       { return len(b), nil }
func (c *testOpContext) SetValue(k, v string)               {}
func (c *testOpContext) Value(k string) string             { return "" }
func (c *testOpContext) Param(k string) string             { return "" }
func (c *testOpContext) SetCookie(ck router.Cookie)        {}
func (c *testOpContext) Cookie(k string) (router.Cookie, bool) { return router.Cookie{}, false }
func (c *testOpContext) SetUserID(id string)                {}
func (c *testOpContext) UserID() string                     { return "user1" }

func (c *testOpContext) Decode(into model.Decodable) error {
	if c.decodeErr != nil {
		return c.decodeErr
	}
	switch dst := into.(type) {
	case *roomlayout.ListFloorsArgs:
		if src, ok := c.args.(*roomlayout.ListFloorsArgs); ok {
			*dst = *src
		}
	case *roomlayout.Floor:
		if src, ok := c.args.(*roomlayout.Floor); ok {
			*dst = *src
		}
	case *roomlayout.DeleteFloorArgs:
		if src, ok := c.args.(*roomlayout.DeleteFloorArgs); ok {
			*dst = *src
		}
	case *roomlayout.ListRoomsArgs:
		if src, ok := c.args.(*roomlayout.ListRoomsArgs); ok {
			*dst = *src
		}
	case *roomlayout.GetRoomArgs:
		if src, ok := c.args.(*roomlayout.GetRoomArgs); ok {
			*dst = *src
		}
	case *roomlayout.Room:
		if src, ok := c.args.(*roomlayout.Room); ok {
			*dst = *src
		}
	case *roomlayout.DeactivateRoomArgs:
		if src, ok := c.args.(*roomlayout.DeactivateRoomArgs); ok {
			*dst = *src
		}
	case *roomlayout.ListEquipmentArgs:
		if src, ok := c.args.(*roomlayout.ListEquipmentArgs); ok {
			*dst = *src
		}
	case *roomlayout.Equipment:
		if src, ok := c.args.(*roomlayout.Equipment); ok {
			*dst = *src
		}
	case *roomlayout.DeleteEquipmentArgs:
		if src, ok := c.args.(*roomlayout.DeleteEquipmentArgs); ok {
			*dst = *src
		}
	case *roomlayout.ListRoomEquipmentArgs:
		if src, ok := c.args.(*roomlayout.ListRoomEquipmentArgs); ok {
			*dst = *src
		}
	case *roomlayout.SetRoomEquipmentArgs:
		if src, ok := c.args.(*roomlayout.SetRoomEquipmentArgs); ok {
			*dst = *src
		}
	case *roomlayout.ListOptionsArgs:
		if src, ok := c.args.(*roomlayout.ListOptionsArgs); ok {
			*dst = *src
		}
	case *roomlayout.ListRoomCategoriesArgs:
		if src, ok := c.args.(*roomlayout.ListRoomCategoriesArgs); ok {
			*dst = *src
		}
	case *roomlayout.SetRoomCategoriesArgs:
		if src, ok := c.args.(*roomlayout.SetRoomCategoriesArgs); ok {
			*dst = *src
		}
	case *roomlayout.ListRoomShiftsArgs:
		if src, ok := c.args.(*roomlayout.ListRoomShiftsArgs); ok {
			*dst = *src
		}
	case *roomlayout.RoomShiftForm:
		if src, ok := c.args.(*roomlayout.RoomShiftForm); ok {
			*dst = *src
		}
	case *roomlayout.DeleteRoomShiftArgs:
		if src, ok := c.args.(*roomlayout.DeleteRoomShiftArgs); ok {
			*dst = *src
		}
	case *roomlayout.CancelShiftOccurrenceArgs:
		if src, ok := c.args.(*roomlayout.CancelShiftOccurrenceArgs); ok {
			*dst = *src
		}
	case *roomlayout.MoveShiftOccurrenceArgs:
		if src, ok := c.args.(*roomlayout.MoveShiftOccurrenceArgs); ok {
			*dst = *src
		}
	case *roomlayout.ListRoomDayArgs:
		if src, ok := c.args.(*roomlayout.ListRoomDayArgs); ok {
			*dst = *src
		}
	case *roomlayout.ListBoardArgs:
		if src, ok := c.args.(*roomlayout.ListBoardArgs); ok {
			*dst = *src
		}
	case *roomlayout.FindFreeRoomsArgs:
		if src, ok := c.args.(*roomlayout.FindFreeRoomsArgs); ok {
			*dst = *src
		}
	case *roomlayout.ListPublicAvailabilityArgs:
		if src, ok := c.args.(*roomlayout.ListPublicAvailabilityArgs); ok {
			*dst = *src
		}
	}
	return nil
}

func (c *testOpContext) Encode(v model.Encodable) error {
	c.resp = v
	return nil
}

func TestAllOperationsExecution(t *testing.T) {
	env := newTestEnv("tenantA")
	reg := &testExecRegistry{}
	env.mod.MountOperations(reg)

	// 1. Save floor via op
	ctx := &testOpContext{args: &roomlayout.Floor{TenantId: "tenantA", Name: "Floor Op 1"}}
	reg.handlers[roomlayout.OpSaveFloor](ctx)
	if ctx.resp == nil {
		t.Fatalf("save_floor op failed to encode response")
	}
	fRes := ctx.resp.(*roomlayout.Floor)

	// 2. List floors via op
	ctx = &testOpContext{args: &roomlayout.ListFloorsArgs{TenantId: "tenantA"}}
	reg.handlers[roomlayout.OpListFloors](ctx)

	// 3. Save room via op
	ctx = &testOpContext{args: &roomlayout.Room{TenantId: "tenantA", FloorId: fRes.Id, Code: "ROp1", Name: "Room Op 1", IsActive: true}}
	reg.handlers[roomlayout.OpSaveRoom](ctx)
	rRes := ctx.resp.(*roomlayout.Room)

	// 4. Get room via op
	ctx = &testOpContext{args: &roomlayout.GetRoomArgs{TenantId: "tenantA", Id: rRes.Id}}
	reg.handlers[roomlayout.OpGetRoom](ctx)

	// 5. List rooms via op
	ctx = &testOpContext{args: &roomlayout.ListRoomsArgs{TenantId: "tenantA"}}
	reg.handlers[roomlayout.OpListRooms](ctx)

	// 6. Save equipment via op
	ctx = &testOpContext{args: &roomlayout.Equipment{TenantId: "tenantA", Name: "Eq Op 1"}}
	reg.handlers[roomlayout.OpSaveEquipment](ctx)
	eqRes := ctx.resp.(*roomlayout.Equipment)

	// 7. List equipment via op
	ctx = &testOpContext{args: &roomlayout.ListEquipmentArgs{TenantId: "tenantA"}}
	reg.handlers[roomlayout.OpListEquipment](ctx)

	// 8. Set room equipment via op
	ctx = &testOpContext{args: &roomlayout.SetRoomEquipmentArgs{TenantId: "tenantA", RoomId: rRes.Id, EquipmentIds: []roomlayout.IdRef{{Id: eqRes.Id}}}}
	reg.handlers[roomlayout.OpSetRoomEquipment](ctx)

	// 9. List room equipment via op
	ctx = &testOpContext{args: &roomlayout.ListRoomEquipmentArgs{TenantId: "tenantA", RoomId: rRes.Id}}
	reg.handlers[roomlayout.OpListRoomEquipment](ctx)

	// 10. List categories via op
	ctx = &testOpContext{args: &roomlayout.ListOptionsArgs{TenantId: "tenantA"}}
	reg.handlers[roomlayout.OpListCategories](ctx)

	// 11. Set room categories via op
	ctx = &testOpContext{args: &roomlayout.SetRoomCategoriesArgs{TenantId: "tenantA", RoomId: rRes.Id, CategoryIds: []roomlayout.IdRef{{Id: "cat1"}}}}
	reg.handlers[roomlayout.OpSetRoomCategories](ctx)

	// 12. List room categories via op
	ctx = &testOpContext{args: &roomlayout.ListRoomCategoriesArgs{TenantId: "tenantA", RoomId: rRes.Id}}
	reg.handlers[roomlayout.OpListRoomCategories](ctx)

	// 13. List occupants via op
	ctx = &testOpContext{args: &roomlayout.ListOptionsArgs{TenantId: "tenantA"}}
	reg.handlers[roomlayout.OpListOccupants](ctx)

	// 14. Save room shift via op
	ctx = &testOpContext{args: &roomlayout.RoomShiftForm{RoomId: rRes.Id, CategoryId: "cat1", OccupantLabel: "Doc Op", DayOfWeek: roomlayout.WeekdayMonday, Start: "09:00", End: "13:00"}}
	reg.handlers[roomlayout.OpSaveRoomShift](ctx)
	sRes := ctx.resp.(*roomlayout.RoomShiftForm)

	// 15. List room shifts via op
	ctx = &testOpContext{args: &roomlayout.ListRoomShiftsArgs{TenantId: "tenantA", RoomId: rRes.Id}}
	reg.handlers[roomlayout.OpListRoomShifts](ctx)

	// 16. Cancel shift occurrence via op
	ctx = &testOpContext{args: &roomlayout.CancelShiftOccurrenceArgs{TenantId: "tenantA", ShiftId: sRes.Id, Date: "2026-03-16"}}
	reg.handlers[roomlayout.OpCancelShiftOccurrence](ctx)

	// 17. Save second room via op
	ctx = &testOpContext{args: &roomlayout.Room{TenantId: "tenantA", FloorId: fRes.Id, Code: "ROp2", Name: "Room Op 2", IsActive: true}}
	reg.handlers[roomlayout.OpSaveRoom](ctx)
	rRes2 := ctx.resp.(*roomlayout.Room)
	_ = env.mod.SetRoomCategories("tenantA", rRes2.Id, []string{"cat1"})

	// 18. Move shift occurrence via op
	ctx = &testOpContext{args: &roomlayout.MoveShiftOccurrenceArgs{TenantId: "tenantA", ShiftId: sRes.Id, Date: "2026-03-16", TargetRoomId: rRes2.Id}}
	reg.handlers[roomlayout.OpMoveShiftOccurrence](ctx)

	// 19. List room day via op
	ctx = &testOpContext{args: &roomlayout.ListRoomDayArgs{TenantId: "tenantA", RoomId: rRes2.Id, Date: "2026-03-16"}}
	reg.handlers[roomlayout.OpListRoomDay](ctx)

	// 20. List board via op
	ctx = &testOpContext{args: &roomlayout.ListBoardArgs{TenantId: "tenantA"}}
	reg.handlers[roomlayout.OpListBoard](ctx)

	// 21. Find free rooms via op
	ctx = &testOpContext{args: &roomlayout.FindFreeRoomsArgs{TenantId: "tenantA", Date: "2026-03-16", Start: "14:00", End: "16:00"}}
	reg.handlers[roomlayout.OpFindFreeRooms](ctx)

	// 22. List public availability via op
	ctx = &testOpContext{args: &roomlayout.ListPublicAvailabilityArgs{Date: "2026-03-16"}}
	reg.handlers[roomlayout.OpListPublicAvailability](ctx)

	// 23. Delete room shift via op
	ctx = &testOpContext{args: &roomlayout.DeleteRoomShiftArgs{TenantId: "tenantA", Id: sRes.Id}}
	reg.handlers[roomlayout.OpDeleteRoomShift](ctx)

	// 24. Deactivate room via op
	ctx = &testOpContext{args: &roomlayout.DeactivateRoomArgs{TenantId: "tenantA", Id: rRes.Id}}
	reg.handlers[roomlayout.OpDeactivateRoom](ctx)

	// 25. Delete equipment via op
	ctx = &testOpContext{args: &roomlayout.DeleteEquipmentArgs{TenantId: "tenantA", Id: eqRes.Id}}
	reg.handlers[roomlayout.OpDeleteEquipment](ctx)

	// 26. Delete floor via op
	_ = env.mod.DeactivateRoom("tenantA", rRes2.Id)
	ctx = &testOpContext{args: &roomlayout.DeleteFloorArgs{TenantId: "tenantA", Id: fRes.Id}}
	reg.handlers[roomlayout.OpDeleteFloor](ctx)

	// Error branch tests for Op handlers
	decErrCtx := &testOpContext{decodeErr: fmt.Err("decode error")}
	reg.handlers[roomlayout.OpListFloors](decErrCtx)
	reg.handlers[roomlayout.OpSaveFloor](decErrCtx)
	reg.handlers[roomlayout.OpDeleteFloor](decErrCtx)
	reg.handlers[roomlayout.OpListRooms](decErrCtx)
	reg.handlers[roomlayout.OpGetRoom](decErrCtx)
	reg.handlers[roomlayout.OpSaveRoom](decErrCtx)
	reg.handlers[roomlayout.OpDeactivateRoom](decErrCtx)
	reg.handlers[roomlayout.OpListEquipment](decErrCtx)
	reg.handlers[roomlayout.OpSaveEquipment](decErrCtx)
	reg.handlers[roomlayout.OpDeleteEquipment](decErrCtx)
	reg.handlers[roomlayout.OpListRoomEquipment](decErrCtx)
	reg.handlers[roomlayout.OpSetRoomEquipment](decErrCtx)
	reg.handlers[roomlayout.OpListCategories](decErrCtx)
	reg.handlers[roomlayout.OpListRoomCategories](decErrCtx)
	reg.handlers[roomlayout.OpSetRoomCategories](decErrCtx)
	reg.handlers[roomlayout.OpListOccupants](decErrCtx)
	reg.handlers[roomlayout.OpListRoomShifts](decErrCtx)
	reg.handlers[roomlayout.OpSaveRoomShift](decErrCtx)
	reg.handlers[roomlayout.OpDeleteRoomShift](decErrCtx)
	reg.handlers[roomlayout.OpCancelShiftOccurrence](decErrCtx)
	reg.handlers[roomlayout.OpMoveShiftOccurrence](decErrCtx)
	reg.handlers[roomlayout.OpListRoomDay](decErrCtx)
	reg.handlers[roomlayout.OpListBoard](decErrCtx)
	reg.handlers[roomlayout.OpFindFreeRooms](decErrCtx)
	reg.handlers[roomlayout.OpListPublicAvailability](decErrCtx)

	// NotFound error branch
	notFoundCtx := &testOpContext{args: &roomlayout.GetRoomArgs{TenantId: "tenantA", Id: "non-existent-id"}}
	reg.handlers[roomlayout.OpGetRoom](notFoundCtx)
	if notFoundCtx.status != 404 {
		t.Fatalf("expected status 404 for not found room, got %d", notFoundCtx.status)
	}

	// Invalid date error branch
	invDateCtx := &testOpContext{args: &roomlayout.CancelShiftOccurrenceArgs{TenantId: "tenantA", ShiftId: "s1", Date: "invalid-date"}}
	reg.handlers[roomlayout.OpCancelShiftOccurrence](invDateCtx)
	if invDateCtx.status != 400 {
		t.Fatalf("expected status 400 for invalid date, got %d", invDateCtx.status)
	}
}
