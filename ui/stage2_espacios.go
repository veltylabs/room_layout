package ui

import (
	"webtyp.com/components/stepindicator"
	"webtyp.com/components/themetoggle"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
)

type Stage2Espacios struct {
	dom.Element
	Root *RootView

	IsDrafting   bool
	DraftFloorID string
	DraftName    string
	DraftType    string // "box", "shared", "store"
	DraftCells   []string

	isDragging  bool
	dragMode    bool
	sliderNodes *dom.SignalNodes
}

func (s *Stage2Espacios) Render() *dom.Element {
	root := s.Root

	if s.DraftType == "" {
		s.DraftType = RoomTypeBox
	}
	if s.DraftFloorID == "" && len(root.Floors) > 0 {
		s.DraftFloorID = root.Floors[0].ID
	}
	if s.DraftName == "" {
		s.DraftName = "Box " + fmt.Sprint(len(root.Rooms)+101)
	}

	stepper := &stepindicator.StepIndicator{
		Steps: []stepindicator.Step{
			{Label: "Planta"},
			{Label: "Espacios"},
			{Label: "Artefactos"},
			{Label: "Operación"},
		},
		Active: 1,
		OnChange: func(idx int) {
			root.GoToStage(idx)
		},
	}

	header := html.Header().Class("rm-header").
		Child(
			html.Div().Class("rm-brand").
				Child(
					html.Div().Class("rm-brand-title").Text("Room Map"),
					html.Span().Class("rm-brand-sub").Text("Clínica Central"),
				),
			stepper,
			html.Div().Class("rm-header-actions").
				Child(
					&themetoggle.ThemeToggle{},
					html.Button().Class("btn").
						Child(html.Span().Text("Guardar")).
						OnClick(func(dom.Event) {
							root.SaveAll()
						}),
				),
		)

	// Legend
	legend := html.Div().Class("rm-legend").
		Child(
			html.Span().Class("rm-legend-item").
				Child(html.Span().Class("sw", "sw-box").Attr("style", "display:inline-block;width:14px;height:14px;border-radius:3px;background:var(--color-primary);margin-right:6px;"), html.Span().Text("Atención")),
			html.Span().Class("rm-legend-item").
				Child(html.Span().Class("sw", "sw-shared").Attr("style", "display:inline-block;width:14px;height:14px;border-radius:3px;background:var(--color-muted);margin-right:6px;"), html.Span().Text("Compartido")),
			html.Span().Class("rm-legend-item").
				Child(html.Span().Class("sw", "sw-store").Attr("style", "display:inline-block;width:14px;height:14px;border-radius:3px;background:var(--color-accent, #e8a33d);margin-right:6px;"), html.Span().Text("Soporte")),
			html.Span().Class("rm-legend-item").
				Child(html.Span().Class("sw", "sw-free").Attr("style", "display:inline-block;width:14px;height:14px;border-radius:3px;border:1px dashed var(--color-outline);background:var(--color-surface);margin-right:6px;"), html.Span().Text("Libre")),
		)

	topBar := html.Section().Class("rm-section-title").
		Child(
			html.Div().Class("rm-title-box").
				Child(
					html.Span().Class("rm-step-tag").Text("Paso 2 de 4"),
					html.H1().Text("Ubica los boxes y espacios compartidos"),
					html.P().Text("Elige la planta, ponle nombre al espacio y márcalo sobre las celdas habitables. Las celdas ya ocupadas por otro espacio no se pueden marcar."),
				),
			legend,
		)

	// Two-column layout: content & aside
	columns := html.Div().Class("rm-columns")

	// Left: Content (Floor Slider)
	content := html.Div().Class("rm-content")

	if s.IsDrafting {
		banner := html.Div().Class("rm-banner-draft").
			Child(
				html.Span().Class("rm-draft-title").
					BindTextFunc(func() string {
						_ = root.VerSig.Get()
						return "✎ Marcando «" + s.DraftName + "» (" + fmt.Sprint(len(s.DraftCells)) + " celdas seleccionadas)"
					}),
				html.Div().Class("rm-draft-actions").
					Child(
						html.Button().Class("btn", "btn-primary", "btn-sm").
							Text("✓ Guardar espacio").
							OnClick(func(dom.Event) {
								if len(s.DraftCells) > 0 {
									newRoom := RoomData{
										ID:       root.NewID(),
										FloorID:  s.DraftFloorID,
										Code:     "B-" + fmt.Sprint(len(root.Rooms)+101),
										Name:     s.DraftName,
										RoomType: s.DraftType,
										Cells:    s.DraftCells,
									}
									root.Rooms = append(root.Rooms, newRoom)
									s.IsDrafting = false
									s.DraftCells = nil
									s.DraftName = "Box " + fmt.Sprint(len(root.Rooms)+101)
									root.Refresh()
								}
							}),
						html.Button().Class("btn", "btn-sm").
							Text("✕ Cancelar").
							OnClick(func(dom.Event) {
								s.IsDrafting = false
								s.DraftCells = nil
								root.Refresh()
							}),
					),
			)
		content.Child(banner)
	}

	s.sliderNodes = dom.NewNodes(s.buildFloorCards()...)
	slider := html.Div().Class("rm-slider").BindChildren(s.sliderNodes)

	content.Child(slider)

	// Right: Aside (Form & List)
	aside := html.Aside().Class("rm-aside")

	// Panel 1: Nuevo Espacio
	newPanel := html.Div().Class("panel").
		Child(
			html.Span().Class("rm-panel-title").Text("Nuevo espacio"),
			html.Label().Class("field").Text("Planta").
				Child(
					dom.NewElement("select").Class("input").
						Child(func() []dom.Component {
							var opts []dom.Component
							for _, f := range root.Floors {
								opt := html.Option(f.ID, f.Name)
								if f.ID == s.DraftFloorID {
									opt.Attr("selected", "selected")
								}
								opts = append(opts, opt)
							}
							return opts
						}()...).
						OnChange(func(e dom.Event) {
							// Update target floor
						}),
				),
			html.Label().Class("field").Text("Nombre del espacio").
				Child(
					html.Input("text").Class("input").
						Attr("value", s.DraftName).
						OnChange(func(e dom.Event) {
							// Update name
						}),
				),
			html.Label().Class("field").Text("Tipo de espacio").
				Child(
					html.Div().Class("seg").
						Child(
							func() *dom.Element {
								btn := html.Button().Class("btn-seg").Text("Atención")
								if s.DraftType == RoomTypeBox {
									btn.Class("on")
								}
								return btn.OnClick(func(dom.Event) {
									s.DraftType = RoomTypeBox
									root.Refresh()
								})
							}(),
							func() *dom.Element {
								btn := html.Button().Class("btn-seg").Text("Compartido")
								if s.DraftType == RoomTypeShared {
									btn.Class("on")
								}
								return btn.OnClick(func(dom.Event) {
									s.DraftType = RoomTypeShared
									root.Refresh()
								})
							}(),
							func() *dom.Element {
								btn := html.Button().Class("btn-seg").Text("Soporte")
								if s.DraftType == RoomTypeStore {
									btn.Class("on")
								}
								return btn.OnClick(func(dom.Event) {
									s.DraftType = RoomTypeStore
									root.Refresh()
								})
							}(),
						),
				),
			html.Button().Class("btn", "btn-primary").
				BindTextFunc(func() string {
					_ = root.VerSig.Get()
					if s.IsDrafting {
						return "✓ Guardar espacio (" + fmt.Sprint(len(s.DraftCells)) + " celdas)"
					}
					return "✎ Marcar en la planta"
				}).
				OnClick(func(dom.Event) {
					if !s.IsDrafting {
						s.IsDrafting = true
						s.DraftCells = nil
						root.Refresh()
					} else if len(s.DraftCells) > 0 {
						newRoom := RoomData{
							ID:       root.NewID(),
							FloorID:  s.DraftFloorID,
							Code:     "B-" + fmt.Sprint(len(root.Rooms)+101),
							Name:     s.DraftName,
							RoomType: s.DraftType,
							Cells:    s.DraftCells,
						}
						root.Rooms = append(root.Rooms, newRoom)
						s.IsDrafting = false
						s.DraftCells = nil
						s.DraftName = "Box " + fmt.Sprint(len(root.Rooms)+101)
						root.Refresh()
					}
				}),
		)

	// Panel 2: Espacios creados
	listPanel := html.Div().Class("panel").
		Child(html.Span().Class("rm-panel-title").
			BindTextFunc(func() string {
				_ = root.VerSig.Get()
				return "Espacios (" + fmt.Sprint(len(root.Rooms)) + ")"
			}))

	for rIdx := range root.Rooms {
		rm := &root.Rooms[rIdx]
		idx := rIdx

		swCls := "sw-box"
		swColor := "var(--color-primary)"
		if rm.RoomType == RoomTypeShared {
			swCls = "sw-shared"
			swColor = "var(--color-muted)"
		} else if rm.RoomType == RoomTypeStore {
			swCls = "sw-store"
			swColor = "var(--color-accent, #e8a33d)"
		}

		item := html.Div().Class("room-item").
			Child(
				html.Span().Class("sw", swCls).
					Attr("style", fmt.Sprintf("display:inline-block;width:14px;height:14px;border-radius:3px;background:%s;margin-right:8px;flex:none;", swColor)),
				html.Div().Class("rm-room-info").
					Child(
						html.Span().Class("rm-room-name").Text(rm.Name),
						html.Span().Class("rm-room-range").Text(BoundingBoxLabel(rm.Cells)),
					),
				html.Button().Class("btn", "btn-sm", "btn-danger").Text("✕").
					OnClick(func(dom.Event) {
						var updated []RoomData
						for j, room := range root.Rooms {
							if j != idx {
								updated = append(updated, room)
							}
						}
						root.Rooms = updated
						root.Refresh()
					}),
			)
		listPanel.Child(item)
	}

	aside.Child(newPanel, listPanel)
	columns.Child(content, aside)

	// Footer
	footer := html.Footer().Class("rm-footer").
		Child(
			html.Button().Class("btn").Text("← Planta").
				OnClick(func(dom.Event) {
					root.GoToStage(0)
				}),
			html.Span().Class("rm-footer-summary").
				BindTextFunc(func() string {
					_ = root.VerSig.Get()
					totalOccupied := 0
					for _, rm := range root.Rooms {
						totalOccupied += len(rm.Cells)
					}
					return fmt.Sprint(len(root.Rooms)) + " espacios definidos · " + fmt.Sprint(totalOccupied) + " celdas ocupadas"
				}),
			html.Button().Class("btn", "btn-primary").
				Child(html.Span().Text("Siguiente: Artefactos →")).
				OnClick(func(dom.Event) {
					root.GoToStage(2)
				}),
		)

	layout := html.Div().Class("rm-stage").
		Child(header, topBar, columns, footer)

	return layout
}

