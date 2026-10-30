package ui

import (
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

type Stage3Artefactos struct {
	dom.Element
	Root *RootView

	IsPlacing          bool
	PlacingKind        string
	SelectedArtifactID string
	sliderNodes        *dom.SignalNodes
}

func (s *Stage3Artefactos) Render() *dom.Element {
	root := s.Root

	header := root.header(2)

	topBar := html.Section().Class(clsTitleBar).
		Child(
			html.Div().Class(clsTitle).
				Child(
					html.Span().Class(clsStepTag).Text(lang.Translate("Step", 3, "of", 4).String()),
					html.H2().Class(clsHeading).Text(lang.Translate("Equip each space").String()),
					html.P().Text(lang.Translate("Place the equipment in each space. Touch a piece of equipment to select it, change its status or remove it.").String()),
				),
		)

	columns := html.Div().Class(clsColumns)

	// Left: Content (Floor Slider)
	content := html.Div().Class(clsContent)

	if s.IsPlacing {
		kindName := artifactKindLabel(s.PlacingKind)
		banner := html.Div().Class("rm-banner-draft").
			Child(
				html.Span().Class("rm-draft-title").
					Text("📍 "+kindName+" · "+lang.Translate("Touch a free cell inside a room to place it").String()),
				html.Button().Class(clsButton).
					Text("✕ "+lang.Translate("Cancel").String()).
					OnClick(func(dom.Event) {
						s.IsPlacing = false
						root.Refresh()
					}),
			)
		content.Child(banner)
	}

	s.sliderNodes = dom.NewNodes(s.buildFloorCards()...)
	slider := html.Div().Class(clsSlider).BindChildren(s.sliderNodes)

	content.Child(slider)

	// Right: Aside (Selected Inspector & Catalog)
	aside := html.Aside().Class(clsAside)

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

		inspectorPanel := html.Div().Class(clsPanel).
			Child(
				html.Div().Class("rm-sel-header").
					Child(
						html.Span().Class(clsPanelTitle).Text(lang.Translate("Selected equipment").String()),
						html.Button().Class(clsButton).Text("✕").
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
								html.Span().Class(clsField).Text(lang.Translate("Location").String()+": "+roomName+" ("+selectedArt.Cell+")"),
							),
					),
				html.Label().Class(clsField).Text(lang.Translate("Status").String()).
					Child(
						html.Div().Class(clsSeg).
							Child(
								func() *dom.Element {
									btn := html.Button().Class(clsSegItem).Text(lang.Translate("Normal").String())
									if selectedArt.Status == StatusOk {
										btn.SetState(widget.Selected)
									}
									return btn.OnClick(func(dom.Event) {
										setArtifactStatus(root, *selectedArt, StatusOk, "")
									})
								}(),
								func() *dom.Element {
									btn := html.Button().Class(clsSegItem).Text(lang.Translate("Warning").String())
									if selectedArt.Status == StatusWarn {
										btn.SetState(widget.Selected)
									}
									return btn.OnClick(func(dom.Event) {
										setArtifactStatus(root, *selectedArt, StatusWarn, lang.Translate("Needs periodic review").String())
									})
								}(),
								func() *dom.Element {
									btn := html.Button().Class(clsSegItem).Text(lang.Translate("Critical").String())
									if selectedArt.Status == StatusCrit {
										btn.SetState(widget.Selected)
									}
									return btn.OnClick(func(dom.Event) {
										setArtifactStatus(root, *selectedArt, StatusCrit, lang.Translate("Out of service").String())
									})
								}(),
							),
					),
			)

		if selectedArt.Status != StatusOk {
			reasonInput := html.Label().Class(clsField).Text(lang.Translate("Alert reason").String()).
				Child(
					html.Input("text").Class(clsInput).
						Attr("value", selectedArt.Reason).
						OnChange(func(e dom.Event) {
							a := *selectedArt
							a.Reason = e.TargetValue()
							root.saveArtifact(a)
						}),
				)
			inspectorPanel.Child(reasonInput)
		}

		deleteBtn := html.Button().Class(clsDanger).
			Text(lang.Translate("Remove from map").String()).
			OnClick(func(dom.Event) {
				s.SelectedArtifactID = ""
				root.deleteArtifact(selectedArt.ID)
			})
		inspectorPanel.Child(deleteBtn)

		aside.Child(inspectorPanel)
	}

	// Catalog Panel
	catalogPanel := html.Div().Class(clsPanel).
		Child(
			html.Span().Class(clsPanelTitle).Text(lang.Translate("Equipment catalog").String()),
			html.P().Class(clsField).Text(lang.Translate("Touch an item to place it on the map").String()+":"),
			html.Div().Class("cat").
				Child(
					catalogItem(ArtifactAgenda, "AG", s, root),
					catalogItem(ArtifactPC, "PC", s, root),
					catalogItem(ArtifactPrinter, "IMP", s, root),
					catalogItem(ArtifactGastro, "GAS", s, root),
					catalogItem(ArtifactChair, "SD", s, root),
					catalogItem(ArtifactECG, "ECG", s, root),
					catalogItem(ArtifactMonitor, "MON", s, root),
					catalogItem(ArtifactSupply, "INS", s, root),
				),
		)

	aside.Child(catalogPanel)
	columns.Child(content, aside)

	// Footer
	footer := html.Footer().Class(clsFooter).
		Child(
			html.Button().Class(clsButton).Text("← "+lang.Translate("Spaces").String()).
				OnClick(func(dom.Event) {
					root.GoToStage(1)
				}),
			html.Span().Class(clsSummary).
				BindTextFunc(func() string {
					_ = root.VerSig.Get()
					alertCount := 0
					for _, a := range root.Artifacts {
						if a.Status != StatusOk {
							alertCount++
						}
					}
					return lang.Translate(len(root.Artifacts), "equipment", "·", len(root.Rooms), "spaces", "·", alertCount, "alerts").String()
				}),
			html.Button().Class(clsPrimary).
				Child(html.Span().Text(lang.Translate("Finish").String()+" →")).
				OnClick(func(dom.Event) {
					root.GoToStage(3)
				}),
		)

	layout := html.Div().Class(clsStage).
		Child(header, footer, topBar, columns)

	return layout
}

