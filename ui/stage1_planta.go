package ui

import (
	"webtyp.com/components/stepindicator"
	"webtyp.com/components/themetoggle"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
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

	stepper := &stepindicator.StepIndicator{
		Steps: []stepindicator.Step{
			{Label: "Planta"},
			{Label: "Espacios"},
			{Label: "Artefactos"},
			{Label: "Operación"},
		},
		Active: 0,
		OnChange: func(idx int) {
			root.GoToStage(idx)
		},
	}

	// 1. Top Header with ThemeToggle
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

	decColBtn := html.Button().Class("icon-btn").
		Attr("type", "button").
		Attr("title", "Quitar columna").
		Attr("style", "padding: 2px 7px; font-weight: 700; cursor: pointer; border-radius: 4px; border: 1px solid var(--color-outline); background: var(--color-surface);").
		Text("−").
		OnClick(func(dom.Event) {
			if root.Cols > 6 {
				root.GridLocked = false
				root.Cols--
				root.Refresh()
			}
		})

	incColBtn := html.Button().Class("icon-btn").
		Attr("type", "button").
		Attr("title", "Añadir columna").
		Attr("style", "padding: 2px 7px; font-weight: 700; cursor: pointer; border-radius: 4px; border: 1px solid var(--color-outline); background: var(--color-surface);").
		Text("+").
		OnClick(func(dom.Event) {
			if root.Cols < 40 {
				root.GridLocked = false
				root.Cols++
				root.Refresh()
			}
		})

	decRowBtn := html.Button().Class("icon-btn").
		Attr("type", "button").
		Attr("title", "Quitar fila").
		Attr("style", "padding: 2px 7px; font-weight: 700; cursor: pointer; border-radius: 4px; border: 1px solid var(--color-outline); background: var(--color-surface);").
		Text("−").
		OnClick(func(dom.Event) {
			if root.Rows > 4 {
				root.GridLocked = false
				root.Rows--
				root.Refresh()
			}
		})

	incRowBtn := html.Button().Class("icon-btn").
		Attr("type", "button").
		Attr("title", "Añadir fila").
		Attr("style", "padding: 2px 7px; font-weight: 700; cursor: pointer; border-radius: 4px; border: 1px solid var(--color-outline); background: var(--color-surface);").
		Text("+").
		OnClick(func(dom.Event) {
			if root.Rows < 40 {
				root.GridLocked = false
				root.Rows++
				root.Refresh()
			}
		})

	colInput := html.Input("number").Class("num-input").
		Attr("min", "6").Attr("max", "40").
		BindAttrFunc("value", func() string {
			_ = root.VerSig.Get()
			return fmt.Sprint(root.Cols)
		}).
		BindAttrFunc("style", func() string {
			_ = root.VerSig.Get()
			if root.GridLocked {
				return "opacity: 0.85; width: 44px; text-align: center; border-radius: 4px; border: 1px solid var(--color-outline); background: var(--color-surface);"
			}
			return "opacity: 1; cursor: text; width: 44px; text-align: center; border-radius: 4px; border: 1.5px solid var(--color-primary); background: var(--color-surface); font-weight: 700;"
		}).
		OnInput(func(e dom.Event) {
			v := ParsePositiveInt(e.TargetValue())
			if v >= 6 && v <= 40 && v != root.Cols {
				root.GridLocked = false
				root.Cols = v
				root.Refresh()
			}
		}).
		OnChange(func(e dom.Event) {
			v := ParsePositiveInt(e.TargetValue())
			if v >= 6 && v <= 40 && v != root.Cols {
				root.GridLocked = false
				root.Cols = v
				root.Refresh()
			}
		})

	rowInput := html.Input("number").Class("num-input").
		Attr("min", "4").Attr("max", "40").
		BindAttrFunc("value", func() string {
			_ = root.VerSig.Get()
			return fmt.Sprint(root.Rows)
		}).
		BindAttrFunc("style", func() string {
			_ = root.VerSig.Get()
			if root.GridLocked {
				return "opacity: 0.85; width: 44px; text-align: center; border-radius: 4px; border: 1px solid var(--color-outline); background: var(--color-surface);"
			}
			return "opacity: 1; cursor: text; width: 44px; text-align: center; border-radius: 4px; border: 1.5px solid var(--color-primary); background: var(--color-surface); font-weight: 700;"
		}).
		OnInput(func(e dom.Event) {
			v := ParsePositiveInt(e.TargetValue())
			if v >= 4 && v <= 40 && v != root.Rows {
				root.GridLocked = false
				root.Rows = v
				root.Refresh()
			}
		}).
		OnChange(func(e dom.Event) {
			v := ParsePositiveInt(e.TargetValue())
			if v >= 4 && v <= 40 && v != root.Rows {
				root.GridLocked = false
				root.Rows = v
				root.Refresh()
			}
		})

	lockIcon := html.Span().Class("icon-padlock").
		BindTextFunc(func() string {
			_ = root.VerSig.Get()
			if root.GridLocked {
				return "🔒"
			}
			return "🔓"
		})

	lockBtn := html.Button().Class("icon-btn").
		Attr("type", "button").
		Attr("title", "Bloquear/Desbloquear cuadrícula").
		Child(lockIcon).
		OnClick(func(dom.Event) {
			root.GridLocked = !root.GridLocked
			root.Refresh()
		})

	gridCtrl := html.Div().Class("rm-grid-control").
		Child(
			html.Span().Class("rm-grid-lbl").Text("Cuadrícula del edificio:"),
			html.Span().Class("rm-grid-field").Text("Col ").Child(decColBtn, colInput, incColBtn),
			html.Span().Text(" × "),
			html.Span().Class("rm-grid-field").Text("Filas ").Child(decRowBtn, rowInput, incRowBtn),
			lockBtn,
		)

	dotsNav := html.Div().Class("rm-nav-dots")
	prevBtn := html.Button().Class("icon-btn").Text("◀").
		OnClick(func(dom.Event) {
			if root.ActiveFloor > 0 {
				root.ActiveFloor--
				root.Refresh()
			}
		})
	nextBtn := html.Button().Class("icon-btn").Text("▶").
		OnClick(func(dom.Event) {
			if root.ActiveFloor < len(root.Floors)-1 {
				root.ActiveFloor++
				root.Refresh()
			}
		})
	dotsNav.Child(prevBtn)
	for i := range root.Floors {
		idx := i
		dot := html.Button().Class("dot").
			BindClassFunc("active", func() bool {
				_ = root.VerSig.Get()
				return root.ActiveFloor == idx
			}).
			OnClick(func(dom.Event) {
				root.ActiveFloor = idx
				root.Refresh()
			})
		dotsNav.Child(dot)
	}
	dotsNav.Child(nextBtn)

	topBar := html.Section().Class("rm-section-title").
		Child(
			html.Div().Class("rm-title-box").
				Child(
					html.Span().Class("rm-step-tag").Text("Paso 1 de 4"),
					html.H1().Text("Dibuja el espacio habitable de cada planta"),
					html.P().Text("Haz clic sobre la cuadrícula para habilitar o deshabilitar celdas habitables."),
				),
			html.Div().Class("rm-title-actions").Child(gridCtrl, dotsNav),
		)

	// 3. Unlocked Warning Banner
	warningBanner := html.Div().Class("rm-banner-warn").
		BindAttrFunc("style", func() string {
			_ = root.VerSig.Get()
			if root.GridLocked {
				return "display: none;"
			}
			return "background: rgba(232, 163, 61, 0.15); border: 1px solid var(--color-accent, #e8a33d); color: var(--color-on-surface); padding: 8px 14px; border-radius: 8px; font-size: 13px; margin: 8px 0;"
		}).
		Text("⚠️ Cuadrícula desbloqueada. El tamaño se comparte entre todas las plantas y define las coordenadas (A1, B2…) de cada espacio. Modifícalo con precaución.")

	// 4. Floor Slider
	s.sliderNodes = dom.NewNodes(s.buildFloorCards()...)
	slider := html.Div().Class("rm-slider").BindChildren(s.sliderNodes)

	// 5. Total count summary & Footer
	footer := html.Footer().Class("rm-footer").
		Child(
			html.Span().Class("rm-footer-summary").
				BindTextFunc(func() string {
					_ = root.VerSig.Get()
					totalHabitable := 0
					for _, fl := range root.Floors {
						totalHabitable += len(fl.HabitableCells)
					}
					return fmt.Sprint(len(root.Floors)) + " plantas configuradas · " + fmt.Sprint(totalHabitable) + " celdas habitables en total"
				}),
			html.Button().Class("btn", "btn-primary").
				Child(html.Span().Text("Siguiente: Espacios →")).
				OnClick(func(dom.Event) {
					root.GoToStage(1)
				}),
		)

	layout := html.Div().Class("rm-stage")
	layout.Child(header, topBar, warningBanner, slider, footer)

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

	for i := range root.Floors {
		fIdx := i
		f := &root.Floors[fIdx]

		cardHeader := html.Div().Class("rm-floor-header").
			Child(
				html.Span().Class("rm-floor-badge").Text(fmt.Sprint(f.Position)),
				html.Input("text").Class("name-input").Attr("value", f.Name).
					OnChange(func(e dom.Event) {
						f.Name = e.TargetValue()
					}),
				html.Span().Class("rm-floor-count").
					BindTextFunc(func() string {
						_ = root.VerSig.Get()
						return fmt.Sprint(len(f.HabitableCells)) + " celdas hab."
					}),
			)

		// Grid element with letter headers and row numbers
		gridEl := html.Div().Class("grid").
			Attr("style", fmt.Sprintf("display: grid; grid-template-columns: 24px repeat(%d, minmax(22px, 1fr)); gap: 2px; padding: 10px; background: var(--color-background); border: 1px solid var(--color-outline); border-radius: 12px; overflow-x: auto; touch-action: none; user-select: none; -webkit-user-select: none;", root.Cols)).
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
					if CellInList(f.HabitableCells, coord) {
						return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-primary); background: var(--color-primary); color: #ffffff; border-radius: 4px; cursor: pointer; box-shadow: 0 0 3px var(--color-primary); touch-action: none; user-select: none; -webkit-user-select: none;"
					}
					return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1px solid var(--color-outline); background: var(--color-surface); cursor: pointer; border-radius: 4px; touch-action: none; user-select: none; -webkit-user-select: none;"
				})

				cellBtn.OnPointerDown(func(e dom.Event) {
					e.ReleasePointerCapture()
					isHab := CellInList(f.HabitableCells, coord)
					s.dragMode = !isHab
					s.isDragging = true
					if s.dragMode {
						f.HabitableCells = AddCell(f.HabitableCells, coord)
					} else {
						f.HabitableCells = RemoveCell(f.HabitableCells, coord)
					}
					root.Refresh()
				})

				cellBtn.OnPointerEnter(func(e dom.Event) {
					if s.isDragging {
						if s.dragMode {
							f.HabitableCells = AddCell(f.HabitableCells, coord)
						} else {
							f.HabitableCells = RemoveCell(f.HabitableCells, coord)
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

		cardFooter := html.Div().Class("rm-floor-footer").
			Child(
				html.Span().Class("rm-floor-stat").
					BindTextFunc(func() string {
						_ = root.VerSig.Get()
						total := root.Cols * root.Rows
						pct := 0
						if total > 0 {
							pct = (len(f.HabitableCells) * 100) / total
						}
						return "Habitable: " + fmt.Sprint(pct) + "% de la planta"
					}),
				html.Button().Class("btn", "btn-sm").Text("Limpiar").
					OnClick(func(dom.Event) {
						f.HabitableCells = nil
						root.Refresh()
					}),
			)

		floorCard := html.Article().Class("floor").
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
	addFloorCard := html.Button().Class("floor-add").
		Key("floor-add-btn").
		Child(
			html.Span().Class("icon-add").Text("+"),
			html.Span().Text("Agregar planta"),
		).
		OnClick(func(dom.Event) {
			root.AddFloor("Piso " + fmt.Sprint(len(root.Floors)+1))
		})
	cards = append(cards, addFloorCard)

	return cards
}
