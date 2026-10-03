package ui

import (
	"webtyp.com/components/stepindicator"
	"webtyp.com/dom"
	"webtyp.com/html"
)

type Stage3Artefactos struct {
	dom.Element
}

func (s *Stage3Artefactos) Render() *dom.Element {
	stepper := &stepindicator.StepIndicator{
		Steps: []stepindicator.Step{
			{Label: "1"}, {Label: "2"}, {Label: "3"}, {Label: "4"},
		},
		Active: 2,
	}
	return html.Div().Child(stepper)
}
