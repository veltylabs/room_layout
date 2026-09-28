---
PLAN: "fix(ui): the room_layout screen actually paints, loads its choices and saves; the demo mounts it"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 3479941765748529990
PR: https://github.com/veltylabs/room_layout/pull/2
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Read `AGENTS.md` at the repo root first: its import rules (no stdlib `fmt`/`strings`/`strconv`/`errors`,
> no Go maps, `webtyp.com/fmt` instead) apply to every file you touch.

# Plan — `room_layout/ui`: a screen that works

## 0. Context (verified — do not re-diagnose)

`v0.1.0` shipped `ui/` and `web/` but the screen is not usable. Verified defects:

| # | File | Defect |
|---|---|---|
| D1 | `ui/catalogs.go` (`buildRoomsTab`, `buildCatalogsTab`) | Each `crudview.New(...)` result is discarded (`_ = cv`, `_, _ =`). The tab paints an empty `div` with a hand-composed id. No room, floor or equipment can be listed or created. |
| D2 | `ui/shifts.go` | Same: the shift `crudview` is built inside a callback and thrown away (`_ = cv`); the tab paints only a `<select>`. |
| D3 | `ui/board.go`, `ui/assign.go`, `ui/shifts.go` | A `<select>`'s chosen value is read with `GetID()`. That returns the element's DOM id, never the chosen option. Filters and room pickers never work. |
| D4 | `ui/board.go` | `reloadBoard` appends a new block to `boardDeck` on every reload (`boardDeck.Child(innerDiv)`), so each filter change duplicates the board. And children appended after render never reach the live DOM anyway. |
| D5 | `ui/assign.go` | The checkbox containers are **reassigned** inside callbacks (`catBoxContainer = dom.NewElement(...)`): the new element is never attached. "Guardar" sends empty category lists and never saves equipment. |
| D6 | `ui/freesearch.go` | `FindFreeRoomsArgs` has no widgets (base `model.Text()` kinds), so `form.New` paints no inputs; the "Asignar" button has no handler. |
| D7 | selects inside crudview forms | `Room.floor_id` and `RoomShiftForm.room_id/category_id/occupant_id/day_of_week` are `input.Select()` with **no options**: nothing can be chosen, so nothing can be saved. |
| D8 | `web/client.go` | Calls `ui.Browser(...)` and discards the result: the demo mounts nothing. |
| D9 | `tests/ui_view_wasm_test.go` | Only checks that `Browser` returns non-nil — why D1–D8 passed. |

**The failing tests already exist**: `tests/ui_live_wasm_test.go` (committed with this plan). It mounts
`ui.Browser` into a live DOM over the real module + seed + `router/loopback`. All four tests are red
today. **Do not edit that file** except to fix a compile error caused by a rename this plan orders
(there is none). They are the acceptance criterion.

Upstream pieces this plan relies on, already published and already in `go.mod`:

- `webtyp.com/layout v0.3.2`: `(*crudview.CrudView).SetOptions(fieldName string, opts ...fmt.KeyValue)` —
  sets the choices of a select in the crudview's form; called after render it repaints the live control.
- `webtyp.com/form v0.4.19`: the mechanism behind it.

## 1. Rules for this plan (repeat of `AGENTS.md`, applied here)

- `ui/` may import `webtyp.com/layout/*`, `webtyp.com/components/*`, `dom`, `html`, `css`, `svg`, `widget`.
  The root package `roomlayout` must not change in this plan **except** the one helper in Stage 1.
- **Never compose an element id** (`ID("rl-...")`, `"x-"+id`). `dom` mints ids. Delete every `.ID(...)`
  call in `ui/`. For a hook tests or CSS need, use `Attr("data-...", value)`.
- **Read a select's value from the event**: `OnChange(func(e dom.Event) { v := e.TargetValue() ... })`.
  `GetID()` is never used to read a value.
- **A list that changes after render is a `*dom.SignalNodes`** bound with `BindChildren`, and it is
  replaced with `.Set(...)` — never appended to with `.Child(...)` after render.
