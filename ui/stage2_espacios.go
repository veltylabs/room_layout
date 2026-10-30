package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/components/modaldialog"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

type Stage2Espacios struct {
	dom.Element
	Root *RootView

	IsDrafting  bool
	started     bool
	EditingID   string // el espacio guardado que se está editando; "" = uno nuevo
	editingCode string
	pendingCode string // código del espacio nuevo ya enviado, mientras llega su id
	dirty       bool   // hubo cambios mientras llegaba ese id
	DraftName   string
	DraftType   string // "box", "shared", "store"
	DraftCells  []string

	isDragging  bool
	dragMode    bool
	sliderNodes *dom.SignalNodes
	formNodes   *dom.SignalNodes
	listNodes   *dom.SignalNodes

	confirm    *modaldialog.ModalDialog // «¿Eliminar «Psicología»?» antes de quitar un espacio
	deleteName *dom.SignalString
	deleteID   string

	outlineNodes *dom.SignalNodes // contornos y nombres del piso en pantalla
	outlineFloor string
}

func (s *Stage2Espacios) Render() *dom.Element {
	root := s.Root

	if s.DraftType == "" {
		s.DraftType = RoomTypeBox
	}
	if !s.started { // al entrar a «Espacios» ya se puede marcar: sin un paso previo escondido
		s.started = true
		s.IsDrafting = true
	}
	if s.DraftName == "" {
		s.DraftName = "Box " + fmt.Sprint(len(root.Rooms)+101)
	}

	header := root.header(1)

	legendItem := func(swatch string, label lang.Text) *dom.Element {
		return html.Span().Class(clsLegendItem).Child(html.Span().Class(swatch), html.Span().Text(lang.Translate(label).String()))
	}
	legend := html.Div().Class(clsLegend).Child(
		legendItem(clsSwatchBox, "Clinical"),
		legendItem(clsSwatchShared, "Shared"),
		legendItem(clsSwatchStore, "Support"),
		legendItem(clsSwatchFree, "Free"),
	)

	topBar := html.Section().Class(clsTitleBar).
		Child(
			html.Div().Class(clsTitle).
				Child(
					html.Span().Class(clsStepTag).Text(lang.Translate("Step", 2, "of", 4).String()),
					html.H2().Class(clsHeading).Text(lang.Translate("Rooms and shared spaces").String()),
					html.P().Text(lang.Translate("Draw a space on the floor plan: it is saved when you lift your finger. Drawing next to the selected space makes it bigger; drawing elsewhere starts a new one. Touch a space to select it.").String()),
				),
			legend,
		)

	s.sliderNodes = dom.NewNodes(s.buildFloorCards()...)
	content := html.Div().Class(clsContent).Child(html.Div().Class(clsSlider).BindChildren(s.sliderNodes))

	s.formNodes = dom.NewNodes(s.buildForm())
	s.listNodes = dom.NewNodes(s.buildList())
	aside := html.Aside().Class(clsAside).Child(
		html.Div().BindChildren(s.formNodes),
		html.Div().BindChildren(s.listNodes),
	)

	footer := html.Footer().Class(clsFooter).
		Child(
			html.Button().Class(clsButton).Text("← "+lang.Translate("Floor plan").String()).
				OnClick(func(dom.Event) {
					root.GoToStage(0)
				}),
			html.Span().Class(clsSummary).
				BindTextFunc(func() string {
					_ = root.VerSig.Get()
					return lang.Translate(len(root.Rooms), "spaces").String()
				}),
			html.Button().Class(clsPrimary).
				Child(html.Span().Text(lang.Translate("Next").String()+": "+lang.Translate("Equipment").String()+" →")).
				OnClick(func(dom.Event) {
					root.GoToStage(2)
				}),
		)

	return html.Div().Class(clsStage).
		Child(header, footer, topBar, html.Div().Class(clsColumns).Child(content, aside), s.deleteDialog())
}

