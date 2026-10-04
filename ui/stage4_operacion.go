package ui

import (
	"webtyp.com/components/segmentedcontrol"
	"webtyp.com/components/themetoggle"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
)

type Stage4Operacion struct {
	dom.Element
	Root *RootView

	Mode             string // "plan", "now"
	PlanPeriod       string // "week", "month"
	NowDay           int    // 1=Lun .. 6=Sáb
	NowShift         string // "morning", "afternoon"
	Layer            string // "occupancy", "equipment"
	SelectedRoomID   string
	SelectedDoctorID string
	DoctorSearch     string
}

func (s *Stage4Operacion) Render() *dom.Element {
	root := s.Root

	if s.Mode == "" {
		s.Mode = "plan"
	}
	if s.PlanPeriod == "" {
		s.PlanPeriod = "week"
	}
	if s.NowDay == 0 {
		s.NowDay = 1 // Lunes
	}
	if s.NowShift == "" {
		s.NowShift = "morning"
	}
	if s.Layer == "" {
		s.Layer = "occupancy"
	}

	modeSeg := &segmentedcontrol.SegmentedControl{
		Options: []segmentedcontrol.Option{
			{Value: "plan", Label: "Planificar"},
			{Value: "now", Label: "Ahora"},
		},
		Selected: s.Mode,
		OnChange: func(val string) {
			s.Mode = val
			root.Refresh()
		},
	}

	configBtn := html.Button().Class("btn").
		Child(
			html.Span().Class("icon-gear").Text("⚙"),
			html.Span().Text("Configurar mapa"),
		).
		OnClick(func(dom.Event) {
			root.GoToStage(0)
		})

	header := html.Header().Class("rm-header").
		Child(
			html.Div().Class("rm-brand").
				Child(
					html.Div().Class("rm-brand-title").Text("Room Map"),
					html.Span().Class("rm-brand-sub").Text("Clínica Central"),
				),
			modeSeg,
			html.Div().Class("rm-header-actions").
				Child(
					&themetoggle.ThemeToggle{},
					configBtn,
				),
		)

	// Toolbar
	toolbar := html.Div().Class("rm-toolbar")

	if s.Mode == "plan" {
		periodSeg := &segmentedcontrol.SegmentedControl{
			Options: []segmentedcontrol.Option{
				{Value: "week", Label: "Semana"},
				{Value: "month", Label: "Mes"},
			},
			Selected: s.PlanPeriod,
			OnChange: func(val string) {
				s.PlanPeriod = val
				root.Refresh()
			},
		}

		totalBlocks := len(root.Rooms) * 12 // 6 days * 2 shifts
		occBlocks := len(root.Shifts)
		occPct := 0
		if totalBlocks > 0 {
			occPct = (occBlocks * 100) / totalBlocks
		}
		freeBlocks := totalBlocks - occBlocks
		if freeBlocks < 0 {
			freeBlocks = 0
		}

		stats := html.Div().Class("stats").
			Child(
				html.Span().Child(html.Span().Text("Ocupación: "), html.Strong().Text(fmt.Sprint(occPct)+"%")),
				html.Span().Child(html.Strong().Text(fmt.Sprint(freeBlocks)), html.Span().Text(" bloques libres")),
			)

		toolbar.Child(periodSeg, stats)
	} else {
		// "now" mode: Day chips + Shift toggle
		dayChips := html.Div().Class("rm-day-chips")
		dayNames := []string{"Lun", "Mar", "Mié", "Jue", "Vie", "Sáb"}
		for i, name := range dayNames {
			dayIdx := i + 1
			chip := html.Button().Class("chip")
			if s.NowDay == dayIdx {
				chip.Class("chip", "on")
			}
			chip.Text(name).OnClick(func(dom.Event) {
				s.NowDay = dayIdx
				root.Refresh()
			})
			dayChips.Child(chip)
		}

		shiftSeg := &segmentedcontrol.SegmentedControl{
			Options: []segmentedcontrol.Option{
				{Value: "morning", Label: "Mañana (09:00 - 13:00)"},
				{Value: "afternoon", Label: "Tarde (14:00 - 18:00)"},
			},
			Selected: s.NowShift,
			OnChange: func(val string) {
				s.NowShift = val
				root.Refresh()
			},
		}

		toolbar.Child(dayChips, shiftSeg)
	}

	// Layer switch (Ocupación / Funcionamiento)
	layerSeg := &segmentedcontrol.SegmentedControl{
		Options: []segmentedcontrol.Option{
			{Value: "occupancy", Label: "Ocupación"},
			{Value: "equipment", Label: "Funcionamiento"},
		},
		Selected: s.Layer,
		OnChange: func(val string) {
			s.Layer = val
			root.Refresh()
		},
	}
	toolbar.Child(layerSeg)

	columns := html.Div().Class("rm-columns")

	// Left: Content (Floor Map Canvas)
	content := html.Div().Class("rm-content")
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
				cellEl := html.Div().Class("cell")

				if !CellInList(fl.HabitableCells, coord) {
					cellEl.Attr("style", "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1px dashed var(--color-outline); opacity: 0.12; pointer-events: none; border-radius: 4px;")
					gridEl.Child(cellEl)
					continue
				}

				room := FindRoomByCell(root.Rooms, fl.ID, coord)
				if room == nil {
					cellEl.Attr("style", "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px dashed var(--color-outline); background: var(--color-surface); border-radius: 4px;")
					gridEl.Child(cellEl)
					continue
				}

				// Room cell
				cellStyle := "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-primary); background: var(--color-primary); opacity: 0.85; border-radius: 4px; display: flex; align-items: center; justify-content: center; position: relative;"
				if room.RoomType == RoomTypeShared {
					cellStyle = "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-muted); background: var(--color-muted); opacity: 0.75; border-radius: 4px; display: flex; align-items: center; justify-content: center; position: relative;"
				} else if room.RoomType == RoomTypeStore {
					cellStyle = "aspect-ratio: 1/1; min-width: 22px; min-height: 22px; border: 1.5px solid var(--color-accent, #e8a33d); background: var(--color-accent, #e8a33d); opacity: 0.85; border-radius: 4px; display: flex; align-items: center; justify-content: center; position: relative;"
				}
				cellEl.Attr("style", cellStyle)

				minCol, minRow, _, _ := BoundingBox(room.Cells)
				if rowIdx == minRow && colIdx == minCol {
					ov := s.renderRoomOverlay(room, fl.ID)
					cellEl.Child(ov)
				}

				gridEl.Child(cellEl)
			}
		}

		floorCard.Child(fHeader, gridEl)
		slider.Child(floorCard)
	}

	content.Child(slider)

	// Right: Aside (Selected Room Detail OR Doctors & Alerts)
	aside := html.Aside().Class("rm-aside")

	var selectedRoom *RoomData
	if s.SelectedRoomID != "" {
		for i := range root.Rooms {
			if root.Rooms[i].ID == s.SelectedRoomID {
				selectedRoom = &root.Rooms[i]
				break
			}
		}
	}

	if selectedRoom != nil {
		roomDetailPanel := s.renderRoomDetailPanel(selectedRoom)
		aside.Child(roomDetailPanel)
	} else {
		// Professionals & Alerts Panel
		proPanel := s.renderProfessionalsPanel()
		alertsPanel := s.renderAlertsPanel()
		aside.Child(proPanel, alertsPanel)
	}

	columns.Child(content, aside)

	// Footer
	footer := html.Footer().Class("rm-footer").
		Child(
			html.Button().Class("btn").Text("← Artefactos").
				OnClick(func(dom.Event) {
					root.GoToStage(2)
				}),
			html.Span().Class("rm-footer-summary").
				Text("Dashboard de Operación · Clínica Central"),
			html.Button().Class("btn", "btn-primary").
				Child(html.Span().Text("Guardar cambios")).
				OnClick(func(dom.Event) {
					root.SaveAll()
				}),
		)

	layout := html.Div().Class("rm-stage").
		Child(header, toolbar, columns, footer)

	return layout
}

