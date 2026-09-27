package roomlayout

import (
	"webtyp.com/model"
	"webtyp.com/router"
)

var _ router.OperationModule = (*Module)(nil)

func (m *Module) ModelName() string {
	return ModelName
}

const (
	OpListFloors             = "list_floors"
	OpSaveFloor              = "save_floor"
	OpDeleteFloor            = "delete_floor"
	OpListRooms              = "list_rooms"
	OpGetRoom               = "get_room"
	OpSaveRoom               = "save_room"
	OpDeactivateRoom         = "deactivate_room"
	OpListEquipment          = "list_equipment"
	OpSaveEquipment          = "save_equipment"
	OpDeleteEquipment        = "delete_equipment"
	OpListRoomEquipment      = "list_room_equipment"
	OpSetRoomEquipment       = "set_room_equipment"
	OpListCategories         = "list_categories"
	OpListRoomCategories     = "list_room_categories"
	OpSetRoomCategories      = "set_room_categories"
	OpListOccupants          = "list_occupants"
	OpListRoomShifts         = "list_room_shifts"
	OpSaveRoomShift          = "save_room_shift"
	OpDeleteRoomShift        = "delete_room_shift"
	OpCancelShiftOccurrence  = "cancel_shift_occurrence"
	OpMoveShiftOccurrence    = "move_shift_occurrence"
	OpListRoomDay            = "list_room_day"
	OpListBoard              = "list_board"
	OpFindFreeRooms          = "find_free_rooms"
	OpListPublicAvailability = "list_public_availability"
)

func mapErrorStatus(err error) int {
	if err == nil {
		return 200
	}
	if err == ErrNotFound {
		return 404
	}
	if err == ErrCodeAlreadyExists || err == ErrFloorInUse || err == ErrCategoryInUse || err == ErrRoomOverlap || err == ErrOccupantOverlap {
		return 409
	}
	if _, ok := err.(*ValidationError); ok || err == ErrTenantRequired || err == ErrInvalidRange || err == ErrInvalidDate || err == ErrInvalidWeekday || err == ErrUnknownCategory || err == ErrUnknownOccupant || err == ErrOccupantRequired || err == ErrCategoryNotAllowed || err == ErrOutsideBounds || err == ErrNotWeekly {
		return 400
	}
	return 500
}

func (m *Module) writeError(ctx router.Context, err error) {
	code := mapErrorStatus(err)
	ctx.WriteStatus(code)
	ctx.Write([]byte(err.Error()))
}