// buildForm es el panel del espacio que se marca: uno nuevo, o uno guardado que se tocó para
// editarlo (en la planta o en la lista).
func (s *Stage2Espacios) buildForm() *dom.Element {
	title := lang.Translate("New space").String()
	if s.EditingID != "" {
		title = lang.Translate("Edit space").String()
	}
	typeBtn := func(roomType string, label lang.Text) *dom.Element {
		btn := html.Button().Class(clsSegItem).Attr("type", "button").Text(lang.Translate(label).String())
		if s.DraftType == roomType {
			btn.SetState(widget.Selected)
		}
		return btn.OnClick(func(dom.Event) {
			s.DraftType = roomType
			s.commit()
			s.refresh()
		})
	}
	panel := html.Div().Class(clsPanel).Child(
		html.H3().Class(clsPanelTitle).Text(title),
		html.Label().Class(clsField).Text(lang.Translate("Space name").String()).
			Child(html.Input("text").Class(clsInput).Attr("value", s.DraftName).
				OnChange(func(e dom.Event) {
					s.DraftName = e.TargetValue()
					s.commit()
				})),
		html.Div().Class(clsField).Text(lang.Translate("Space type").String()).
			Child(html.Div().Class(clsSeg).Child(
				typeBtn(RoomTypeBox, "Clinical"),
				typeBtn(RoomTypeShared, "Shared"),
				typeBtn(RoomTypeStore, "Support"),
			)),
	)
	if s.EditingID != "" || s.pendingCode != "" {
		panel.Child(html.Button().Class(clsButton).Attr("type", "button").Text("+ " + lang.Translate("New space").String()).
			OnClick(func(dom.Event) { s.startNew() }))
	}
	return panel
}

// buildList son los espacios guardados: tocar uno lo abre para editar; ✕ lo quita del mapa.
func (s *Stage2Espacios) buildList() *dom.Element {
	root := s.Root
	list := html.Div().Class(clsList)
	for i := range root.Rooms {
		rm := root.Rooms[i]
		item := html.Div().Class(clsItem).Attr("role", "button").
			Child(
				html.Span().Class(swatchClass(rm.RoomType)),
				html.Span().Class(clsItemName).Text(rm.Name),
				html.Button().Class(clsDanger).Attr("type", "button").Attr("title", lang.Translate("Delete").String()).Text("✕").
					OnClick(func(e dom.Event) {
						e.StopPropagation()
						s.askDelete(rm)
					}),
			).
			OnClick(func(dom.Event) { s.edit(rm) })
		if rm.ID == s.EditingID {
			item.SetState(widget.Selected)
		}
		list.Child(item)
	}
	return html.Div().Class(clsPanel).Child(
		html.H3().Class(clsPanelTitle).Text(lang.Translate("Spaces").String()+" ("+fmt.Sprint(len(root.Rooms))+")"),
		list,
	)
}

// edit abre un espacio guardado en el panel: sus celdas pasan a marcarse en color de selección.
func (s *Stage2Espacios) edit(rm RoomData) {
	root := s.Root
	s.EditingID, s.editingCode = rm.ID, rm.Code
	s.DraftName, s.DraftType = rm.Name, rm.RoomType
	s.DraftCells = append([]string(nil), rm.Cells...)
	s.IsDrafting = true
	for i, f := range root.Floors {
		if f.ID == rm.FloorID && root.ActiveFloor != i {
			root.ActiveFloor = i
			root.RebuildAllFloors()
		}
	}
	s.refresh()
}

// startNew deja el panel listo para un espacio nuevo.
func (s *Stage2Espacios) startNew() {
	s.EditingID, s.editingCode, s.pendingCode, s.dirty = "", "", "", false
	s.DraftName, s.DraftType, s.DraftCells = "Box "+fmt.Sprint(len(s.Root.Rooms)+101), RoomTypeBox, nil
	s.refresh()
}

// refresh redibuja el panel, la lista y las celdas.
func (s *Stage2Espacios) refresh() {
	if s.outlineNodes != nil {
		s.outlineNodes.Set(s.roomOutlines(s.outlineFloor))
	}
	if s.formNodes != nil {
		s.formNodes.Set([]*dom.Element{s.buildForm()})
		s.listNodes.Set([]*dom.Element{s.buildList()})
	}
	s.Root.Refresh()
}

