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

type roomsTab struct {
	dom.Element
	cv       *crudview.CrudView
	caller   router.Caller
	tenantID string
}

func newRoomsTab(caller router.Caller, ids model.IDGenerator, tenantID string) (*roomsTab, error) {
	presenter := roomlayout.NewRoomView(caller)
	cv, err := crudview.New(crudview.Config{
		ParentID:  ID + ".rooms",
		Presenter: presenter,
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}
	t := &roomsTab{
		Element:  *dom.NewElement("div"),
		cv:       cv,
		caller:   caller,
		tenantID: tenantID,
	}
	return t, nil
}

func (t *roomsTab) Init(ctx dom.Ctx) {
	var floorsRes roomlayout.FloorList
	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListFloors, &roomlayout.ListFloorsArgs{TenantId: t.tenantID}, &floorsRes, func(err error) {
		if err == nil {
			opts := make([]fmt.KeyValue, len(floorsRes))
			for i, f := range floorsRes {
				opts[i] = fmt.KeyValue{Key: f.Id, Value: f.Name}
			}
			t.cv.SetOptions("floor_id", opts...)
		}
	})
}

func (t *roomsTab) Render() *dom.Element {
	return html.Div().Child(t.cv)
}

func catalogsPanel(caller router.Caller, ids model.IDGenerator) (*dom.Element, error) {
	fPresenter := roomlayout.NewFloorView(caller)
	floorsCV, err := crudview.New(crudview.Config{
		ParentID:  ID + ".floors",
		Presenter: fPresenter,
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}

	eqPresenter := roomlayout.NewEquipmentView(caller)
	eqCV, err := crudview.New(crudview.Config{
		ParentID:  ID + ".equipment",
		Presenter: eqPresenter,
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}

	return html.Div().Child(floorsCV, eqCV).Class("rl-catalogs-tab"), nil
}
