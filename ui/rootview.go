package ui

import (
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
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

	verCount int
	VerSig   *dom.SignalString

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
