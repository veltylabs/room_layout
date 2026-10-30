package ui

import (
	"webtyp.com/components/stepindicator"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/layout"
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
	notifier layout.Notifier // los avisos de la plataforma: «Guardado» o el error al leer o guardar

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
	r.reload()
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

// header es el encabezado de las etapas de diseño: los cuatro pasos y, a la derecha, lo que la
// etapa agregue (el selector de piso).
func (r *RootView) header(active int, actions ...dom.Component) *dom.Element {
	stepper := &stepindicator.StepIndicator{
		Steps:    []stepindicator.Step{{Label: "Floor plan"}, {Label: "Spaces"}, {Label: "Equipment"}, {Label: "Operation"}},
		Active:   active,
		OnChange: r.GoToStage,
	}
	return html.Header().Class(clsHeader).Child(stepper, html.Div().Class(clsActions).Child(actions...))
}

// shownFloors son los índices de los pisos que se dibujan: solo el activo, uno a la vez.
func (r *RootView) shownFloors() []int {
	if len(r.Floors) == 0 {
		return nil
	}
	if r.ActiveFloor < 0 || r.ActiveFloor >= len(r.Floors) {
		r.ActiveFloor = 0
	}
	return []int{r.ActiveFloor}
}

// floorNav es el selector de piso (◀ 1 de 2 ▶): va en la cabecera de la tarjeta del piso, junto a
// su nombre, igual en las cuatro etapas. La tarjeta se rehace al cambiar de piso.
func (r *RootView) floorNav() *dom.Element {
	nav := html.Div().Class(clsNav)
	if len(r.Floors) < 2 {
		return nav
	}
	step := func(delta int, glyph string) *dom.Element {
		next := r.ActiveFloor + delta
		btn := html.Button().Class(clsButton).Attr("type", "button").Text(glyph)
		if next < 0 || next >= len(r.Floors) {
			return btn.Attr("disabled", "true")
		}
		return btn.OnClick(func(dom.Event) {
			r.ActiveFloor = next
			r.RebuildAllFloors()
		})
	}
	return nav.Child(
		step(-1, "◀"),
		html.Span().Text(lang.Translate(r.ActiveFloor+1, "of", len(r.Floors)).String()),
		step(1, "▶"),
	)
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

// AddFloor crea un piso nuevo y lo guarda en el momento; el servidor le pone su id.
//
// Ejemplo: mapa.AddFloor("Piso 3")
func (r *RootView) AddFloor(name string) {
	r.ActiveFloor = len(r.Floors)
	r.saveFloor(FloorData{Name: name, Position: len(r.Floors) + 1})
}

func (r *RootView) Render() *dom.Element {
	if r.ActiveStageSig == nil {
		r.Init(nil)
	}

	container := html.Div().Class(NameRoomMap.Root().String())

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