func (s *Stage4Operacion) renderRoomOverlay(room *RoomData, floorID string) *dom.Element {
	root := s.Root
	ov := html.Button().Class("ov")
	if s.SelectedRoomID == room.ID {
		ov.Class("ov", "sel")
	}

	ov.Child(html.Span().Class("ov-name").Text(room.Name))

	if s.Layer == "equipment" {
		// Funcionamiento layer: show equipment alert status
		toks := html.Div().Class("rm-ov-tokens")
		for _, art := range root.Artifacts {
			if art.RoomID == room.ID {
				tok := html.Span().Class("tok-mini").
					Text(artifactKindCode(art.Kind))
				if art.Status == StatusWarn {
					tok.Class("tok-mini", "warn")
				} else if art.Status == StatusCrit {
					tok.Class("tok-mini", "crit")
				}
				toks.Child(tok)
			}
		}
		ov.Child(toks)
	} else if s.Mode == "plan" && s.PlanPeriod == "week" {
		// Mini calendar 6x2
		mc := html.Div().Class("mc")
		days := []int{1, 2, 3, 4, 5, 6}
		for _, d := range days {
			// Morning
			sqM := html.Div().Class("sq")
			if s.isRoomOccupied(room.ID, d, "09:00") {
				sqM.Class("sq", "occ")
			}
			// Afternoon
			sqA := html.Div().Class("sq")
			if s.isRoomOccupied(room.ID, d, "14:00") {
				sqA.Class("sq", "occ")
			}
			mc.Child(sqM, sqA)
		}
		ov.Child(mc)
	} else if s.Mode == "plan" && s.PlanPeriod == "month" {
		// Month occupancy %
		occCount := 0
		for _, sh := range root.Shifts {
			if sh.RoomID == room.ID {
				occCount++
			}
		}
		pct := (occCount * 100) / 12
		ov.Child(html.Span().Class("pct").Text(fmt.Sprint(pct) + "%"))
	} else if s.Mode == "now" {
		// Ahora mode
		hour := "09:00"
		if s.NowShift == "afternoon" {
			hour = "14:00"
		}
		doctorName := s.getOccupantName(room.ID, s.NowDay, hour)
		if doctorName != "" {
			docBox := html.Div().Class("now").
				Child(
					html.Span().Class("av", "av-s").Text(doctorName[:2]),
					html.Span().Class("now-name").Text(doctorName),
				)
			ov.Child(docBox)
		} else {
			ov.Child(html.Span().Class("freep").Text("Libre"))
		}
	}

	ov.OnClick(func(dom.Event) {
		s.SelectedRoomID = room.ID
		root.Refresh()
	})

	return ov
}

