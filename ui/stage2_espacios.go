package ui

import (
	"webtyp.com/components/stepindicator"
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
			html.Button().Class("btn").
				Child(html.Span().Text("Guardar")).
				OnClick(func(dom.Event) {
					root.SaveAll()
				}),
		)

	// Legend
	legend := html.Div().Class("rm-legend").
		Child(
			html.Span().Class("rm-legend-item").
				Child(html.Span().Class("sw", "sw-box"), html.Span().Text("Atención")),
			html.Span().Class("rm-legend-item").
				Child(html.Span().Class("sw", "sw-shared"), html.Span().Text("Compartido")),
			html.Span().Class("rm-legend-item").
				Child(html.Span().Class("sw", "sw-store"), html.Span().Text("Soporte")),
			html.Span().Class("rm-legend-item").
				Child(html.Span().Class("sw", "sw-free"), html.Span().Text("Libre")),
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
					Text("✎ Marcando «"+s.DraftName+"» ("+fmt.Sprint(len(s.DraftCells))+" celdas seleccionadas)"),
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

	slider := html.Div().Class("rm-slider")

	for i := range root.Floors {
		fl := &root.Floors[i]
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
				html.Span().Class("rm-floor-count").Text(fmt.Sprint(len(fl.HabitableCells))+" hab."),
			)

		// Grid: renders only habitable cells with room colors
		gridEl := html.Div().Class("grid").
			Attr("style", fmt.Sprintf("display: grid; grid-template-columns: repeat(%d, 1fr); grid-template-rows: repeat(%d, 1fr);", root.Cols, root.Rows))

		for r := 0; r < root.Rows; r++ {
			for c := 0; c < root.Cols; c++ {
				coord := FormatCell(r, c)
				cellEl := html.Div().Class("cell").
					Attr("style", fmt.Sprintf("grid-row: %d; grid-column: %d;", r+1, c+1))

				if !CellInList(fl.HabitableCells, coord) {
					// Non-habitable cell: invisible
					cellEl.Class("cell", "non-habitable")
					gridEl.Child(cellEl)
					continue
				}

				// Habitable cell: check if in draft, existing room, or free
				if s.IsDrafting && isCurrentDraft && CellInList(s.DraftCells, coord) {
					cellEl.Class("cell", "draft")
				} else if room := FindRoomByCell(root.Rooms, fl.ID, coord); room != nil {
					switch room.RoomType {
					case RoomTypeShared:
						cellEl.Class("cell", "t-shared")
					case RoomTypeStore:
						cellEl.Class("cell", "t-store")
					default:
						cellEl.Class("cell", "t-box")
					}
					cellEl.Attr("title", room.Name)
				} else {
					cellEl.Class("cell", "free")
					if s.IsDrafting && isCurrentDraft {
						cellEl.Class("cell", "free", "paint")
					}
				}

				cellEl.OnClick(func(dom.Event) {
					if s.IsDrafting && isCurrentDraft {
						// Overlap guard: cannot select if in another room
						if otherRoom := FindRoomByCell(root.Rooms, fl.ID, coord); otherRoom != nil {
							return
						}
						if CellInList(s.DraftCells, coord) {
							s.DraftCells = RemoveCell(s.DraftCells, coord)
						} else {
							s.DraftCells = AddCell(s.DraftCells, coord)
						}
						root.Refresh()
					}
				})

				gridEl.Child(cellEl)
			}
		}

		floorCard.Child(fHeader, gridEl)
		slider.Child(floorCard)
	}

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
				Text(func() string {
					if s.IsDrafting {
						return "✓ Guardar espacio (" + fmt.Sprint(len(s.DraftCells)) + " celdas)"
					}
					return "✎ Marcar en la planta"
				}()).
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
		Child(html.Span().Class("rm-panel-title").Text("Espacios (" + fmt.Sprint(len(root.Rooms)) + ")"))

	for rIdx := range root.Rooms {
		rm := &root.Rooms[rIdx]
		idx := rIdx

		swCls := "sw-box"
		if rm.RoomType == RoomTypeShared {
			swCls = "sw-shared"
		} else if rm.RoomType == RoomTypeStore {
			swCls = "sw-store"
		}

		item := html.Div().Class("room-item").
			Child(
				html.Span().Class("sw", swCls),
				html.Div().Class("rm-room-info").
					Child(
						html.Span().Class("rm-room-name").Text(rm.Name),
						html.Span().Class("rm-room-range").Text(BoundingBoxLabel(rm.Cells)),
					),
				html.Button().Class("btn", "btn-sm", "btn-danger").Text("✕").
					OnClick(func(dom.Event) {
						// Delete room
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
	totalOccupied := 0
	for _, rm := range root.Rooms {
		totalOccupied += len(rm.Cells)
	}
	summaryText := fmt.Sprint(len(root.Rooms)) + " espacios definidos · " + fmt.Sprint(totalOccupied) + " celdas ocupadas"

	footer := html.Footer().Class("rm-footer").
		Child(
			html.Button().Class("btn").Text("← Planta").
				OnClick(func(dom.Event) {
					root.GoToStage(0)
				}),
			html.Span().Class("rm-footer-summary").Text(summaryText),
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
