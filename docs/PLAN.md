---
PLAN: "feat(ui): the floor-plan map reads and saves the real data — every change is saved at once through the module's operations"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

# PLAN — R2b: the map works with the real data

## Why

Phase R2 of `veltylabs/mjosefa-cms/docs/MVP_DEPLOYMENT_ROADMAP.md`, decision D34: «Espacios» enters the
app only when the map works. Today `ui/rootview.go` keeps floors, rooms, artifacts and shifts in
memory slices (`Floors`, `Rooms`, `Artifacts`, `Shifts`, `Doctors`), never reads them from the server,
and `SaveAll` is empty (`ui/rootview.go:137`): whatever is drawn is lost. The server side already has
every operation (part 1, v0.3.64 added the artifact ones).

**Decided by the user: every change is saved at once**, like the list screens. There is no «Save»
button. After each save or delete the map reads the list again from the server, so the map always
shows what the server has and never invents ids: a new record is sent with `Id: ""` and the server
gives it its id.

## APIs used

| Symbol | Exact | Where |
|---|---|---|
| `router.Caller.Call` | `Call(op string, args model.Encodable, out model.Decodable, done func(error))`; `op` is `ModelName + "." + OpX` | `webtyp.com/router` (as in `view.go:64`, `roomShiftLister.List`) |
| `view.NewCallerLister` | `func NewCallerLister(caller router.Caller, ops view.Ops, newList func() model.ModelSlice) view.Presenter` | `webtyp.com/view` (as in `tests/room_artifact_ops_test.go`) |
| `view.Saver` / `view.Deleter` | `Save(recs []model.Model, done func(error))` / `Delete(ids []string, done func(error))` | `webtyp.com/view` — the only way to write: the wire shape is private to `view` |
| list ops | `OpListFloors` → `FloorList`; `OpListRooms` → `RoomList`; `OpListRoomArtifacts` (args `nil`) → `RoomArtifactList`; `OpListOccupants` → `OptionList` (`Option{Id, Label}`); `OpListRoomShifts` (args `&ListRoomShiftsArgs{RoomId}`) → `RoomShiftFormList` | `ops.go` |
| write ops | `OpSaveFloor`/`OpDeleteFloor`, `OpSaveRoom`/`OpDeactivateRoom`, `OpSaveRoomArtifact`/`OpDeleteRoomArtifact`, `OpSaveRoomShift`/`OpDeleteRoomShift` | `ops.go` |

Check each signature in the source before using it; if one differs, stop and say so in the PR.

## Stage 1 — conversions (`ui/data.go`, new)

Unexported functions, one per direction, between the module models and the map types of
`ui/types.go`:

- Cells are stored on the server as **one text, comma-separated**: `"A1,B1,C1"`. `splitCells(s string) []string`
  (empty text → `nil`) and `joinCells(cells []string) string`. Use `webtyp.com/fmt`, never `strings`.
- `Floor` ↔ `FloorData` (`HabitableCells` via split/join; `Cols`, `Rows` go to `RootView.Cols/Rows`: the
  largest of the floors read; when saving a floor, write the current `RootView.Cols/Rows`).
- `Room` ↔ `RoomData` (`Cells`, `RoomType`, `FloorId`, `Code`, `Name`); `IsActive: true` on save.
- `RoomArtifact` ↔ `ArtifactData` (`FloorID` is the floor of its room, looked up in `RootView.Rooms`).
- `RoomShiftForm` ↔ `ShiftData`; `Option` → `DoctorData{ID, Name, Initials}` (initials: first letter of
  the first two words).

## Stage 2 — reading (`ui/data.go`, `ui/rootview.go`)

- `func NewMap(caller router.Caller) *RootView` — exported constructor, comment in Spanish with an
  example. `Init` calls `r.reload()`.
- `func (r *RootView) reload()`: calls the five list ops above (shifts: one call per room after the rooms
  arrive), fills the slices through Stage 1, then `r.RebuildAllFloors()`. On an error, keep the previous
  data and show the error text in the map's header (one `*dom.SignalString`), never `panic`.
- Delete `NewID` and the `ids` field: ids come from the server.

## Stage 3 — writing at once (`ui/data.go` + the stages)