- **Network calls happen in `Init(dom.Ctx)`**, not in constructors. Each tab is a struct that embeds
  `dom.Element` **by value** and implements `Init(dom.Ctx)` and `Render() *dom.Element`.
  Reference: `github.com/veltylabs/appointment_booking/ui/serviceconfigview.go` and `staffpicker.go` —
  the patterns below are copied from them.
- No Go maps. No stdlib `fmt`/`strings`/`strconv`. Texts shown to the user are Spanish (as today).

## 2. Stages

### Stage 1 — weekday options (root package)

New file `weekday.go` in package `roomlayout`:

```go
// WeekdayOptions: the choices for RoomShiftForm.day_of_week. The empty key
// means "dated shift" (the date field is used instead).
func WeekdayOptions() []fmt.KeyValue {
	return []fmt.KeyValue{
		{Key: "", Value: "— fecha específica —"},
		{Key: WeekdayMonday, Value: "Lunes"},
		{Key: WeekdayTuesday, Value: "Martes"},
		{Key: WeekdayWednesday, Value: "Miércoles"},
		{Key: WeekdayThursday, Value: "Jueves"},
		{Key: WeekdayFriday, Value: "Viernes"},
		{Key: WeekdaySaturday, Value: "Sábado"},
		{Key: WeekdaySunday, Value: "Domingo"},
	}
}
```

Test in `tests/weekday_test.go`: 8 entries, first key `""`, every other key is one of the `Weekday*` constants.

### Stage 2 — shared room picker: `ui/roompicker.go` (new)

Copy the shape of `appointment_booking/ui/staffpicker.go`:

```go
type roomPicker struct {
	caller   router.Caller
	tenantID string
	sel      *dom.SignalString   // chosen room id
	opts     *dom.SignalNodes    // <option> nodes
	rooms    []roomlayout.Room
	onChange func(id string)
}
func newRoomPicker(caller router.Caller, tenantID string, onChange func(id string)) *roomPicker
func (p *roomPicker) load(then func())   // calls ModelName+"."+OpListRooms with ListRoomsArgs{TenantId, ActiveOnly: true};
                                         // fills p.rooms, rebuilds options, selects the first room if none selected, then then()
func (p *roomPicker) render() *dom.Element // Div(Label "Espacio", <select> BindChildren(p.opts) OnChange(e.TargetValue()))
```

Option text: `room.Code + " · " + room.Name`. Use `html.Option(id, label)` / `html.SelectedOption(id, label)`
exactly as `staffpicker.go` does. Empty list → one disabled option `"No hay espacios activos"`.

### Stage 3 — Espacios and catalogs tabs: `ui/catalogs.go` (rewrite)

- `roomsPanel(caller, ids, tenantID) (dom.Component, error)`: build the crudview **and return it** (it is the panel):
  ```go
  cv, err := crudview.New(crudview.Config{ParentID: ID + ".rooms", Presenter: roomlayout.NewRoomView(caller), IDs: ids})
  ```
  Then load floors and feed the room form's floor select:
  `caller.Call(ModelName+"."+OpListFloors, &ListFloorsArgs{TenantId: tenantID}, &floors, ...)` →
  `cv.SetOptions("floor_id", <one KeyValue per floor: Key=Id, Value=Name>...)`. Do this call from the
  wrapper's `Init` (see below), not at construction.
- `catalogsPanel(caller, ids)`: a `Div` holding **two** crudviews (floors: `ParentID: ID + ".floors"`;
  equipment: `ParentID: ID + ".equipment"`), each a child of the div — not discarded.

Because a crudview returned as-is has no `Init` hook for the floors call, wrap the rooms crudview in a
small struct `roomsTab{ dom.Element; cv *crudview.CrudView; caller; tenantID }` with `Init` (issue the
floors call → `cv.SetOptions`) and `Render()` returning `dom.NewElement("div").Child(t.cv)`.
When a floor is saved in the catalogs tab the room form's options go stale until reload — accepted.

### Stage 4 — Turnos tab: `ui/shifts.go` (rewrite)

Copy `appointment_booking/ui/serviceconfigview.go`: struct `shiftsTab{ dom.Element; caller; ids; tenantID; picker *roomPicker; panel *dom.SignalNodes }`.

