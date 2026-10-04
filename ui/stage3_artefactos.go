package ui

import (
	"webtyp.com/components/stepindicator"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
)

type Stage3Artefactos struct {
	dom.Element
	Root *RootView

	IsPlacing          bool
	PlacingKind        string
	SelectedArtifactID string
}

func (s *Stage3Artefactos) Render() *dom.Element {
	root := s.Root

	stepper := &stepindicator.StepIndicator{
		Steps: []stepindicator.Step{
			{Label: "Planta"},
			{Label: "Espacios"},
			{Label: "Artefactos"},
			{Label: "Operación"},
		},
		Active: 2,
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

	topBar := html.Section().Class("rm-section-title").
		Child(
			html.Div().Class("rm-title-box").
				Child(
					html.Span().Class("rm-step-tag").Text("Paso 3 de 4"),
					html.H1().Text("Equipa cada espacio"),
					html.P().Text("Agrega artefactos dentro de los espacios. Haz clic en un artefacto para seleccionarlo, cambiar su estado de alerta o quitarlo."),
				),
		)

	columns := html.Div().Class("rm-columns")

	// Left: Content (Floor Slider)
	content := html.Div().Class("rm-content")

	if s.IsPlacing {
		kindName := artifactKindLabel(s.PlacingKind)
		banner := html.Div().Class("rm-banner-draft").
			Child(
				html.Span().Class("rm-draft-title").
					Text("📍 Colocando «"+kindName+"» · Haz clic en una celda libre dentro de un box para ubicarlo"),
				html.Button().Class("btn", "btn-sm").
					Text("✕ Cancelar").
					OnClick(func(dom.Event) {
						s.IsPlacing = false
						root.Refresh()
					}),
			)
		content.Child(banner)
	}

	slider := html.Div().Class("rm-slider")

	for i := range root.Floors {
		fl := &root.Floors[i]

		floorCard := html.Article().Class("floor")
		fHeader := html.Div().Class("rm-floor-header").
			Child(
				html.Span().Class("rm-floor-badge").Text(fmt.Sprint(fl.Position)),
				html.Span().Class("name-input").Text(fl.Name),
				html.Span().Class("rm-floor-count").Text(fmt.Sprint(len(fl.HabitableCells))+" hab."),
			)

		gridEl := html.Div().Class("grid").
			Attr("style", fmt.Sprintf("display: grid; grid-template-columns: repeat(%d, 1fr); grid-template-rows: repeat(%d, 1fr);", root.Cols, root.Rows))

		for r := 0; r < root.Rows; r++ {
			for c := 0; c < root.Cols; c++ {
				coord := FormatCell(r, c)
				cellEl := html.Div().Class("cell").
					Attr("style", fmt.Sprintf("grid-row: %d; grid-column: %d;", r+1, c+1))

				if !CellInList(fl.HabitableCells, coord) {
					cellEl.Class("cell", "non-habitable")
					gridEl.Child(cellEl)
					continue
				}

				room := FindRoomByCell(root.Rooms, fl.ID, coord)
				if room != nil {
					switch room.RoomType {
					case RoomTypeShared:
						cellEl.Class("cell", "g-shared")
					case RoomTypeStore:
						cellEl.Class("cell", "g-store")
					default:
						cellEl.Class("cell", "g-box")
					}
				} else {
					cellEl.Class("cell", "free")
				}

				// Check if artifact is placed here
				art := FindArtifactByCell(root.Artifacts, fl.ID, coord)
				if art != nil {
					tokEl := html.Button().Class("tok").
						Attr("title", art.Code+" ("+artifactKindLabel(art.Kind)+")").
						Child(html.Span().Class("tok-code").Text(artifactKindCode(art.Kind)))

					if art.Status == StatusWarn {
						tokEl.Class("tok", "warn")
						tokEl.Child(html.Span().Class("bang").Text("!"))
					} else if art.Status == StatusCrit {
						tokEl.Class("tok", "crit")
						tokEl.Child(html.Span().Class("bang").Text("!"))
					}

					if s.SelectedArtifactID == art.ID {
						tokEl.Class("tok", "sel")
					}

					tokEl.OnClick(func(dom.Event) {
						s.SelectedArtifactID = art.ID
						s.IsPlacing = false
						root.Refresh()
					})

					cellEl.Child(tokEl)
				} else if s.IsPlacing && room != nil {
					// Drop target available
					cellEl.Class("cell", "drop")
					cellEl.OnClick(func(dom.Event) {
						// Overlap guard: Max 1 agenda per room
						if s.PlacingKind == ArtifactAgenda {
							for _, a := range root.Artifacts {
								if a.RoomID == room.ID && a.Kind == ArtifactAgenda {
									// Already has agenda
									return
								}
							}
						}

						codeCount := 1
						for _, a := range root.Artifacts {
							if a.Kind == s.PlacingKind {
								codeCount++
							}
						}
						code := artifactKindCode(s.PlacingKind) + "-" + fmt.Sprintf("%02d", codeCount)

						newArt := ArtifactData{
							ID:      root.NewID(),
							RoomID:  room.ID,
							FloorID: fl.ID,
							Kind:    s.PlacingKind,
							Code:    code,
							Cell:    coord,
							Status:  StatusOk,
						}
						root.Artifacts = append(root.Artifacts, newArt)
						s.SelectedArtifactID = newArt.ID
						s.IsPlacing = false
						root.Refresh()
					})
				}

				gridEl.Child(cellEl)
			}
		}

		floorCard.Child(fHeader, gridEl)
		slider.Child(floorCard)
	}

	content.Child(slider)

	// Right: Aside (Selected Inspector & Catalog)
	aside := html.Aside().Class("rm-aside")

	var selectedArt *ArtifactData
	if s.SelectedArtifactID != "" {
		for i := range root.Artifacts {
			if root.Artifacts[i].ID == s.SelectedArtifactID {
				selectedArt = &root.Artifacts[i]
				break
			}
		}
	}

	if selectedArt != nil {
		roomName := "Sin asignar"
		for _, rm := range root.Rooms {
			if rm.ID == selectedArt.RoomID {
				roomName = rm.Name
				break
			}
		}

		inspectorPanel := html.Div().Class("panel").
			Child(
				html.Div().Class("rm-sel-header").
					Child(
						html.Span().Class("rm-panel-title").Text("Artefacto seleccionado"),
						html.Button().Class("btn", "btn-sm").Text("✕").
							OnClick(func(dom.Event) {
								s.SelectedArtifactID = ""
								root.Refresh()
							}),
					),
				html.Div().Class("sel-card").
					Child(
						html.Div().Class("tok", "sel").
							Child(html.Span().Text(artifactKindCode(selectedArt.Kind))),
						html.Div().Class("rm-sel-details").
							Child(
								html.Strong().Text(selectedArt.Code+" · "+artifactKindLabel(selectedArt.Kind)),
								html.Span().Class("field").Text("Ubicación: "+roomName+" ("+selectedArt.Cell+")"),
							),
					),
				html.Label().Class("field").Text("Estado operacional").
					Child(
						html.Div().Class("seg").
							Child(
								func() *dom.Element {
									btn := html.Button().Class("btn-seg").Text("Normal")
									if selectedArt.Status == StatusOk {
										btn.Class("on")
									}
									return btn.OnClick(func(dom.Event) {
										selectedArt.Status = StatusOk
										root.Refresh()
									})
								}(),
								func() *dom.Element {
									btn := html.Button().Class("btn-seg").Text("Aviso")
									if selectedArt.Status == StatusWarn {
										btn.Class("on")
									}
									return btn.OnClick(func(dom.Event) {
										selectedArt.Status = StatusWarn
										if selectedArt.Reason == "" {
											selectedArt.Reason = "Requiere revisión periódica"
										}
										root.Refresh()
									})
								}(),
								func() *dom.Element {
									btn := html.Button().Class("btn-seg").Text("Crítico")
									if selectedArt.Status == StatusCrit {
										btn.Class("on")
									}
									return btn.OnClick(func(dom.Event) {
										selectedArt.Status = StatusCrit
										if selectedArt.Reason == "" {
											selectedArt.Reason = "Equipo fuera de servicio"
										}
										root.Refresh()
									})
								}(),
							),
					),
			)

		if selectedArt.Status != StatusOk {
			reasonInput := html.Label().Class("field").Text("Motivo de alerta").
				Child(
					html.Input("text").Class("input").
						Attr("value", selectedArt.Reason).
						OnChange(func(e dom.Event) {
							// Update reason
						}),
				)
			inspectorPanel.Child(reasonInput)
		}

		deleteBtn := html.Button().Class("btn", "btn-danger").
			Text("Quitar del mapa").
			OnClick(func(dom.Event) {
				var updated []ArtifactData
				for _, a := range root.Artifacts {
					if a.ID != selectedArt.ID {
						updated = append(updated, a)
					}
				}
				root.Artifacts = updated
				s.SelectedArtifactID = ""
				root.Refresh()
			})
		inspectorPanel.Child(deleteBtn)

		aside.Child(inspectorPanel)
	}

	// Catalog Panel
	catalogPanel := html.Div().Class("panel").
		Child(
			html.Span().Class("rm-panel-title").Text("Catálogo de artefactos"),
			html.P().Class("field").Text("Haz clic en un ítem para colocarlo sobre el mapa:"),
			html.Div().Class("cat").
				Child(
					catalogItem(ArtifactAgenda, "Agenda médica", "AG", s, root),
					catalogItem(ArtifactPC, "Computador", "PC", s, root),
					catalogItem(ArtifactPrinter, "Impresora", "IMP", s, root),
					catalogItem(ArtifactGastro, "Gastroscopio", "GAS", s, root),
					catalogItem(ArtifactChair, "Sillón dental", "SD", s, root),
					catalogItem(ArtifactECG, "ECG portátil", "ECG", s, root),
					catalogItem(ArtifactMonitor, "Monitor signos", "MON", s, root),
					catalogItem(ArtifactSupply, "Insumos clínicos", "INS", s, root),
				),
		)

	aside.Child(catalogPanel)
	columns.Child(content, aside)

	// Footer
	alertCount := 0
	for _, a := range root.Artifacts {
		if a.Status != StatusOk {
			alertCount++
		}
	}
	summaryText := fmt.Sprint(len(root.Artifacts)) + " artefactos en " + fmt.Sprint(len(root.Rooms)) + " espacios · " + fmt.Sprint(alertCount) + " alertas de mantención"

	footer := html.Footer().Class("rm-footer").
		Child(
			html.Button().Class("btn").Text("← Espacios").
				OnClick(func(dom.Event) {
					root.GoToStage(1)
				}),
			html.Span().Class("rm-footer-summary").Text(summaryText),
			html.Button().Class("btn", "btn-primary").
				Child(html.Span().Text("Finalizar configuración →")).
				OnClick(func(dom.Event) {
					root.GoToStage(3)
				}),
		)

	layout := html.Div().Class("rm-stage").
		Child(header, topBar, columns, footer)

	return layout
}

func catalogItem(kind, label, code string, s *Stage3Artefactos, root *RootView) *dom.Element {
	btn := html.Button().Class("btn-cat")
	if s.IsPlacing && s.PlacingKind == kind {
		btn.Class("btn-cat", "on")
	}
	btn.Child(
		html.Span().Class("ic").Text(code),
		html.Span().Text(label),
	)
	btn.OnClick(func(dom.Event) {
		s.IsPlacing = true
		s.PlacingKind = kind
		s.SelectedArtifactID = ""
		root.Refresh()
	})
	return btn
}

func artifactKindCode(kind string) string {
	switch kind {
	case ArtifactAgenda:
		return "AG"
	case ArtifactPC:
		return "PC"
	case ArtifactPrinter:
		return "IMP"
	case ArtifactGastro:
		return "GAS"
	case ArtifactChair:
		return "SD"
	case ArtifactECG:
		return "ECG"
	case ArtifactMonitor:
		return "MON"
	case ArtifactSupply:
		return "INS"
	default:
		return "EQ"
	}
}

func artifactKindLabel(kind string) string {
	switch kind {
	case ArtifactAgenda:
		return "Agenda médica"
	case ArtifactPC:
		return "Computador"
	case ArtifactPrinter:
		return "Impresora"
	case ArtifactGastro:
		return "Gastroscopio"
	case ArtifactChair:
		return "Sillón dental"
	case ArtifactECG:
		return "Electrocardiógrafo"
	case ArtifactMonitor:
		return "Monitor de signos"
	case ArtifactSupply:
		return "Insumos clínicos"
	default:
		return "Equipo"
	}
}
