package ui

import (
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/dom"
	"webtyp.com/html"
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
}

func (r *RootView) Render() *dom.Element {
	return html.Div().Child(&Stage1Planta{})
}

func Browser(caller router.Caller, ids model.IDGenerator, tenantID string, opts ...Option) (platformd.UIModule, error) {
	o := options{label: DefaultLabel}
	for _, opt := range opts {
		opt(&o)
	}

	return platformd.NewUIModule(ID, o.label, Icon(ID), &RootView{}), nil
}
