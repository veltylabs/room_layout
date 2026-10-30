package ui

import (
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/lang"
)

type Stage1Planta struct {
	dom.Element
	Root *RootView

	isDragging  bool
	dragMode    bool
	sliderNodes *dom.SignalNodes
}

func (s *Stage1Planta) Render() *dom.Element {
	root := s.Root
	activeFloor := root.ActiveFloor
	if activeFloor < 0 || activeFloor >= len(root.Floors) {
		activeFloor = 0
	}

	header := root.header(0)

	decColBtn := html.Button().Class(clsButton).
		Attr("type", "button").
		Attr("title", lang.Translate("Remove column").String()).
		Attr("style", "padding: 2px 7px; font-weight: 700; cursor: pointer; border-radius: 4px; border: 1px solid var(--color-outline); background: var(--color-surface);").
		Text("−").
		OnClick(func(dom.Event) {
			if root.Cols > 6 {
				root.ResizeGrid(root.Cols-1, root.Rows)
			}
		})

	incColBtn := html.Button().Class(clsButton).
		Attr("type", "button").
		Attr("title", lang.Translate("Add column").String()).
		Attr("style", "padding: 2px 7px; font-weight: 700; cursor: pointer; border-radius: 4px; border: 1px solid var(--color-outline); background: var(--color-surface);").
		Text("+").
		OnClick(func(dom.Event) {
			if root.Cols < 40 {
				root.ResizeGrid(root.Cols+1, root.Rows)
			}
		})

	decRowBtn := html.Button().Class(clsButton).
		Attr("type", "button").
		Attr("title", lang.Translate("Remove row").String()).
		Attr("style", "padding: 2px 7px; font-weight: 700; cursor: pointer; border-radius: 4px; border: 1px solid var(--color-outline); background: var(--color-surface);").
		Text("−").
		OnClick(func(dom.Event) {
			if root.Rows > 4 {
				root.ResizeGrid(root.Cols, root.Rows-1)
			}
		})

	incRowBtn := html.Button().Class(clsButton).
		Attr("type", "button").
		Attr("title", lang.Translate("Add row").String()).
		Attr("style", "padding: 2px 7px; font-weight: 700; cursor: pointer; border-radius: 4px; border: 1px solid var(--color-outline); background: var(--color-surface);").
		Text("+").
		OnClick(func(dom.Event) {
			if root.Rows < 40 {
				root.ResizeGrid(root.Cols, root.Rows+1)
			}
		})

	colInput := html.Input("number").Class(clsInput).
		Attr("min", "6").Attr("max", "40").
		Attr("style", "width: 64px; text-align: center; border-radius: 4px; border: 1.5px solid var(--color-primary); background: var(--color-surface); font-weight: 700;").
		BindAttrFunc("value", func() string {
			_ = root.VerSig.Get()
			return fmt.Sprint(root.Cols)
		}).
		OnChange(func(e dom.Event) {
			v := ParsePositiveInt(e.TargetValue())
			if v >= 6 && v <= 40 && v != root.Cols {
				root.ResizeGrid(v, root.Rows)
			}
		})

	rowInput := html.Input("number").Class(clsInput).
		Attr("min", "4").Attr("max", "40").
		Attr("style", "width: 64px; text-align: center; border-radius: 4px; border: 1.5px solid var(--color-primary); background: var(--color-surface); font-weight: 700;").
		BindAttrFunc("value", func() string {
			_ = root.VerSig.Get()
			return fmt.Sprint(root.Rows)
		}).
		OnChange(func(e dom.Event) {
			v := ParsePositiveInt(e.TargetValue())
			if v >= 4 && v <= 40 && v != root.Rows {
				root.ResizeGrid(root.Cols, v)
			}
		})

	gridCtrl := html.Div().Class(clsActions).
		Child(
			html.Span().Text(lang.Translate("Building grid").String()+":"),
			html.Span().Class(clsGridField).Text(lang.Translate("Columns").String()+" ").Child(decColBtn, colInput, incColBtn),
			html.Span().Text(" × "),
			html.Span().Class(clsGridField).Text(lang.Translate("Rows").String()+" ").Child(decRowBtn, rowInput, incRowBtn),
		)

	topBar := html.Section().Class(clsTitleBar).
		Child(
			html.Div().Class(clsTitle).
				Child(
					html.Span().Class(clsStepTag).Text(lang.Translate("Step", 1, "of", 4).String()),
					html.H2().Class(clsHeading).Text(lang.Translate("Draw the usable area of each floor").String()),
					html.P().Text(lang.Translate("Touch or drag over the grid to mark or clear usable cells.").String()),
				),
			html.Div().Class(clsActions).Child(gridCtrl),
		)

	// Floor Slider
	s.sliderNodes = dom.NewNodes(s.buildFloorCards()...)
	slider := html.Div().Class(clsSlider).BindChildren(s.sliderNodes)

	// Total count summary & Footer
	footer := html.Footer().Class(clsFooter).
		Child(
			html.Span().Class(clsSummary).
				BindTextFunc(func() string {
					_ = root.VerSig.Get()
					return lang.Translate(len(root.Floors), "floors").String()
				}),
			html.Button().Class(clsPrimary).
				Child(html.Span().Text(lang.Translate("Next").String()+": "+lang.Translate("Spaces").String()+" →")).
				OnClick(func(dom.Event) {
					root.GoToStage(1)
				}),
		)

	layout := html.Div().Class(clsStage)
	layout.Child(header, footer, topBar, slider)

	return layout
}

