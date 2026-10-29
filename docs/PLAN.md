---
PLAN: "refactor: list screens save and delete through view.MountSave/MountDelete; errors through rpc.WriteError; resources from the tables; English keys translated"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 5492341499030472707
---

> Read AGENTS.md at the repo root before any change.

# PLAN — `room_layout` passes the module review (step R1)

## Why

`room_layout` joins the MVP wave of `veltylabs/mjosefa-cms` (it registers the clinic's boxes, who
works where and when, and the equipment of each room). Before it enters the app it must pass the
6-point module review (skill **app-modules**, «Before a module enters an app»). Today its list
screens (Rooms, Floors, Equipment) **cannot save**: the generic screen sends `{"records":[…]}` and
deletes with `{"ids":[…]}`, but `save_floor`/`save_room`/`save_equipment` decode one record and the
deletes decode their own args. It also has its own error mapping (`mapErrorStatus` +
`writeError`, which sends `err.Error()` even on a 500), permissions as string constants, Spanish
translation keys (`"Domingo"`…), and 64 untranslated keys in fr/de/ru. Step R2 (next plan) puts the
floor-plan map in the menu.

## APIs used

| Symbol | Exact | Tag |
|---|---|---|
| `view.MountSave` | `func MountSave[T any, PT interface{ *T; model.Model }](reg router.OperationRegistry, op string, save func(PT) error) router.Route` | view v0.6.53 |
| `view.MountDelete` | `func MountDelete(reg router.OperationRegistry, op string, remove func(ids []string) error) router.Route` | view v0.6.53 |
| `rpc.WriteError` | `func WriteError(ctx router.Context, err error)` | rpc v0.2.10 |
| `model.ResourceOf` | `func ResourceOf(def *Definition) Resource` | model v0.2.12 |
| `model.NotFound`, `model.Conflict`, `model.Invalid` | `type X string` with `Error()` and `Kind()` | model v0.2.12 |
| `model.KindOf` | `func KindOf(err error) ErrorKind` | model v0.2.12 |

## Tools

```
go get webtyp.com/view@v0.6.53 webtyp.com/rpc@v0.2.10 webtyp.com/model@v0.2.12
```

## Stage 1 — errors (`model.go`, `ops.go`)

Delete `domainError`, `ValidationError`, `mapErrorStatus`, `writeError`. Re-type every `Err*`
constant keeping its text, so statuses do not change (today's `mapErrorStatus`, `ops.go:43-66`):

| Constants | Type |
|---|---|
| `ErrNotFound` | `model.NotFound` |
| `ErrCodeAlreadyExists`, `ErrFloorInUse`, `ErrCategoryInUse`, `ErrRoomOverlap`, `ErrOccupantOverlap` | `model.Conflict` |
| `ErrTenantRequired`, `ErrInvalidRange`, `ErrInvalidDate`, `ErrInvalidWeekday`, `ErrUnknownCategory`, `ErrUnknownOccupant`, `ErrOccupantRequired`, `ErrCategoryNotAllowed`, `ErrOutsideBounds`, `ErrNotWeekly` | `model.Invalid` |

Add `ErrCannotRead model.Invalid = "room_layout: cannot read the request"`. A `ValidationError{Err: err}`
whose `err` comes from a model `Validate` returns `err` as is. Every handler writes errors with
`rpc.WriteError(ctx, err)` (decode failure → `ErrCannotRead`). No `==` between errors anywhere (use
`model.KindOf`). Fix the two `New` errors with odd texts (`module.go:45`: «Deps.Categories no
provided_2» and its sibling) to plain English: «room_layout: Deps.Categories is required».

## Stage 2 — resources

Delete `ResourceFloor`, `ResourceRoom`, `ResourceEquipment`, `ResourceShift`; every use (in `ops.go`
and `ui/browser.go`) becomes `model.ResourceOf(&FloorModel)`, `&RoomModel`, `&EquipmentModel`,
`&RoomShiftModel`. Names do not change (`floor`, `room`, `equipment`, `room_shift`).

## Stage 3 — save and delete of the list screens (`ops.go`)

- `OpSaveFloor`, `OpSaveRoom`, `OpSaveEquipment`: `view.MountSave(reg, op, func(x *T) error { out, err := m.SaveX(*x); *x = out; return err })`
  with the same `.Requires(...)` as today; the tenant defaulting each handler does today
  (`if x.TenantId == "" { x.TenantId = m.tenantID }`) moves inside the function.
- `OpDeleteFloor`, `OpDeleteEquipment`: `view.MountDelete(reg, op, func(ids []string) error { for each id: m.DeleteX(m.tenantID, id) })`.
- `OpSaveRoomShift`, `OpDeleteRoomShift` are used by the map, not by a list screen: leave their
  shapes, change only their errors (Stage 1).
- Delete the replaced handlers and any args type only they used (regenerate `model_orm.go` with
  `ormc` if a model definition goes away).

## Stage 4 — translation keys (`model.go`, `ui/`, `lang.json`)

- Weekday options (`model.go`, `weekdayInput`): English keys (`"Sunday"`…`"Saturday"`).
- Grep `ui/` and the root for Spanish string literals used as `lang.Text`/labels; make them English
  keys.
- `lang.json`: translate every key to every language the file has — **real translations**, never
  placeholders; `langc sync .` → 0 pending, 0 reuse warnings, 0 identical values.

## Stage 5 — tests

- Update tests that used the deleted types/handlers.
- New `tests/screen_save_test.go`: through `view.NewRecords().Caller(loopback.ActingAs(...))` and
  `view.NewCallerLister` with each list presenter's ops (as an app does), save and delete a floor,
  a room and an equipment: no error; rows stored, then gone. Saving a room with a duplicate code
  gives `model.KindOf(err) == model.ErrorConflict` on the client.
- The demo (`web/browser.go`) still opens; screenshot of the three list screens in the PR.

## Stage 6 — docs (Spanish)

`README.md`: guardar/borrar con `view.MountSave`/`MountDelete`, errores por `rpc.WriteError`,
permisos por tabla con `model.ResourceOf`.

## Executor notes
- `view.MountSave` operations were fully implemented for floors, rooms, and equipment.
- All testing fixes have been successfully applied and all tests are passing locally (no panic errors, id checks pass). Test coverage execution has run correctly across UI component rendering constraints.
- Spanish language translation keys (`ui/` files) were mapped to proper English keys. `langc` synced seamlessly across supported languages.
- `ormc` regenerated model schema after removing redundant Arg structs successfully.
