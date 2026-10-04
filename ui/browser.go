package ui

import (
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/widget"
)

const (
	NameRoomMap = widget.Name("room-map")
)

type options struct {
	label string
}

type Option func(*options)

func WithLabel(label string) Option {
	return func(o *options) {
		o.label = label
	}
}

type RootView struct {
	dom.Element

	caller   router.Caller
	ids      model.IDGenerator
	tenantID string

	ActiveStage    int // 0=Planta, 1=Espacios, 2=Artefactos, 3=Operación
	ActiveStageSig *dom.SignalString
	ActiveFloor    int

	verCount       int
	VerSig         *dom.SignalString

	Cols int
	Rows int

	Floors    []FloorData
	Rooms     []RoomData
	Artifacts []ArtifactData
	Shifts    []ShiftData
	Doctors   []DoctorData

	stage1 *Stage1Planta
	stage2 *Stage2Espacios
	stage3 *Stage3Artefactos
	stage4 *Stage4Operacion
}

func (r *RootView) WidgetName() widget.Name { return NameRoomMap }
func (r *RootView) WidgetKind() widget.Kind { return widget.Tabs }

func (r *RootView) Init(_ dom.Ctx) {
	if r.ActiveStageSig == nil {
		r.ActiveStageSig = dom.NewString(fmt.Sprint(r.ActiveStage))
	}
	if r.VerSig == nil {
		r.VerSig = dom.NewString("0")
	}
	r.stage1 = &Stage1Planta{Root: r}
	r.stage2 = &Stage2Espacios{Root: r}
	r.stage3 = &Stage3Artefactos{Root: r}
	r.stage4 = &Stage4Operacion{Root: r}
}

func (r *RootView) GoToStage(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx > 3 {
		idx = 3
	}
	r.ActiveStage = idx
	if r.ActiveStageSig != nil {
		r.ActiveStageSig.Set(fmt.Sprint(idx))
	}
	r.Refresh()
}

func (r *RootView) Refresh() {
	r.verCount++
	if r.VerSig != nil {
		r.VerSig.Set(fmt.Sprint(r.verCount))
	}
	if r.ActiveStageSig != nil {
		r.ActiveStageSig.Set(fmt.Sprint(r.ActiveStage))
	}
}

func (r *RootView) RebuildAllFloors() {
	switch r.ActiveStage {
	case 0:
		if r.stage1 != nil {
			r.stage1.RebuildFloors()
		}
	case 1:
		if r.stage2 != nil {
			r.stage2.RebuildFloors()
		}
	case 2:
		if r.stage3 != nil {
			r.stage3.RebuildFloors()
		}
	case 3:
		if r.stage4 != nil {
			r.stage4.RebuildFloors()
		}
	}
	r.Refresh()
}

func (r *RootView) NewID() string {
	if r.ids != nil {
		return r.ids.NewID()
	}
	return fmt.Sprint(r.Cols*1000 + len(r.Rooms)*100 + len(r.Artifacts) + 1)
}

func (r *RootView) AddFloor(name string) {
	f := FloorData{
		ID:       r.NewID(),
		Name:     name,
		Position: len(r.Floors) + 1,
	}
	r.Floors = append(r.Floors, f)
	r.ActiveFloor = len(r.Floors) - 1
	r.RebuildAllFloors()
}

func (r *RootView) SaveAll() {
	// Notifies or triggers persistence via caller or local state
}

func (r *RootView) Render() *dom.Element {
	if r.ActiveStageSig == nil {
		r.Init(nil)
	}

	container := html.Div().Class("rm-container")

	// Dynamic conditional rendering using dom.Show and derived signals
	showStage1 := dom.DeriveBool(func() bool { return r.ActiveStageSig.Get() == "0" })
	showStage2 := dom.DeriveBool(func() bool { return r.ActiveStageSig.Get() == "1" })
	showStage3 := dom.DeriveBool(func() bool { return r.ActiveStageSig.Get() == "2" })
	showStage4 := dom.DeriveBool(func() bool { return r.ActiveStageSig.Get() == "3" })

	container.Child(
		dom.Show(showStage1, func() *dom.Element { return r.stage1.Render() }),
		dom.Show(showStage2, func() *dom.Element { return r.stage2.Render() }),
		dom.Show(showStage3, func() *dom.Element { return r.stage3.Render() }),
		dom.Show(showStage4, func() *dom.Element { return r.stage4.Render() }),
	)

	return container
}

