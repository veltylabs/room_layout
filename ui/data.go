package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

// cellSep separa las celdas cuando se guardan en el servidor como un solo texto: "A1,B1,C1".
const cellSep = ","

// NewMap arma el mapa del recinto sobre las operaciones del módulo: al abrir lee los pisos,
// boxes, equipos y personal del servidor, y cada cambio se guarda en el momento.
//
// Ejemplo: mapa := ui.NewMap(caller)
func NewMap(caller router.Caller) *RootView {
	return &RootView{caller: caller}
}

// reload vuelve a leer todo del servidor y redibuja: el mapa muestra siempre lo guardado.
func (r *RootView) reload() {
	if r.caller == nil {
		return
	}
	floors := &roomlayout.FloorList{}
	r.caller.Call(op(roomlayout.OpListFloors), &roomlayout.ListFloorsArgs{}, floors, func(err error) {
		if r.failed(err) {
			return
		}
		rooms := &roomlayout.RoomList{}
		r.caller.Call(op(roomlayout.OpListRooms), &roomlayout.ListRoomsArgs{ActiveOnly: true}, rooms, func(err error) {
			if r.failed(err) {
				return
			}
			arts := &roomlayout.RoomArtifactList{}
			r.caller.Call(op(roomlayout.OpListRoomArtifacts), nil, arts, func(err error) {
				if r.failed(err) {
					return
				}
				occ := &roomlayout.OptionList{}
				r.caller.Call(op(roomlayout.OpListOccupants), &roomlayout.ListOptionsArgs{}, occ, func(err error) {
					if r.failed(err) {
						return
					}
					r.fill(*floors, *rooms, *arts, *occ)
					r.status.Set("")
					r.RebuildAllFloors()
				})
			})
		})
	})
}

// failed muestra el error en la cabecera del mapa; los datos que había se conservan.
func (r *RootView) failed(err error) bool {
	if err == nil {
		return false
	}
	r.status.Set(err.Error())
	return true
}

func (r *RootView) fill(floors roomlayout.FloorList, rooms roomlayout.RoomList, arts roomlayout.RoomArtifactList, occ roomlayout.OptionList) {
	r.Floors = make([]FloorData, 0, len(floors))
	for _, f := range floors {
		r.Floors = append(r.Floors, FloorData{ID: f.Id, Name: f.Name, Position: int(f.Position), HabitableCells: splitCells(f.HabitableCells)})
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
		r.Rooms = append(r.Rooms, RoomData{ID: rm.Id, FloorID: rm.FloorId, Code: rm.Code, Name: rm.Name, RoomType: rm.RoomType, Cells: splitCells(rm.Cells)})
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
	r.save(roomlayout.OpListFloors, roomlayout.OpSaveFloor, roomlayout.OpDeleteFloor,
		func() model.ModelSlice { return &roomlayout.FloorList{} },
		&roomlayout.Floor{Id: f.ID, Name: f.Name, Position: int64(f.Position), Cols: int64(r.Cols), Rows: int64(r.Rows), HabitableCells: joinCells(f.HabitableCells)})
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
	r.save(roomlayout.OpListRooms, roomlayout.OpSaveRoom, "",
		func() model.ModelSlice { return &roomlayout.RoomList{} },
		&roomlayout.Room{Id: rm.ID, FloorId: rm.FloorID, Code: rm.Code, Name: rm.Name, RoomType: rm.RoomType, Cells: joinCells(rm.Cells), IsActive: true})
}

// saveArtifact guarda un equipo del mapa. Uno nuevo va con ID "".
func (r *RootView) saveArtifact(a ArtifactData) {
	r.save(roomlayout.OpListRoomArtifacts, roomlayout.OpSaveRoomArtifact, roomlayout.OpDeleteRoomArtifact,
		func() model.ModelSlice { return &roomlayout.RoomArtifactList{} },
		&roomlayout.RoomArtifact{Id: a.ID, RoomId: a.RoomID, Kind: a.Kind, Code: a.Code, Cell: a.Cell, Status: a.Status, Reason: a.Reason, NextMaintenance: a.NextMaintenance, Items: a.Items})
}

// deactivateRoom quita un box del mapa (se desactiva: sus turnos pasados siguen existiendo).
func (r *RootView) deactivateRoom(id string) {
	r.caller.Call(op(roomlayout.OpDeactivateRoom), &roomlayout.DeactivateRoomArgs{Id: id}, nil, func(err error) {
		if !r.failed(err) {
			r.reload()
		}
	})
}

// deleteArtifact borra un equipo del mapa.
func (r *RootView) deleteArtifact(id string) {
	lister := view.NewCallerLister(r.caller,
		view.Ops{Module: roomlayout.ModelName, List: roomlayout.OpListRoomArtifacts, Save: roomlayout.OpSaveRoomArtifact, Delete: roomlayout.OpDeleteRoomArtifact},
		func() model.ModelSlice { return &roomlayout.RoomArtifactList{} })
	lister.(view.Deleter).Delete([]string{id}, func(err error) {
		if !r.failed(err) {
			r.reload()
		}
	})
}

// save escribe un registro por el mismo camino que las pantallas de lista y relee al terminar.
func (r *RootView) save(list, save, del string, newList func() model.ModelSlice, rec model.Model) {
	if r.caller == nil {
		return
	}
	lister := view.NewCallerLister(r.caller, view.Ops{Module: roomlayout.ModelName, List: list, Save: save, Delete: del}, newList)
	lister.(view.Saver).Save([]model.Model{rec}, func(err error) {
		if !r.failed(err) {
			r.reload()
		}
	})
}

func op(name string) string { return roomlayout.ModelName + "." + name }

func splitCells(s string) []string {
	if s == "" {
		return nil
	}
	return fmt.Convert(s).Split(cellSep)
}

func joinCells(cells []string) string { return fmt.Convert(cells).Join(cellSep).String() }

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

var _ dom.Component = (*RootView)(nil)