func (s *Stage2Espacios) RebuildFloors() {
	if s.sliderNodes != nil {
		s.sliderNodes.Set(s.buildFloorCards())
	}
}

func (s *Stage2Espacios) buildFloorCards() []*dom.Element {
	root := s.Root
	var cards []*dom.Element

	for i := range root.Floors {
		fl := &root.Floors[i]
		fIdx := i
		isCurrentDraft := s.IsDrafting && fl.ID == s.DraftFloorID

		floorCard := html.Article().Class("floor")
		if s.IsDrafting && !isCurrentDraft {
			floorCard.Class("floor", "dim")
		}
		if isCurrentDraft {
			floorCard.Class("floor", "drawing")
		}

		fHeader := html.Div().Class("rm-floor-header").
			Child(
				html.Span().Class("rm-floor-badge").Text(fmt.Sprint(fl.Position)),
				html.Span().Class("name-input").Text(fl.Name),
				html.Span().Class("rm-floor-count").
					BindTextFunc(func() string {
						_ = root.VerSig.Get()
						return fmt.Sprint(len(fl.HabitableCells)) + " hab."
					}),
			)

		// Grid: renders only habitable cells with room colors
		gridEl := html.Div().Class("grid").
			Attr("style", fmt.Sprintf("display: grid; grid-template-columns: 24px repeat(%d, minmax(22px, 1fr)); gap: 2px; padding: 10px; background: var(--color-background); border: 1px solid var(--color-outline); border-radius: 12px; overflow-x: auto; touch-action: none; user-select: none; -webkit-user-select: none;", root.Cols)).
			OnPointerUp(func(dom.Event) {
				s.isDragging = false
			}).
			OnPointerLeave(func(dom.Event) {
				s.isDragging = false
			})

		// Column headers
		gridEl.Child(html.Span())
		for c := 0; c < root.Cols; c++ {
			gridEl.Child(html.Span().Class("lbl-col").
				Attr("style", "font-size: 11px; font-weight: 700; color: var(--color-muted); text-align: center; height: 20px; display: flex; align-items: center; justify-content: center;").
				Text(ColName(c)))
		}

		for r := 0; r < root.Rows; r++ {
			rowIdx := r
			gridEl.Child(html.Span().Class("lbl-row").
				Attr("style", "font-size: 11px; font-weight: 700; color: var(--color-muted); display: flex; align-items: center; justify-content: center;").
				Text(fmt.Sprint(rowIdx + 1)))

			for c := 0; c < root.Cols; c++ {
				colIdx := c
				coord := FormatCell(rowIdx, colIdx)

				cellBtn := html.Button().Class("cell").
					Attr("type", "button").
					Attr("title", coord)

				cellBtn.BindAttrFunc("style", func() string {
					_ = root.VerSig.Get()
					if !CellInList(fl.HabitableCells, coord) {
						return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1px dashed var(--color-outline); opacity: 0.12; pointer-events: none; border-radius: 4px;"
					}

					// Habitable cell
					if s.IsDrafting && isCurrentDraft && CellInList(s.DraftCells, coord) {
						return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 2px solid #ffffff; background: var(--color-primary); color: #ffffff; border-radius: 4px; cursor: pointer; box-shadow: 0 0 5px var(--color-primary); touch-action: none; user-select: none; -webkit-user-select: none;"
					}

					room := FindRoomByCell(root.Rooms, fl.ID, coord)
					if room != nil {
						switch room.RoomType {
						case RoomTypeShared:
							return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-muted); background: var(--color-muted); opacity: 0.75; color: #ffffff; border-radius: 4px; cursor: pointer; touch-action: none; user-select: none; -webkit-user-select: none;"
						case RoomTypeStore:
							return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-accent, #e8a33d); background: var(--color-accent, #e8a33d); opacity: 0.85; color: #ffffff; border-radius: 4px; cursor: pointer; touch-action: none; user-select: none; -webkit-user-select: none;"
						default:
							return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-primary); background: var(--color-primary); opacity: 0.85; color: #ffffff; border-radius: 4px; cursor: pointer; touch-action: none; user-select: none; -webkit-user-select: none;"
						}
					}

					// Free habitable cell
					return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px dashed var(--color-outline); background: var(--color-surface); cursor: pointer; border-radius: 4px; touch-action: none; user-select: none; -webkit-user-select: none;"
				})

				cellBtn.OnPointerDown(func(e dom.Event) {
					if s.IsDrafting && isCurrentDraft {
						e.ReleasePointerCapture()
						if otherRoom := FindRoomByCell(root.Rooms, fl.ID, coord); otherRoom != nil {
							return
						}
						isDrafted := CellInList(s.DraftCells, coord)
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
						if otherRoom := FindRoomByCell(root.Rooms, fl.ID, coord); otherRoom != nil {
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

				cellBtn.OnPointerUp(func(dom.Event) {
					s.isDragging = false
				})

				gridEl.Child(cellBtn)
			}
		}

		floorCard.Key("floor-" + fl.ID).
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