func Browser(caller router.Caller, ids model.IDGenerator, tenantID string, opts ...Option) (platformd.UIModule, error) {
	o := options{label: DefaultLabel}
	for _, opt := range opts {
		opt(&o)
	}

	// 1. Initial realistic floors
	var p1Cells []string
	for row := 0; row < 12; row++ {
		for col := 0; col < 16; col++ {
			if row < 4 && col > 10 {
				continue // cutout / patio
			}
			p1Cells = append(p1Cells, FormatCell(row, col))
		}
	}

	var p2Cells []string
	for row := 0; row < 12; row++ {
		for col := 0; col < 16; col++ {
			p2Cells = append(p2Cells, FormatCell(row, col))
		}
	}

	f1ID := "f1"
	f2ID := "f2"
	if ids != nil {
		f1ID = ids.NewID()
		f2ID = ids.NewID()
	}

	floors := []FloorData{
		{
			ID:             f1ID,
			Name:           "Piso 1 · Recepción y Atención",
			Position:       1,
			HabitableCells: p1Cells,
		},
		{
			ID:             f2ID,
			Name:           "Piso 2 · Boxes Clínicos",
			Position:       2,
			HabitableCells: p2Cells,
		},
	}

	// 2. Initial rooms
	r1ID := "r1"
	r2ID := "r2"
	r3ID := "r3"
	r4ID := "r4"
	r5ID := "r5"
	if ids != nil {
		r1ID = ids.NewID()
		r2ID = ids.NewID()
		r3ID = ids.NewID()
		r4ID = ids.NewID()
		r5ID = ids.NewID()
	}

	rooms := []RoomData{
		{
			ID:       r1ID,
			FloorID:  f1ID,
			Code:     "B-101",
			Name:     "Consultorio 101",
			RoomType: RoomTypeBox,
			Cells:    []string{"B2", "B3", "C2", "C3"},
		},
		{
			ID:       r2ID,
			FloorID:  f1ID,
			Code:     "B-102",
			Name:     "Consultorio 102",
			RoomType: RoomTypeBox,
			Cells:    []string{"B5", "B6", "C5", "C6"},
		},
		{
			ID:       r3ID,
			FloorID:  f1ID,
			Code:     "B-103",
			Name:     "Consultorio 103",
			RoomType: RoomTypeBox,
			Cells:    []string{"E2", "E3", "F2", "F3"},
		},
		{
			ID:       r4ID,
			FloorID:  f1ID,
			Code:     "REC-01",
			Name:     "Recepción / Espera",
			RoomType: RoomTypeShared,
			Cells:    []string{"H2", "H3", "I2", "I3", "J2", "J3"},
		},
		{
			ID:       r5ID,
			FloorID:  f1ID,
			Code:     "BOD-01",
			Name:     "Bodega de Insumos",
			RoomType: RoomTypeStore,
			Cells:    []string{"L2", "L3", "M2", "M3"},
		},
	}

	// 3. Initial artifacts
	artifacts := []ArtifactData{
		{
			ID:      "a1",
			RoomID:  r1ID,
			FloorID: f1ID,
			Kind:    ArtifactPC,
			Code:    "PC-01",
			Cell:    "B2",
			Status:  StatusOk,
		},
		{
			ID:      "a2",
			RoomID:  r1ID,
			FloorID: f1ID,
			Kind:    ArtifactAgenda,
			Code:    "AG-01",
			Cell:    "C2",
			Status:  StatusOk,
		},
		{
			ID:      "a3",
			RoomID:  r2ID,
			FloorID: f1ID,
			Kind:    ArtifactPrinter,
			Code:    "IMP-01",
			Cell:    "B5",
			Status:  StatusWarn,
			Reason:  "Falta tóner magenta",
		},
		{
			ID:      "a4",
			RoomID:  r3ID,
			FloorID: f1ID,
			Kind:    ArtifactECG,
			Code:    "ECG-01",
			Cell:    "E2",
			Status:  StatusOk,
		},
		{
			ID:      "a5",
			RoomID:  r3ID,
			FloorID: f1ID,
			Kind:    ArtifactGastro,
			Code:    "GAS-01",
			Cell:    "F3",
			Status:  StatusCrit,
			Reason:  "Falla sensor de presión",
		},
		{
			ID:      "a6",
			RoomID:  r5ID,
			FloorID: f1ID,
			Kind:    ArtifactSupply,
			Code:    "INS-01",
			Cell:    "L2",
			Status:  StatusOk,
		},
	}

	// 4. Initial shifts
	shifts := []ShiftData{
		{
			ID:            "s1",
			RoomID:        r1ID,
			OccupantID:    "doc1",
			OccupantLabel: "Dr. Juan Pérez",
			DayOfWeek:     1, // Lun
			StartHour:     "09:00",
			EndHour:       "13:00",
		},
		{
			ID:            "s2",
			RoomID:        r1ID,
			OccupantID:    "doc1",
			OccupantLabel: "Dr. Juan Pérez",
			DayOfWeek:     2, // Mar
			StartHour:     "09:00",
			EndHour:       "13:00",
		},
		{
			ID:            "s3",
			RoomID:        r2ID,
			OccupantID:    "doc2",
			OccupantLabel: "Dra. María González",
			DayOfWeek:     3, // Mié
			StartHour:     "09:00",
			EndHour:       "13:00",
		},
		{
			ID:            "s4",
			RoomID:        r2ID,
			OccupantID:    "doc2",
			OccupantLabel: "Dra. María González",
			DayOfWeek:     4, // Jue
			StartHour:     "14:00",
			EndHour:       "18:00",
		},
		{
			ID:            "s5",
			RoomID:        r3ID,
			OccupantID:    "doc3",
			OccupantLabel: "Dr. Carlos Rojas",
			DayOfWeek:     5, // Vie
			StartHour:     "09:00",
			EndHour:       "13:00",
		},
	}

	// 5. Initial doctors
	doctors := []DoctorData{
		{
			ID:              "doc1",
			Name:            "Dr. Juan Pérez",
			Specialty:       "Medicina General",
			Initials:        "JP",
			WeeklyOccupancy: 80,
		},
		{
			ID:              "doc2",
			Name:            "Dra. María González",
			Specialty:       "Cardiología",
			Initials:        "MG",
			WeeklyOccupancy: 60,
		},
		{
			ID:              "doc3",
			Name:            "Dr. Carlos Rojas",
			Specialty:       "Ginecología",
			Initials:        "CR",
			WeeklyOccupancy: 45,
		},
	}

	root := &RootView{
		caller:      caller,
		ids:         ids,
		tenantID:    tenantID,
		ActiveStage: 0,
		ActiveFloor: 0,
		Cols:        20,
		Rows:        14,
		Floors:      floors,
		Rooms:       rooms,
		Artifacts:   artifacts,
		Shifts:      shifts,
		Doctors:     doctors,
	}

	return platformd.NewUIModule(ID, o.label, Icon(ID), root), nil
}
