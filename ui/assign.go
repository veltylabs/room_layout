package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/dom"
	"webtyp.com/html"
	"webtyp.com/router"
)

func containsID(list []string, id string) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}

func toggleID(list []string, id string, on bool) []string {
	if on {
		if !containsID(list, id) {
			return append(list, id)
		}
		return list
	}
	out := make([]string, 0, len(list))
	for _, v := range list {
		if v != id {
			out = append(out, v)
		}
	}
	return out
}

type assignTab struct {
	dom.Element
	caller   router.Caller
	tenantID string
	picker   *roomPicker
	boxes    *dom.SignalNodes
	msg      *dom.SignalString
	cats     roomlayout.OptionList
	eqs      roomlayout.EquipmentList
	catOn    []string
	eqOn     []string
}

func newAssignTab(caller router.Caller, tenantID string) *assignTab {
	t := &assignTab{
		Element:  *dom.NewElement("div"),
		caller:   caller,
		tenantID: tenantID,
		boxes:    dom.NewNodes(),
		msg:      dom.NewString(""),
	}
	t.picker = newRoomPicker(caller, tenantID, func(id string) {
		t.loadRoom()
	})
	return t
}

func (t *assignTab) Init(ctx dom.Ctx) {
	// Load categories and equipment once, then load picker
	var catRes roomlayout.OptionList
	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListCategories, &roomlayout.ListOptionsArgs{TenantId: t.tenantID}, &catRes, func(err error) {
		if err == nil {
			t.cats = catRes
		}
		var eqRes roomlayout.EquipmentList
		t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListEquipment, &roomlayout.ListEquipmentArgs{TenantId: t.tenantID}, &eqRes, func(err error) {
			if err == nil {
				t.eqs = eqRes
			}
			t.picker.load(t.loadRoom)
		})
	})
}

func (t *assignTab) loadRoom() {
	t.msg.Set("")
	roomID := t.picker.sel.Get()
	if roomID == "" {
		t.boxes.Set(nil)
		return
	}

	var roomCats roomlayout.OptionList
	var roomEq roomlayout.EquipmentList
	pending := 2

	renderBoxes := func() {
		pending--
		if pending > 0 {
			return
		}

		t.catOn = nil
		for _, c := range roomCats {
			t.catOn = append(t.catOn, c.Id)
		}

		t.eqOn = nil
		for _, e := range roomEq {
			t.eqOn = append(t.eqOn, e.Id)
		}

		// Build category fieldset
		catFieldset := html.Fieldset().Key("cat-fieldset").Child(html.Legend().Text("Áreas habilitadas")).Class("rl-cat-boxes")
		for _, cat := range t.cats {
			catID := cat.Id
			chk := html.Input("checkbox").
				Attr("data-category-id", catID).
				BindAttrBool("checked", dom.DeriveBool(func() bool {
					return containsID(t.catOn, catID)
				})).
				OnChange(func(e dom.Event) {
					t.catOn = toggleID(t.catOn, catID, e.TargetChecked())
				})
			catFieldset.Child(html.Label().Child(chk, html.Span().Text(" "+cat.Label)))
		}

		// Build equipment fieldset
		eqFieldset := html.Fieldset().Key("eq-fieldset").Child(html.Legend().Text("Equipamiento")).Class("rl-eq-boxes")
		for _, eq := range t.eqs {
			eqID := eq.Id
			chk := html.Input("checkbox").
				Attr("data-equipment-id", eqID).
				BindAttrBool("checked", dom.DeriveBool(func() bool {
					return containsID(t.eqOn, eqID)
				})).
				OnChange(func(e dom.Event) {
					t.eqOn = toggleID(t.eqOn, eqID, e.TargetChecked())
				})
			eqFieldset.Child(html.Label().Child(chk, html.Span().Text(" "+eq.Name)))
		}

		t.boxes.Set([]*dom.Element{catFieldset, eqFieldset})
	}

	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRoomCategories, &roomlayout.ListRoomCategoriesArgs{TenantId: t.tenantID, RoomId: roomID}, &roomCats, func(err error) {
		renderBoxes()
	})

	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRoomEquipment, &roomlayout.ListRoomEquipmentArgs{TenantId: t.tenantID, RoomId: roomID}, &roomEq, func(err error) {
		renderBoxes()
	})
}

func (t *assignTab) save() {
	roomID := t.picker.sel.Get()
	if roomID == "" {
		return
	}

	catRefs := make([]roomlayout.IdRef, len(t.catOn))
	for i, id := range t.catOn {
		catRefs[i] = roomlayout.IdRef{Id: id}
	}

	eqRefs := make([]roomlayout.IdRef, len(t.eqOn))
	for i, id := range t.eqOn {
		eqRefs[i] = roomlayout.IdRef{Id: id}
	}

	catArgs := roomlayout.SetRoomCategoriesArgs{
		TenantId:    t.tenantID,
		RoomId:      roomID,
		CategoryIds: catRefs,
	}

	t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpSetRoomCategories, &catArgs, nil, func(err error) {
		if err != nil {
			t.msg.Set(err.Error())
			return
		}

		eqArgs := roomlayout.SetRoomEquipmentArgs{
			TenantId:     t.tenantID,
			RoomId:       roomID,
			EquipmentIds: eqRefs,
		}
		t.caller.Call(roomlayout.ModelName+"."+roomlayout.OpSetRoomEquipment, &eqArgs, nil, func(err error) {
			if err != nil {
				t.msg.Set(err.Error())
				return
			}
			t.msg.Set("Guardado.")
		})
	})
}

func (t *assignTab) Render() *dom.Element {
	saveBtn := html.Button().Class("rl-btn-save").Text("Guardar")
	saveBtn.OnClick(func(_ dom.Event) {
		t.save()
	})

	msgDiv := html.Div().BindText(t.msg).Class("rl-assign-msg")

	return html.Div().Child(
		t.picker.render(),
		html.Div().BindChildren(t.boxes),
		saveBtn,
		msgDiv,
	).Class("rl-assign-tab")
}
