package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/dom"
	"webtyp.com/html"
	"webtyp.com/router"
)

type roomPicker struct {
	caller   router.Caller
	tenantID string
	sel      *dom.SignalString
	opts     *dom.SignalNodes
	rooms    []*roomlayout.Room
	onChange func(id string)
}

func newRoomPicker(caller router.Caller, tenantID string, onChange func(id string)) *roomPicker {
	return &roomPicker{
		caller:   caller,
		tenantID: tenantID,
		sel:      dom.NewString(""),
		opts:     dom.NewNodes(),
		onChange: onChange,
	}
}

func (p *roomPicker) load(then func()) {
	var roomsRes roomlayout.RoomList
	p.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRooms, &roomlayout.ListRoomsArgs{TenantId: p.tenantID, ActiveOnly: true}, &roomsRes, func(err error) {
		if err != nil || len(roomsRes) == 0 {
			p.rooms = nil
			p.sel.Set("")
			p.opts.Set([]*dom.Element{
				html.Option("", "No hay espacios activos").Key("empty").Attr("disabled", "disabled"),
			})
			if then != nil {
				then()
			}
			return
		}

		p.rooms = roomsRes
		current := p.sel.Get()
		var found bool
		for _, r := range roomsRes {
			if r.Id == current {
				found = true
				break
			}
		}
		if !found {
			current = roomsRes[0].Id
			p.sel.Set(current)
		}

		var nodes []*dom.Element
		for _, r := range roomsRes {
			label := r.Code + " · " + r.Name
			if r.Id == current {
				nodes = append(nodes, html.SelectedOption(r.Id, label).Key(r.Id))
			} else {
				nodes = append(nodes, html.Option(r.Id, label).Key(r.Id))
			}
		}
		p.opts.Set(nodes)

		if then != nil {
			then()
		}
	})
}

func (p *roomPicker) render() *dom.Element {
	return html.Div().Child(
		html.Label().Text("Espacio"),
		dom.NewElement("select").
			BindChildren(p.opts).
			OnChange(func(e dom.Event) {
				id := e.TargetValue()
				p.sel.Set(id)
				if p.onChange != nil {
					p.onChange(id)
				}
			}),
	)
}
