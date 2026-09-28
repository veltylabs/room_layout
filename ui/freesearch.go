package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/components/contentcard"
	"webtyp.com/dom"
	"webtyp.com/html"
	"webtyp.com/router"
)

type freeSearchTab struct {
	dom.Element
	caller    router.Caller
	tenantID  string
	date      *dom.SignalString
	start     *dom.SignalString
	end       *dom.SignalString
	category  *dom.SignalString
	shiftID   *dom.SignalString
	msg       *dom.SignalString
	catOpts   *dom.SignalNodes
	shiftOpts *dom.SignalNodes
	results   *dom.SignalNodes
	dayShifts []*roomlayout.DayShift
}

func newFreeSearchTab(caller router.Caller, tenantID string) *freeSearchTab {
	return &freeSearchTab{
		Element:   *dom.NewElement("div"),
		caller:    caller,
		tenantID:  tenantID,
		date:      dom.NewString(""),
		start:     dom.NewString(""),
		end:       dom.NewString(""),
		category:  dom.NewString(""),
		shiftID:   dom.NewString(""),
		msg:       dom.NewString(""),
		catOpts:   dom.NewNodes(),
		shiftOpts: dom.NewNodes(),
		results:   dom.NewNodes(),
	}
}

func (t *freeSearchTab) Init(ctx dom.Ctx) {
	// Populate category options
	var catRes roomlayout.OptionList
	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListCategories, &roomlayout.ListOptionsArgs{TenantId: t.tenantID}, &catRes, func(err error) {
		if err == nil {
			nodes := []*dom.Element{html.Option("", "Cualquier área").Key("cat-any")}
			for _, opt := range catRes {
				nodes = append(nodes, html.Option(opt.Id, opt.Label).Key("cat-"+opt.Id))
			}
			t.catOpts.Set(nodes)
		}
	})

	t.shiftOpts.Set([]*dom.Element{html.Option("", "— ninguno —").Key("shift-none")})
}

func (t *freeSearchTab) loadDayShifts() {
	d := t.date.Get()
	if d == "" {
		t.dayShifts = nil
		t.shiftID.Set("")
		t.shiftOpts.Set([]*dom.Element{html.Option("", "— ninguno —").Key("shift-none")})
		return
	}

	args := roomlayout.ListRoomDayArgs{
		TenantId: t.tenantID,
		Date:     d,
	}
	var shiftsRes roomlayout.DayShiftList
	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRoomDay, &args, &shiftsRes, func(err error) {
		if err != nil {
			t.dayShifts = nil
			t.shiftID.Set("")
			t.shiftOpts.Set([]*dom.Element{html.Option("", "— ninguno —").Key("shift-none")})
			return
		}

		t.dayShifts = shiftsRes
		nodes := []*dom.Element{html.Option("", "— ninguno —").Key("shift-none")}
		for _, s := range shiftsRes {
			label := s.RoomCode + " · " + s.OccupantLabel + " " + s.Start + "–" + s.End
			nodes = append(nodes, html.Option(s.ShiftId, label).Key("shift-"+s.ShiftId))
		}
		t.shiftOpts.Set(nodes)
	})
}

func (t *freeSearchTab) onShiftSelected(sID string) {
	t.shiftID.Set(sID)
	if sID == "" {
		return
	}
	for _, s := range t.dayShifts {
		if s.ShiftId == sID {
			t.start.Set(s.Start)
			t.end.Set(s.End)
			t.category.Set(s.CategoryId)
			break
		}
	}
}

