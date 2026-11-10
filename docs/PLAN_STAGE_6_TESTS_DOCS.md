← [Stage 5](PLAN_STAGE_5_BROWSER.md)

# Stage 6 — tests, translations, documentation

## 6.1 What already exists

| File | What to do |
|---|---|
| `tests/setup_test.go` (`newTestEnv(tenantID)`, the doubles) | **Keep as it is.** |
| `tests/floor_test.go`, `room_test.go`, `equipment_test.go`, `equipment_device_test.go`, `category_test.go`, `area_test.go`, `board_test.go`, `day_test.go`, `free_test.go`, `public_test.go`, `shift_test.go`, `overlap_test.go`, `occupant_room_test.go`, `tenant_test.go`, `error_kinds_test.go`, `deps_test.go`, `seed_test.go`, `map_data_test.go`, `localnow_test.go`, `weekday_test.go`, `stage_rerender_test.go`, `widgets_test.go`, `ui_css_test.go` | **Keep.** They test the kept methods and the map. Fix only what Stages 1–5 change: `Floor/Room/Equipment.UpdatedAt` is a `time.Nano` now (a test that compares it with seconds compares with `env.clockSec.Nano()`); an empty tenant answers `ErrTenantRequired`. Never delete a test of these. |
| `tests/screen_save_test.go`, `tests/view_test.go` | **Keep and port:** the caller is `view.NewRecords().Caller(crud.LoopbackActingAs(nil, "u1", env.mod))` (not `loopback.WithTenant(...)`); a record carries NO `TenantId` and no `Id` on create (`TestScreenSaveIgnoresTenantFromRequest` becomes: a record carrying a `TenantId` is refused with `model.ErrorInvalid`). |
| `tests/ui_view_test.go` (`browser` tag) | **Keep.** Its caller double takes `router.Operation`: `Call(op router.Operation, args model.Encodable, into model.Decodable, done func(err error))`, `Dispatch(op router.Operation, args model.Encodable)`. |
| `tests/ops_test.go` | **Delete**, but first MOVE `TestEventPublishingOnMutations` to a new `tests/events_test.go` unchanged (it tests the events of the kept methods); `TestMountOperationsAndPublicCheck` is replaced by `TestMount_MatchesDefinition` below. |
| `tests/ops_exec_test.go`, `tests/list_args_test.go`, `tests/room_artifact_ops_test.go` | **Delete** (they drive the removed handlers `op*` and `view.ListArgs`); their domain checks are the crud tests below. |
| `tests/orm_coverage_test.go` | Keep if it still compiles. Delete it ONLY if it fails because the helpers `ReadAll<Entity>` it covers are no longer generated, and say so in the PR. |

## 6.2 New crud tests, `tests/rules_test.go` (package `tests`), with these helpers

