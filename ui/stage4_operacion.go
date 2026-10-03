package ui

import (
	"webtyp.com/components/segmentedcontrol"
	"webtyp.com/dom"
	"webtyp.com/html"
)

type Stage4Operacion struct {
	dom.Element
}

func (s *Stage4Operacion) Render() *dom.Element {
	seg := &segmentedcontrol.SegmentedControl{
		Options: []segmentedcontrol.Option{
			{Value: "plan", Label: "Planificar"},
			{Value: "now", Label: "Ahora"},
		},
		Selected: "plan",
	}
	return html.Div().Child(seg)
}
