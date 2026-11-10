package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/crud"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/lang"
	"webtyp.com/layout"
	"webtyp.com/model"
	"webtyp.com/msgtype"
	"webtyp.com/router"
	"webtyp.com/view"
)

// NewMap arma el mapa del recinto sobre las operaciones del módulo: al abrir lee los pisos,
// boxes, equipos y personal del servidor, y cada cambio se guarda en el momento.
//
// Ejemplo: mapa := ui.NewMap(caller)
func NewMap(caller router.Caller) *RootView {
	return &RootView{caller: caller}
}

// operation arma la operación de una entidad del módulo.
func operation(entity model.Model, action model.Action) router.Operation {
	return router.Operation{Module: roomlayout.ModelName, Entity: entity.ModelName(), Action: action}
}

// readAll es lo que pide una lectura completa (el tope de una página de crud).
var readAll = &crud.ReadArgs{Limit: crud.MaxPage}

// UseNotifier recibe los avisos de la plataforma: la app la llama al montar la pantalla.
func (r *RootView) UseNotifier(n layout.Notifier) { r.notifier = n }

// savedAndReload confirma un guardado: avisa «Guardado» y relee.
func (r *RootView) savedAndReload(err error) {
	if r.failed(err) {
		return
	}
	r.notify(msgtype.Success, lang.Translate("Saved").String())
	r.reload()
}

// notify muestra un aviso de la plataforma; sin plataforma (pruebas) no hay dónde mostrarlo.
func (r *RootView) notify(t msgtype.Type, msg string) {
	if r.notifier != nil {
		r.notifier.Notify(t, msg, layout.Auto())
	}
}

// reload vuelve a leer todo del servidor y redibuja: el mapa muestra siempre lo guardado.
func (r *RootView) reload() {
	if r.caller == nil {
		return
	}
	floors := &roomlayout.FloorList{}
	r.caller.Call(operation(&roomlayout.Floor{}, model.Read), readAll, floors, func(err error) {
		if r.failed(err) {
			return
		}
		rooms := &roomlayout.RoomList{}
		r.caller.Call(operation(&roomlayout.Room{}, model.Read), readAll, rooms, func(err error) {
			if r.failed(err) {
				return
			}
			arts := &roomlayout.RoomArtifactList{}
			r.caller.Call(operation(&roomlayout.RoomArtifact{}, model.Read), readAll, arts, func(err error) {
				if r.failed(err) {
					return
				}
				occ := &roomlayout.OccupantOptionList{}
				r.caller.Call(operation(&roomlayout.OccupantOption{}, model.Read), &crud.ReadArgs{Limit: crud.MaxPage, Filter: &roomlayout.OccupantOption{}}, occ, func(err error) {
					if r.failed(err) {
						return
					}
					if r.failed(r.fill(*floors, *rooms, *arts, *occ)) {
						return
					}
					r.RebuildAllFloors()
				})
			})
		})
	})
}

// failed muestra el error como aviso de la plataforma (queda hasta que se cierra); los datos
// que había se conservan.
func (r *RootView) failed(err error) bool {
	if err == nil {
		return false
	}
	r.notify(msgtype.Error, err.Error())
	return true
}

// fill pasa lo leído del servidor al mapa; un piso o box con celdas mal guardadas es un error.
func (r *RootView) fill(floors roomlayout.FloorList, rooms roomlayout.RoomList, arts roomlayout.RoomArtifactList, occ roomlayout.OccupantOptionList) error {
	r.Floors = make([]FloorData, 0, len(floors))
	for _, f := range floors {
		cells, err := splitCells(f.HabitableCells)
		if err != nil {
			return err
		}
		r.Floors = append(r.Floors, FloorData{ID: f.Id, Name: f.Name, Position: int(f.Position), HabitableCells: cells})
		if int(f.Cols) > r.Cols {
			r.Cols = int(f.Cols)
		}
		if int(f.Rows) > r.Rows {
			r.Rows = int(f.Rows)
		}
	}
	if r.ActiveFloor >= len(r.Floors) {
		r.ActiveFloor = 0
	}
	r.Rooms = make([]RoomData, 0, len(rooms))
	for _, rm := range rooms {
		cells, err := splitCells(rm.Cells)
		if err != nil {
			return err
		}
		r.Rooms = append(r.Rooms, RoomData{ID: rm.Id, FloorID: rm.FloorId, Code: rm.Code, Name: rm.Name, RoomType: rm.RoomType, Cells: cells})
	}
	r.Artifacts = make([]ArtifactData, 0, len(arts))
	for _, a := range arts {
		r.Artifacts = append(r.Artifacts, ArtifactData{
			ID: a.Id, RoomID: a.RoomId, FloorID: r.floorOfRoom(a.RoomId), Kind: a.Kind, Code: a.Code, Cell: a.Cell,
			Status: a.Status, Reason: a.Reason, NextMaintenance: a.NextMaintenance, Items: a.Items,
		})
	}
	r.Doctors = make([]DoctorData, 0, len(occ))
	for _, o := range occ {
		r.Doctors = append(r.Doctors, DoctorData{ID: o.Id, Name: o.Label, Initials: initials(o.Label)})
	}
	return nil
}

func (r *RootView) floorOfRoom(roomID string) string {
	for _, rm := range r.Rooms {
		if rm.ID == roomID {
			return rm.FloorID
		}
	}
	return ""
}

