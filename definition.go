package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"
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

// Errores de dominio: cada uno lleva su tipo (model.NotFound, model.Conflict o
// model.Invalid), y crud responde con el estado que corresponde.
//
// Ejemplo: if model.KindOf(err) == model.ErrorConflict { ... }
const (
	ErrNotFound           model.NotFound = "not found"
	ErrTenantRequired     model.Invalid  = "tenant id is required"
	ErrCodeAlreadyExists  model.Conflict = "code already exists for a room in this tenant"
	ErrFloorInUse         model.Conflict = "floor still has rooms"
	ErrInvalidRange       model.Invalid  = "start must be before end, both within 00:00 and 24:00"
	ErrInvalidDate        model.Invalid  = "invalid date, expected YYYY-MM-DD"
	ErrInvalidWeekday     model.Invalid  = "day_of_week must be 0..6"
	ErrUnknownCategory    model.Invalid  = "unknown category"
	ErrUnknownOccupant    model.Invalid  = "unknown occupant"
	ErrOccupantRequired   model.Invalid  = "an occupant is required (id or label)"
	ErrCategoryNotAllowed model.Invalid  = "room is not enabled for this category"
	ErrCategoryInUse      model.Conflict = "category still has active shifts in this space"
	ErrRoomOverlap        model.Conflict = "room already has a shift in that time range"
	ErrOccupantOverlap    model.Conflict = "occupant has a conflicting shift at the requested time"
	ErrOutsideBounds      model.Invalid  = "shift is outside the establishment's opening hours for that date"
	ErrInvalidArea        model.Invalid  = "invalid cells, expected column,row,width,height separated by ;"
	ErrNotWeekly          model.Invalid  = "only a weekly shift can have one occurrence cancelled"
	ErrBlankName          model.Invalid  = "the name cannot be blank"
	ErrFieldNotEditable   model.Invalid  = "this field cannot be changed"
)

func weekdayInput() input.Input {
	return input.Select(
		fmt.KeyValue{Key: WeekdaySunday, Value: "Sunday"},
		fmt.KeyValue{Key: WeekdayMonday, Value: "Monday"},
		fmt.KeyValue{Key: WeekdayTuesday, Value: "Tuesday"},
		fmt.KeyValue{Key: WeekdayWednesday, Value: "Wednesday"},
		fmt.KeyValue{Key: WeekdayThursday, Value: "Thursday"},
		fmt.KeyValue{Key: WeekdayFriday, Value: "Friday"},
		fmt.KeyValue{Key: WeekdaySaturday, Value: "Saturday"},
	)
}

var FloorModel = model.Definition{
	Name:  "room_floor",
	Allow: model.Create | model.Read | model.Update | model.Delete,
	Rules: model.Rules{
		Create: model.AddMyBusinessRule,
		Read:   model.HaveNoBusinessRule,
		Update: model.AddMyBusinessRule,
		Delete: model.AddMyBusinessRule,
	},
	Fields: model.Fields{
		{Name: "cols", Type: model.Int()},
		{Name: "rows", Type: model.Int()},
		{Name: "grid_locked", Type: model.Bool()},
		{Name: "habitable_cells", Type: model.Text()},
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true, ServerOnly: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "name", Type: input.Text(), Item: model.ItemLabel, NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 60}},
		{Name: "position", Type: input.Number(), OmitEmpty: true},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
	},
}

var RoomModel = model.Definition{
	Name:  "room_space",
	Allow: model.Create | model.Read | model.Update | model.Delete,
	Rules: model.Rules{
		Create: model.AddMyBusinessRule,
		Read:   model.HaveNoBusinessRule,
		Update: model.AddMyBusinessRule,
		Delete: model.HaveNoBusinessRule,
	},
	Fields: model.Fields{
		{Name: "room_type", Type: model.Text()},
		{Name: "cells", Type: model.Text()},
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true, ServerOnly: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "floor_id", Type: input.Select(), Ref: &FloorModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-', '_'}}},
		{Name: "code", Type: input.Text(), Item: model.ItemLabel, NotNull: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-'}, Minimum: 1, Maximum: 20}},
		{Name: "name", Type: input.Text(), Item: model.ItemDescription, NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 120}},
		{Name: "notes", Type: input.Textarea(), OmitEmpty: true},
		{Name: "is_active", Type: input.Checkbox()},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
	},
}

var RoomArtifactModel = model.Definition{
	Name:  "room_artifact",
	Allow: model.Create | model.Read | model.Update | model.Delete,
	Rules: model.Rules{
		Create: model.HaveNoBusinessRule,
		Read:   model.AddMyBusinessRule,
		Update: model.HaveNoBusinessRule,
		Delete: model.HaveNoBusinessRule,
	},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true, ServerOnly: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "room_id", Type: model.Text(), NotNull: true},
		{Name: "kind", Type: model.Text(), NotNull: true, Item: model.ItemDescription},
		{Name: "code", Type: model.Text(), NotNull: true, Item: model.ItemLabel},
		{Name: "cell", Type: model.Text(), NotNull: true},
		{Name: "status", Type: model.Text(), NotNull: true},
		{Name: "reason", Type: model.Text()},
		{Name: "next_maintenance", Type: model.Text()},
		{Name: "items", Type: model.Text()},
	},
}