func (s *Stage2Espacios) RebuildFloors() {
	s.resolvePending()
	if s.sliderNodes != nil {
		s.sliderNodes.Set(s.buildFloorCards())
		s.formNodes.Set([]*dom.Element{s.buildForm()})
		s.listNodes.Set([]*dom.Element{s.buildList()})
	}
}

func (s *Stage2Espacios) buildFloorCards() []*dom.Element {
	root := s.Root
	var cards []*dom.Element

	for _, i := range root.shownFloors() {
		fl := &root.Floors[i]
		fIdx := i
		isCurrentDraft := s.IsDrafting // solo se dibuja el piso activo

		floorCard := html.Article().Class(clsFloor)

		fHeader := html.Div().Class(clsFloorHeader).
			Child(
				html.Span().Class(clsFloorBadge).Text(fmt.Sprint(fl.Position)),
				html.Span().Class(clsFloorName).Text(fl.Name),
				root.floorNav(),
			)

		// Grid: renders only habitable cells with room colors
		gridEl := html.Div().Class("grid").
			Attr("style", fmt.Sprintf("display: grid; grid-template-columns: 24px repeat(%d, 28px); gap: 2px; padding: 10px; background: var(--color-background); border: 1px solid var(--color-outline); border-radius: 12px; overflow-x: auto; touch-action: none; user-select: none; -webkit-user-select: none;", root.Cols)).
			OnPointerUp(func(dom.Event) { s.endStroke() }).
			OnPointerLeave(func(dom.Event) { s.endStroke() })

		// Column headers
		gridEl.Child(html.Span())
		for c := 0; c < root.Cols; c++ {
			gridEl.Child(html.Span().Class("lbl-col").
				Attr("style", gridPos(0, c+1)+"font-size: 11px; font-weight: 700; color: var(--color-muted); text-align: center; height: 20px; display: flex; align-items: center; justify-content: center;").
				Text(fmt.Sprint(c + 1)))
		}

		for r := 0; r < root.Rows; r++ {
			rowIdx := r
			gridEl.Child(html.Span().Class("lbl-row").
				Attr("style", gridPos(rowIdx+1, 0)+"font-size: 11px; font-weight: 700; color: var(--color-muted); display: flex; align-items: center; justify-content: center;").
				Text(fmt.Sprint(rowIdx + 1)))

			for c := 0; c < root.Cols; c++ {
				colIdx := c
				coord := FormatCell(rowIdx, colIdx)

				cellBtn := html.Button().Class("cell").
					Attr("type", "button").
					Attr("title", coord)

				cellBtn.BindAttrFunc("style", func() string {
					_ = root.VerSig.Get()
					return gridPos(rowIdx+1, colIdx+1) + s.cellStyle(fl, coord)
				})

				cellBtn.OnPointerDown(func(e dom.Event) {
					if s.IsDrafting && isCurrentDraft {
						e.ReleasePointerCapture()
						if other := s.otherRoom(fl.ID, coord); other != nil {
							s.edit(*other) // tocar un espacio guardado lo abre para editarlo
							return
						}
						isDrafted := CellInList(s.DraftCells, coord)
						if !isDrafted && !touches(s.DraftCells, coord) {
							s.startNew() // dibujar lejos del espacio seleccionado empieza uno nuevo
						}
						s.dragMode = !isDrafted
						s.isDragging = true
						if s.dragMode {
							s.DraftCells = AddCell(s.DraftCells, coord)
						} else {
							s.DraftCells = RemoveCell(s.DraftCells, coord)
						}
						root.Refresh()
					}
				})

				cellBtn.OnPointerEnter(func(e dom.Event) {
					if s.isDragging && s.IsDrafting && isCurrentDraft {
						if s.otherRoom(fl.ID, coord) != nil {
							return
						}
						if s.dragMode {
							s.DraftCells = AddCell(s.DraftCells, coord)
						} else {
							s.DraftCells = RemoveCell(s.DraftCells, coord)
						}
						root.Refresh()
					}
				})

				cellBtn.OnPointerUp(func(dom.Event) { s.endStroke() })

				gridEl.Child(cellBtn)
			}
		}

		if s.outlineNodes == nil {
			s.outlineNodes = dom.NewNodes()
		}
		s.outlineFloor = fl.ID
		s.outlineNodes.Set(s.roomOutlines(fl.ID))
		gridEl.Child(html.Div().Attr("style", "display: contents;").BindChildren(s.outlineNodes))

		floorCard.Key(fmt.Sprintf("floor-s2-%s-%dx%d", fl.ID, root.Cols, root.Rows)).
			BindClassFunc("active", func() bool {
				_ = root.VerSig.Get()
				return root.ActiveFloor == fIdx
			}).
			BindAttrFunc("style", func() string {
				_ = root.VerSig.Get()
				if root.ActiveFloor == fIdx {
					return "display: flex; flex-direction: column; gap: 8px;"
				}
				return "display: none;"
			}).
			Child(fHeader, gridEl)

		cards = append(cards, floorCard)
	}

	return cards
}

