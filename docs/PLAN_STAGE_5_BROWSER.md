← [Stage 4](PLAN_STAGE_4_METHODS.md) | Next → [Stage 6](PLAN_STAGE_6_TESTS_DOCS.md)

# Stage 5 — `view.go`, `ui/data.go`, `module/`, demo

`view.Ops` has only `Module`, `Actions` and `Page` (the old `List`, `Save`, `Delete` fields are gone); the operation is derived from the module and the entity; the list constructor takes `func() model.TableSlice`. If `view.NewCallerLister` or `view.Ops` differ in `view@v0.6.74`, run `go doc webtyp.com/view Ops` and name the difference in the PR; the reference is https://github.com/veltylabs/device_manager/blob/main/view.go. The map (`ui/stage*.go`, `ui/rootview.go`, `ui/types.go`, `ui/parts.go`, `ui/plan.go`) is NOT touched: only `ui/data.go`, which talks to the server.

## 5.1 `view.go` — replace the whole file with exactly

```go
package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

// errNotShiftForm es un error de programación (se pasó otro modelo): dato, no texto a traducir.
var errNotShiftForm = "room_layout: expected *RoomShiftForm"

// NewFloorView construye el Presenter de los pisos.
//
// Ejemplo: p := roomlayout.NewFloorView(caller)
func NewFloorView(caller router.Caller, opts ...view.Option) view.Presenter {
	lister := view.NewCallerLister(view.NewRecords().Caller(caller),
		view.Ops{Module: ModelName, Page: view.Page{First: view.All}},
		func() model.TableSlice { return &FloorList{} })
	allOpts := append([]view.Option{view.WithTitle("Floors")}, opts...)
	return view.New(lister, &Floor{}, allOpts...)
}

// NewRoomView construye el Presenter de las salas.
//
// Ejemplo: p := roomlayout.NewRoomView(caller)
func NewRoomView(caller router.Caller, opts ...view.Option) view.Presenter {
	lister := view.NewCallerLister(view.NewRecords().Caller(caller),
		view.Ops{Module: ModelName, Page: view.Page{First: view.All}},
		func() model.TableSlice { return &RoomList{} })
	allOpts := append([]view.Option{view.WithTitle("Rooms")}, opts...)
	return view.New(lister, &Room{}, allOpts...)
}

// NewEquipmentView construye el Presenter del equipamiento.
//
// Ejemplo: p := roomlayout.NewEquipmentView(caller)
func NewEquipmentView(caller router.Caller, opts ...view.Option) view.Presenter {
	lister := view.NewCallerLister(view.NewRecords().Caller(caller),
		view.Ops{Module: ModelName, Page: view.Page{First: view.All}},
		func() model.TableSlice { return &EquipmentList{} })
	allOpts := append([]view.Option{view.WithTitle("Equipment")}, opts...)
	return view.New(lister, &Equipment{}, allOpts...)
}

// roomShiftLister lista los turnos de una sala como formularios (horas «HH:MM»). STATUS (remove this line when plan 2
// closes): guardar y borrar turnos todavía no se ofrece por el cable; el plan 2 los activa.
type roomShiftLister struct {
	caller router.Caller
	roomID string
}

func (l *roomShiftLister) Page() view.Page {
	return view.Page{First: view.All}
}

func (l *roomShiftLister) List(_ view.ListArgs, done func(rows []view.Record, err error)) {
	var res RoomShiftList
	args := &crud.ReadArgs{Limit: crud.MaxPage, Filter: &RoomShift{RoomId: l.roomID}}
	op := router.Operation{Module: ModelName, Entity: (&RoomShift{}).ModelName(), Action: model.Read}
	l.caller.Call(op, args, &res, func(err error) {
		if err != nil {
			done(nil, err)
			return
		}
		var rows []view.Record
		for _, s := range res {
			form := shiftToForm(*s)
			rows = append(rows, &form)
		}
		done(rows, nil)
	})
}

func (l *roomShiftLister) Save(recs []view.Record, done func(err error)) {
	if len(recs) == 0 {
		done(nil)
		return
	}
	if _, ok := recs[0].(*RoomShiftForm); !ok {
		done(fmt.Err(errNotShiftForm))
		return
	}
	done(ErrShiftWritesPending)
}

func (l *roomShiftLister) Delete(ids []string, done func(err error)) {
	if len(ids) == 0 {
		done(nil)
		return
	}
	done(ErrShiftWritesPending)
}

func (l *roomShiftLister) Records() *view.Records {
	return nil
}

func (m *RoomShiftForm) Item() model.Item {
	return model.Item{ID: m.Id, Label: m.OccupantLabel, Description: model.JoinItem(m.DayOfWeek, m.Date, m.Start, m.End)}
}

// NewShiftView construye el Presenter de los turnos de una sala.
//
// Ejemplo: p := roomlayout.NewShiftView(caller, "room-1")
func NewShiftView(caller router.Caller, roomID string) view.Presenter {
	lister := &roomShiftLister{caller: caller, roomID: roomID}
	return view.New(lister, &RoomShiftForm{})
}
```
`view.ListArgs` is still the argument of the `Lister` interface's `List` in `view@v0.6.74` if the compiler accepts it; if the interface of `view.Lister` has changed, copy its exact `List` signature from `go doc webtyp.com/view Lister` and name the difference in the PR. `ErrShiftWritesPending` is the Stage 1 error.

