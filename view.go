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

// roomShiftLister lista los turnos de una sala como formularios (horas «HH:MM»).
type roomShiftLister struct {
	records *view.Records
	caller  router.Caller
	roomID  string
}

func (l *roomShiftLister) Page() view.Page {
	return view.Page{First: view.All}
}

func (l *roomShiftLister) List(_ crud.ReadArgs, done func(rows []view.Record, err error)) {
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
		l.records.Put(rows...)
		done(rows, nil)
	})
}

func (l *roomShiftLister) Save(recs []model.Model, done func(err error)) {
	if len(recs) == 0 {
		done(nil)
		return
	}
	form, ok := recs[0].(*RoomShiftForm)
	if !ok {
		done(fmt.Err(errNotShiftForm))
		return
	}
	form.RoomId = l.roomID
	shift, err := formToShift(*form)
	if err != nil {
		done(err)
		return
	}
	entity := (&RoomShift{}).ModelName()
	if form.Id == "" {
		op := router.Operation{Module: ModelName, Entity: entity, Action: model.Create}
		var saved RoomShift
		l.caller.Call(op, &crud.CreateArgs{Records: []model.Model{&shift}}, &saved, func(err error) {
			if err != nil {
				done(err)
				return
			}
			f := shiftToForm(saved)
			l.records.Put(&f)
			done(nil)
		})
		return
	}
	op := router.Operation{Module: ModelName, Entity: entity, Action: model.Update}
	var saved RoomShift
	l.caller.Call(op, &crud.UpdateArgs{IDs: []string{form.Id}, Fields: editableShiftFields, Record: &shift}, &saved, func(err error) {
		if err != nil {
			done(err)
			return
		}
		f := shiftToForm(saved)
		l.records.Put(&f)
		done(nil)
	})
}

func (l *roomShiftLister) Delete(ids []string, done func(err error)) {
	if len(ids) == 0 {
		done(nil)
		return
	}
	op := router.Operation{Module: ModelName, Entity: (&RoomShift{}).ModelName(), Action: model.Delete}
	l.caller.Call(op, &crud.DeleteArgs{IDs: ids[:1]}, nil, func(err error) {
		if err == nil {
			l.records.Remove((&RoomShiftForm{}).ModelName(), ids[:1]...)
		}
		done(err)
	})
}

func (l *roomShiftLister) Records() *view.Records {
	return l.records
}

func (m *RoomShiftForm) Item() model.Item {
	return model.Item{ID: m.Id, Label: m.OccupantLabel, Description: model.JoinItem(m.DayOfWeek, m.Date, m.Start, m.End)}
}

// NewShiftView construye el Presenter de los turnos de una sala.
//
// Ejemplo: p := roomlayout.NewShiftView(caller, "room-1")
func NewShiftView(caller router.Caller, roomID string) view.Presenter {
	lister := &roomShiftLister{
		records: view.NewRecords(),
		caller:  caller,
		roomID:  roomID,
	}
	return view.New(lister, &RoomShiftForm{})
}