// endStroke cierra un trazo (soltar el dedo o el mouse) y lo guarda: no hay botón «Guardar».
func (s *Stage2Espacios) endStroke() {
	if !s.isDragging {
		return
	}
	s.isDragging = false
	s.commit()
}

// commit guarda en el momento el espacio del panel (nuevo o editado). Un espacio nuevo sin celdas
// todavía no existe; con su primer trazo se crea, y mientras el servidor no devuelve su id los
// cambios esperan (dirty) para no crearlo dos veces.
func (s *Stage2Espacios) commit() {
	root := s.Root
	floors := root.shownFloors()
	if len(s.DraftCells) == 0 || len(floors) == 0 {
		return
	}
	if s.EditingID == "" && s.pendingCode != "" {
		s.dirty = true
		return
	}
	code := s.editingCode
	if code == "" {
		code = s.newCode()
		s.pendingCode = code
	}
	root.saveRoom(RoomData{
		ID:       s.EditingID,
		FloorID:  root.Floors[floors[0]].ID,
		Code:     code,
		Name:     s.DraftName,
		RoomType: s.DraftType,
		Cells:    s.DraftCells,
	})
}

// resolvePending: tras releer, el espacio recién creado ya tiene id y queda en edición.
func (s *Stage2Espacios) resolvePending() {
	if s.pendingCode == "" {
		return
	}
	for _, rm := range s.Root.Rooms {
		if rm.Code == s.pendingCode {
			s.EditingID, s.editingCode, s.pendingCode = rm.ID, rm.Code, ""
			if s.dirty {
				s.dirty = false
				s.commit()
			}
			return
		}
	}
}

// newCode es un código que ningún espacio del mapa usa todavía.
func (s *Stage2Espacios) newCode() string {
	for n := len(s.Root.Rooms) + 101; ; n++ {
		code := "B-" + fmt.Sprint(n)
		taken := false
		for _, rm := range s.Root.Rooms {
			taken = taken || rm.Code == code
		}
		if !taken {
			return code
		}
	}
}

// otherRoom es el espacio guardado en esa celda, salvo el que se está editando.
func (s *Stage2Espacios) otherRoom(floorID, coord string) *RoomData {
	rm := FindRoomByCell(s.Root.Rooms, floorID, coord)
	if rm == nil || rm.ID == s.EditingID {
		return nil
	}
	return rm
}

// Estilo de cada celda: el espacio seleccionado va en tono ámbar suave (su contorno, en ámbar);
// los demás, en el tono de su tipo; lo libre, punteado.
const (
	cellBase     = "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border-radius: 4px; cursor: pointer; touch-action: none; user-select: none; -webkit-user-select: none; "
	cellOff      = "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1px dashed var(--color-outline); opacity: 0.12; pointer-events: none; border-radius: 4px;"
	cellDraft    = cellBase + "border: 1.5px solid transparent; background: var(--color-accent-wash);"
	cellBox      = cellBase + "border: 1.5px solid transparent; background: var(--color-selection);"
	cellShared   = cellBase + "border: 1.5px solid transparent; background: var(--color-surface-sunken);"
	cellStore    = cellBase + "border: 1.5px solid transparent; background: var(--color-accent-wash);"
	cellFreeCell = cellBase + "border: 1.5px dashed var(--color-outline); background: var(--color-surface);"
)