## 5.2 `ui/data.go` — the server calls (five sites)

1. Add, below `NewMap`, written exactly (imports `webtyp.com/crud` and keep `router`, `view`, `model`):

```go
// operation arma la operación de una entidad del módulo.
func operation(entity model.Model, action model.Action) router.Operation {
	return router.Operation{Module: roomlayout.ModelName, Entity: entity.ModelName(), Action: action}
}

// readAll es lo que pide una lectura completa (el tope de una página de crud).
var readAll = &crud.ReadArgs{Limit: crud.MaxPage}
```
2. In `reload`, replace the four calls:
   - `r.caller.Call(op(roomlayout.OpListFloors), &view.ListArgs{}, floors, …)` → `r.caller.Call(operation(&roomlayout.Floor{}, model.Read), readAll, floors, …)`;
   - the same for `OpListRooms` → `operation(&roomlayout.Room{}, model.Read)` and `OpListRoomArtifacts` → `operation(&roomlayout.RoomArtifact{}, model.Read)`;
   - occupants: `occ := &roomlayout.OccupantOptionList{}` and `r.caller.Call(operation(&roomlayout.OccupantOption{}, model.Read), &crud.ReadArgs{Limit: crud.MaxPage, Filter: &roomlayout.OccupantOption{}}, occ, …)`.
   The signature of `fill` changes its last parameter to `occ roomlayout.OccupantOptionList` (its body reads `.Id` and `.Label`, which the new type has).
3. `saveFloor`, `saveRoom`, `saveArtifact`: the call `r.save(roomlayout.OpListFloors, roomlayout.OpSaveFloor, roomlayout.OpDeleteFloor, func() model.ModelSlice {…}, rec)` becomes `r.save(func() model.TableSlice { return &roomlayout.FloorList{} }, rec)` (same record, only the first three arguments disappear and the list constructor returns `model.TableSlice`); the same for `RoomList` and `RoomArtifactList`. A new record still goes with `Id: ""` (the server stamps it: Stage 1 `ServerOnly`); the records NEVER carry `TenantId`.
4. `save` and `deleteArtifact`, replace with exactly:

```go
// save escribe un registro por el mismo camino que las pantallas de lista y relee al terminar.
func (r *RootView) save(newList func() model.TableSlice, rec model.Model) {
	if r.caller == nil {
		return
	}
	lister := view.NewCallerLister(r.caller, view.Ops{Module: roomlayout.ModelName, Page: view.Page{First: view.All}}, newList)
	lister.(view.Saver).Save([]model.Model{rec}, r.savedAndReload)
}

// deleteArtifact borra un equipo del mapa.
func (r *RootView) deleteArtifact(id string) {
	lister := view.NewCallerLister(r.caller, view.Ops{Module: roomlayout.ModelName, Page: view.Page{First: view.All}},
		func() model.TableSlice { return &roomlayout.RoomArtifactList{} })
	lister.(view.Deleter).Delete([]string{id}, r.savedAndReload)
}
```
5. `deactivateRoom` becomes (a deactivation is an update of `is_active`):

