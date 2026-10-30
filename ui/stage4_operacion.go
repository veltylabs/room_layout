package ui

import (
	"webtyp.com/components/segmentedcontrol"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/lang"
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
	sliderNodes      *dom.SignalNodes
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
			{Value: "plan", Label: "Plan"},
			{Value: "now", Label: "Now"},
		},
		Selected: s.Mode,
		OnChange: func(val string) {
			s.Mode = val
			root.Refresh()
		},
	}

	configBtn := html.Button().Class(clsButton).
		Child(
			html.Span().Class("icon-gear").Text("⚙"),
			html.Span().Text(lang.Translate("Configure map").String()),
		).
		OnClick(func(dom.Event) {
			root.GoToStage(0)
		})

	header := html.Header().Class(clsHeader).
		Child(
			modeSeg,
			html.Div().Class(clsActions).
				Child(configBtn),
		)

	// Toolbar
	toolbar := html.Div().Class("rm-toolbar")

	if s.Mode == "plan" {
		periodSeg := &segmentedcontrol.SegmentedControl{
			Options: []segmentedcontrol.Option{
				{Value: "week", Label: "Week"},
				{Value: "month", Label: "Month"},
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
				html.Span().Child(html.Span().Text(lang.Translate("Occupancy").String()+": "), html.Strong().Text(fmt.Sprint(occPct)+"%")),
				html.Span().Child(html.Strong().Text(fmt.Sprint(freeBlocks)), html.Span().Text(" "+lang.Translate("free blocks").String())),
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
				{Value: "morning", Label: "Morning"},
				{Value: "afternoon", Label: "Afternoon"},
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
			{Value: "occupancy", Label: "Occupancy"},
			{Value: "equipment", Label: "Operation"},
		},
		Selected: s.Layer,
		OnChange: func(val string) {
			s.Layer = val
			root.Refresh()
		},
	}
	toolbar.Child(layerSeg)

	columns := html.Div().Class(clsColumns)

	// Left: Content (Floor Map Canvas)
	content := html.Div().Class(clsContent)
	s.sliderNodes = dom.NewNodes(s.buildFloorCards()...)
	slider := html.Div().Class(clsSlider).BindChildren(s.sliderNodes)
	content.Child(slider)

	// Right: Aside (Selected Room Detail OR Doctors & Alerts)
	aside := html.Aside().Class(clsAside)

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
	footer := html.Footer().Class(clsFooter).
		Child(
			html.Button().Class(clsButton).Text("← "+lang.Translate("Equipment").String()).
				OnClick(func(dom.Event) {
					root.GoToStage(2)
				}),
			html.Span().Class(clsSummary).
				Text(lang.Translate("Operation").String()),
		)

	layout := html.Div().Class(clsStage).
		Child(header, footer, toolbar, columns)

	return layout
}

func (s *Stage4Operacion) RebuildFloors() {
	if s.sliderNodes != nil {
		s.sliderNodes.Set(s.buildFloorCards())
	}
}

