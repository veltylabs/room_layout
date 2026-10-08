package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"
)

// ModelName es la identidad del módulo: prefijo de cada op en el cable.
const ModelName = "room_layout"

// Recursos RBAC.
const (
	ResourceFloor     = "floor"
	ResourceRoom      = "room"
	ResourceEquipment = "equipment"
	ResourceShift     = "room_shift"
)

// Constantes de estado del tablero.
const (
	BoardStatusFree   = "free"
	BoardStatusBusy   = "busy"
	BoardStatusClosed = "closed"
)

// Duración mínima de slot público.
const MinPublicSlotMinutes = 15

// Constantes para días de la semana.
const (
	WeekdaySunday    = "0"
	WeekdayMonday    = "1"
	WeekdayTuesday   = "2"
	WeekdayWednesday = "3"
	WeekdayThursday  = "4"
	WeekdayFriday    = "5"
	WeekdaySaturday  = "6"
)

// Topics de eventos.
const (
	TopicFloorSaved               = "room_layout.floor.saved"
	TopicFloorDeleted             = "room_layout.floor.deleted"
	TopicRoomSaved                = "room_layout.room.saved"
	TopicRoomDeactivated          = "room_layout.room.deactivated"
	TopicEquipmentSaved           = "room_layout.equipment.saved"
	TopicEquipmentDeleted         = "room_layout.equipment.deleted"
	TopicShiftSaved               = "room_layout.shift.saved"
	TopicShiftDeleted             = "room_layout.shift.deleted"
	TopicShiftOccurrenceCancelled = "room_layout.shift.occurrence_cancelled"
	TopicShiftOccurrenceMoved     = "room_layout.shift.occurrence_moved"
)

// Errores de dominio.
var (
	ErrNotFound           = fmt.Err("room_layout: not found")
	ErrTenantRequired     = fmt.Err("room_layout: tenant_id is required")
	ErrCodeAlreadyExists  = fmt.Err("room_layout: room code already exists in this tenant")
	ErrFloorInUse         = fmt.Err("room_layout: floor still has rooms")
	ErrInvalidRange       = fmt.Err("room_layout: start must be before end, both within 00:00 and 24:00")
	ErrInvalidDate        = fmt.Err("room_layout: invalid date, expected YYYY-MM-DD")
	ErrInvalidWeekday     = fmt.Err("room_layout: day_of_week must be 0..6")
	ErrUnknownCategory    = fmt.Err("room_layout: unknown category")
	ErrUnknownOccupant    = fmt.Err("room_layout: unknown occupant")
	ErrOccupantRequired   = fmt.Err("room_layout: occupant_id or occupant_label is required")
	ErrCategoryNotAllowed = fmt.Err("room_layout: room is not enabled for this category")
	ErrCategoryInUse      = fmt.Err("room_layout: category still has active shifts in this room")
	ErrRoomOverlap        = fmt.Err("room_layout: room already has a shift in that time range")
	ErrOccupantOverlap    = fmt.Err("room_layout: occupant already has a shift in that time range")
	ErrOutsideBounds      = fmt.Err("room_layout: shift is outside the establishment's opening hours for that date")
	ErrNotWeekly          = fmt.Err("room_layout: only a weekly shift can have one occurrence cancelled")
)

type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string {
	if e == nil || e.Err == nil {
		return "room_layout: validation error"
	}
	return e.Err.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func weekdayInput() input.Input {
	return input.Select(
		fmt.KeyValue{Key: WeekdaySunday, Value: "Domingo"},
		fmt.KeyValue{Key: WeekdayMonday, Value: "Lunes"},
		fmt.KeyValue{Key: WeekdayTuesday, Value: "Martes"},
		fmt.KeyValue{Key: WeekdayWednesday, Value: "Miércoles"},
		fmt.KeyValue{Key: WeekdayThursday, Value: "Jueves"},
		fmt.KeyValue{Key: WeekdayFriday, Value: "Viernes"},
		fmt.KeyValue{Key: WeekdaySaturday, Value: "Sábado"},
	)
}

var FloorModel = model.Definition{
	Name: "floor",
	Fields: model.Fields{
		{Name: "cols", Type: model.Int()},
		{Name: "rows", Type: model.Int()},
		{Name: "grid_locked", Type: model.Bool()},
		{Name: "habitable_cells", Type: model.Text()},
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "name", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 60}},
		{Name: "position", Type: input.Number(), OmitEmpty: true},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}