func (t *freeSearchTab) search() {
	d := t.date.Get()
	st := t.start.Get()
	en := t.end.Get()

	if d == "" || st == "" || en == "" {
		t.msg.Set("Indique fecha, desde y hasta.")
		t.results.Set(nil)
		return
	}

	t.msg.Set("")
	args := roomlayout.FindFreeRoomsArgs{
		TenantId:   t.tenantID,
		Date:       d,
		Start:      st,
		End:        en,
		CategoryId: t.category.Get(),
	}

	var freeRes roomlayout.FreeRoomList
	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpFindFreeRooms, &args, &freeRes, func(err error) {
		if err != nil {
			t.msg.Set(err.Error())
			t.results.Set(nil)
			return
		}

		if len(freeRes) == 0 {
			t.results.Set([]*dom.Element{
				html.Div().Key("empty").Text("No hay espacios libres."),
			})
			return
		}

		chosenShiftID := t.shiftID.Get()
		var cardNodes []*dom.Element
		for _, fr := range freeRes {
			room := fr
			headerText := room.Code + " · " + room.Name + " (" + room.FloorName + ")"
			header := html.Div().Text(headerText)
			body := html.Div().Text(room.EquipmentLabels)

			card := html.Div().
				Key("free-" + room.RoomId).
				Child(
					&contentcard.ContentCard{
						Header: header,
						Body:   body,
					},
				).Attr("data-free-room", room.RoomId)

			if chosenShiftID != "" {
				moveBtn := html.Button().Class("rl-btn-move").Text("Mover aquí")
				moveBtn.OnClick(func(_ dom.Event) {
					moveArgs := roomlayout.MoveShiftOccurrenceArgs{
						TenantId:     t.tenantID,
						ShiftId:      chosenShiftID,
						Date:         d,
						TargetRoomId: room.RoomId,
					}
					var movedShift roomlayout.RoomShift
					t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpMoveShiftOccurrence, &moveArgs, &movedShift, func(err error) {
						if err != nil {
							t.msg.Set(err.Error())
							return
						}
						t.msg.Set("Turno movido a " + room.Code)
						t.search()
					})
				})
				footer := html.Div().Child(moveBtn)
				// Re-create card with footer
				card = html.Div().
					Key("free-" + room.RoomId).
					Child(
						&contentcard.ContentCard{
							Header: header,
							Body:   body,
							Footer: footer,
						},
					).Attr("data-free-room", room.RoomId)
			}

			cardNodes = append(cardNodes, card)
		}

		t.results.Set(cardNodes)
	})
}

func (t *freeSearchTab) Render() *dom.Element {
	dateInput := html.Input("date").Bind(t.date)
	dateInput.OnChange(func(e dom.Event) {
		t.date.Set(e.TargetValue())
		t.loadDayShifts()
	})

	startInput := html.Input("time").Bind(t.start)
	startInput.OnChange(func(e dom.Event) {
		t.start.Set(e.TargetValue())
	})

	endInput := html.Input("time").Bind(t.end)
	endInput.OnChange(func(e dom.Event) {
		t.end.Set(e.TargetValue())
	})

	catSelect := dom.NewElement("select").
		BindChildren(t.catOpts).
		Bind(t.category).
		OnChange(func(e dom.Event) {
			t.category.Set(e.TargetValue())
		})

	shiftSelect := dom.NewElement("select").
		BindChildren(t.shiftOpts).
		Bind(t.shiftID).
		OnChange(func(e dom.Event) {
			t.onShiftSelected(e.TargetValue())
		})

	searchBtn := html.Button().Class("rl-btn-search").Text("Buscar")
	searchBtn.OnClick(func(_ dom.Event) {
		t.search()
	})

	formDiv := html.Div().Child(
		html.Div().Child(html.Label().Text("Fecha"), dateInput),
		html.Div().Child(html.Label().Text("Desde"), startInput),
		html.Div().Child(html.Label().Text("Hasta"), endInput),
		html.Div().Child(html.Label().Text("Área"), catSelect),
		html.Div().Child(html.Label().Text("Turno a mover"), shiftSelect),
		searchBtn,
	).Class("rl-free-form")

	msgDiv := html.Div().BindText(t.msg).Class("rl-free-msg")

	return html.Div().Child(
		formDiv,
		msgDiv,
		html.Div().BindChildren(t.results).Class("rl-free-results"),
	).Class("rl-free-tab")
}