func (s *Stage3Artefactos) RebuildFloors() {
	if s.sliderNodes != nil {
		s.sliderNodes.Set(s.buildFloorCards())
	}
}

func (s *Stage3Artefactos) buildFloorCards() []*dom.Element {
	root := s.Root
	var cards []*dom.Element

	for _, i := range root.shownFloors() {
		fl := &root.Floors[i]
		fIdx := i

		floorCard := html.Article().Class(clsFloor)
		fHeader := html.Div().Class(clsFloorHeader).
			Child(
				html.Span().Class(clsFloorBadge).Text(fmt.Sprint(fl.Position)),
				html.Span().Class(clsFloorName).Text(fl.Name),
				root.floorNav(),
			)

		gridEl := html.Div().Class("grid").
			Attr("style", fmt.Sprintf(planGrid, root.Cols))

		for r := 0; r < root.Rows; r++ {
			rowIdx := r
			for c := 0; c < root.Cols; c++ {
				colIdx := c
				coord := FormatCell(rowIdx, colIdx)

				cellBtn := html.Div().Class("cell")

				cellBtn.BindAttrFunc("style", func() string {
					_ = root.VerSig.Get()
					room := FindRoomByCell(root.Rooms, fl.ID, coord)
					if room != nil && s.IsPlacing && FindArtifactByCell(root.Artifacts, fl.ID, coord) == nil {
						return gridPos(rowIdx, colIdx) + "border: 2px dashed var(--color-primary); background: var(--color-selection); cursor: copy;" + planContent
					}
					return gridPos(rowIdx, colIdx) + planCell(fl, room, coord) + planContent
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

						s.IsPlacing = false
						root.saveArtifact(ArtifactData{
							RoomID:  room.ID,
							FloorID: fl.ID,
							Kind:    s.PlacingKind,
							Code:    code,
							Cell:    coord,
							Status:  StatusOk,
						})
					})
				}

				gridEl.Child(cellBtn)
			}
		}
		for _, outline := range root.roomOutlines(fl.ID, "", plain, nameTop) {
			gridEl.Child(outline)
		}

		floorCard.Key(fmt.Sprintf("floor-s3-%s-%dx%d", fl.ID, root.Cols, root.Rows)).
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

func catalogItem(kind, code string, s *Stage3Artefactos, root *RootView) *dom.Element {
	btn := html.Button().Class("btn-cat")
	if s.IsPlacing && s.PlacingKind == kind {
		btn.Class("btn-cat", "on")
	}
	btn.Child(
		html.Span().Class("ic").Text(code),
		html.Span().Text(artifactKindLabel(kind)),
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

// artifactKindLabel es el nombre de un tipo de equipo, traducido. Un solo lugar para cada nombre:
// lo usan el catálogo, la etiqueta del equipo seleccionado y las alertas.
func artifactKindLabel(kind string) string {
	return lang.Translate(string(artifactKindKey(kind))).String()
}

func artifactKindKey(kind string) lang.Text {
	switch kind {
	case ArtifactAgenda:
		return "Medical agenda"
	case ArtifactPC:
		return "Computer"
	case ArtifactPrinter:
		return "Printer"
	case ArtifactGastro:
		return "Gastroscope"
	case ArtifactChair:
		return "Dental chair"
	case ArtifactECG:
		return "Electrocardiograph"
	case ArtifactMonitor:
		return "Vital signs monitor"
	case ArtifactSupply:
		return "Clinical supplies"
	default:
		return "Equipment"
	}
}

// setArtifactStatus cambia el estado de un equipo y lo guarda; al pasar a alerta sin motivo,
// le pone el motivo por defecto de ese estado.
func setArtifactStatus(root *RootView, a ArtifactData, status, defaultReason string) {
	a.Status = status
	if a.Reason == "" {
		a.Reason = defaultReason
	}
	root.saveArtifact(a)
}
