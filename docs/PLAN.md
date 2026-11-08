---
PLAN: "refactor!: room_layout on the closed crud contract — server tenant, transactional batch, single declarative validation, public read via registry, seed.Demo"
EXECUTOR: jules
REVIEWER: none
STATUS: draft
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.
> Never edit `model_orm.go` or `model_crud.go`: a build error there means the hand-written code is wrong.

# Plan: `room_layout` on `crud.Registry`

Context: master `webtyp/docs/ARCHITECTURE_MASTER_PLAN.md` D2–D18 (phases A4.2 + L5). Follows the template in `patient_directory`.
After this plan the module mounts operations through generated helpers on `crud.Registry`,
tenant comes from `Deps`, validation of the schema lives in one place (`crud`), `PublicSlot` does not call `.Public()` (D10: public reads are configured by the app), and `seed.Demo` provides the demo without `webtyp.com/layout`.

## Tools and local dependencies

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go install webtyp.com/ormc/cmd/ormc@v0.1.37
```
```bash
go get webtyp.com/model@v0.2.14 webtyp.com/router@v0.4.19 webtyp.com/crud@v0.0.3 webtyp.com/rpc@v0.2.17 webtyp.com/mcp@v0.3.23 \
  webtyp.com/dbsync@v0.3.47 webtyp.com/view@v0.6.67 webtyp.com/appkit@v0.2.28
```
`view@v0.6.67` and `appkit@v0.2.28` contain `view.Demo` and `appkit` without `layout` (plans L1, L2).
Remove every `replace` from `go.mod`; `go mod tidy`; run `ormc` (generates `model_crud.go`).
After the delivery check, `go.mod` does not list `webtyp.com/layout`.

## APIs used

| Symbol | Signature |
|---|---|
| generated `Mount<Entity>Read/Create/Update/Delete` | typed signatures on `crud.Registry` |
| `crud.ReadArgs` | `struct{ ID, Search, After string; Limit int64; Filter model.Model }` |
| `crud.Loopback` / `crud.LoopbackActingAs` | `func Loopback(mods ...crud.Module) router.Caller` / `func LoopbackActingAs(userID string, mods ...crud.Module) router.Caller` |
| `crud.Op` | `func Op(m model.ModuleNaming, action model.Action) string` |
| `model.ValidateFields` | `func ValidateFields(action byte, f Fielder) error` |
| `orm.DB.Tx` | `func (d *DB) Tx(fn func(tx *DB) error) error` |
| `view.Notifier` | notification contract (D16) |

## Explicit deletions (grep must be empty)

- `ops.go`: raw `reg.Operation` calls, `var _ router.OperationModule`, `.Public()` calls (D10: room_layout drops `.Public()`, ops.go:529), imports `webtyp.com/router`, `webtyp.com/rpc`.
  `grep -rn "reg.Operation\|router.Context\|rpc\.\|\.Public(" --include=*.go . | grep -v _test` → empty.
- `model.go`: manual argument models for ops if any; errors covered by schema/crud.
- `module.go`: operations logic moved to `crud.go` and `queries.go`.
- `web/`: directory deleted (`test ! -d web`).

## Stage 1 — `model.go` and `errors.go`

- `tenant_id` and `updated_at` get `ServerOnly: true` on models.
- Remove manual argument models (`ListFloorsArgsModel`, `ListRoomsArgsModel`, `GetRoomArgsModel`, etc.) that are superseded by `crud.ReadArgs`.
- Verify `errors.go` has doc comments in Spanish with usage examples (`// Ejemplo:`).
- Install `ormc@v0.1.37` and run `ormc generate .`.

## Stage 2 — `crud.go` (writes) and `queries.go` (reads)

- `crud.go`: Create, Update, Delete methods for Floor, Room, Equipment, Category, Shift, etc. wrapped in `m.db.Tx`.
  Signatures accept `(c crud.Caller, list <Entity>List) error`.
- `queries.go`: Read methods for entities accepting `(c crud.Caller, args crud.ReadArgs, filter model.Fielder) (model.ModelSlice, error)`.
  Filter always scopes by `tenant_id` from `m.tenantID`.
  `PublicSlot` read method implements availability read without calling `.Public()` (public exposure is configured at app level).

## Stage 3 — `ops.go` and `module.go`

- `ops.go`:
```go
const ModelName = "room_layout"

func (m *Module) ModelName() string { return ModelName }

var _ crud.Module = (*Module)(nil)

func (m *Module) MountOperations(reg crud.Registry) {
    MountFloorRead(reg, nil, m.ReadFloors)
    MountFloorCreate(reg, m.CreateFloors)
    MountFloorUpdate(reg, m.UpdateFloors)
    MountFloorDelete(reg, m.DeleteFloors)

    MountRoomRead(reg, nil, m.ReadRooms)
    MountRoomCreate(reg, m.CreateRooms)
    MountRoomUpdate(reg, m.UpdateRooms)
    MountRoomDelete(reg, m.DeleteRooms)

    MountEquipmentRead(reg, nil, m.ReadEquipment)
    MountEquipmentCreate(reg, m.CreateEquipment)
    MountEquipmentUpdate(reg, m.UpdateEquipment)
    MountEquipmentDelete(reg, m.DeleteEquipment)

    MountRoomCategoryRead(reg, nil, m.ReadRoomCategories)
    MountRoomCategoryCreate(reg, m.CreateRoomCategories)
    MountRoomCategoryUpdate(reg, m.UpdateRoomCategories)
    MountRoomCategoryDelete(reg, m.DeleteRoomCategories)

    MountRoomShiftRead(reg, nil, m.ReadRoomShifts)
    MountRoomShiftCreate(reg, m.CreateRoomShifts)
    MountRoomShiftUpdate(reg, m.UpdateRoomShifts)
    MountRoomShiftDelete(reg, m.DeleteRoomShifts)

    MountPublicSlotRead(reg, nil, m.ReadPublicSlots)
}
```
- Adapt generated helpers according to `model_crud.go`.

## Stage 4 — `view.go`, `seed/demo.go`, and remove `web/`

- `view.go`: update `view.Ops` to closed contract: `Allow: model.Read | model.Create | model.Update` (no string op names). Do NOT add `//go:build browser` so tests compile cleanly.
- `seed/demo.go`:
```go
//go:build wasm

package seed

import (
    "github.com/veltylabs/room_layout/ui"
    "webtyp.com/crud"
    "webtyp.com/ddl"
    "webtyp.com/view"
)

func Demo() (view.Demo, error) {
    // create mem DB and seed module
    ...
    return view.Demo{
        Label:    "Room Layout",
        Timezone: "America/Santiago",
        ActingAs: "u1",
        Modules:  []crud.Module{mod},
        Screens:  func(d view.BrowserDeps) []view.Screen { return ui.Screens(...) },
    }, nil
}
```
- Delete `web/` directory (`rm -rf web`).

## Stage 5 — tests and delivery check

- Update tests to use `crud.Loopback(m)` or `crud.LoopbackActingAs(userID, m)` or test direct methods.
- Run `gotest`.
- Run `langc sync .`.
- Verify explicit deletions grep is empty.
