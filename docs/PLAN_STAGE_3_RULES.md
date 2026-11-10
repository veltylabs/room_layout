← [Stage 2](PLAN_STAGE_2_MODULE_STORE.md) | Next → [Stage 4](PLAN_STAGE_4_METHODS.md)

# Stage 3 — generate, shared checks, rule files

## 3.1 Generate

1. Delete `model_orm.go` and `definition_crud.go` if they exist (not `ops.go`: Stage 4 deletes it). Run `definitionc` in the repo root. It writes `definition_gen.go` (the types, `<Entity>List`, the column constants `Floor_`, `Room_`, `RoomEquipment_`…, `ModelName`, `ModelName()`, `Definition()`, `MountOperations()`) and one skeleton per action declared `AddMyBusinessRule`. Then repeat the `go get` line of Stage 1 and `go mod tidy`.
2. Expected skeletons (if a name differs, keep the generated one): `rule_room_floor_{create,update,delete}.go`, `rule_room_space_{create,update}.go`, `rule_room_equipment_{create,update,delete}.go`, `rule_room_artifact_read.go`, `rule_room_space_equipment_{create,read}.go`, `rule_room_category_{create,read,delete}.go`, `rule_room_shift_read.go`, and the read models `rule_category_option_read.go`, `rule_occupant_option_read.go`, `rule_day_shift_read.go`, `rule_board_room_read.go`, `rule_free_room_read.go`, `rule_public_slot_read.go`. **Keep the signatures `definitionc` wrote**; replace only the bodies with the ones below.

## 3.2 `checks.go` (new) — what a rule and its kept method share, written exactly

```go
package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/storage"
)

// trimmed quita los espacios de un texto y rechaza que quede vacío: es el «Trim» de los antiguos Save* como regla.
// Ejemplo: name, err := trimmed("name", "  Box 1 ")   // "Box 1"
func trimmed(field, value string) (string, error) {
	v := fmt.TrimSpace(value)
	if v == "" {
		return "", &model.FieldError{Field: field, Err: ErrBlankName}
	}
	return v, nil
}

// checkCodeFree rechaza un código de sala que ya usa otra sala (exceptId es la propia sala al actualizar, "" al crear).
// Ejemplo: err := checkCodeFree(tx, "B-101", "")   // ErrCodeAlreadyExists si ya existe
func checkCodeFree(tx crud.Tx, code, exceptId string) error {
	rows, err := allRows(tx, func() model.Model { return &Room{} }, storage.Eq(Room_.Code, code))
	if err != nil {
		return err
	}
	for _, r := range rows {
		if r.(*Room).Id != exceptId {
			return ErrCodeAlreadyExists
		}
	}
	return nil
}

// checkFloorFree rechaza borrar un piso que todavía tiene salas.
// Ejemplo: err := checkFloorFree(tx, "f1")   // ErrFloorInUse si tiene salas
func checkFloorFree(tx crud.Tx, floorId string) error {
	_, found, err := firstRow(tx, func() model.Model { return &Room{} }, storage.Eq(Room_.FloorId, floorId))
	if err != nil {
		return err
	}
	if found {
		return ErrFloorInUse
	}
	return nil
}

// checkRoomExists rechaza un vínculo con una sala que no existe.
// Ejemplo: err := checkRoomExists(tx, "r1")   // ErrNotFound si no existe
func checkRoomExists(tx crud.Tx, roomId string) error {
	var r Room
	if err := tx.ReadOne(&r, roomId); err != nil {
		return notFound(err, ErrNotFound)
	}
	return nil
}

// categoryIn dice si la categoría está entre las opciones que entrega el lector de categorías de la app.
// Ejemplo: ok := categoryIn(options, "cat1")
func categoryIn(options []fmt.KeyValue, categoryId string) bool {
	for _, opt := range options {
		if opt.Key == categoryId {
			return true
		}
	}
	return false
}

// checkCategoryNotInUse rechaza quitar a una sala una categoría que un turno activo suyo todavía usa.
// Ejemplo: err := checkCategoryNotInUse(tx, "r1", "cat1")   // ErrCategoryInUse si hay un turno activo
func checkCategoryNotInUse(tx crud.Tx, roomId, categoryId string) error {
	_, found, err := firstRow(tx, func() model.Model { return &RoomShift{} },
		storage.Eq(RoomShift_.RoomId, roomId), storage.Eq(RoomShift_.CategoryId, categoryId), storage.Eq(RoomShift_.IsActive, true))
	if err != nil {
		return err
	}
	if found {
		return ErrCategoryInUse
	}
	return nil
}

// deleteEquipmentLinks borra todos los vínculos sala–equipamiento de un equipo (se llama al borrar el equipo).
// Ejemplo: err := deleteEquipmentLinks(tx, "e1")
func deleteEquipmentLinks(tx crud.Tx, equipmentId string) error {
	for {
		rows, err := allRows(tx, func() model.Model { return &RoomEquipment{} }, storage.Eq(RoomEquipment_.EquipmentId, equipmentId))
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		for _, r := range rows {
			link := r.(*RoomEquipment)
			if err := tx.Delete(&RoomEquipment{}, model.JoinKey(link.RoomId, link.EquipmentId)); err != nil {
				return err
			}
		}
	}
}
```

