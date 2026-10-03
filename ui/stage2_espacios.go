package ui

import (
	"webtyp.com/components/stepindicator"
	"webtyp.com/dom"
	"webtyp.com/html"
)

type Stage2Espacios struct {
	dom.Element
}

func (s *Stage2Espacios) Render() *dom.Element {
	stepper := &stepindicator.StepIndicator{
		Steps: []stepindicator.Step{
			{Label: "1"}, {Label: "2"}, {Label: "3"}, {Label: "4"},
		},
		Active: 1,
	}
	return html.Div().Child(stepper)
}
