package ui

import (
	"webtyp.com/components/cellgrid"
	"webtyp.com/components/stepindicator"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
)

type Stage1Planta struct {
	dom.Element
	Root *RootView
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

	// 1. Top Header
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

	// 2. Title and Building Grid controls
	colInput := html.Input("number").Class("num-input").
		Attr("min", "6").Attr("max", "40").
		Attr("value", fmt.Sprint(root.Cols))
	if root.GridLocked {
		colInput.Attr("disabled", "disabled")
	}
	colInput.OnChange(func(e dom.Event) {
		val := colInput.String()
		if v, _, ok := ParseCell("A" + val); ok && v > 4 && v <= 40 {
			root.Cols = v
			root.Refresh()
		}
	})

	rowInput := html.Input("number").Class("num-input").
		Attr("min", "4").Attr("max", "40").
		Attr("value", fmt.Sprint(root.Rows))
	if root.GridLocked {
		rowInput.Attr("disabled", "disabled")
	}
	rowInput.OnChange(func(e dom.Event) {
		val := rowInput.String()
		if v, _, ok := ParseCell("A" + val); ok && v > 2 && v <= 40 {
			root.Rows = v
			root.Refresh()
		}
	})

	lockBtn := html.Button().Class("icon-btn").
		Attr("title", "Bloquear/Desbloquear cuadrícula").
		OnClick(func(dom.Event) {
			root.GridLocked = !root.GridLocked
			root.Refresh()
		})
	if root.GridLocked {
		lockBtn.Child(html.Span().Class("icon-padlock").Text("🔒"))
	} else {
		lockBtn.Child(html.Span().Class("icon-padlock-unlocked").Text("🔓"))
	}

	gridCtrl := html.Div().Class("rm-grid-control").
		Child(
			html.Span().Class("rm-grid-lbl").Text("Cuadrícula del edificio:"),
			html.Label().Class("rm-grid-field").Text("Col ").Child(colInput),
			html.Span().Text(" × "),
			html.Label().Class("rm-grid-field").Text("Filas ").Child(rowInput),
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
		dot := html.Button().Class("dot")
		if idx == activeFloor {
			dot.Class("dot", "active")
		}
		dot.OnClick(func(dom.Event) {
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
					html.P().Text("Haz clic o arrastra sobre la cuadrícula para habilitar celdas habitables. Si empiezas sobre una habilitada, arrastrar la quita."),
				),
			html.Div().Class("rm-title-actions").Child(gridCtrl, dotsNav),
		)

	// 3. Unlocked Warning Banner
	var warningBanner *dom.Element
	if !root.GridLocked {
		warningBanner = html.Div().Class("rm-banner-warn").
			Text("⚠️ Cuadrícula desbloqueada. El tamaño se comparte entre todas las plantas y define las coordenadas (A1, B2…) de cada espacio. Modifícalo con precaución.")
	}

	// 4. Floor Slider
	slider := html.Div().Class("rm-slider")

	for i := range root.Floors {
		fIdx := i
		f := &root.Floors[fIdx]

		pct := 0
		totalCells := root.Cols * root.Rows
		if totalCells > 0 {
			pct = (len(f.HabitableCells) * 100) / totalCells
		}

		cardHeader := html.Div().Class("rm-floor-header").
			Child(
				html.Span().Class("rm-floor-badge").Text(fmt.Sprint(f.Position)),
				html.Input("text").Class("name-input").Attr("value", f.Name).
					OnChange(func(e dom.Event) {
						// Inline rename
						// value updated via DOM input
					}),
				html.Span().Class("rm-floor-count").Text(fmt.Sprint(len(f.HabitableCells))+" celdas hab."),
			)

		// Grid with cellgrid component
		grid := &cellgrid.CellGrid{
			Cols: root.Cols,
			Rows: root.Rows,
			IsActive: func(r, c int) bool {
				return CellInList(f.HabitableCells, FormatCell(r, c))
			},
			OnCellClick: func(r, c int) {
				coord := FormatCell(r, c)
				if CellInList(f.HabitableCells, coord) {
					f.HabitableCells = RemoveCell(f.HabitableCells, coord)
				} else {
					f.HabitableCells = AddCell(f.HabitableCells, coord)
				}
				root.Refresh()
			},
		}

		cardFooter := html.Div().Class("rm-floor-footer").
			Child(
				html.Span().Class("rm-floor-stat").
					Text("Habitable: "+fmt.Sprint(pct)+"% de la planta"),
				html.Button().Class("btn", "btn-sm").Text("Limpiar").
					OnClick(func(dom.Event) {
						f.HabitableCells = nil
						root.Refresh()
					}),
			)

		floorCard := html.Article().Class("floor").
			Child(cardHeader, grid, cardFooter)

		slider.Child(floorCard)
	}

	// "+ Agregar planta" Card
	addFloorCard := html.Button().Class("floor-add").
		Child(
			html.Span().Class("icon-add").Text("+"),
			html.Span().Text("Agregar planta"),
		).
		OnClick(func(dom.Event) {
			root.AddFloor("Piso " + fmt.Sprint(len(root.Floors)+1))
		})
	slider.Child(addFloorCard)

	// 5. Total count summary & Footer
	totalHabitable := 0
	for _, fl := range root.Floors {
		totalHabitable += len(fl.HabitableCells)
	}
	summaryText := fmt.Sprint(len(root.Floors)) + " plantas configuradas · " + fmt.Sprint(totalHabitable) + " celdas habitables en total"

	footer := html.Footer().Class("rm-footer").
		Child(
			html.Span().Class("rm-footer-summary").Text(summaryText),
			html.Button().Class("btn", "btn-primary").
				Child(html.Span().Text("Siguiente: Espacios →")).
				OnClick(func(dom.Event) {
					root.GoToStage(1)
				}),
		)

	layout := html.Div().Class("rm-stage")
	layout.Child(header, topBar)
	if warningBanner != nil {
		layout.Child(warningBanner)
	}
	layout.Child(slider, footer)

	return layout
}