func (m *Module) MountOperations(reg router.OperationRegistry) {
	// list_floors
	reg.Operation(OpListFloors, func(ctx router.Context) {
		var args ListFloorsArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		floors, err := m.ListFloors(tenantID)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := FloorList{}
		for i := range floors {
			list = append(list, &floors[i])
		}
		ctx.Encode(&list)
	}).Requires(ResourceFloor, model.Read).Accepts(&ListFloorsArgs{})

	// save_floor
	reg.Operation(OpSaveFloor, func(ctx router.Context) {
		var floor Floor
		if err := ctx.Decode(&floor); err != nil {
			m.writeError(ctx, err)
			return
		}
		if floor.TenantId == "" {
			floor.TenantId = m.tenantID
		}
		res, err := m.SaveFloor(floor)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.Encode(&res)
	}).Requires(ResourceFloor, model.Create|model.Update).Accepts(&Floor{})

	// delete_floor
	reg.Operation(OpDeleteFloor, func(ctx router.Context) {
		var args DeleteFloorArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		if err := m.DeleteFloor(tenantID, args.Id); err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourceFloor, model.Delete).Accepts(&DeleteFloorArgs{})

	// list_rooms
	reg.Operation(OpListRooms, func(ctx router.Context) {
		var args ListRoomsArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		rooms, err := m.ListRooms(tenantID, args.ActiveOnly)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := RoomList{}
		for i := range rooms {
			list = append(list, &rooms[i])
		}
		ctx.Encode(&list)
	}).Requires(ResourceRoom, model.Read).Accepts(&ListRoomsArgs{})

	// get_room
	reg.Operation(OpGetRoom, func(ctx router.Context) {
		var args GetRoomArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		res, err := m.GetRoom(tenantID, args.Id)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.Encode(&res)
	}).Requires(ResourceRoom, model.Read).Accepts(&GetRoomArgs{})

	// save_room
	reg.Operation(OpSaveRoom, func(ctx router.Context) {
		var room Room
		if err := ctx.Decode(&room); err != nil {
			m.writeError(ctx, err)
			return
		}
		if room.TenantId == "" {
			room.TenantId = m.tenantID
		}
		res, err := m.SaveRoom(room)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.Encode(&res)
	}).Requires(ResourceRoom, model.Create|model.Update).Accepts(&Room{})

	// deactivate_room
	reg.Operation(OpDeactivateRoom, func(ctx router.Context) {
		var args DeactivateRoomArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		if err := m.DeactivateRoom(tenantID, args.Id); err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourceRoom, model.Update).Accepts(&DeactivateRoomArgs{})

	// list_equipment
	reg.Operation(OpListEquipment, func(ctx router.Context) {
		var args ListEquipmentArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		eqs, err := m.ListEquipment(tenantID)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := EquipmentList{}
		for i := range eqs {
			list = append(list, &eqs[i])
		}
		ctx.Encode(&list)
	}).Requires(ResourceEquipment, model.Read).Accepts(&ListEquipmentArgs{})

	// save_equipment
	reg.Operation(OpSaveEquipment, func(ctx router.Context) {
		var eq Equipment
		if err := ctx.Decode(&eq); err != nil {
			m.writeError(ctx, err)
			return
		}
		if eq.TenantId == "" {
			eq.TenantId = m.tenantID
		}
		res, err := m.SaveEquipment(eq)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.Encode(&res)
	}).Requires(ResourceEquipment, model.Create|model.Update).Accepts(&Equipment{})

	// delete_equipment
	reg.Operation(OpDeleteEquipment, func(ctx router.Context) {
		var args DeleteEquipmentArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		if err := m.DeleteEquipment(tenantID, args.Id); err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourceEquipment, model.Delete).Accepts(&DeleteEquipmentArgs{})

	// list_room_equipment
	reg.Operation(OpListRoomEquipment, func(ctx router.Context) {
		var args ListRoomEquipmentArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		eqs, err := m.ListRoomEquipment(tenantID, args.RoomId)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := EquipmentList{}
		for i := range eqs {
			list = append(list, &eqs[i])
		}
		ctx.Encode(&list)
	}).Requires(ResourceRoom, model.Read).Accepts(&ListRoomEquipmentArgs{})

	// set_room_equipment
	reg.Operation(OpSetRoomEquipment, func(ctx router.Context) {
		var args SetRoomEquipmentArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		var ids []string
		for _, ref := range args.EquipmentIds {
			ids = append(ids, ref.Id)
		}
		if err := m.SetRoomEquipment(tenantID, args.RoomId, ids); err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourceRoom, model.Update).Accepts(&SetRoomEquipmentArgs{})

	// list_categories
	reg.Operation(OpListCategories, func(ctx router.Context) {
		var args ListOptionsArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		opts, err := m.deps.Categories.CategoryOptions(tenantID)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := OptionList{}
		for _, kv := range opts {
			list = append(list, &Option{Id: kv.Key, Label: kv.Value})
		}
		ctx.Encode(&list)
	}).Requires(ResourceRoom, model.Read).Accepts(&ListOptionsArgs{})

	// list_room_categories
	reg.Operation(OpListRoomCategories, func(ctx router.Context) {
		var args ListRoomCategoriesArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		opts, err := m.ListRoomCategories(tenantID, args.RoomId)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := OptionList{}
		for _, kv := range opts {
			list = append(list, &Option{Id: kv.Key, Label: kv.Value})
		}
		ctx.Encode(&list)
	}).Requires(ResourceRoom, model.Read).Accepts(&ListRoomCategoriesArgs{})

	// set_room_categories
	reg.Operation(OpSetRoomCategories, func(ctx router.Context) {
		var args SetRoomCategoriesArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		var ids []string
		for _, ref := range args.CategoryIds {
			ids = append(ids, ref.Id)
		}
		if err := m.SetRoomCategories(tenantID, args.RoomId, ids); err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourceRoom, model.Update).Accepts(&SetRoomCategoriesArgs{})

	// list_occupants
	reg.Operation(OpListOccupants, func(ctx router.Context) {
		var args ListOptionsArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		opts, err := m.deps.Occupants.OccupantOptions(tenantID)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := OptionList{}
		for _, kv := range opts {
			list = append(list, &Option{Id: kv.Key, Label: kv.Value})
		}
		ctx.Encode(&list)
	}).Requires(ResourceShift, model.Read).Accepts(&ListOptionsArgs{})

	// list_room_shifts
	reg.Operation(OpListRoomShifts, func(ctx router.Context) {
		var args ListRoomShiftsArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		shifts, err := m.ListRoomShifts(tenantID, args.RoomId)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := RoomShiftFormList{}
		for _, s := range shifts {
			form := shiftToForm(s)
			list = append(list, &form)
		}
		ctx.Encode(&list)
	}).Requires(ResourceShift, model.Read).Accepts(&ListRoomShiftsArgs{})

	// save_room_shift
	reg.Operation(OpSaveRoomShift, func(ctx router.Context) {
		var form RoomShiftForm
		if err := ctx.Decode(&form); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := m.tenantID
		shift, err := m.SaveShift(tenantID, form)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		resForm := shiftToForm(shift)
		ctx.Encode(&resForm)
	}).Requires(ResourceShift, model.Create|model.Update).Accepts(&RoomShiftForm{})

	// delete_room_shift
	reg.Operation(OpDeleteRoomShift, func(ctx router.Context) {
		var args DeleteRoomShiftArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		if err := m.DeleteShift(tenantID, args.Id); err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourceShift, model.Delete).Accepts(&DeleteRoomShiftArgs{})

	// cancel_shift_occurrence
	reg.Operation(OpCancelShiftOccurrence, func(ctx router.Context) {
		var args CancelShiftOccurrenceArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		dateSec, err := parseDate(args.Date)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		if err := m.CancelShiftOccurrence(tenantID, args.ShiftId, dateSec); err != nil {
			m.writeError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourceShift, model.Update).Accepts(&CancelShiftOccurrenceArgs{})

	// move_shift_occurrence
	reg.Operation(OpMoveShiftOccurrence, func(ctx router.Context) {
		var args MoveShiftOccurrenceArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		dateSec, err := parseDate(args.Date)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		shift, err := m.MoveShiftOccurrence(tenantID, args.ShiftId, dateSec, args.TargetRoomId)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		resForm := shiftToForm(shift)
		ctx.Encode(&resForm)
	}).Requires(ResourceShift, model.Create|model.Update).Accepts(&MoveShiftOccurrenceArgs{})

	// list_room_day
	reg.Operation(OpListRoomDay, func(ctx router.Context) {
		var args ListRoomDayArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		var dateSec int64
		if args.Date != "" {
			d, err := parseDate(args.Date)
			if err != nil {
				m.writeError(ctx, err)
				return
			}
			dateSec = d
		}
		shifts, err := m.ListRoomDay(tenantID, args.RoomId, dateSec)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := DayShiftList{}
		for i := range shifts {
			list = append(list, &shifts[i])
		}
		ctx.Encode(&list)
	}).Requires(ResourceShift, model.Read).Accepts(&ListRoomDayArgs{})

	// list_board
	reg.Operation(OpListBoard, func(ctx router.Context) {
		var args ListBoardArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		board, err := m.ListBoard(tenantID, args.CategoryId, args.EquipmentId)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := BoardRoomList{}
		for i := range board {
			list = append(list, &board[i])
		}
		ctx.Encode(&list)
	}).Requires(ResourceRoom, model.Read).Accepts(&ListBoardArgs{})

	// find_free_rooms
	reg.Operation(OpFindFreeRooms, func(ctx router.Context) {
		var args FindFreeRoomsArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		tenantID := args.TenantId
		if tenantID == "" {
			tenantID = m.tenantID
		}
		dateSec, errDate := parseDate(args.Date)
		startMin, errStart := parseHour(args.Start)
		endMin, errEnd := parseHour(args.End)
		if errDate != nil || errStart != nil || errEnd != nil {
			m.writeError(ctx, ErrInvalidRange)
			return
		}
		freeRooms, err := m.FindFreeRooms(tenantID, dateSec, startMin, endMin, args.CategoryId)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := FreeRoomList{}
		for i := range freeRooms {
			list = append(list, &freeRooms[i])
		}
		ctx.Encode(&list)
	}).Requires(ResourceRoom, model.Read).Accepts(&FindFreeRoomsArgs{})

	// list_public_availability (Public)
	reg.Operation(OpListPublicAvailability, func(ctx router.Context) {
		var args ListPublicAvailabilityArgs
		if err := ctx.Decode(&args); err != nil {
			m.writeError(ctx, err)
			return
		}
		dateSec, errDate := parseDate(args.Date)
		if errDate != nil {
			m.writeError(ctx, errDate)
			return
		}
		slots, err := m.ListPublicAvailability(dateSec, args.CategoryId)
		if err != nil {
			m.writeError(ctx, err)
			return
		}
		list := PublicSlotList{}
		for i := range slots {
			list = append(list, &slots[i])
		}
		ctx.Encode(&list)
	}).Public().Accepts(&ListPublicAvailabilityArgs{})
}
