package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/dom"
	"webtyp.com/layout/crudview"
	"webtyp.com/model"
	"webtyp.com/router"
)

func buildShiftsTab(caller router.Caller, ids model.IDGenerator, tenantID string) (*dom.Element, error) {
	container := dom.NewElement("div").Class("rl-shifts-tab")

	roomSelect := dom.NewElement("select").Class("rl-room-select")

	var selectedRoomID string
	var cv *crudview.CrudView

	reloadShiftsView := func(roomID string) {
		selectedRoomID = roomID
		presenter := roomlayout.NewShiftView(caller, tenantID, roomID)
		c, err := crudview.New(crudview.Config{
			ParentID:  "rl-shifts-crud",
			Presenter: presenter,
			IDs:       ids,
		})
		if err == nil {
			cv = c
		}
	}

	var roomsRes roomlayout.RoomList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRooms, &roomlayout.ListRoomsArgs{TenantId: tenantID, ActiveOnly: true}, &roomsRes, func(err error) {
		if err == nil && len(roomsRes) > 0 {
			for _, r := range roomsRes {
				roomSelect.Child(dom.NewElement("option").Attr("value", r.Id).Text(r.Code + " · " + r.Name))
			}
			selectedRoomID = roomsRes[0].Id
			reloadShiftsView(selectedRoomID)
		}
	})

	roomSelect.OnChange(func(_ dom.Event) {
		reloadShiftsView(roomSelect.GetID())
	})

	crudDiv := dom.NewElement("div").ID("rl-shifts-crud")
	_ = cv
	container.Child(roomSelect, crudDiv)
	return container, nil
}
