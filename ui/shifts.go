package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/layout/crudview"
	"webtyp.com/model"
	"webtyp.com/router"
)

type shiftsTab struct {
	dom.Element
	caller   router.Caller
	ids      model.IDGenerator
	tenantID string
	picker   *roomPicker
	panel    *dom.SignalNodes
}

func newShiftsTab(caller router.Caller, ids model.IDGenerator, tenantID string) *shiftsTab {
	t := &shiftsTab{
		Element:  *dom.NewElement("div"),
		caller:   caller,
		ids:      ids,
		tenantID: tenantID,
		panel:    dom.NewNodes(),
	}
	t.picker = newRoomPicker(caller, tenantID, func(id string) {
		t.rebuild()
	})
	return t
}

func (t *shiftsTab) Init(ctx dom.Ctx) {
	t.picker.load(t.rebuild)
}

func (t *shiftsTab) rebuild() {
	roomID := t.picker.sel.Get()
	if roomID == "" {
		t.panel.Set([]*dom.Element{
			html.Div().Text("No hay espacios activos."),
		})
		return
	}

	presenter := roomlayout.NewShiftView(t.caller, t.tenantID, roomID)
	cv, err := crudview.New(crudview.Config{
		ParentID:  ID + ".shifts." + roomID,
		Presenter: presenter,
		IDs:       t.ids,
		NewRecord: func() model.Model {
			return &roomlayout.RoomShiftForm{RoomId: roomID}
		},
	})
	if err != nil {
		t.panel.Set([]*dom.Element{
			html.Div().Text("Error al cargar turnos: " + err.Error()),
		})
		return
	}

	t.panel.Set([]*dom.Element{html.Div().Child(cv)})

	// 1. room_id options from picker.rooms
	roomOpts := make([]fmt.KeyValue, len(t.picker.rooms))
	for i, r := range t.picker.rooms {
		roomOpts[i] = fmt.KeyValue{Key: r.Id, Value: r.Code + " · " + r.Name}
	}
	cv.SetOptions("room_id", roomOpts...)

	// 2. category_id options
	var catRes roomlayout.OptionList
	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListCategories, &roomlayout.ListOptionsArgs{TenantId: t.tenantID}, &catRes, func(err error) {
		if err == nil {
			opts := make([]fmt.KeyValue, len(catRes))
			for i, o := range catRes {
				opts[i] = fmt.KeyValue{Key: o.Id, Value: o.Label}
			}
			cv.SetOptions("category_id", opts...)
		}
	})

	// 3. occupant_id options
	var occRes roomlayout.OptionList
	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListOccupants, &roomlayout.ListOptionsArgs{TenantId: t.tenantID}, &occRes, func(err error) {
		if err == nil {
			opts := make([]fmt.KeyValue, len(occRes)+1)
			opts[0] = fmt.KeyValue{Key: "", Value: "Externo (escribir nombre)"}
			for i, o := range occRes {
				opts[i+1] = fmt.KeyValue{Key: o.Id, Value: o.Label}
			}
			cv.SetOptions("occupant_id", opts...)
		}
	})

	// 4. day_of_week options
	cv.SetOptions("day_of_week", roomlayout.WeekdayOptions()...)
}

func (t *shiftsTab) Render() *dom.Element {
	return html.Div().Child(
		t.picker.render(),
		html.Div().BindChildren(t.panel),
	).Class("rl-shifts-tab")
}