var RoomModel = model.Definition{
	Name: "room",
	Fields: model.Fields{
		{Name: "room_type", Type: model.Text()},
		{Name: "cells", Type: model.Text()},
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "floor_id", Type: input.Select(), Ref: &FloorModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-', '_'}}},
		{Name: "code", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-'}, Minimum: 1, Maximum: 20}},
		{Name: "name", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 120}},
		{Name: "notes", Type: input.Textarea(), OmitEmpty: true},
		{Name: "is_active", Type: input.Checkbox(), NotNull: true},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}


var RoomArtifactModel = model.Definition{
	Name: "room_artifact",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), NotNull: true},
		{Name: "kind", Type: model.Text(), NotNull: true},
		{Name: "code", Type: model.Text(), NotNull: true},
		{Name: "cell", Type: model.Text(), NotNull: true},
		{Name: "status", Type: model.Text(), NotNull: true},
		{Name: "reason", Type: model.Text()},
		{Name: "next_maintenance", Type: model.Text()},
		{Name: "items", Type: model.Text()},
	},
}

var EquipmentModel = model.Definition{
	Name: "equipment",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "name", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 80}},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}

var RoomEquipmentModel = model.Definition{
	Name: "room_equipment",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), Ref: &RoomModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
		{Name: "equipment_id", Type: model.Text(), Ref: &EquipmentModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
	},
}

var RoomCategoryModel = model.Definition{
	Name: "room_category",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), Ref: &RoomModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
		{Name: "category_id", Type: model.Text(), DB: &model.FieldDB{PK: true}, NotNull: true},
	},
}

var RoomShiftModel = model.Definition{
	Name: "room_shift",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), Ref: &RoomModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "category_id", Type: model.Text(), NotNull: true},
		{Name: "occupant_id", Type: model.Text(), OmitEmpty: true},
		{Name: "occupant_label", Type: model.Text(), NotNull: true},
		{Name: "day_of_week", Type: model.Int()},
		{Name: "specific_date", Type: model.Int()},
		{Name: "start_min", Type: model.Int(), NotNull: true},
		{Name: "end_min", Type: model.Int(), NotNull: true},
		{Name: "is_active", Type: model.Bool(), NotNull: true},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}

var RoomShiftCancellationModel = model.Definition{
	Name: "room_shift_cancellation",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "shift_id", Type: model.Text(), Ref: &RoomShiftModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
		{Name: "specific_date", Type: model.Int(), DB: &model.FieldDB{PK: true}, NotNull: true},
	},
}

var RoomShiftFormModel = model.Definition{
	Name: "room_shift_form",
	Fields: model.Fields{
		{Name: "id", Type: input.Text(), OmitEmpty: true},
		{Name: "room_id", Type: input.Select(), NotNull: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-', '_'}}},
		{Name: "category_id", Type: input.Select(), NotNull: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-', '_'}}},
		{Name: "occupant_id", Type: input.Select(), OmitEmpty: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-', '_'}}},
		{Name: "occupant_label", Type: input.Text(), OmitEmpty: true, Permitted: model.Permitted{Maximum: 120}},
		{Name: "day_of_week", Type: input.Select(), OmitEmpty: true},
		{Name: "date", Type: input.Date(), OmitEmpty: true},
		{Name: "start", Type: input.Hour(), NotNull: true},
		{Name: "end", Type: input.Hour(), NotNull: true},
	},
}

var IdRefModel = model.Definition{
	Name: "id_ref",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), NotNull: true},
	},
}

// Argument definitions.
var ListFloorsArgsModel = model.Definition{
	Name: "list_floors_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
	},
}

var DeleteFloorArgsModel = model.Definition{
	Name: "delete_floor_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "id", Type: model.Text(), NotNull: true},
	},
}

var ListRoomsArgsModel = model.Definition{
	Name: "list_rooms_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "active_only", Type: model.Bool()},
	},
}

var GetRoomArgsModel = model.Definition{
	Name: "get_room_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "id", Type: model.Text(), NotNull: true},
	},
}

var DeactivateRoomArgsModel = model.Definition{
	Name: "deactivate_room_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "id", Type: model.Text(), NotNull: true},
	},
}

var ListEquipmentArgsModel = model.Definition{
	Name: "list_equipment_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
	},
}

var DeleteEquipmentArgsModel = model.Definition{
	Name: "delete_equipment_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "id", Type: model.Text(), NotNull: true},
	},
}

var ListRoomEquipmentArgsModel = model.Definition{
	Name: "list_room_equipment_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), NotNull: true},
	},
}

var SetRoomEquipmentArgsModel = model.Definition{
	Name: "set_room_equipment_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), NotNull: true},
		{Name: "equipment_ids", Type: model.StructSlice(&IdRefModel)},
	},
}

var ListRoomCategoriesArgsModel = model.Definition{
	Name: "list_room_categories_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), NotNull: true},
	},
}

