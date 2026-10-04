package ui

import (
	"webtyp.com/components/stepindicator"
	"webtyp.com/components/themetoggle"
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
				html.Span().Class("rm-floor-count").
					BindTextFunc(func() string {
						_ = root.VerSig.Get()
						return fmt.Sprint(len(fl.HabitableCells)) + " hab."
					}),
			)

		gridEl := html.Div().Class("grid").
			Attr("style", fmt.Sprintf("display: grid; grid-template-columns: 24px repeat(%d, minmax(22px, 1fr)); gap: 2px; padding: 10px; background: var(--color-background); border: 1px solid var(--color-outline); border-radius: 12px; overflow-x: auto;", root.Cols))

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

				cellBtn := html.Div().Class("cell")

				cellBtn.BindAttrFunc("style", func() string {
					_ = root.VerSig.Get()
					if !CellInList(fl.HabitableCells, coord) {
						return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1px dashed var(--color-outline); opacity: 0.12; pointer-events: none; border-radius: 4px;"
					}

					room := FindRoomByCell(root.Rooms, fl.ID, coord)
					if room != nil {
						if s.IsPlacing && FindArtifactByCell(root.Artifacts, fl.ID, coord) == nil {
							return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 2px dashed var(--color-primary); background: var(--color-selection); cursor: copy; border-radius: 4px; display: flex; align-items: center; justify-content: center;"
						}
						switch room.RoomType {
						case RoomTypeShared:
							return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-muted); background: var(--color-muted); opacity: 0.75; border-radius: 4px; display: flex; align-items: center; justify-content: center;"
						case RoomTypeStore:
							return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-accent, #e8a33d); background: var(--color-accent, #e8a33d); opacity: 0.85; border-radius: 4px; display: flex; align-items: center; justify-content: center;"
						default:
							return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-primary); background: var(--color-primary); opacity: 0.85; border-radius: 4px; display: flex; align-items: center; justify-content: center;"
						}
					}

					return "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px dashed var(--color-outline); background: var(--color-surface); border-radius: 4px; display: flex; align-items: center; justify-content: center;"
				})

				room := FindRoomByCell(root.Rooms, fl.ID, coord)

				// Check if artifact is placed here
				art := FindArtifactByCell(root.Artifacts, fl.ID, coord)
				if art != nil {
					artifact := *art
					tokEl := html.Button().Class("tok").
						Attr("type", "button").
						Attr("title", artifact.Code+" ("+artifactKindLabel(artifact.Kind)+")").
						Child(html.Span().Class("tok-code").Text(artifactKindCode(artifact.Kind)))

					tokStyle := "width: 24px; height: 24px; border-radius: 50%; font-weight: 800; font-size: 10px; display: inline-flex; align-items: center; justify-content: center; background: var(--color-background); border: 2px solid var(--color-primary); color: var(--color-primary); box-shadow: 0 1px 4px rgba(0,0,0,0.3); cursor: pointer;"
					if artifact.Status == StatusWarn {
						tokStyle = "width: 24px; height: 24px; border-radius: 50%; font-weight: 800; font-size: 10px; display: inline-flex; align-items: center; justify-content: center; background: var(--color-background); border: 2px solid var(--color-accent, #e8a33d); color: var(--color-accent, #e8a33d); box-shadow: 0 1px 4px rgba(0,0,0,0.3); cursor: pointer;"
						tokEl.Child(html.Span().Class("bang").Attr("style", "position:absolute;top:-4px;right:-4px;width:12px;height:12px;border-radius:50%;background:var(--color-accent, #e8a33d);color:#fff;font-size:8px;font-weight:800;display:flex;align-items:center;justify-content:center;").Text("!"))
					} else if artifact.Status == StatusCrit {
						tokStyle = "width: 24px; height: 24px; border-radius: 50%; font-weight: 800; font-size: 10px; display: inline-flex; align-items: center; justify-content: center; background: rgba(186, 44, 13, 0.2); border: 2px solid var(--color-danger, #ba2c0d); color: var(--color-danger, #ba2c0d); box-shadow: 0 1px 4px rgba(0,0,0,0.3); cursor: pointer;"
						tokEl.Child(html.Span().Class("bang").Attr("style", "position:absolute;top:-4px;right:-4px;width:12px;height:12px;border-radius:50%;background:var(--color-danger, #ba2c0d);color:#fff;font-size:8px;font-weight:800;display:flex;align-items:center;justify-content:center;").Text("!"))
					}

					if s.SelectedArtifactID == artifact.ID {
						tokStyle += " box-shadow: 0 0 0 3px #ffffff, 0 0 6px var(--color-primary); background: var(--color-primary); color: #ffffff;"
					}
					tokEl.Attr("style", tokStyle)

					tokEl.OnClick(func(dom.Event) {
						s.SelectedArtifactID = artifact.ID
						s.IsPlacing = false
						root.Refresh()
					})

					cellBtn.Child(tokEl)
				} else if s.IsPlacing && room != nil {
					// Drop target available
					cellBtn.OnClick(func(dom.Event) {
						// Overlap guard: Max 1 agenda per room
						if s.PlacingKind == ArtifactAgenda {
							for _, a := range root.Artifacts {
								if a.RoomID == room.ID && a.Kind == ArtifactAgenda {
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

				gridEl.Child(cellBtn)
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
							Attr("style", "width:28px;height:28px;border-radius:50%;background:var(--color-primary);color:#fff;display:flex;align-items:center;justify-content:center;font-weight:800;").
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
	footer := html.Footer().Class("rm-footer").
		Child(
			html.Button().Class("btn").Text("← Espacios").
				OnClick(func(dom.Event) {
					root.GoToStage(1)
				}),
			html.Span().Class("rm-footer-summary").
				BindTextFunc(func() string {
					_ = root.VerSig.Get()
					alertCount := 0
					for _, a := range root.Artifacts {
						if a.Status != StatusOk {
							alertCount++
						}
					}
					return fmt.Sprint(len(root.Artifacts)) + " artefactos en " + fmt.Sprint(len(root.Rooms)) + " espacios · " + fmt.Sprint(alertCount) + " alertas de mantención"
				}),
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