func (s *Stage4Operacion) renderRoomDetailPanel(room *RoomData) *dom.Element {
	root := s.Root
	panel := html.Div().Class("panel").
		Child(
			html.Div().Class("rm-sel-header").
				Child(
					html.Div().
						Child(
							html.Strong().Text(room.Name),
							html.Span().Class("rm-room-range").Text(" ("+room.Code+")"),
						),
					html.Button().Class("btn", "btn-sm").Text("✕").
						OnClick(func(dom.Event) {
							s.SelectedRoomID = ""
							root.Refresh()
						}),
				),
			html.Span().Class("field").Text("Horario semanal de este box:"),
		)

	// List weekly shifts
	shiftsList := html.Div().Class("rm-shifts-list")
	for _, sh := range root.Shifts {
		if sh.RoomID == room.ID {
			shift := sh
			dayName := dayOfWeekName(shift.DayOfWeek)
			row := html.Div().Class("slot", "occ").
				Child(
					html.Strong().Text(dayName+": "),
					html.Span().Text(shift.OccupantLabel+" ("+shift.StartHour+" - "+shift.EndHour+")"),
					html.Button().Class("btn", "btn-sm", "btn-danger").Text("✕").
						OnClick(func(dom.Event) {
							var updated []ShiftData
							for _, item := range root.Shifts {
								if item.ID != shift.ID {
									updated = append(updated, item)
								}
							}
							root.Shifts = updated
							root.Refresh()
						}),
				)
			shiftsList.Child(row)
		}
	}
	panel.Child(shiftsList)

	// Quick shift assign form
	var selectedDocID string
	if len(root.Doctors) > 0 {
		selectedDocID = root.Doctors[0].ID
	}
	assignForm := html.Div().Class("rm-assign-box").
		Child(
			html.Span().Class("field").Text("Asignar turno a profesional:"),
			dom.NewElement("select").Class("input").
				Child(func() []dom.Component {
					var opts []dom.Component
					for _, doc := range root.Doctors {
						opts = append(opts, html.Option(doc.ID, doc.Name+" ("+doc.Specialty+")"))
					}
					return opts
				}()...),
			html.Button().Class("btn", "btn-primary", "btn-sm").
				Text("+ Asignar lunes mañana").
				OnClick(func(dom.Event) {
					docName := "Dr. Juan Pérez"
					for _, d := range root.Doctors {
						if d.ID == selectedDocID {
							docName = d.Name
							break
						}
					}
					newShift := ShiftData{
						ID:            root.NewID(),
						RoomID:        room.ID,
						OccupantID:    selectedDocID,
						OccupantLabel: docName,
						DayOfWeek:     1, // Lunes
						StartHour:     "09:00",
						EndHour:       "13:00",
					}
					root.Shifts = append(root.Shifts, newShift)
					root.Refresh()
				}),
		)
	panel.Child(assignForm)

	return panel
}

