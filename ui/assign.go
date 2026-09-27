package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/router"
)

func isOptionActive(roomCats roomlayout.OptionList, optionID string) bool {
	for _, rc := range roomCats {
		if rc.Id == optionID {
			return true
		}
	}
	return false
}

func isEquipmentActive(roomEq roomlayout.EquipmentList, eqID string) bool {
	for _, eq := range roomEq {
		if eq.Id == eqID {
			return true
		}
	}
	return false
}

func buildAssignTab(caller router.Caller, tenantID string) *dom.Element {
	container := dom.NewElement("div").Class("rl-assign-tab")

	roomSelect := dom.NewElement("select").Class("rl-room-select")
	catBoxContainer := dom.NewElement("div").Class("rl-cat-boxes")
	eqBoxContainer := dom.NewElement("div").Class("rl-eq-boxes")
	msgDiv := dom.NewElement("div").Class("rl-msg")

	saveBtn := dom.NewElement("button").Class("rl-btn-save").Text("Guardar")

	reloadRoomAssignments := func(roomID string) {
		msgDiv.Text("")
		// Get categories
		var catOpts roomlayout.OptionList
		caller.Call(roomlayout.ModelName+"."+roomlayout.OpListCategories, &roomlayout.ListOptionsArgs{TenantId: tenantID}, &catOpts, func(err error) {
			if err == nil {
				var roomCats roomlayout.OptionList
				caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRoomCategories, &roomlayout.ListRoomCategoriesArgs{TenantId: tenantID, RoomId: roomID}, &roomCats, func(err error) {
					catBoxContainer = dom.NewElement("div").Class("rl-cat-boxes")
					for _, opt := range catOpts {
						chk := dom.NewElement("input").Attr("type", "checkbox").Attr("value", opt.Id)
						if isOptionActive(roomCats, opt.Id) {
							chk.Attr("checked", "checked")
						}
						lbl := dom.NewElement("label").Child(chk, dom.NewElement("span").Text(" "+opt.Label))
						catBoxContainer.Child(lbl)
					}
				})
			}
		})

		// Get equipment
		var eqList roomlayout.EquipmentList
		caller.Call(roomlayout.ModelName+"."+roomlayout.OpListEquipment, &roomlayout.ListEquipmentArgs{TenantId: tenantID}, &eqList, func(err error) {
			if err == nil {
				var roomEq roomlayout.EquipmentList
				caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRoomEquipment, &roomlayout.ListRoomEquipmentArgs{TenantId: tenantID, RoomId: roomID}, &roomEq, func(err error) {
					eqBoxContainer = dom.NewElement("div").Class("rl-eq-boxes")
					for _, eq := range eqList {
						chk := dom.NewElement("input").Attr("type", "checkbox").Attr("value", eq.Id)
						if isEquipmentActive(roomEq, eq.Id) {
							chk.Attr("checked", "checked")
						}
						lbl := dom.NewElement("label").Child(chk, dom.NewElement("span").Text(" "+eq.Name))
						eqBoxContainer.Child(lbl)
					}
				})
			}
		})
	}

	var roomsRes roomlayout.RoomList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRooms, &roomlayout.ListRoomsArgs{TenantId: tenantID, ActiveOnly: true}, &roomsRes, func(err error) {
		if err == nil && len(roomsRes) > 0 {
			for _, r := range roomsRes {
				roomSelect.Child(dom.NewElement("option").Attr("value", r.Id).Text(r.Code + " · " + r.Name))
			}
			reloadRoomAssignments(roomsRes[0].Id)
		}
	})

	roomSelect.OnChange(func(_ dom.Event) {
		reloadRoomAssignments(roomSelect.GetID())
	})

	saveBtn.OnClick(func(_ dom.Event) {
		roomID := roomSelect.GetID()
		// Save categories & equipment
		argsCats := roomlayout.SetRoomCategoriesArgs{TenantId: tenantID, RoomId: roomID}
		caller.Call(roomlayout.ModelName+"."+roomlayout.OpSetRoomCategories, &argsCats, nil, func(err error) {
			if err != nil {
				msgDiv.Text(fmt.Sprintf("Error: %v", err))
			} else {
				msgDiv.Text("Guardado correctamente")
			}
		})
	})

	container.Child(roomSelect, catBoxContainer, eqBoxContainer, saveBtn, msgDiv)
	return container
}
