← [Index](PLAN.md) | Next → [Stage 2](PLAN_STAGE_2_MODULE_STORE.md)

# Stage 1 — tools and `definition.go`

## 1.1 Tools (run first, exactly)

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go install webtyp.com/definitionc/cmd/definitionc@v0.1.42
export PATH="$PATH:$(go env GOPATH)/bin"
go get webtyp.com/model@v0.2.19 webtyp.com/crud@v0.0.12 webtyp.com/orm@v0.12.33 webtyp.com/storage@v0.2.1 webtyp.com/ddl@v0.0.49 webtyp.com/router@v0.4.22 webtyp.com/definitionc@v0.1.42 webtyp.com/view@v0.6.74 webtyp.com/appkit@v0.2.35
go mod tidy
```

`GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`; remove every `replace` from `go.mod`. **`definitionc` rewrites `go.mod` to older versions every time it runs: after EVERY run repeat the `go get` line and `go mod tidy`.** If a module cannot be fetched, stop and say so in the PR; never invent a hash.

Facts you need (asked before; do not ask again):

- Arguments: `crud.CreateArgs{Records []model.Model}`, `crud.UpdateArgs{IDs []string, Fields []string, Record model.Model}`, `crud.DeleteArgs{IDs []string}`, `crud.ReadArgs{ID, Search, After string, Limit int64, Filter model.Model}`; never a bare entity pointer, a `*List` or a map. Kinds are flat (`model.Pointer`/`List`/`Map` do not exist).
- Inside a rule `tx` is `crud.Tx`: `ReadOne(&rec, id)`, `ReadAll(newRow, crud.Query{Conditions, Order, After, Limit}, onRow)`, `Create`, `UpdateFields(&rec, id, cols)`, `Delete(&rec, id)`. Conditions: `storage.Eq/Neq/Gt/Gte/Lt/Lte/In/Like` (AND only); the tenant is applied by the store: **never add a `tenant_id` condition**; at most `crud.MaxPage` rows per read.
- **Update rules:** `crud` writes `next` after the rule with exactly the columns in `Fields`. A column the rule derives and the caller did not send is written by the rule with `tx.UpdateFields(&row, id, cols)`; a sent column is rewritten by `crud` from `next`, so the rule assigns `next.X = computed` for it.
- **The `id` of a table:** `ValidateInput` always requires a Text primary key unless the field is `ServerOnly`. A table whose rows the server creates: `id` is `OmitEmpty: true, ServerOnly: true` and Create records OMIT it; never send a `ServerOnly` field.
- `NotNull` on an `Int` or `Bool` rejects the zero value: `is_active` (false) must not be `NotNull`.

- **`ModelName` is written by `definitionc`** (`const ModelName` and `func (m *Module) ModelName() string`): delete the hand-written ones (`definition.go` line 10 and `ops.go`), or the build says `ModelName redeclared`.
- **A table's `updated_at` that `crud` stamps must be `model.NanoTime()`**: `crud` writes the current instant in NANOSECONDS into `updated_at`; a `model.UnixTime()` column would hold nanoseconds read as seconds. Its Go type is `time.Nano`, so every `x.UpdatedAt = m.now().Unix()` of a table this plan migrates becomes `x.UpdatedAt = m.now()`. (`RoomShift.updated_at` stays `UnixTime` in this plan: `crud` does not write `RoomShift` yet.)
- **Composite keys** (`room_id` + `equipment_id`, `room_id` + `category_id`, `shift_id` + `specific_date`) are client-sent: the wire id of a row is `model.JoinKey(a, b)`; a table with a composite key is read in ONE page (at most `crud.MaxPage` rows; a room never has that many links).
- A table whose `Read` is declared `AddMyBusinessRule` has a `Search` rule: `func (p *X) Search(a crud.CurrentUser, filter *X, text string, m *Module) ([]storage.Condition, error)`; `filter` carries the fields the caller sent (D52: the entity is its own filter); `text` may be `""`.

## 1.2 `definition.go` — exact edits (keep every other constant, `weekdayInput`, `RoomShiftFormModel`)

1. **Delete** `const ModelName`, `const ErrCannotRead`, and these models (they only carried removed operations): `IdRefModel`, `ListFloorsArgsModel`, `ListRoomsArgsModel`, `GetRoomArgsModel`, `DeactivateRoomArgsModel`, `ListEquipmentArgsModel`, `ListRoomEquipmentArgsModel`, `SetRoomEquipmentArgsModel`, `ListRoomCategoriesArgsModel`, `SetRoomCategoriesArgsModel`, `ListOptionsArgsModel`, `ListRoomShiftsArgsModel`, `DeleteRoomShiftArgsModel`, `CancelShiftOccurrenceArgsModel`, `MoveShiftOccurrenceArgsModel`, `ListRoomDayArgsModel`, `ListBoardArgsModel`, `FindFreeRoomsArgsModel`, `ListPublicAvailabilityArgsModel`, `OptionModel`. Check: `grep -n "ArgsModel\|IdRefModel\|OptionModel" definition.go` → only the two `…OptionModel` of 1.2.6.
2. **Add two errors** to the `const (…)` block of errors: `ErrBlankName model.Invalid = "room_layout: the name cannot be blank"` and `ErrShiftWritesPending model.Invalid = "room_layout: shifts cannot be changed from the screen yet"`.
3. **Every table** (`FloorModel`, `RoomModel`, `RoomArtifactModel`, `EquipmentModel`, `RoomEquipmentModel`, `RoomCategoryModel`, `RoomShiftModel`, `RoomShiftCancellationModel`): `tenant_id` gets `ServerOnly: true`. The tables whose rows the server creates (`FloorModel`, `RoomModel`, `RoomArtifactModel`, `EquipmentModel`, `RoomShiftModel`): `id` becomes `{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true, ServerOnly: true}`.
4. **`updated_at`** of `FloorModel`, `RoomModel`, `EquipmentModel` becomes `{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true}`. `RoomShiftModel.updated_at` is NOT changed.
5. **`Allow` and `Rules`** (set exactly; a table not listed has none):

| Table | `Allow` | `Rules` |
|---|---|---|
| `FloorModel` | `model.Create \| model.Read \| model.Update \| model.Delete` | `Create: AddMyBusinessRule, Read: HaveNoBusinessRule, Update: AddMyBusinessRule, Delete: AddMyBusinessRule` |
| `RoomModel` | same four | `Create: AddMyBusinessRule, Read: HaveNoBusinessRule, Update: AddMyBusinessRule, Delete: HaveNoBusinessRule` |
| `RoomArtifactModel` | same four | `Create`, `Update`, `Delete`: `HaveNoBusinessRule`; `Read: AddMyBusinessRule` |
| `EquipmentModel` | same four | `Create: AddMyBusinessRule, Read: HaveNoBusinessRule, Update: AddMyBusinessRule, Delete: AddMyBusinessRule` |
| `RoomEquipmentModel` | `model.Create \| model.Read \| model.Delete` | `Create: AddMyBusinessRule, Read: AddMyBusinessRule, Delete: HaveNoBusinessRule` |
| `RoomCategoryModel` | `model.Create \| model.Read \| model.Delete` | `Create: AddMyBusinessRule, Read: AddMyBusinessRule, Delete: AddMyBusinessRule` |
| `RoomShiftModel` | `model.Read` | `Read: AddMyBusinessRule` |
| `RoomShiftCancellationModel` | `model.Read` | `Read: HaveNoBusinessRule` |

   (`Rules` is written `Rules: model.Rules{Create: model.AddMyBusinessRule, …}`, placed after `Name` of each model.)
6. **Read models.** Replace `DayShiftModel`, `BoardRoomModel`, `FreeRoomModel`, `PublicSlotModel` with the literals below and add the two option models (the old `OptionModel` is deleted). A read model has `Wire: true`, `Allow: model.Read`, `Rules: model.Rules{Read: model.AddMyBusinessRule}`, NO `NotNull` on any field and `OmitEmpty` on all; the values the caller sends (the filter) are fields of the model:

```go
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
```
7. **`Definition`** (new), at the end of the file:

```go
// Definition declara el módulo (plan 1: los turnos solo se leen por el cable; escribirlos llega en el plan 2).
var Definition = model.ModuleDefinition{
	Name: "room_layout",
	Entities: []model.Definition{
		FloorModel, RoomModel, RoomArtifactModel, EquipmentModel, RoomEquipmentModel, RoomCategoryModel,
		RoomShiftModel, RoomShiftCancellationModel,
		CategoryOptionModel, OccupantOptionModel, DayShiftModel, BoardRoomModel, FreeRoomModel, PublicSlotModel,
	},
}
```
`RoomShiftFormModel` stays a plain definition (no `Allow`, not in `Entities`): the shift screen converts it with `shiftToForm` (`convert.go`, unchanged).

`go vet ./...` fails until Stage 4 (`ops.go`, the old bodies): expected. Commit and push.