var EquipmentModel = model.Definition{
	Name:  "room_equipment",
	Allow: model.Create | model.Read | model.Update | model.Delete,
	Rules: model.Rules{
		Create: model.AddMyBusinessRule,
		Read:   model.HaveNoBusinessRule,
		Update: model.AddMyBusinessRule,
		Delete: model.AddMyBusinessRule,
	},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true, ServerOnly: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "name", Type: input.Text(), Item: model.ItemLabel, NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 80}},
		// Id suelto del equipo de red (device_manager) al que corresponde este equipamiento; vacío si
		// no tiene red (un gastroscopio). Este módulo no importa device_manager: la app le entrega los
		// equipos por una interfaz, y el id se asigna en el mapa, no en el formulario de la lista.
		{Name: "device_id", Type: model.Text(), OmitEmpty: true},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
	},
}

var RoomEquipmentModel = model.Definition{
	Name:  "room_space_equipment",
	Allow: model.Create | model.Read | model.Delete,
	Rules: model.Rules{
		Create: model.AddMyBusinessRule,
		Read:   model.AddMyBusinessRule,
		Delete: model.HaveNoBusinessRule,
	},
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "room_id", Type: model.Text(), Ref: &RoomModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
		{Name: "equipment_id", Type: model.Text(), Ref: &EquipmentModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
	},
}

var RoomCategoryModel = model.Definition{
	Name:  "room_category",
	Allow: model.Create | model.Read | model.Delete,
	Rules: model.Rules{
		Create: model.AddMyBusinessRule,
		Read:   model.AddMyBusinessRule,
		Delete: model.AddMyBusinessRule,
	},
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "room_id", Type: model.Text(), Ref: &RoomModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
		{Name: "category_id", Type: model.Text(), DB: &model.FieldDB{PK: true}, NotNull: true},
	},
}

var RoomShiftModel = model.Definition{
	Name:  "room_shift",
	Allow: model.Create | model.Read | model.Update | model.Delete,
	Rules: model.Rules{
		Create: model.AddMyBusinessRule,
		Read:   model.AddMyBusinessRule,
		Update: model.AddMyBusinessRule,
		Delete: model.AddMyBusinessRule,
	},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true, ServerOnly: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "room_id", Type: model.Text(), Ref: &RoomModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "category_id", Type: model.Text(), NotNull: true},
		{Name: "occupant_id", Type: model.Text(), OmitEmpty: true},
		{Name: "occupant_label", Type: model.Text(), OmitEmpty: true}, // lo completa la regla con el nombre del ocupante; sin ocupante lo envía quien llama
		{Name: "day_of_week", Type: model.Int(), OmitEmpty: true},     // turno semanal: 0 = domingo … 6 = sábado; en uno con fecha lo calcula la regla
		{Name: "specific_date", Type: model.UnixTime(), OmitEmpty: true},
		{Name: "start_min", Type: model.Int(), OmitEmpty: true}, // minutos desde la medianoche; la regla exige 0 ≤ inicio < fin ≤ 1440
		{Name: "end_min", Type: model.Int(), NotNull: true},
		{Name: "is_active", Type: model.Bool(), OmitEmpty: true, ServerOnly: true}, // la regla lo deja en true
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
	},
}

var RoomShiftCancellationModel = model.Definition{
	Name:  "room_shift_cancellation",
	Allow: model.Create | model.Read,
	Rules: model.Rules{
		Create: model.AddMyBusinessRule,
		Read:   model.HaveNoBusinessRule,
	},
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "shift_id", Type: model.Text(), Ref: &RoomShiftModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
		{Name: "specific_date", Type: model.UnixTime(), DB: &model.FieldDB{PK: true}, NotNull: true},
		{Name: "moved_to_room_id", Type: model.Text(), Label: "Target room", OmitEmpty: true, Ref: &RoomModel, DB: &model.FieldDB{RefColumn: "id"}}, // si la ocurrencia se movió: la sala destino
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

// CategoryOptionModel es una opción de categoría de atención (lo que el lector de categorías de la app responde).
var CategoryOptionModel = model.Definition{
	Name: "category_option", Wire: true, Allow: model.Read, Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), OmitEmpty: true},
		{Name: "label", Type: model.Text(), OmitEmpty: true},
	},
}

// OccupantOptionModel es una opción de ocupante (profesional o persona) que puede tener un turno.
var OccupantOptionModel = model.Definition{
	Name: "occupant_option", Wire: true, Allow: model.Read, Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), OmitEmpty: true},
		{Name: "label", Type: model.Text(), OmitEmpty: true},
	},
}

