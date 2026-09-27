package ui

import (
	"webtyp.com/components/decktabs"
	"webtyp.com/dom"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/router"
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

func Browser(caller router.Caller, ids model.IDGenerator, tenantID string, opts ...Option) (platformd.UIModule, error) {
	o := options{label: DefaultLabel}
	for _, opt := range opts {
		opt(&o)
	}

	tabBoard := buildBoardTab(caller, tenantID)
	tabFree := buildFreeSearchTab(caller, ids, tenantID)
	tabShifts, err := buildShiftsTab(caller, ids, tenantID)
	if err != nil {
		return nil, err
	}
	tabRooms, err := buildRoomsTab(caller, ids)
	if err != nil {
		return nil, err
	}
	tabAssign := buildAssignTab(caller, tenantID)
	tabCatalogs, err := buildCatalogsTab(caller, ids)
	if err != nil {
		return nil, err
	}

	tabs := &decktabs.DeckTabs{
		Items: []decktabs.Item{
			{ID: "board", Label: "Tablero", Panel: tabBoard},
			{ID: "freesearch", Label: "Buscar libre", Panel: tabFree},
			{ID: "shifts", Label: "Turnos", Panel: tabShifts},
			{ID: "rooms", Label: "Espacios", Panel: tabRooms},
			{ID: "assign", Label: "Áreas y equipos", Panel: tabAssign},
			{ID: "catalogs", Label: "Niveles y equipamiento", Panel: tabCatalogs},
		},
	}

	deck := dom.NewElement("div").Child(tabs)
	return platformd.NewUIModule(ID, o.label, Icon(ID), deck), nil
}