// saveFloor guarda un piso con el tamaño de cuadrícula actual. Un piso nuevo va con ID "".
func (r *RootView) saveFloor(f FloorData) {
	r.save(func() model.TableSlice { return &roomlayout.FloorList{} },
		&roomlayout.Floor{Id: f.ID, Name: f.Name, Position: int64(f.Position), Cols: int64(r.Cols), Rows: int64(r.Rows), HabitableCells: joinCells(f.HabitableCells)})
}

// ResizeGrid cambia el tamaño de la cuadrícula del edificio y guarda los pisos. Nunca deja fuera
// lo dibujado: si la cuadrícula quedaría más chica que alguna planta, espacio o equipo, no cambia
// y lo avisa.
//
// Ejemplo: mapa.ResizeGrid(22, 38)
func (r *RootView) ResizeGrid(cols, rows int) {
	minCols, minRows := r.usedExtent()
	if cols < minCols || rows < minRows {
		r.notify(msgtype.Warning, lang.Translate(msgGridTooSmall).String())
		r.Refresh()
		return
	}
	r.Cols, r.Rows = cols, rows
	r.gridChanged()
}

// msgGridTooSmall avisa que achicar la cuadrícula dejaría fuera parte del plano.
const msgGridTooSmall lang.Text = "The grid cannot be smaller than what is already drawn"

// usedExtent es la columna y la fila más lejanas que ocupa algo dibujado, en cualquier piso.
func (r *RootView) usedExtent() (cols, rows int) {
	grow := func(cells []string) {
		for _, s := range cells {
			if c, ok := roomlayout.ParseCell(s); ok {
				cols, rows = max(cols, c.Col), max(rows, c.Row)
			}
		}
	}
	for _, f := range r.Floors {
		grow(f.HabitableCells)
	}
	for _, rm := range r.Rooms {
		grow(rm.Cells)
	}
	for _, a := range r.Artifacts {
		grow([]string{a.Cell})
	}
	return cols, rows
}

// gridChanged redibuja y guarda todos los pisos con el nuevo tamaño de cuadrícula.
func (r *RootView) gridChanged() {
	r.RebuildAllFloors()
	for _, f := range r.Floors {
		r.saveFloor(f)
	}
}

// saveRoom guarda un box. Uno nuevo va con ID "".
func (r *RootView) saveRoom(rm RoomData) {
	r.save(func() model.TableSlice { return &roomlayout.RoomList{} },
		&roomlayout.Room{Id: rm.ID, FloorId: rm.FloorID, Code: rm.Code, Name: rm.Name, RoomType: rm.RoomType, Cells: joinCells(rm.Cells), IsActive: true})
}

// saveArtifact guarda un equipo del mapa. Uno nuevo va con ID "".
func (r *RootView) saveArtifact(a ArtifactData) {
	r.save(func() model.TableSlice { return &roomlayout.RoomArtifactList{} },
		&roomlayout.RoomArtifact{Id: a.ID, RoomId: a.RoomID, Kind: a.Kind, Code: a.Code, Cell: a.Cell, Status: a.Status, Reason: a.Reason, NextMaintenance: a.NextMaintenance, Items: a.Items})
}

// deactivateRoom quita un box del mapa (se desactiva: sus turnos pasados siguen existiendo).
func (r *RootView) deactivateRoom(id string) {
	args := &crud.UpdateArgs{IDs: []string{id}, Fields: []string{roomlayout.Room_.IsActive}, Record: &roomlayout.Room{}}
	r.caller.Call(operation(&roomlayout.Room{}, model.Update), args, nil, r.savedAndReload)
}

// deleteArtifact borra un equipo del mapa.
func (r *RootView) deleteArtifact(id string) {
	lister := view.NewCallerLister(r.caller, view.Ops{Module: roomlayout.ModelName, Page: view.Page{First: view.All}},
		func() model.TableSlice { return &roomlayout.RoomArtifactList{} })
	lister.(view.Deleter).Delete([]string{id}, r.savedAndReload)
}

// save escribe un registro por el mismo camino que las pantallas de lista y relee al terminar.
func (r *RootView) save(newList func() model.TableSlice, rec model.Model) {
	if r.caller == nil {
		return
	}
	lister := view.NewCallerLister(r.caller, view.Ops{Module: roomlayout.ModelName, Page: view.Page{First: view.All}}, newList)
	lister.(view.Saver).Save([]model.Model{rec}, r.savedAndReload)
}

// splitCells lee los rectángulos guardados (roomlayout.EncodeArea) como celdas del mapa.
func splitCells(s string) ([]string, error) {
	area, err := roomlayout.DecodeArea(s)
	if err != nil {
		return nil, err
	}
	cells := make([]string, len(area))
	for i, c := range area {
		cells[i] = c.String()
	}
	return cells, nil
}

// joinCells guarda las celdas del mapa como rectángulos (roomlayout.EncodeArea).
func joinCells(cells []string) string {
	area := make([]roomlayout.Cell, 0, len(cells))
	for _, s := range cells {
		if c, ok := roomlayout.ParseCell(s); ok {
			area = append(area, c)
		}
	}
	return roomlayout.EncodeArea(area)
}

// initials: la primera letra de las dos primeras palabras ("Ana Torres" → "AT").
func initials(name string) string {
	out := ""
	for _, w := range fmt.Convert(name).Split(" ") {
		if w != "" && len(out) < 2 {
			out += w[:1]
		}
	}
	return out
}

var (
	_ dom.Component       = (*RootView)(nil)
	_ layout.UsesNotifier = (*RootView)(nil)
)