var SetRoomCategoriesArgsModel = model.Definition{
	Name: "set_room_categories_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), NotNull: true},
		{Name: "category_ids", Type: model.StructSlice(&IdRefModel)},
	},
}

var ListOptionsArgsModel = model.Definition{
	Name: "list_options_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
	},
}

var ListRoomShiftsArgsModel = model.Definition{
	Name: "list_room_shifts_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), NotNull: true},
	},
}

var DeleteRoomShiftArgsModel = model.Definition{
	Name: "delete_room_shift_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "id", Type: model.Text(), NotNull: true},
	},
}

var CancelShiftOccurrenceArgsModel = model.Definition{
	Name: "cancel_shift_occurrence_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "shift_id", Type: model.Text(), NotNull: true},
		{Name: "date", Type: model.Text(), NotNull: true},
	},
}

var MoveShiftOccurrenceArgsModel = model.Definition{
	Name: "move_shift_occurrence_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "shift_id", Type: model.Text(), NotNull: true},
		{Name: "date", Type: model.Text(), NotNull: true},
		{Name: "target_room_id", Type: model.Text(), NotNull: true},
	},
}

var ListRoomDayArgsModel = model.Definition{
	Name: "list_room_day_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text()},
		{Name: "date", Type: model.Text()},
	},
}

var ListBoardArgsModel = model.Definition{
	Name: "list_board_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "category_id", Type: model.Text()},
		{Name: "equipment_id", Type: model.Text()},
	},
}

var FindFreeRoomsArgsModel = model.Definition{
	Name: "find_free_rooms_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "date", Type: model.Text(), NotNull: true},
		{Name: "start", Type: model.Text(), NotNull: true},
		{Name: "end", Type: model.Text(), NotNull: true},
		{Name: "category_id", Type: model.Text()},
	},
}

var ListPublicAvailabilityArgsModel = model.Definition{
	Name: "list_public_availability_args",
	Fields: model.Fields{
		{Name: "date", Type: model.Text(), NotNull: true},
		{Name: "category_id", Type: model.Text()},
	},
}

// Result / Output Definitions.
var OptionModel = model.Definition{
	Name: "option",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), NotNull: true},
		{Name: "label", Type: model.Text(), NotNull: true},
	},
}

var DayShiftModel = model.Definition{
	Name: "day_shift",
	Fields: model.Fields{
		{Name: "shift_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), NotNull: true},
		{Name: "room_code", Type: model.Text(), NotNull: true},
		{Name: "category_id", Type: model.Text(), NotNull: true},
		{Name: "category_label", Type: model.Text(), NotNull: true},
		{Name: "occupant_id", Type: model.Text()},
		{Name: "occupant_label", Type: model.Text(), NotNull: true},
		{Name: "start", Type: model.Text(), NotNull: true},
		{Name: "end", Type: model.Text(), NotNull: true},
		{Name: "is_weekly", Type: model.Bool(), NotNull: true},
	},
}

var BoardRoomModel = model.Definition{
	Name: "board_room",
	Fields: model.Fields{
		{Name: "room_id", Type: model.Text(), NotNull: true},
		{Name: "code", Type: model.Text(), NotNull: true},
		{Name: "name", Type: model.Text(), NotNull: true},
		{Name: "floor_name", Type: model.Text(), NotNull: true},
		{Name: "floor_position", Type: model.Int(), NotNull: true},
		{Name: "category_labels", Type: model.Text(), NotNull: true},
		{Name: "equipment_labels", Type: model.Text(), NotNull: true},
		{Name: "status", Type: model.Text(), NotNull: true},
		{Name: "current_occupant_label", Type: model.Text()},
		{Name: "current_until", Type: model.Text()},
		{Name: "next_start", Type: model.Text()},
		{Name: "next_occupant_label", Type: model.Text()},
		{Name: "today", Type: model.Text(), NotNull: true},
	},
}

var FreeRoomModel = model.Definition{
	Name: "free_room",
	Fields: model.Fields{
		{Name: "room_id", Type: model.Text(), NotNull: true},
		{Name: "code", Type: model.Text(), NotNull: true},
		{Name: "name", Type: model.Text(), NotNull: true},
		{Name: "floor_name", Type: model.Text(), NotNull: true},
		{Name: "equipment_labels", Type: model.Text(), NotNull: true},
	},
}

var PublicSlotModel = model.Definition{
	Name: "public_slot",
	Fields: model.Fields{
		{Name: "room_code", Type: model.Text(), NotNull: true},
		{Name: "room_name", Type: model.Text(), NotNull: true},
		{Name: "floor_name", Type: model.Text(), NotNull: true},
		{Name: "start", Type: model.Text(), NotNull: true},
		{Name: "end", Type: model.Text(), NotNull: true},
	},
}