## 3.3 The rule files (keep each generated signature; write the Spanish comment on top; each file imports what it uses: `webtyp.com/crud`, `webtyp.com/model`, `webtyp.com/storage`)

**Floors** (`rule_room_floor_*.go`). `crud` already validated the schema; the rule trims and checks. `crud` writes `next` after an update with exactly `fields`; a derived column the caller did not send (`updated_at`) is written with `tx.UpdateFields`:

```go
// Create es la regla de negocio de crear room_floor: el nombre se recorta y no puede quedar vacío.
func (p *Floor) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	name, err := trimmed(Floor_.Name, p.Name)
	if err != nil {
		return 0, err
	}
	p.Name = name
	p.UpdatedAt = m.now()
	return crud.ApplyAndPublish, nil
}

// Update es la regla de negocio de actualizar room_floor: si cambia el nombre se recorta; updated_at siempre.
func (next *Floor) Update(a crud.CurrentUser, tx crud.Tx, old *Floor, fields []string, m *Module) (crud.RuleAnswer, error) {
	if containsString(fields, Floor_.Name) {
		name, err := trimmed(Floor_.Name, next.Name)
		if err != nil {
			return 0, err
		}
		next.Name = name
	}
	if err := tx.UpdateFields(&Floor{UpdatedAt: m.now()}, old.Id, []string{Floor_.UpdatedAt}); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}

// Delete es la regla de negocio de borrar room_floor: un piso con salas no se borra. El receptor es la fila guardada.
func (p *Floor) Delete(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := checkFloorFree(tx, p.Id); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
```

**Rooms** (`rule_room_space_*.go`): deactivate is an update of `["is_active"]`, which these rules let through.

```go
// Create es la regla de negocio de crear room_space: código y nombre recortados, y un código que ninguna otra sala usa.
func (p *Room) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	code, err := trimmed(Room_.Code, p.Code)
	if err != nil {
		return 0, err
	}
	name, err := trimmed(Room_.Name, p.Name)
	if err != nil {
		return 0, err
	}
	if err := checkCodeFree(tx, code, ""); err != nil {
		return 0, err
	}
	p.Code, p.Name = code, name
	p.UpdatedAt = m.now()
	return crud.ApplyAndPublish, nil
}

// Update es la regla de negocio de actualizar room_space: si cambian código o nombre se recortan y el código sigue libre.
func (next *Room) Update(a crud.CurrentUser, tx crud.Tx, old *Room, fields []string, m *Module) (crud.RuleAnswer, error) {
	if containsString(fields, Room_.Code) {
		code, err := trimmed(Room_.Code, next.Code)
		if err != nil {
			return 0, err
		}
		if err := checkCodeFree(tx, code, old.Id); err != nil {
			return 0, err
		}
		next.Code = code
	}
	if containsString(fields, Room_.Name) {
		name, err := trimmed(Room_.Name, next.Name)
		if err != nil {
			return 0, err
		}
		next.Name = name
	}
	if err := tx.UpdateFields(&Room{UpdatedAt: m.now()}, old.Id, []string{Room_.UpdatedAt}); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
```

**Equipment** (`rule_room_equipment_*.go`; note: the entity `Equipment` has `Name: "room_equipment"`, the link table `RoomEquipment` has `Name: "room_space_equipment"`):

```go
// Create es la regla de negocio de crear room_equipment: el nombre se recorta y no puede quedar vacío.
func (p *Equipment) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	name, err := trimmed(Equipment_.Name, p.Name)
	if err != nil {
		return 0, err
	}
	p.Name = name
	p.UpdatedAt = m.now()
	return crud.ApplyAndPublish, nil
}

// Update es la regla de negocio de actualizar room_equipment: si cambia el nombre se recorta; updated_at siempre.
func (next *Equipment) Update(a crud.CurrentUser, tx crud.Tx, old *Equipment, fields []string, m *Module) (crud.RuleAnswer, error) {
	if containsString(fields, Equipment_.Name) {
		name, err := trimmed(Equipment_.Name, next.Name)
		if err != nil {
			return 0, err
		}
		next.Name = name
	}
	if err := tx.UpdateFields(&Equipment{UpdatedAt: m.now()}, old.Id, []string{Equipment_.UpdatedAt}); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}

// Delete es la regla de negocio de borrar room_equipment: sus vínculos con salas se borran primero, en la misma transacción.
func (p *Equipment) Delete(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := deleteEquipmentLinks(tx, p.Id); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
```