Helpers in `ui/data.go`: `saveFloor(f FloorData)`, `saveRoom(rm RoomData)`, `saveArtifact(a ArtifactData)`,
`saveShift(s ShiftData)`, `deleteFloor(id)`, `deactivateRoom(id)`, `deleteArtifact(id)`, `deleteShift(id)`.
Each builds the model (Stage 1), writes through `view.NewCallerLister(...)` `Save`/`Delete` (or `Call` for
`OpDeactivateRoom`, which is not a list op), and on success calls `r.reload()`. A new record has `Id: ""`.

Replace every in-memory change with the helper:

| Where | Today | Becomes |
|---|---|---|
| `rootview.go:126` `AddFloor` | appends a `FloorData` with a local id | `saveFloor(FloorData{Name, Position})` |
| `stage1_planta.go:268` floor name input | `f.Name = …` | `saveFloor(*f)` after setting it |
| `stage1_planta.go:342` `OnPointerUp` | `s.isDragging = false` | also `saveFloor(*f)` (one save per drag, not per cell) |
| `stage1_planta.go:356` «Clear» | `f.HabitableCells = nil` | then `saveFloor(*f)` |
| `stage2_espacios.go:123`, `:238` | `root.Rooms = append(…, newRoom)` | `saveRoom(newRoom)` with `ID: ""` |
| `stage2_espacios.go:286` | `root.Rooms = updated` (removal) | `deactivateRoom(id)` |
| any other edit of a room's name, type or cells in `stage2_espacios.go` | in-memory | `saveRoom(…)` |
| `stage3_artefactos.go:396` | `append(…, newArt)` | `saveArtifact(newArt)` with `ID: ""` |
| `stage3_artefactos.go:146,156,169` status changes | in-memory | `saveArtifact(*selectedArt)` after setting it |
| `stage3_artefactos.go:201` | `root.Artifacts = updated` (removal) | `deleteArtifact(id)` |
| `stage4_operacion.go:465` | `append(…, newShift)` | `saveShift(newShift)` with `ID: ""` |
| `stage4_operacion.go:421` | `root.Shifts = updated` (removal) | `deleteShift(id)` |

Delete `SaveAll` and the four «Guardar» buttons that call it (`stage1_planta.go:55`,
`stage2_espacios.go:66`, `stage3_artefactos.go:51`, `stage4_operacion.go:207`). After this stage
`grep -n "root\.\(Floors\|Rooms\|Artifacts\|Shifts\) *= " ui/*.go` lists only `reload` in `ui/data.go`.

## Stage 4 — tests (`tests/map_data_test.go`, new)

Same setup as `tests/room_artifact_ops_test.go` (`newTestEnv` + `view.NewRecords().Caller(loopback.WithTenant(…))`).

- `TestMap_LoadsServerData`: create through the module one floor (`HabitableCells: "A1,B1"`), one room
  (`Cells: "A1"`), one artifact; `NewMap(caller)` + `Init(nil)` → `Floors[0].HabitableCells` is
  `["A1","B1"]`, `Rooms[0].Cells` is `["A1"]`, one artifact with its `FloorID`.
- `TestMap_AddFloorIsSaved`: `m.AddFloor("Piso 1")` → `env.mod.ListFloors(tenant)` has it, and `m.Floors`
  has it with the server's id.
- `TestMap_ArtifactRoundTrip`: save one through the helper path the stage uses, delete it → the
  module has 0.

## Delivery check

```
GOPRIVATE=webtyp.com/*,github.com/veltylabs/* GOFLAGS=-mod=mod gotest
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never pause to ask permission: when the check passes, commit, push and open the PR. Stuck 30
  minutes: open the PR listing what fails.
- `git diff --stat main` lists only: `ui/data.go`, `ui/rootview.go`, `ui/stage1_planta.go`,
  `ui/stage2_espacios.go`, `ui/stage3_artefactos.go`, `ui/stage4_operacion.go`,
  `tests/map_data_test.go`, this plan. No stray scripts. Texts and the menu entry are the next plan
  (R2c): do not touch `ui/browser.go` or `lang.json` here.

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `ui/data.go` | compiles |
| 2 | `ui/data.go`, `ui/rootview.go` | compiles |
| 3 | `ui/data.go`, `ui/stage*.go` | grep above clean |
| 4 | `tests/map_data_test.go` | suite green |
