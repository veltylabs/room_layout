package tests

import (
	"testing"

	"github.com/veltylabs/room_layout/ui"
	"webtyp.com/dom"
)

// Volver a una etapa la vuelve a dibujar: nada de lo dibujado antes se reutiliza (un elemento
// tiene un solo padre; reutilizarlo hacía caer la app al entrar por segunda vez a «Espacios»).
func TestStages_RenderTwice(t *testing.T) {
	m := ui.NewMap(nil)
	m.Init(nil)
	stages := []struct {
		name  string
		stage interface{ Render() *dom.Element }
	}{
		{"planta", &ui.Stage1Planta{Root: m}},
		{"espacios", &ui.Stage2Espacios{Root: m}},
		{"artefactos", &ui.Stage3Artefactos{Root: m}},
		{"operacion", &ui.Stage4Operacion{Root: m}},
	}
	for _, st := range stages {
		for i := 0; i < 2; i++ {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("%s: render %d panicked: %v", st.name, i+1, r)
					}
				}()
				drawAll(st.stage.Render())
			}()
		}
	}
}

// drawAll dibuja como el navegador: cada componente anidado se inicia y se dibuja.
func drawAll(c dom.Component) {
	if in, ok := c.(interface{ Init(dom.Ctx) }); ok {
		if _, isEl := c.(*dom.Element); !isEl {
			in.Init(nil)
		}
	}
	if vr, ok := c.(interface{ Render() *dom.Element }); ok {
		if _, isEl := c.(*dom.Element); !isEl {
			drawAll(vr.Render())
		}
	}
	if el, ok := c.(interface{ Children() []dom.Component }); ok {
		for _, ch := range el.Children() {
			drawAll(ch)
		}
	}
}