- `Init`: `panel = dom.NewNodes()`; `picker.load(t.rebuild)`. Picker `onChange` → `t.rebuild()`.
- `rebuild()`: no room → `panel.Set(Div().Text("No hay espacios activos."))`. Otherwise
  `crudview.New(Config{ParentID: ID + ".shifts." + roomID, Presenter: roomlayout.NewShiftView(caller, tenantID, roomID), IDs: ids})`,
  then `panel.Set([]*dom.Element{dom.NewElement("div").Child(cv)})`, then feed the four selects:
  - `room_id` ← `picker.rooms` (Key=Id, Value=Code+" · "+Name)
  - `category_id` ← op `OpListCategories` (`ListOptionsArgs{TenantId}` → `OptionList`; Key=Id, Value=Label)
  - `occupant_id` ← `{Key: "", Value: "Externo (escribir nombre)"}` **first**, then op `OpListOccupants`
    (same args/result types). RL-5: an external occupant has no id, only `occupant_label`.
  - `day_of_week` ← `roomlayout.WeekdayOptions()`
  - `NewRecord: func() model.Model { return &roomlayout.RoomShiftForm{RoomId: roomID} }` in the Config, so a
    new shift starts in the picked room.
- `Render()`: `Div().Child(t.picker.render()).Child(Div().BindChildren(t.panel))`.

### Stage 5 — Tablero: `ui/board.go` (rewrite)

Struct `boardTab{ dom.Element; caller; tenantID; category, equipment *dom.SignalString; catOpts, eqOpts, cards *dom.SignalNodes }`.

- `Init`: load categories (`OpListCategories`) and equipment (`OpListEquipment`, `ListEquipmentArgs{TenantId}`)
  into `catOpts`/`eqOpts` (first option `Option("", "Todas las áreas")` / `Option("", "Todo el equipamiento")`),
  then `reload()`.
- Filter selects: `Attr("data-board-filter", "category")` and `Attr("data-board-filter", "equipment")`;
  `OnChange(e) { t.category.Set(e.TargetValue()); t.reload() }` (same for equipment).
- `reload()`: `OpListBoard` with `ListBoardArgs{TenantId, CategoryId: t.category.Get(), EquipmentId: t.equipment.Get()}`.
  On error: `cards.Set([Div().Text("No se pudo cargar el tablero: " + err.Error())])`. On success build the
  whole list **fresh** and `cards.Set(...)` it — group by `FloorName` in arrival order (keep the existing
  `floorGroup` slice helper), an `h2` per floor, then one card per room.
- Each card is `dom.NewElement("div").Attr("data-board-room", br.RoomId).Attr("data-status", br.Status).Child(&contentcard.ContentCard{...})`
  with the same header/body/footer texts as today. Empty result → `Div().Text("No hay espacios que coincidan.")`.
- `Render()`: controls div (two selects bound to `catOpts`/`eqOpts`) + `Div().BindChildren(t.cards)`.

### Stage 6 — Buscar libre: `ui/freesearch.go` (rewrite)

Hand-built controls (no `form.New`: `FindFreeRoomsArgs` is an op-args model, not a form model):
`<input type="date">`, two `<input type="time">` (Desde / Hasta), a category `<select>` (first option
`""` = "Cualquier área"), and an optional **"Turno a mover"** `<select>`. Each control writes its value into
a `*dom.SignalString` from `OnChange(e.TargetValue())`. Default date: empty until the user picks one.

- "Turno a mover" options: when the date changes, call `OpListRoomDay` with `ListRoomDayArgs{TenantId, Date}`
  (no room: all rooms) → `DayShiftList`; options `""` = "— ninguno —" then one per shift:
  `RoomCode + " · " + OccupantLabel + " " + Start + "–" + End`, Key = `ShiftId`. Choosing a shift copies its
  `Start`/`End`/`CategoryId` into the three signals (the inputs are bound with `Bind(signal)`).
- "Buscar" button: validate date/start/end non-empty (else show `"Indique fecha, desde y hasta."`), then
  `OpFindFreeRooms` with `FindFreeRoomsArgs{TenantId, Date, Start, End, CategoryId}` → `FreeRoomList`;
  `results.Set(...)` fresh (never append). Error → show `err.Error()` in the message area.