// DayShiftModel es un turno de un día. Filtro: room_id (opcional) y date ("AAAA-MM-DD", opcional: hoy).
var DayShiftModel = model.Definition{
	Name: "day_shift", Wire: true, Allow: model.Read, Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "date", Type: model.Text(), OmitEmpty: true}, // filtro
		{Name: "shift_id", Type: model.Text(), OmitEmpty: true},
		{Name: "room_id", Type: model.Text(), OmitEmpty: true}, // filtro y dato
		{Name: "room_code", Type: model.Text(), OmitEmpty: true},
		{Name: "category_id", Type: model.Text(), OmitEmpty: true},
		{Name: "category_label", Type: model.Text(), OmitEmpty: true},
		{Name: "occupant_id", Type: model.Text(), OmitEmpty: true},
		{Name: "occupant_label", Type: model.Text(), OmitEmpty: true},
		{Name: "start", Type: model.Text(), OmitEmpty: true},
		{Name: "end", Type: model.Text(), OmitEmpty: true},
		{Name: "is_weekly", Type: model.Bool(), OmitEmpty: true},
	},
}

// BoardRoomModel es una fila del tablero de salas. Filtro: category_id y equipment_id (opcionales).
var BoardRoomModel = model.Definition{
	Name: "board_room", Wire: true, Allow: model.Read, Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "category_id", Type: model.Text(), OmitEmpty: true},  // filtro
		{Name: "equipment_id", Type: model.Text(), OmitEmpty: true}, // filtro
		{Name: "room_id", Type: model.Text(), OmitEmpty: true},
		{Name: "code", Type: model.Text(), OmitEmpty: true},
		{Name: "name", Type: model.Text(), OmitEmpty: true},
		{Name: "floor_name", Type: model.Text(), OmitEmpty: true},
		{Name: "floor_position", Type: model.Int(), OmitEmpty: true},
		{Name: "category_labels", Type: model.Text(), OmitEmpty: true},
		{Name: "equipment_labels", Type: model.Text(), OmitEmpty: true},
		{Name: "status", Type: model.Text(), OmitEmpty: true},
		{Name: "current_occupant_label", Type: model.Text(), Label: "Current occupant", OmitEmpty: true},
		{Name: "current_until", Type: model.Text(), OmitEmpty: true},
		{Name: "next_start", Type: model.Text(), OmitEmpty: true},
		{Name: "next_occupant_label", Type: model.Text(), Label: "Next occupant", OmitEmpty: true},
		{Name: "today", Type: model.Text(), OmitEmpty: true},
	},
}

// FreeRoomModel es una sala libre. Filtro: date, start y end ("AAAA-MM-DD", "HH:MM", "HH:MM": los tres requeridos por la regla) y category_id (opcional).
var FreeRoomModel = model.Definition{
	Name: "free_room", Wire: true, Allow: model.Read, Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "date", Type: model.Text(), OmitEmpty: true},        // filtro
		{Name: "start", Type: model.Text(), OmitEmpty: true},       // filtro
		{Name: "end", Type: model.Text(), OmitEmpty: true},         // filtro
		{Name: "category_id", Type: model.Text(), OmitEmpty: true}, // filtro
		{Name: "room_id", Type: model.Text(), OmitEmpty: true},
		{Name: "code", Type: model.Text(), OmitEmpty: true},
		{Name: "name", Type: model.Text(), OmitEmpty: true},
		{Name: "floor_name", Type: model.Text(), OmitEmpty: true},
		{Name: "equipment_labels", Type: model.Text(), OmitEmpty: true},
	},
}

// PublicSlotModel es un hueco libre que el público puede ver (la app lo declara público). Filtro: date (requerida por la regla) y category_id (opcional).
var PublicSlotModel = model.Definition{
	Name: "public_slot", Wire: true, Allow: model.Read, Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "date", Type: model.Text(), OmitEmpty: true},        // filtro
		{Name: "category_id", Type: model.Text(), OmitEmpty: true}, // filtro
		{Name: "room_code", Type: model.Text(), OmitEmpty: true},
		{Name: "room_name", Type: model.Text(), OmitEmpty: true},
		{Name: "floor_name", Type: model.Text(), OmitEmpty: true},
		{Name: "start", Type: model.Text(), OmitEmpty: true},
		{Name: "end", Type: model.Text(), OmitEmpty: true},
	},
}

// Definition declara el módulo room_layout en el marco crud-by-declaration.
var Definition = model.ModuleDefinition{
	Name: "room_layout",
	Entities: []model.Definition{
		FloorModel, RoomModel, RoomArtifactModel, EquipmentModel, RoomEquipmentModel, RoomCategoryModel,
		RoomShiftModel, RoomShiftCancellationModel,
		CategoryOptionModel, OccupantOptionModel, DayShiftModel, BoardRoomModel, FreeRoomModel, PublicSlotModel,
	},
}
