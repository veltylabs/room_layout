package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/components/contentcard"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/model"
	"webtyp.com/router"
)

func buildFreeSearchTab(caller router.Caller, ids model.IDGenerator, tenantID string) *dom.Element {
	container := dom.NewElement("div").Class("rl-free-tab")

	args := &roomlayout.FindFreeRoomsArgs{TenantId: tenantID}
	f, err := form.New("rl-free-form", args, ids)
	if err != nil {
		return container
	}

	// Set category options
	var catRes roomlayout.OptionList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListCategories, &roomlayout.ListOptionsArgs{TenantId: tenantID}, &catRes, func(err error) {
		if err == nil {
			var opts []fmt.KeyValue
			for _, opt := range catRes {
				opts = append(opts, fmt.KeyValue{Key: opt.Id, Value: opt.Label})
			}
			f.SetOptions("category_id", opts...)
		}
	})

	resultsDiv := dom.NewElement("div").Class("rl-free-results")

	searchBtn := dom.NewElement("button").Class("rl-btn-search").Text("Buscar")
	searchBtn.OnClick(func(_ dom.Event) {
		var freeRes roomlayout.FreeRoomList
		caller.Call(roomlayout.ModelName+"."+roomlayout.OpFindFreeRooms, args, &freeRes, func(err error) {
			if err != nil {
				return
			}
			innerDiv := dom.NewElement("div")
			for _, fr := range freeRes {
				header := dom.NewElement("div").Text(fmt.Sprintf("%s · %s (%s)", fr.Code, fr.Name, fr.FloorName))
				body := dom.NewElement("div").Text(fr.EquipmentLabels)

				assignBtn := dom.NewElement("button").Class("rl-btn-assign").Text("Asignar")
				footer := dom.NewElement("div").Child(assignBtn)

				card := &contentcard.ContentCard{
					Header: header,
					Body:   body,
					Footer: footer,
				}
				innerDiv.Child(card)
			}
			resultsDiv.Child(innerDiv)
		})
	})

	container.Child(f, searchBtn, resultsDiv)
	return container
}