- Each result card: `Attr("data-free-room", fr.RoomId)`. Only when a shift is chosen, the card carries a
  button **"Mover aquí"** → `OpMoveShiftOccurrence` with
  `MoveShiftOccurrenceArgs{TenantId, ShiftId, Date, TargetRoomId: fr.RoomId}`; success → message
  `"Turno movido a " + fr.Code` and search again; error → `err.Error()`. (RL-4: moving is manual.)
- Delete the old "Asignar" button.

### Stage 7 — Áreas y equipos: `ui/assign.go` (rewrite)

Struct `assignTab{ dom.Element; caller; tenantID; picker *roomPicker; boxes *dom.SignalNodes; msg *dom.SignalString; cats, eqs []roomlayout.Option/Equipment; catOn, eqOn []string }`.

- `Init`: load all categories + all equipment once, then `picker.load(t.loadRoom)`. Picker change → `loadRoom`.
- `loadRoom()`: `OpListRoomCategories` + `OpListRoomEquipment` for the picked room (count both answers
  before painting, like `scheduleview.go`'s `pending` counter), fill `catOn`/`eqOn` (slices of ids), then
  `boxes.Set(...)` fresh: a fieldset "Áreas habilitadas" with one checkbox per category
  (`Attr("data-category-id", id)`, checked when in `catOn`, `OnChange(e) → toggle id in catOn by e.TargetChecked()`)
  and a fieldset "Equipamiento" with one per equipment (`Attr("data-equipment-id", id)`, same with `eqOn`).
- "Guardar": `OpSetRoomCategories` with `SetRoomCategoriesArgs{TenantId, RoomId, CategoryIds: []IdRef built from catOn}`,
  **then** `OpSetRoomEquipment` with `SetRoomEquipmentArgs{..., EquipmentIds: ...}`. Message `"Guardado."` when
  both succeed; otherwise the first `err.Error()` (the module refuses to disable an area with active shifts —
  that message must reach the user).

### Stage 8 — `ui/browser.go`

Tabs in this order, panels are the new structs/components: Tablero (`boardTab`), Buscar libre, Turnos,
Espacios (`roomsTab`), Áreas y equipos (`assignTab`), Niveles y equipamiento (`catalogsPanel`).
Keep `Browser(caller, ids, tenantID, opts ...Option)`, `WithLabel`, `ID`, `DefaultLabel` unchanged
(the app calls them). Delete the `deck := dom.NewElement("div").Child(tabs)` wrapper: pass `tabs` directly
to `platformd.NewUIModule`.

### Stage 9 — `web/client.go`: the demo mounts the screen

Copy `github.com/veltylabs/patient_directory/web/client.go`: keep the current module/seed setup, then
build a `platformd.Platform{AppName: ui.DefaultLabel + " — demo", User: demoUser{}, Modules: []platformd.UIModule{v}, DefaultID: ui.ID}`,
`dom.Append("body", p)` and `select {}`. `panic(err)` on every error (it is a demo). Keep the seed's
`CategoryIDs`/`OccupantIDs` matching the demo readers.

### Stage 10 — tests and docs

- `tests/ui_live_wasm_test.go` (given): all green.
- `tests/ui_view_wasm_test.go`: keep it.
- `docs/ARCHITECTURE.md`: add a short "Pantalla (`ui/`)" section: the six tabs and what each calls; that
  select choices in crudview forms come from `CrudView.SetOptions`; that a move is manual (RL-4).
- `README.md`: "Demo: run `webtyp` at the repo root".

## 3. Acceptance criteria

```bash
gotest                                   # full suite green, wasm included (the wasm build is the one that decides)
grep -rn "GetID()" ui/                   # → empty
grep -rn '\.ID("' ui/                    # → empty
grep -rn "_ = cv\|_, _ = crudview" ui/   # → empty
grep -n "platformd.Platform" web/client.go   # → one hit
```

## 4. Out of scope

- Any change to ops, models (other than `weekday.go`), migrations or the seed.
- The boxes × hours grid and the floor plan (RL-7: later).