**Links and the reads that carry `room_id` in the filter** (`rule_room_space_equipment_*.go`, `rule_room_category_*.go`, `rule_room_artifact_read.go`, `rule_room_shift_read.go`). The four `Search` rules have the same body, shown once for `RoomCategory`; the other three replace the type and the `…_` constants (`RoomEquipment_.RoomId`, `RoomArtifact_.RoomId`, `RoomShift_.RoomId`):

```go
// Search es la regla de negocio de leer room_category: con room_id en el filtro, solo los vínculos de esa sala.
func (p *RoomCategory) Search(a crud.CurrentUser, filter *RoomCategory, text string, m *Module) ([]storage.Condition, error) {
	if filter.RoomId == "" {
		return nil, nil
	}
	return []storage.Condition{storage.Eq(RoomCategory_.RoomId, filter.RoomId)}, nil
}

// Create es la regla de negocio de crear room_category: la sala existe y la categoría es una de las que entrega la app.
func (p *RoomCategory) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := checkRoomExists(tx, p.RoomId); err != nil {
		return 0, &model.FieldError{Field: RoomCategory_.RoomId, Err: err}
	}
	options, err := m.deps.Categories.CategoryOptions(m.tenantID)
	if err != nil {
		return 0, err
	}
	if !categoryIn(options, p.CategoryId) {
		return 0, &model.FieldError{Field: RoomCategory_.CategoryId, Err: ErrUnknownCategory}
	}
	return crud.ApplyAndPublish, nil
}

// Delete es la regla de negocio de borrar room_category: no se quita una categoría que un turno activo de la sala usa.
func (p *RoomCategory) Delete(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := checkCategoryNotInUse(tx, p.RoomId, p.CategoryId); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
```
`RoomEquipment` has only `Create` (`checkRoomExists`, same wrapping) and `Search`; `RoomArtifact` and `RoomShift` only `Search`.

**Read models** (`rule_*_read.go`). The filter is `filter` (D52). Each calls the KEPT method, which already does the work; the list type is the generated `<Model>List` (`[]*<Model>`; if it is not, build it as `service_catalog/rule_currency_option_read.go` does):

```go
// Read es la regla de negocio de leer category_option: las categorías que entrega el lector de categorías de la app.
func (p *CategoryOption) Read(a crud.CurrentUser, filter *CategoryOption, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	opts, err := m.deps.Categories.CategoryOptions(m.tenantID)
	if err != nil {
		return nil, err
	}
	list := CategoryOptionList{}
	for _, kv := range opts {
		list = append(list, &CategoryOption{Id: kv.Key, Label: kv.Value})
	}
	return &list, nil
}
```
`OccupantOption`: the same with `m.deps.Occupants.OccupantOptions(m.tenantID)`. `DayShift`:

```go
// Read es la regla de negocio de leer day_shift: los turnos de una sala (o de todas) en un día; sin date, hoy.
func (p *DayShift) Read(a crud.CurrentUser, filter *DayShift, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	date, err := parseDate(filter.Date)
	if err != nil {
		return nil, &model.FieldError{Field: DayShift_.Date, Err: err}
	}
	shifts, err := m.ListRoomDay(m.tenantID, filter.RoomId, date)
	if err != nil {
		return nil, err
	}
	list := DayShiftList{}
	for i := range shifts {
		list = append(list, &shifts[i])
	}
	return &list, nil
}
```
`BoardRoom`: `board, err := m.ListBoard(m.tenantID, filter.CategoryId, filter.EquipmentId)`, list of `&board[i]`. `FreeRoom`:

```go
// Read es la regla de negocio de leer free_room: las salas libres en un día y un tramo de horas (date, start y end son requeridos).
func (p *FreeRoom) Read(a crud.CurrentUser, filter *FreeRoom, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	if filter.Date == "" {
		return nil, &model.FieldError{Field: FreeRoom_.Date, Err: ErrInvalidDate}
	}
	date, errDate := parseDate(filter.Date)
	startMin, errStart := parseHour(filter.Start)
	endMin, errEnd := parseHour(filter.End)
	if errDate != nil || errStart != nil || errEnd != nil {
		return nil, ErrInvalidRange
	}
	free, err := m.FindFreeRooms(m.tenantID, date, startMin, endMin, filter.CategoryId)
	if err != nil {
		return nil, err
	}
	list := FreeRoomList{}
	for i := range free {
		list = append(list, &free[i])
	}
	return &list, nil
}
```
`PublicSlot`: `if filter.Date == ""` → field error `PublicSlot_.Date` / `ErrInvalidDate`; `date, err := parseDate(filter.Date)` (`err` → `ErrInvalidDate` field error); `slots, err := m.ListPublicAvailability(date, filter.CategoryId)`; list of `&slots[i]`. The app declares `public_slot` public with `model.Read` (README note, Stage 6).

The package does not compile yet (Stage 4). Commit and push; **open the pull request now** with what exists and keep committing to it.