```go
func (r *RootView) deactivateRoom(id string) {
	args := &crud.UpdateArgs{IDs: []string{id}, Fields: []string{roomlayout.Room_.IsActive}, Record: &roomlayout.Room{}}
	r.caller.Call(operation(&roomlayout.Room{}, model.Update), args, nil, r.savedAndReload)
}
```
6. Delete `func op(name string) string` (`grep -n "func op(\|op(roomlayout" ui/*.go` → empty).

## 5.3 `module/`

`module/server.go`: `var browser appkit.ModuleBrowser` → `var browser view.ModuleBrowser` (import `webtyp.com/view`). `module/browser.go`: `var browser = appkit.ModuleBrowser{Screens: func(d appkit.BrowserDeps) []view.Screen {…}}` → `var browser = view.ModuleBrowser{Screens: func(d view.BrowserDeps) []view.Screen { return ui.Screens(d.Caller) }}`; `appkit.ModuleServer` stays. (Same change as https://github.com/veltylabs/patient_directory/blob/main/module/browser.go.)

## 5.4 Demo — `seed/demo.go` (new, `//go:build wasm`), written exactly; then `git rm web/browser/main.go` (`test ! -e web`)

```go
//go:build wasm

package seed

import (
	roomlayout "github.com/veltylabs/room_layout"
	"github.com/veltylabs/room_layout/migrate"
	"github.com/veltylabs/room_layout/ui"
	"webtyp.com/crud"
	"webtyp.com/ddl"
	"webtyp.com/fmt"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/time"
	"webtyp.com/unixid"
	"webtyp.com/view"
)

const (
	demoTenantID = "demo-tenant"
	demoZone     = "America/Santiago"
	demoCategory = "cat-general"
	demoOccupant = "occ-perez"
)

// demoCategories es el lector de categorías de la demostración: una sola categoría.
type demoCategories struct{}

func (demoCategories) CategoryOptions(tenantID string) ([]fmt.KeyValue, error) {
	return []fmt.KeyValue{{Key: demoCategory, Value: "Medicina general"}}, nil
}

// demoOccupants es el lector de ocupantes de la demostración: un solo profesional.
type demoOccupants struct{}

func (demoOccupants) OccupantOptions(tenantID string) ([]fmt.KeyValue, error) {
	return []fmt.KeyValue{{Key: demoOccupant, Value: "Dra. Pérez"}}, nil
}

// Demo crea una demostración ejecutable del módulo room_layout sobre una base en memoria con datos semilla.
//
// Ejemplo: demo, err := seed.Demo()
func Demo() (view.Demo, error) {
	db := orm.New(mem.New())
	ids, err := unixid.NewUnixID()
	if err != nil {
		return view.Demo{}, err
	}
	zone, err := time.LoadZone(demoZone)
	if err != nil {
		return view.Demo{}, err
	}
	mod, err := roomlayout.New(db, roomlayout.Deps{
		IDs: ids, TenantID: demoTenantID, Zone: zone, Categories: demoCategories{}, Occupants: demoOccupants{},
	})
	if err != nil {
		return view.Demo{}, err
	}
	if _, err := Load(mod, demoTenantID, LoadOptions{CategoryIDs: []string{demoCategory}, OccupantIDs: []string{demoOccupant}}); err != nil {
		return view.Demo{}, err
	}
	return view.Demo{
		Label:      "Salas",
		Timezone:   demoZone,
		ActingAs:   "u1",
		Modules:    []crud.Module{mod},
		Migrations: []ddl.Step{{Name: roomlayout.ModelName, Run: migrate.Migrate}},
		Screens: func(d view.BrowserDeps) []view.Screen {
			return ui.Screens(d.Caller)
		},
	}, nil
}
```
`seed/seed.go` (`Load`) does not change (it uses the kept methods). If `go get webtyp.com/unixid@latest` is needed, run it and `go mod tidy`. The `wasm` destination of `gotest` compiles `seed/demo.go`; do not run the demo (headless CI). `grep -rn "webtyp.com/layout\|appkit.BrowserDeps\|appkit.ModuleBrowser\|router/loopback" --include=*.go . | grep -v "^./ui/"` → empty (the `ui/` package legitimately imports `webtyp.com/layout` for its widgets). Commit and push.