func (s *Stage2Espacios) cellStyle(fl *FloorData, coord string) string {
	if !CellInList(fl.HabitableCells, coord) {
		return cellOff
	}
	if s.IsDrafting && CellInList(s.DraftCells, coord) {
		return cellDraft
	}
	if rm := s.otherRoom(fl.ID, coord); rm != nil {
		switch rm.RoomType {
		case RoomTypeShared:
			return cellShared
		case RoomTypeStore:
			return cellStore
		}
		return cellBox
	}
	return cellFreeCell
}

// roomOutlines son los contornos de «Espacios»: todos menos el seleccionado, que se dibuja aparte
// desde lo que se está marcando.
func (s *Stage2Espacios) roomOutlines(floorID string) []*dom.Element {
	return append(s.Root.roomOutlines(floorID, s.EditingID, numbered, nameCenter), s.selectedOutline()...)
}

// selectedOutline es el contorno ámbar del espacio seleccionado, con su nombre.
func (s *Stage2Espacios) selectedOutline() []*dom.Element {
	if s.EditingID == "" && s.pendingCode == "" {
		return nil
	}
	cells := make([]roomlayout.Cell, 0, len(s.DraftCells))
	for _, c := range s.DraftCells {
		if cell, ok := roomlayout.ParseCell(c); ok {
			cells = append(cells, cell)
		}
	}
	rects := roomlayout.Rects(cells)
	var out []*dom.Element
	for i, r := range rects {
		el := html.Div().Attr("aria-hidden", "true").
			Key(fmt.Sprintf("selected-%d-%s-%d,%d,%d,%d", i, s.DraftName, r.Col, r.Row, r.W, r.H)).
			Attr("style", fmt.Sprintf("grid-row: %d / span %d; grid-column: %d / span %d; ", r.Row+1, r.H, r.Col+1, r.W)+roomOutline+numbered.wall()+selectedBorder)
		if i == 0 {
			el.Text(s.DraftName)
		}
		out = append(out, el)
	}
	return out
}

// touches dice si la celda está junto (arriba, abajo o a un lado) a alguna de las celdas.
func touches(cells []string, coord string) bool {
	c, ok := roomlayout.ParseCell(coord)
	if !ok {
		return false
	}
	for _, s := range cells {
		o, ok := roomlayout.ParseCell(s)
		if ok && (o.Col == c.Col && (o.Row == c.Row-1 || o.Row == c.Row+1) || o.Row == c.Row && (o.Col == c.Col-1 || o.Col == c.Col+1)) {
			return true
		}
	}
	return false
}

const selectedBorder = " z-index: 2; border-color: var(--color-accent);" // encima de sus vecinos: su pared se ve ámbar

// deleteDialog es la confirmación antes de quitar un espacio del mapa: nada se quita con un solo
// toque. Es el mismo diálogo que usan las listas de la plataforma. Se arma de nuevo en cada dibujo
// de la etapa: un elemento tiene un solo padre, y volver a «Espacios» la dibuja otra vez.
func (s *Stage2Espacios) deleteDialog() *dom.Element {
	if s.deleteName == nil {
		s.deleteName = dom.NewString("")
	}
	msg := html.P().BindTextFunc(func() string {
		return lang.Translate("Delete").String() + " «" + s.deleteName.Get() + "»?"
	})
	cancel := html.Button().Class(clsButton).Attr("type", "button").Text(lang.Translate("Cancel").String()).
		OnClick(func(dom.Event) { s.confirm.Close() })
	ok := html.Button().Class(clsDelete).Attr("type", "button").Text(lang.Translate("Delete").String()).
		OnClick(func(dom.Event) {
			s.confirm.Close()
			if s.EditingID == s.deleteID {
				s.startNew()
			}
			s.Root.deactivateRoom(s.deleteID)
		})
	s.confirm = &modaldialog.ModalDialog{
		Title:     "Confirm",
		HideClose: true,
		Content:   html.Div().Class(clsPanel).Child(msg, html.Div().Class(clsActions).Child(cancel, ok)),
	}
	return html.Div().Child(s.confirm)
}

// askDelete pide confirmación para quitar un espacio.
func (s *Stage2Espacios) askDelete(rm RoomData) {
	s.deleteID = rm.ID
	s.deleteName.Set(rm.Name)
	s.confirm.Open()
}
