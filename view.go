package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

func NewFloorView(caller router.Caller, opts ...view.Option) view.Presenter {
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
	allOpts := append([]view.Option{view.WithTitle("Floors")}, opts...)
	return view.New(lister, &Floor{}, allOpts...)
}

func NewRoomView(caller router.Caller, opts ...view.Option) view.Presenter {
	lister := view.NewCallerLister(
		caller,
		view.Ops{
			Module: ModelName,
			List:   OpListRooms,
			Save:   OpSaveRoom,
		},
		func() model.ModelSlice { return &RoomList{} },
	)
	allOpts := append([]view.Option{view.WithTitle("Rooms")}, opts...)
	return view.New(lister, &Room{}, allOpts...)
}

func NewEquipmentView(caller router.Caller, opts ...view.Option) view.Presenter {
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
	allOpts := append([]view.Option{view.WithTitle("Equipment")}, opts...)
	return view.New(lister, &Equipment{}, allOpts...)
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

func (l *roomShiftLister) Records() *view.Records {
	return nil
}

func NewShiftView(caller router.Caller, tenantID, roomID string) view.Presenter {
	lister := &roomShiftLister{caller: caller, tenantID: tenantID, roomID: roomID}
	return view.New(lister, &RoomShiftForm{})
}
