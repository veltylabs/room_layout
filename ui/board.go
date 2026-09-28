package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/components/contentcard"
	"webtyp.com/dom"
	"webtyp.com/html"
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

type boardTab struct {
	dom.Element
	caller    router.Caller
	tenantID  string
	category  *dom.SignalString
	equipment *dom.SignalString
	catOpts   *dom.SignalNodes
	eqOpts    *dom.SignalNodes
	cards     *dom.SignalNodes
}

func newBoardTab(caller router.Caller, tenantID string) *boardTab {
	return &boardTab{
		Element:   *dom.NewElement("div"),
		caller:    caller,
		tenantID:  tenantID,
		category:  dom.NewString(""),
		equipment: dom.NewString(""),
		catOpts:   dom.NewNodes(),
		eqOpts:    dom.NewNodes(),
		cards:     dom.NewNodes(),
	}
}

func (t *boardTab) Init(ctx dom.Ctx) {
	// Populate categories
	var catRes roomlayout.OptionList
	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListCategories, &roomlayout.ListOptionsArgs{TenantId: t.tenantID}, &catRes, func(err error) {
		if err == nil {
			nodes := []*dom.Element{html.Option("", "Todas las áreas").Key("cat-all")}
			for _, opt := range catRes {
				nodes = append(nodes, html.Option(opt.Id, opt.Label).Key("cat-"+opt.Id))
			}
			t.catOpts.Set(nodes)
		}
	})

	// Populate equipment
	var eqRes roomlayout.EquipmentList
	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListEquipment, &roomlayout.ListEquipmentArgs{TenantId: t.tenantID}, &eqRes, func(err error) {
		if err == nil {
			nodes := []*dom.Element{html.Option("", "Todo el equipamiento").Key("eq-all")}
			for _, eq := range eqRes {
				nodes = append(nodes, html.Option(eq.Id, eq.Name).Key("eq-"+eq.Id))
			}
			t.eqOpts.Set(nodes)
		}
	})

	t.reload()
}

func (t *boardTab) reload() {
	args := roomlayout.ListBoardArgs{
		TenantId:    t.tenantID,
		CategoryId:  t.category.Get(),
		EquipmentId: t.equipment.Get(),
	}
	var boardRes roomlayout.BoardRoomList
	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListBoard, &args, &boardRes, func(err error) {
		if err != nil {
			t.cards.Set([]*dom.Element{
				html.Div().Key("err").Text("No se pudo cargar el tablero: " + err.Error()),
			})
			return
		}

		if len(boardRes) == 0 {
			t.cards.Set([]*dom.Element{
				html.Div().Key("empty").Text("No hay espacios que coincidan."),
			})
			return
		}

		var groups []floorGroup
		for _, br := range boardRes {
			idx := findOrCreateGroup(&groups, br.FloorName)
			groups[idx].rooms = append(groups[idx].rooms, br)
		}

		var nodes []*dom.Element
		for _, fg := range groups {
			fHeader := html.H2().Key("header-" + fg.name).Text(fg.name).Class("rl-floor-header")
			grid := html.Div().Key("grid-" + fg.name).Class("rl-room-grid")

			for _, br := range fg.rooms {
				headerText := br.Code + " · " + br.Name
				header := html.Div().Text(headerText)

				bodyText := "Libre"
				if br.Status == roomlayout.BoardStatusBusy {
					bodyText = "Ocupado — " + br.CurrentOccupantLabel + " hasta " + br.CurrentUntil
				} else if br.Status == roomlayout.BoardStatusClosed {
					bodyText = "Cerrado"
				}

				body := html.Div().Child(
					html.Div().Class("rl-status-" + br.Status).Text(bodyText),
				)
				if br.NextStart != "" {
					body.Child(html.Div().Class("rl-next").Text("Próximo: " + br.NextStart + " " + br.NextOccupantLabel))
				}

				footerText := br.CategoryLabels + " | " + br.EquipmentLabels
				footer := html.Div().Child(
					html.Div().Text(footerText),
				)

				card := html.Div().
					Key("board-room-" + br.RoomId).
					Child(
						&contentcard.ContentCard{
							Header: header,
							Body:   body,
							Footer: footer,
						},
					).
					Attr("data-board-room", br.RoomId).
					Attr("data-status", br.Status)

				grid.Child(card)
			}

			nodes = append(nodes, fHeader, grid)
		}

		t.cards.Set(nodes)
	})
}

func (t *boardTab) Render() *dom.Element {
	controls := html.Div().Child(
		dom.NewElement("select").
			BindChildren(t.catOpts).
			Attr("data-board-filter", "category").
			OnChange(func(e dom.Event) {
				t.category.Set(e.TargetValue())
				t.reload()
			}),
		dom.NewElement("select").
			BindChildren(t.eqOpts).
			Attr("data-board-filter", "equipment").
			OnChange(func(e dom.Event) {
				t.equipment.Set(e.TargetValue())
				t.reload()
			}),
	).Class("rl-board-controls")

	return html.Div().Child(
		controls,
		html.Div().BindChildren(t.cards),
	).Class("rl-board-tab")
}
