package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/components/contentcard"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/router"
)

type floorGroup struct {
	name  string
	rooms []*roomlayout.BoardRoom
}

func findOrCreateGroup(groups *[]floorGroup, name string) int {
	for i, g := range *groups {
		if g.name == name {
			return i
		}
	}
	*groups = append(*groups, floorGroup{name: name})
	return len(*groups) - 1
}

func buildBoardTab(caller router.Caller, tenantID string) *dom.Element {
	container := dom.NewElement("div").Class("rl-board-tab")

	catSelect := dom.NewElement("select").Class("rl-filter-cat")
	catSelect.Child(dom.NewElement("option").Attr("value", "").Text("Todas las categorías"))

	eqSelect := dom.NewElement("select").Class("rl-filter-eq")
	eqSelect.Child(dom.NewElement("option").Attr("value", "").Text("Todo el equipamiento"))

	// Populate categories
	var catRes roomlayout.OptionList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListCategories, &roomlayout.ListOptionsArgs{TenantId: tenantID}, &catRes, func(err error) {
		if err == nil {
			for _, opt := range catRes {
				catSelect.Child(dom.NewElement("option").Attr("value", opt.Id).Text(opt.Label))
			}
		}
	})

	// Populate equipment
	var eqRes roomlayout.EquipmentList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListEquipment, &roomlayout.ListEquipmentArgs{TenantId: tenantID}, &eqRes, func(err error) {
		if err == nil {
			for _, eq := range eqRes {
				eqSelect.Child(dom.NewElement("option").Attr("value", eq.Id).Text(eq.Name))
			}
		}
	})

	boardDeck := dom.NewElement("div").Class("rl-board-deck")

	reloadBoard := func() {
		catVal := catSelect.GetID()
		eqVal := eqSelect.GetID()
		args := roomlayout.ListBoardArgs{
			TenantId:    tenantID,
			CategoryId:  catVal,
			EquipmentId: eqVal,
		}
		var boardRes roomlayout.BoardRoomList
		caller.Call(roomlayout.ModelName+"."+roomlayout.OpListBoard, &args, &boardRes, func(err error) {
			if err != nil {
				return
			}
			// Group by floor_name using slice of floorGroup
			var groups []floorGroup
			for _, br := range boardRes {
				idx := findOrCreateGroup(&groups, br.FloorName)
				groups[idx].rooms = append(groups[idx].rooms, br)
			}

			innerDiv := dom.NewElement("div")
			for _, fg := range groups {
				fHeader := dom.NewElement("h2").Class("rl-floor-header").Text(fg.name)
				grid := dom.NewElement("div").Class("rl-room-grid")

				for _, br := range fg.rooms {
					header := dom.NewElement("div").Text(fmt.Sprintf("%s · %s", br.Code, br.Name))

					bodyText := "Libre"
					if br.Status == roomlayout.BoardStatusBusy {
						bodyText = fmt.Sprintf("Ocupado — %s hasta %s", br.CurrentOccupantLabel, br.CurrentUntil)
					} else if br.Status == roomlayout.BoardStatusClosed {
						bodyText = "Cerrado"
					}

					body := dom.NewElement("div").Child(
						dom.NewElement("div").Class("rl-status-"+br.Status).Text(bodyText),
					)
					if br.NextStart != "" {
						body.Child(dom.NewElement("div").Class("rl-next").Text(fmt.Sprintf("Próximo: %s %s", br.NextStart, br.NextOccupantLabel)))
					}

					footer := dom.NewElement("div").Child(
						dom.NewElement("div").Text(fmt.Sprintf("%s | %s", br.CategoryLabels, br.EquipmentLabels)),
					)

					card := &contentcard.ContentCard{
						Header: header,
						Body:   body,
						Footer: footer,
					}
					grid.Child(card)
				}

				innerDiv.Child(fHeader, grid)
			}
			boardDeck.Child(innerDiv)
		})
	}

	catSelect.OnChange(func(_ dom.Event) { reloadBoard() })
	eqSelect.OnChange(func(_ dom.Event) { reloadBoard() })

	controls := dom.NewElement("div").Class("rl-board-controls").Child(catSelect, eqSelect)
	container.Child(controls, boardDeck)

	reloadBoard()
	return container
}