func (s *Stage1Planta) RebuildFloors() {
	if s.sliderNodes != nil {
		s.sliderNodes.Set(s.buildFloorCards())
	}
}

func (s *Stage1Planta) buildFloorCards() []*dom.Element {
	root := s.Root
	var cards []*dom.Element

	for _, i := range root.shownFloors() {
		fIdx := i
		f := &root.Floors[fIdx]

		cardHeader := html.Div().Class(clsFloorHeader).
			Child(
				html.Span().Class(clsFloorBadge).Text(fmt.Sprint(f.Position)),
				html.Input("text").Class(clsFloorName).Attr("value", f.Name).
					OnChange(func(e dom.Event) {
						f.Name = e.TargetValue()
						root.saveFloor(*f)
					}),
				root.floorNav(),
			)

		// Grid element with letter headers and row numbers
		gridEl := html.Div().Class("grid").
			Attr("style", fmt.Sprintf("display: grid; grid-template-columns: 24px repeat(%d, 28px); gap: 2px; padding: 10px; background: var(--color-background); border: 1px solid var(--color-outline); border-radius: 12px; overflow-x: auto; touch-action: none; user-select: none; -webkit-user-select: none;", root.Cols)).
			OnPointerUp(func(dom.Event) {
				s.isDragging = false
			}).
			OnPointerLeave(func(dom.Event) {
				s.isDragging = false
			})

		// Column headers
		gridEl.Child(html.Span()) // Corner
		for c := 0; c < root.Cols; c++ {
			gridEl.Child(html.Span().Class("lbl-col").
				Attr("style", "font-size: 11px; font-weight: 700; color: var(--color-muted); text-align: center; height: 20px; display: flex; align-items: center; justify-content: center;").
				Text(fmt.Sprint(c + 1)))
		}

		for r := 0; r < root.Rows; r++ {
			rowIdx := r
			gridEl.Child(html.Span().Class("lbl-row").
				Attr("style", "font-size: 11px; font-weight: 700; color: var(--color-muted); display: flex; align-items: center; justify-content: center;").
				Text(fmt.Sprint(rowIdx + 1)))

			for c := 0; c < root.Cols; c++ {
				colIdx := c
				coord := FormatCell(rowIdx, colIdx)

				isInitiallyHab := CellInList(f.HabitableCells, coord)
				cellSig := dom.NewBool(isInitiallyHab)

				cellBtn := html.Button().Class("cell").
					Attr("type", "button").
					Attr("title", coord)

				cellBtn.BindAttrFunc("style", func() string {
					if cellSig.Get() {
						return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-primary); background: var(--color-primary); color: #ffffff; border-radius: 4px; cursor: pointer; box-shadow: 0 0 3px var(--color-primary); touch-action: none; user-select: none; -webkit-user-select: none;"
					}
					return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1px solid var(--color-outline); background: var(--color-surface); cursor: pointer; border-radius: 4px; touch-action: none; user-select: none; -webkit-user-select: none;"
				})

				cellBtn.OnPointerDown(func(e dom.Event) {
					e.ReleasePointerCapture()
					isHab := cellSig.Get()
					s.dragMode = !isHab
					s.isDragging = true
					cellSig.Set(s.dragMode)
					if s.dragMode {
						f.HabitableCells = AddCell(f.HabitableCells, coord)
					} else {
						f.HabitableCells = RemoveCell(f.HabitableCells, coord)
					}
				})

				cellBtn.OnPointerEnter(func(e dom.Event) {
					if s.isDragging {
						if cellSig.Get() != s.dragMode {
							cellSig.Set(s.dragMode)
							if s.dragMode {
								f.HabitableCells = AddCell(f.HabitableCells, coord)
							} else {
								f.HabitableCells = RemoveCell(f.HabitableCells, coord)
							}
						}
					}
				})

				cellBtn.OnPointerUp(func(dom.Event) {
					s.isDragging = false
					root.saveFloor(*f) // una sola vez por trazo, no por celda
				})

				gridEl.Child(cellBtn)
			}
		}

		cardFooter := html.Div().Class(clsFooter).
			Child(
				html.Button().Class(clsButton).Text(lang.Translate("Clear").String()).
					OnClick(func(dom.Event) {
						f.HabitableCells = nil
						root.saveFloor(*f)
					}),
			)

		floorCard := html.Article().Class(clsFloor).
			Key(fmt.Sprintf("floor-%s-%dx%d", f.ID, root.Cols, root.Rows)).
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
			Child(cardHeader, gridEl, cardFooter)

		cards = append(cards, floorCard)
	}

	// "+ Agregar planta" Card
	addFloorCard := html.Button().Class(clsFloorAdd).
		Key("floor-add-btn").
		Child(
			html.Span().Class("icon-add").Text("+"),
			html.Span().Text(lang.Translate("Add floor").String()),
		).
		OnClick(func(dom.Event) {
			root.AddFloor(lang.Translate("Floor", len(root.Floors)+1).String())
		})
	cards = append(cards, addFloorCard)

	return cards
}
