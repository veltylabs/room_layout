package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

var _ view.Itemizer = (*Floor)(nil)
var _ view.Itemizer = (*Room)(nil)
var _ view.Itemizer = (*Equipment)(nil)
var _ view.Itemizer = (*RoomShiftForm)(nil)

func (f *Floor) Item() view.Item {
	return view.Item{
		ID:    f.Id,
		Label: f.Name,
	}
}

func (r *Room) Item() view.Item {
	return view.Item{
		ID:    r.Id,
		Label: r.Code + " · " + r.Name,
	}
}

func (e *Equipment) Item() view.Item {
	return view.Item{
		ID:    e.Id,
		Label: e.Name,
	}
}

func (s *RoomShiftForm) Item() view.Item {
	var desc string
	if s.Date == "" {
		dowName := s.DayOfWeek
		switch s.DayOfWeek {
		case WeekdaySunday:
			dowName = "Domingo"
		case WeekdayMonday:
			dowName = "Lunes"
		case WeekdayTuesday:
			dowName = "Martes"
		case WeekdayWednesday:
			dowName = "Miércoles"
		case WeekdayThursday:
			dowName = "Jueves"
		case WeekdayFriday:
			dowName = "Viernes"
		case WeekdaySaturday:
			dowName = "Sábado"
		}
		desc = fmt.Sprintf("%s %s–%s", dowName, s.Start, s.End)
	} else {
		desc = fmt.Sprintf("%s %s–%s", s.Date, s.Start, s.End)
	}
	return view.Item{
		ID:          s.Id,
		Label:       s.OccupantLabel,
		Description: desc,
	}
}

func NewFloorView(caller router.Caller) view.Presenter {
	lister := view.NewCallerLister(
		caller,
		view.Ops{
			Module: ModelName,
			List:   OpListFloors,
			Save:   OpSaveFloor,
			Delete: OpDeleteFloor,
		},
		func() model.ModelSlice { return &FloorList{} },
	)
	return view.New(lister, &Floor{})
}

func NewRoomView(caller router.Caller) view.Presenter {
	lister := view.NewCallerLister(
		caller,
		view.Ops{
			Module: ModelName,
			List:   OpListRooms,
			Save:   OpSaveRoom,
		},
		func() model.ModelSlice { return &RoomList{} },
	)
	return view.New(lister, &Room{})
}

func NewEquipmentView(caller router.Caller) view.Presenter {
	lister := view.NewCallerLister(
		caller,
		view.Ops{
			Module: ModelName,
			List:   OpListEquipment,
			Save:   OpSaveEquipment,
			Delete: OpDeleteEquipment,
		},
		func() model.ModelSlice { return &EquipmentList{} },
	)
	return view.New(lister, &Equipment{})
}

type roomShiftLister struct {
	caller   router.Caller
	tenantID string
	roomID   string
}

func (l *roomShiftLister) List(done func(rows []model.Model, err error)) {
	args := ListRoomShiftsArgs{TenantId: l.tenantID, RoomId: l.roomID}
	var res RoomShiftFormList
	l.caller.Call(ModelName+"."+OpListRoomShifts, &args, &res, func(err error) {
		if err != nil {
			done(nil, err)
			return
		}
		var rows []model.Model
		for _, f := range res {
			rows = append(rows, f)
		}
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
		done(fmt.Err("room_layout: expected *RoomShiftForm"))
		return
	}
	form.RoomId = l.roomID
	var resForm RoomShiftForm
	l.caller.Call(ModelName+"."+OpSaveRoomShift, form, &resForm, func(err error) {
		done(err)
	})
}

func (l *roomShiftLister) Delete(ids []string, done func(err error)) {
	if len(ids) == 0 {
		done(nil)
		return
	}
	args := DeleteRoomShiftArgs{TenantId: l.tenantID, Id: ids[0]}
	l.caller.Call(ModelName+"."+OpDeleteRoomShift, &args, nil, done)
}

func NewShiftView(caller router.Caller, tenantID, roomID string) view.Presenter {
	lister := &roomShiftLister{caller: caller, tenantID: tenantID, roomID: roomID}
	return view.New(lister, &RoomShiftForm{})
}
