package ui

import (
	"webtyp.com/components/cellgrid"
	"webtyp.com/components/stepindicator"
	"webtyp.com/dom"
	"webtyp.com/html"
)

type Stage1Planta struct {
	dom.Element
	FloorID string
}

func (s *Stage1Planta) Render() *dom.Element {
	stepper := &stepindicator.StepIndicator{
		Steps: []stepindicator.Step{
			{Label: "1"}, {Label: "2"}, {Label: "3"}, {Label: "4"},
		},
		Active: 0,
	}
	grid := &cellgrid.CellGrid{
		Cols: 10,
		Rows: 10,
	}
	return html.Div().Child(stepper, grid)
}
