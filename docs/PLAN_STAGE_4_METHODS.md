← [Stage 3](PLAN_STAGE_3_RULES.md) | Next → [Stage 5](PLAN_STAGE_5_BROWSER.md)

# Stage 4 — the kept methods on the store; `ops.go` deleted

Other modules and apps call these methods (`seed/`, `mjosefa-cms`): **every exported method keeps its name, parameters and result types, in the file where it is today.** Only the body changes (table of Stage 2.4). A method gets its store from `m.storeFor(tenantID)` (an empty tenant answers `ErrTenantRequired`, another tenant sees nothing). Each is written below; keep the doc comment of today's method above it (add an `Ejemplo:` line if it has none). **`shift.go`, `day.go`, `convert.go`, `cells.go`, `weekday.go`, `interfaces.go` are NOT touched.**

## 4.0 Order helpers — `order.go` (new), written exactly

The old queries ordered in SQL; the store reads by id, so the three lists that had an order sort in Go (no stdlib `sort`, no generics: these lists are a screen's worth of rows).

```go
package roomlayout

// sortFloors ordena los pisos por posición y, a igual posición, por nombre.
func sortFloors(list []Floor) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && floorBefore(list[j], list[j-1]); j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}

func floorBefore(a, b Floor) bool {
	if a.Position != b.Position {
		return a.Position < b.Position
	}
	return a.Name < b.Name
}

// sortRoomsByCode ordena las salas por código.
func sortRoomsByCode(list []Room) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j].Code < list[j-1].Code; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}

// sortEquipmentByName ordena el equipamiento por nombre.
func sortEquipmentByName(list []Equipment) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j].Name < list[j-1].Name; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}
```

## 4.1 `floor.go` — `SaveFloor`, `ListFloors`, `DeleteFloor` (delete `listFloors`)

```go
// floorColumns son las columnas que cambia SaveFloor al actualizar (todas menos id y tenant_id).
var floorColumns = []string{Floor_.Cols, Floor_.Rows, Floor_.GridLocked, Floor_.HabitableCells, Floor_.Name, Floor_.Position, Floor_.UpdatedAt}

func (m *Module) SaveFloor(f Floor) (Floor, error) {
	store, err := m.storeFor(f.TenantId)
	if err != nil {
		return Floor{}, err
	}
	f.Name = fmt.TrimSpace(f.Name)
	isNew := f.Id == ""
	if isNew {
		f.Id = m.deps.IDs.NewID()
	}
	if err := model.ValidateFields(byte(model.Create), &f); err != nil {
		return Floor{}, err
	}
	f.UpdatedAt = m.now()
	err = store.Atomic(func(tx crud.Tx) error {
		if isNew {
			return tx.Create(&f)
		}
		return tx.UpdateFields(&f, f.Id, floorColumns)
	})
	if err != nil {
		return Floor{}, err
	}
	m.publish(TopicFloorSaved, &f)
	return f, nil
}

func (m *Module) ListFloors(tenantID string) ([]Floor, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	rows, err := allRows(store, func() model.Model { return &Floor{} })
	if err != nil {
		return nil, err
	}
	var res []Floor
	for _, r := range rows {
		res = append(res, *r.(*Floor))
	}
	sortFloors(res)
	return res, nil
}

func (m *Module) DeleteFloor(tenantID, id string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if id == "" {
		return ErrNotFound
	}
	err = store.Atomic(func(tx crud.Tx) error {
		if err := checkFloorFree(tx, id); err != nil {
			return err
		}
		return tx.Delete(&Floor{}, id)
	})
	if err != nil {
		return err
	}
	m.publish(TopicFloorDeleted, &Floor{Id: id, TenantId: tenantID})
	return nil
}
```
Imports of `floor.go`: `webtyp.com/crud`, `webtyp.com/fmt`, `webtyp.com/model` (drop `orm`).

## 4.2 `room.go` — `SaveRoom`, `GetRoom`, `ListRooms`, `DeactivateRoom` (delete `listRooms`)

```go
// roomColumns son las columnas que cambia SaveRoom al actualizar (todas menos id y tenant_id).
var roomColumns = []string{Room_.RoomType, Room_.Cells, Room_.FloorId, Room_.Code, Room_.Name, Room_.Notes, Room_.IsActive, Room_.UpdatedAt}

func (m *Module) SaveRoom(r Room) (Room, error) {
	store, err := m.storeFor(r.TenantId)
	if err != nil {
		return Room{}, err
	}
	r.Code = fmt.TrimSpace(r.Code)
	r.Name = fmt.TrimSpace(r.Name)
	isNew := r.Id == ""
	if isNew {
		r.Id = m.deps.IDs.NewID()
	}
	if err := model.ValidateFields(byte(model.Create), &r); err != nil {
		return Room{}, err
	}
	r.UpdatedAt = m.now()
	err = store.Atomic(func(tx crud.Tx) error {
		if err := checkCodeFree(tx, r.Code, r.Id); err != nil {
			return err
		}
		if isNew {
			return tx.Create(&r)
		}
		return tx.UpdateFields(&r, r.Id, roomColumns)
	})
	if err != nil {
		return Room{}, err
	}
	m.publish(TopicRoomSaved, &r)
	return r, nil
}

func (m *Module) GetRoom(tenantID, id string) (Room, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return Room{}, err
	}
	if id == "" {
		return Room{}, ErrNotFound
	}
	var r Room
	if err := store.ReadOne(&r, id); err != nil {
		return Room{}, notFound(err, ErrNotFound)
	}
	return r, nil
}

func (m *Module) ListRooms(tenantID string, activeOnly bool) ([]Room, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	var conds []storage.Condition
	if activeOnly {
		conds = append(conds, storage.Eq(Room_.IsActive, true))
	}
	rows, err := allRows(store, func() model.Model { return &Room{} }, conds...)
	if err != nil {
		return nil, err
	}
	var res []Room
	for _, r := range rows {
		res = append(res, *r.(*Room))
	}
	sortRoomsByCode(res)
	return res, nil
}

func (m *Module) DeactivateRoom(tenantID, id string) error {
	r, err := m.GetRoom(tenantID, id)
	if err != nil {
		return err
	}
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	r.IsActive = false
	r.UpdatedAt = m.now()
	if err := store.UpdateFields(&r, id, []string{Room_.IsActive, Room_.UpdatedAt}); err != nil {
		return err
	}
	m.publish(TopicRoomDeactivated, &r)
	return nil
}
```
Imports of `room.go`: `webtyp.com/crud`, `webtyp.com/fmt`, `webtyp.com/model`, `webtyp.com/storage`.

## 4.3 `equipment.go` — equipment, links, artifacts (delete `listEquipment`, `listRoomArtifacts`)

```go
var equipmentColumns = []string{Equipment_.Name, Equipment_.DeviceId, Equipment_.UpdatedAt}

var artifactColumns = []string{RoomArtifact_.RoomId, RoomArtifact_.Kind, RoomArtifact_.Code, RoomArtifact_.Cell, RoomArtifact_.Status,
	RoomArtifact_.Reason, RoomArtifact_.NextMaintenance, RoomArtifact_.Items}

func (m *Module) SaveEquipment(e Equipment) (Equipment, error) {
	store, err := m.storeFor(e.TenantId)
	if err != nil {
		return Equipment{}, err
	}
	e.Name = fmt.TrimSpace(e.Name)
	isNew := e.Id == ""
	if isNew {
		e.Id = m.deps.IDs.NewID()
	}
	if err := model.ValidateFields(byte(model.Create), &e); err != nil {
		return Equipment{}, err
	}
	e.UpdatedAt = m.now()
	err = store.Atomic(func(tx crud.Tx) error {
		if isNew {
			return tx.Create(&e)
		}
		return tx.UpdateFields(&e, e.Id, equipmentColumns)
	})
	if err != nil {
		return Equipment{}, err
	}
	m.publish(TopicEquipmentSaved, &e)
	return e, nil
}

// equipmentOn responde todo el equipamiento, por nombre.
func equipmentOn(tx crud.Tx) ([]Equipment, error) {
	rows, err := allRows(tx, func() model.Model { return &Equipment{} })
	if err != nil {
		return nil, err
	}
	var res []Equipment
	for _, r := range rows {
		res = append(res, *r.(*Equipment))
	}
	sortEquipmentByName(res)
	return res, nil
}

func (m *Module) ListEquipment(tenantID string) ([]Equipment, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	return equipmentOn(store)
}

func (m *Module) DeleteEquipment(tenantID, id string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if id == "" {
		return ErrNotFound
	}
	err = store.Atomic(func(tx crud.Tx) error {
		if err := deleteEquipmentLinks(tx, id); err != nil {
			return err
		}
		return tx.Delete(&Equipment{}, id)
	})
	if err != nil {
		return err
	}
	m.publish(TopicEquipmentDeleted, &Equipment{Id: id, TenantId: tenantID})
	return nil
}

func (m *Module) ListRoomEquipment(tenantID, roomID string) ([]Equipment, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	links, err := allRows(store, func() model.Model { return &RoomEquipment{} }, storage.Eq(RoomEquipment_.RoomId, roomID))
	if err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []Equipment{}, nil
	}
	all, err := equipmentOn(store)
	if err != nil {
		return nil, err
	}
	var res []Equipment
	for _, eq := range all {
		for _, l := range links {
			if l.(*RoomEquipment).EquipmentId == eq.Id {
				res = append(res, eq)
				break
			}
		}
	}
	return res, nil
}

func (m *Module) SetRoomEquipment(tenantID, roomID string, equipmentIDs []string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if _, err := m.GetRoom(tenantID, roomID); err != nil {
		return err
	}
	return store.Atomic(func(tx crud.Tx) error {
		old, err := allRows(tx, func() model.Model { return &RoomEquipment{} }, storage.Eq(RoomEquipment_.RoomId, roomID))
		if err != nil {
			return err
		}
		for _, r := range old {
			if err := tx.Delete(&RoomEquipment{}, model.JoinKey(roomID, r.(*RoomEquipment).EquipmentId)); err != nil {
				return err
			}
		}
		for _, eqID := range equipmentIDs {
			if err := tx.Create(&RoomEquipment{RoomId: roomID, EquipmentId: eqID}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *Module) SaveRoomArtifact(a RoomArtifact) (RoomArtifact, error) {
	store, err := m.storeFor(a.TenantId)
	if err != nil {
		return RoomArtifact{}, err
	}
	isNew := a.Id == ""
	if isNew {
		a.Id = m.deps.IDs.NewID()
	}
	if err := model.ValidateFields(byte(model.Create), &a); err != nil {
		return RoomArtifact{}, err
	}
	err = store.Atomic(func(tx crud.Tx) error {
		if isNew {
			return tx.Create(&a)
		}
		return tx.UpdateFields(&a, a.Id, artifactColumns)
	})
	if err != nil {
		return RoomArtifact{}, err
	}
	return a, nil
}

func (m *Module) ListRoomArtifacts(tenantID string, roomID string) ([]RoomArtifact, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	var conds []storage.Condition
	if roomID != "" {
		conds = append(conds, storage.Eq(RoomArtifact_.RoomId, roomID))
	}
	rows, err := allRows(store, func() model.Model { return &RoomArtifact{} }, conds...)
	if err != nil {
		return nil, err
	}
	var res []RoomArtifact
	for _, r := range rows {
		res = append(res, *r.(*RoomArtifact))
	}
	return res, nil
}

func (m *Module) DeleteRoomArtifact(tenantID, id string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if id == "" {
		return ErrNotFound
	}
	return store.Delete(&RoomArtifact{}, id)
}
```
Imports of `equipment.go`: `webtyp.com/crud`, `webtyp.com/fmt`, `webtyp.com/model`, `webtyp.com/storage`.

## 4.4 `category.go` — `ListRoomCategories`, `SetRoomCategories`

```go
func (m *Module) ListRoomCategories(tenantID, roomID string) ([]fmt.KeyValue, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	links, err := allRows(store, func() model.Model { return &RoomCategory{} }, storage.Eq(RoomCategory_.RoomId, roomID))
	if err != nil {
		return nil, err
	}
	allOptions, err := m.deps.Categories.CategoryOptions(tenantID)
	if err != nil {
		return nil, err
	}
	var res []fmt.KeyValue
	for _, l := range links {
		id := l.(*RoomCategory).CategoryId
		label := id
		for _, opt := range allOptions {
			if opt.Key == id {
				label = opt.Value
				break
			}
		}
		res = append(res, fmt.KeyValue{Key: id, Value: label})
	}
	return res, nil
}

func (m *Module) SetRoomCategories(tenantID, roomID string, categoryIDs []string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if _, err := m.GetRoom(tenantID, roomID); err != nil {
		return err
	}
	allOptions, err := m.deps.Categories.CategoryOptions(tenantID)
	if err != nil {
		return err
	}
	for _, catID := range categoryIDs {
		if !categoryIn(allOptions, catID) {
			return ErrUnknownCategory
		}
	}
	return store.Atomic(func(tx crud.Tx) error {
		// an active shift of the room cannot lose the category it uses
		shifts, err := allRows(tx, func() model.Model { return &RoomShift{} },
			storage.Eq(RoomShift_.RoomId, roomID), storage.Eq(RoomShift_.IsActive, true))
		if err != nil {
			return err
		}
		for _, r := range shifts {
			if !containsString(categoryIDs, r.(*RoomShift).CategoryId) {
				return ErrCategoryInUse
			}
		}
		old, err := allRows(tx, func() model.Model { return &RoomCategory{} }, storage.Eq(RoomCategory_.RoomId, roomID))
		if err != nil {
			return err
		}
		for _, r := range old {
			if err := tx.Delete(&RoomCategory{}, model.JoinKey(roomID, r.(*RoomCategory).CategoryId)); err != nil {
				return err
			}
		}
		for _, catID := range categoryIDs {
			if err := tx.Create(&RoomCategory{RoomId: roomID, CategoryId: catID}); err != nil {
				return err
			}
		}
		return nil
	})
}
```
Imports of `category.go`: `webtyp.com/crud`, `webtyp.com/fmt`, `webtyp.com/model`, `webtyp.com/storage`.

## 4.5 Deletions (each with its criterion)

| Delete | Criterion |
|---|---|
| `ops.go` (the `Op*` constants, `MountOperations`, every handler, `ModelName()`) | `test ! -e ops.go`; `grep -rn "OpListFloors\|OpSaveRoomShift\|MountOperations(reg" --include=*.go . \| grep -v definition_gen` → empty (the UI and `view.go` are Stage 5) |
| `listFloors`, `listRooms`, `listEquipment`, `listRoomArtifacts` | `grep -rn "listFloors\|listRooms\|listEquipment\|listRoomArtifacts" --include=*.go .` → empty |

`definitionc` wrote the new `MountOperations()`; never write one by hand. `gotest` may still fail on `view.go`, `ui/` and old tests (Stages 5 and 6). Commit and push.
