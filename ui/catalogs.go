package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/dom"
	"webtyp.com/layout/crudview"
	"webtyp.com/model"
	"webtyp.com/router"
)

func buildRoomsTab(caller router.Caller, ids model.IDGenerator) (*dom.Element, error) {
	presenter := roomlayout.NewRoomView(caller)
	cv, err := crudview.New(crudview.Config{
		ParentID:  "rl-rooms-crud",
		Presenter: presenter,
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}
	_ = cv
	container := dom.NewElement("div").ID("rl-rooms-crud")
	return container, nil
}

func buildCatalogsTab(caller router.Caller, ids model.IDGenerator) (*dom.Element, error) {
	container := dom.NewElement("div").Class("rl-catalogs-tab")

	floorsDiv := dom.NewElement("div").ID("rl-floors-crud")
	eqDiv := dom.NewElement("div").ID("rl-eq-crud")

	fPresenter := roomlayout.NewFloorView(caller)
	_, _ = crudview.New(crudview.Config{
		ParentID:  "rl-floors-crud",
		Presenter: fPresenter,
		IDs:       ids,
	})

	eqPresenter := roomlayout.NewEquipmentView(caller)
	_, _ = crudview.New(crudview.Config{
		ParentID:  "rl-eq-crud",
		Presenter: eqPresenter,
		IDs:       ids,
	})

	container.Child(floorsDiv, eqDiv)
	return container, nil
}