func (s *Stage4Operacion) renderProfessionalsPanel() *dom.Element {
	root := s.Root
	panel := html.Div().Class("panel").
		Child(
			html.Span().Class("rm-panel-title").Text("Profesionales"),
			html.Input("search").Class("input").Attr("placeholder", "Buscar profesional...").
				OnChange(func(e dom.Event) {
					// Search filter
				}),
		)

	for _, doc := range root.Doctors {
		doctor := doc
		item := html.Button().Class("pro").
			Child(
				html.Span().Class("av", "av-m").Text(doctor.Initials),
				html.Div().Class("rm-pro-info").
					Child(
						html.Strong().Text(doctor.Name),
						html.Span().Class("field").Text(doctor.Specialty),
						html.Div().Class("meter").
							Child(html.Span().Attr("style", fmt.Sprintf("width: %d%%;", doctor.WeeklyOccupancy))),
					),
			).
			OnClick(func(dom.Event) {
				s.SelectedDoctorID = doctor.ID
				root.Refresh()
			})
		panel.Child(item)
	}

	return panel
}

func (s *Stage4Operacion) renderAlertsPanel() *dom.Element {
	root := s.Root
	panel := html.Div().Class("panel").
		Child(html.Span().Class("rm-panel-title").Text("Alertas operacionales"))

	hasAlerts := false
	for _, art := range root.Artifacts {
		if art.Status != StatusOk {
			hasAlerts = true
			alert := art
			badgeCls := "adot warn"
			if alert.Status == StatusCrit {
				badgeCls = "adot crit"
			}
			item := html.Div().Class("room-item").
				Child(
					html.Span().Class(badgeCls),
					html.Div().Class("rm-room-info").
						Child(
							html.Strong().Text(alert.Code+" · "+artifactKindLabel(alert.Kind)),
							html.Span().Class("field").Text(alert.Reason),
						),
				)
			panel.Child(item)
		}
	}

	if !hasAlerts {
		panel.Child(html.Span().Class("field").Text("No hay alertas activas."))
	}

	return panel
}

func (s *Stage4Operacion) isRoomOccupied(roomID string, day int, startHour string) bool {
	for _, sh := range s.Root.Shifts {
		if sh.RoomID == roomID && sh.DayOfWeek == day && sh.StartHour == startHour {
			return true
		}
	}
	return false
}

func (s *Stage4Operacion) getOccupantName(roomID string, day int, startHour string) string {
	for _, sh := range s.Root.Shifts {
		if sh.RoomID == roomID && sh.DayOfWeek == day && sh.StartHour == startHour {
			return sh.OccupantLabel
		}
	}
	return ""
}

func dayOfWeekName(d int) string {
	switch d {
	case 1:
		return "Lunes"
	case 2:
		return "Martes"
	case 3:
		return "Miércoles"
	case 4:
		return "Jueves"
	case 5:
		return "Viernes"
	case 6:
		return "Sábado"
	default:
		return "Domingo"
	}
}