func (s *Stage4Operacion) buildFloorCards() []*dom.Element {
	root := s.Root
	var cards []*dom.Element

	for _, i := range root.shownFloors() {
		fl := &root.Floors[i]

		floorCard := html.Article().Class(clsFloor).Key(fmt.Sprintf("floor-s4-%s-%dx%d", fl.ID, root.Cols, root.Rows))
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
				cellEl := html.Div().Class("cell")

				room := FindRoomByCell(root.Rooms, fl.ID, coord)
				cellEl.Attr("style", gridPos(rowIdx, colIdx)+planCell(fl, room, coord)+planContent+" position: relative;")
				if room == nil {
					gridEl.Child(cellEl)
					continue
				}

				gridEl.Child(cellEl)
			}
		}

		for _, outline := range root.roomOutlines(fl.ID, "", plain, nameNone) {
			gridEl.Child(outline)
		}
		// el botón de cada espacio cubre su bloque más grande: se toca el espacio entero
		for i := range root.Rooms {
			room := &root.Rooms[i]
			if room.FloorID != fl.ID {
				continue
			}
			if place, ok := roomPlace(room.Cells); ok {
				gridEl.Child(s.renderRoomOverlay(room, fl.ID).Attr("style", place+roomButton))
			}
		}
		floorCard.Child(fHeader, gridEl)
		cards = append(cards, floorCard)
	}

	return cards
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
			ov.Child(html.Span().Class("freep").Text(lang.Translate("Free").String()))
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
	panel := html.Div().Class(clsPanel).
		Child(
			html.Div().Class("rm-sel-header").
				Child(
					html.Div().
						Child(
							html.Strong().Text(room.Name),
							html.Span().Class(clsItemMeta).Text(" ("+room.Code+")"),
						),
					html.Button().Class(clsButton).Text("✕").
						OnClick(func(dom.Event) {
							s.SelectedRoomID = ""
							root.Refresh()
						}),
				),
			html.Span().Class(clsField).Text(lang.Translate("Weekly schedule of this room").String()+":"),
		)

	// List weekly shifts
	shiftsList := html.Div().Class("rm-shifts-list")
	for _, sh := range root.Shifts {
		if sh.RoomID == room.ID {
			shift := sh
			dayName := dayOfWeekName(shift.DayOfWeek)
			row := html.Div().Class("slot", "occ").
				Child(
					html.Strong().Text(lang.Translate(string(dayName)).String()+": "),
					html.Span().Text(shift.OccupantLabel+" ("+shift.StartHour+" - "+shift.EndHour+")"),
					html.Button().Class(clsDanger).Text("✕").
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
			html.Span().Class(clsField).Text(lang.Translate("Assign a shift to a professional").String()+":"),
			dom.NewElement("select").Class(clsInput).
				Child(func() []dom.Component {
					var opts []dom.Component
					for _, doc := range root.Doctors {
						opts = append(opts, html.Option(doc.ID, doc.Name+" ("+doc.Specialty+")"))
					}
					return opts
				}()...),
			html.Button().Class(clsPrimary).
				Text("+ "+lang.Translate("Assign Monday morning").String()).
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
	panel := html.Div().Class(clsPanel).
		Child(
			html.Span().Class(clsPanelTitle).Text(lang.Translate("Professionals").String()),
			html.Input("search").Class(clsInput).Attr("placeholder", lang.Translate("Search professional").String()).
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
						html.Span().Class(clsField).Text(doctor.Specialty),
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
	panel := html.Div().Class(clsPanel).
		Child(html.Span().Class(clsPanelTitle).Text(lang.Translate("Alerts").String()))

	hasAlerts := false
	for _, art := range root.Artifacts {
		if art.Status != StatusOk {
			hasAlerts = true
			alert := art
			badgeCls := "adot warn"
			if alert.Status == StatusCrit {
				badgeCls = "adot crit"
			}
			item := html.Div().Class(clsItem).
				Child(
					html.Span().Class(badgeCls),
					html.Div().Class(clsItemInfo).
						Child(
							html.Strong().Text(alert.Code+" · "+artifactKindLabel(alert.Kind)),
							html.Span().Class(clsField).Text(alert.Reason),
						),
				)
			panel.Child(item)
		}
	}

	if !hasAlerts {
		panel.Child(html.Span().Class(clsField).Text(lang.Translate("No active alerts").String()))
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

// dayOfWeekName es la clave en inglés del día; se muestra con lang.Translate.
func dayOfWeekName(d int) lang.Text {
	switch d {
	case 1:
		return "Monday"
	case 2:
		return "Tuesday"
	case 3:
		return "Wednesday"
	case 4:
		return "Thursday"
	case 5:
		return "Friday"
	case 6:
		return "Saturday"
	default:
		return "Sunday"
	}
}