```go
package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/crud"
	"webtyp.com/events"
	"webtyp.com/model"
	"webtyp.com/router"
)

// call runs one operation of the module as `user` through the same loopback the demo uses and answers the callback error.
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
```
Use `env := newTestEnv("tenantA")` (the module's tenant is `tenantA`; the fake categories answer `cat1`, `cat2`, `cat3`). A record created through `crud` carries NO `Id` and NO `TenantId` or `UpdatedAt` (server fields: sending one answers `ErrServerOnlyField`); the links carry their two key fields. To create a shift use the kept method exactly as `tests/shift_test.go` and `tests/day_test.go` do (`env.mod.SetRoomCategories(...)` then `env.mod.SaveShift("tenantA", roomlayout.RoomShiftForm{RoomId, CategoryId: "cat1", OccupantLabel, DayOfWeek: roomlayout.WeekdayMonday, Start: "09:00", End: "10:00"})`; copy the form fields of those tests).

| Test (all in `tests/rules_test.go`) | setup → call → expected |
|---|---|
| `TestFloor_CreateThroughCrud` | `call("u1", nil, env.mod, "room_floor", model.Create, &crud.CreateArgs{Records: []model.Model{&roomlayout.Floor{Name: "  Piso 1  ", Cols: 10, Rows: 10}}}, nil)` → no error; `env.mod.ListFloors("tenantA")` has 1 row with `Name == "Piso 1"` and `TenantId == "tenantA"` |
| `TestFloor_BlankNameRejected` | the same create with `Name: "   "` → `model.KindOf(err) == model.ErrorInvalid` and `failedField(err) == "name"` |
| `TestFloor_DeleteWithRoomsRefused` | `r := roomOf(t, env, "R1")`; `room_floor.delete` (`&crud.DeleteArgs{IDs: []string{r.FloorId}}`) → `model.KindOf(err) == model.ErrorConflict`; a floor without rooms deletes without error |
| `TestRoom_CodeIsUniquePerTenant` | `roomOf(t, env, "R1")`; `room_space.create` with `{FloorId, Code: "R1", Name: "x", IsActive: true}` → `model.ErrorConflict`; creating `R2` works; `room_space.update` of `R2` with `Fields: []string{"code"}`, `Record: &roomlayout.Room{Code: "R1"}` → `model.ErrorConflict` |
| `TestRoom_DeactivateIsAnUpdate` | `r := roomOf(t, env, "R1")`; `room_space.update` with `Fields: []string{"is_active"}`, `Record: &roomlayout.Room{IsActive: false}` → no error; `env.mod.GetRoom("tenantA", r.Id)` has `IsActive == false` and `Code == "R1"` |
| `TestEquipment_DeleteRemovesItsLinks` | `r := roomOf(...)`, `e, _ := env.mod.SaveEquipment(roomlayout.Equipment{TenantId: "tenantA", Name: "Camilla"})`, `env.mod.SetRoomEquipment("tenantA", r.Id, []string{e.Id})`; `room_equipment.delete` of `e.Id` → no error; `env.mod.ListRoomEquipment("tenantA", r.Id)` is empty and `room_space_equipment.read` with `Filter: &roomlayout.RoomEquipment{RoomId: r.Id}` answers an empty `roomlayout.RoomEquipmentList` |
| `TestRoomEquipment_CompositeKey` | `r`, `e` as above; `room_space_equipment.create` `{RoomId: r.Id, EquipmentId: e.Id}` → no error; again → `model.ErrorConflict`; with `RoomId: "nope"` → `failedField(err) == "room_id"`; `room_space_equipment.delete` ids `[]string{model.JoinKey(r.Id, e.Id)}` → the link is gone |
| `TestRoomCategory_UnknownCategoryRejected` | `r`; `room_category.create` `{RoomId: r.Id, CategoryId: "nope"}` → `failedField(err) == "category_id"`; with `"cat1"` → no error and `env.mod.ListRoomCategories("tenantA", r.Id)` has 1 row |
| `TestRoomCategory_DeleteInUseRefused` | `r`, `SetRoomCategories(…, []string{"cat1"})`, an active shift of `cat1` in `r` (kept method); `room_category.delete` ids `[]string{model.JoinKey(r.Id, "cat1")}` → `model.KindOf(err) == model.ErrorConflict`; after the shift is deleted (`env.mod.DeleteShift("tenantA", shift.Id)`) the same delete works |
| `TestRoomShift_ReadByRoom` | two rooms with one shift each; `var list roomlayout.RoomShiftList`; `room_shift.read` with `&crud.ReadArgs{Filter: &roomlayout.RoomShift{RoomId: r1.Id}}` → 1 row, `RoomId == r1.Id` |
| `TestRoomShift_WritesNotOffered` | `room_shift.create`, `room_shift.delete` and `room_shift_cancellation.create` each → the error text contains `unknown operation` (plan 2 turns them on) |
| `TestReadModel_Options` | `category_option.read` (`Filter: &roomlayout.CategoryOption{}`) answers 3 rows (`cat1`, `cat2`, `cat3`, with the labels of the fake); `occupant_option.read` answers 2 rows (`occ1`, `occ2`) |
| `TestReadModel_DayShift` | the room and weekly Monday shift of `tests/day_test.go`; `day_shift.read` with `Filter: &roomlayout.DayShift{RoomId: r.Id, Date: "2026-03-16"}` → the same shifts (`ShiftId`) that `env.mod.ListRoomDay("tenantA", r.Id, dateSec)` answers |
| `TestReadModel_FreeRoomNeedsItsFilter` | `free_room.read` with an empty filter → `failedField(err) == "date"`; with `Date: "2026-03-16", Start: "xx", End: "10:00"` → `model.ErrorInvalid`; with a valid date and `09:00`–`10:00` → the free rooms of `env.mod.FindFreeRooms` |
| `TestReadModel_PublicSlotNeedsDate` | `public_slot.read` with an empty filter → `failedField(err) == "date"`; with a date → no error |
| `TestTenant_CrudSeesOnlyItsTenant` | `env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantB", Name: "Otro"})` (the kept method writes tenant B through its own sealed store); `room_floor.read` as the module (tenant A) answers an empty list |
| `TestMount_MatchesDefinition` | `crud.Mount(nil, nil, nil, env.mod)` returns nil |
| `TestCreate_PublishesChange` | `pub := &mockPublisher{}` (of `setup_test.go`) as the FIRST argument of `call`; one `room_floor.create` → `pub.events` has one event whose `Payload` is a `*crud.Change` (events reach the publisher given to `LoopbackActingAs`, never `Deps.Publisher`) |
| `TestDemo_Valid` (`tests/demo_test.go`, `//go:build wasm`) | copy https://github.com/veltylabs/device_manager/blob/main/tests/demo_test.go replacing the import path by `github.com/veltylabs/room_layout/seed` |

A rule that refuses is observed through the callback error; assert the field with `failedField` and the kind with `model.KindOf`. `model.Action` of an operation is `model.Create`, `model.Read`, `model.Update`, `model.Delete`.

## 6.3 `docs/TESTS.md`

One row per `func Test…` of the repo (old and new), under the header `| Test | Rule |` and separator `|---|---|`, each row `| [TestName](../tests/file_test.go) | one line rule |` — the link IS the file; **no File column** (it fails the `gotest` tests-index check). Rules in Spanish.

## 6.4 Translations and documentation (Spanish, `veltylabs`)

1. `langc sync .`: zero untranslated keys for `es` and zero reuse warnings (other languages empty is a warning of the repo, not of this plan). `Err…` constants are not `lang.Text`: they need no key; `langc` drops the keys of the removed texts.
2. `README.md` (≤150 lines) with four sections in this order: `## ¿Por qué existe?` (a map of the building: floors, rooms, equipment, who uses which room when; **Sin esto vs. Con esto**; first mentions linked to [`crud`](https://github.com/webtyp/crud) and [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)); `## Antes de comenzar` (Go, `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, a `*orm.DB` with the tables of [migrate/migrate.go](migrate/migrate.go), `Deps.IDs`, `Deps.TenantID`, `Deps.Zone`, `Deps.Categories`, `Deps.Occupants`, optional `Deps.Hours`, `Deps.Publisher`, `Deps.Clock`); `## Guía paso a paso` (Step 1 a complete compilable `main` that imports `"webtyp.com/fmt"`, never stdlib `fmt`, builds the module over `orm.New(mem.New())` with doubles for the two readers, creates a floor and a room with `SaveFloor`/`SaveRoom` and prints `ListRooms`; Step 2 the same through `room_floor.create` (link [tests/rules_test.go](tests/rules_test.go)); Step 3 open the demo with `webtyp` at the repo root, link [seed/demo.go](seed/demo.go)); `## Mapa de documentación` (table `Orden | Documento | Qué pregunta responde y cuándo leerlo` with `docs/ARCHITECTURE.md`, `docs/TESTS.md` and each file of `docs/diagrams/`). Include the operations table of `PLAN.md` («Antes» / «Después»), a permissions table (every entity and read model with its actions) and two notes: `public_slot` is declared public by the app (`model.Read`), and `STATUS (remove this line when plan 2 closes): guardar, borrar, cancelar y mover turnos no se ofrece por el cable hasta el plan 2; los métodos SaveShift, DeleteShift, CancelShiftOccurrence y MoveShiftOccurrence siguen funcionando`.
3. `docs/ARCHITECTURE.md`: keep today's text that is still true and bring it to four sections `## 1. Visión general en un minuto`, `## 2. Piezas y estructura` (a table: `definition.go`, `checks.go`, `rule_*.go`, `floor.go`, `room.go`, `equipment.go`, `category.go`, `shift.go`, `day.go`, `store.go`, `rows.go`, `order.go`, `view.go`, `ui/`, `seed/`), `## 3. Flujos paso a paso` (crear una sala, vincular una categoría, leer el tablero; link the diagram), `## 4. Contratos y límites de seguridad` (the store sealed to one tenant, composite keys, why a rule and its kept method share `checks.go`).
4. `docs/diagrams/ROOM_CATEGORY.md`: context header (what triggers it; `tests/rules_test.go` and `tests/category_test.go` verify it), a `flowchart TD` WITHOUT `subgraph` (create: room exists? → category known? → write; delete: active shift uses it? → delete) and a numbered walkthrough of each branch and its error (`ErrNotFound`, `ErrUnknownCategory`, `ErrCategoryInUse`). Keep `docs/diagrams/database.md` and bring it to the new `updated_at` type (`time.Nano`) only if it names it.
5. Translate to Spanish every README, `docs/*` (except `docs/PLAN*.md`, `LAST_PLAN_EXECUTED.md`, `ui-reference/`), diagram and code comment still in English; every exported symbol without a usage example gets one.
6. Permanent docs never link `docs/PLAN.md` or its stage files.

Run `gotest` and `langc sync .` one last time and read their output, then check the **Delivery check** of `PLAN.md`, commit, push and make sure the pull request is open.
