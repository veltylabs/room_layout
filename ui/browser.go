package ui

import (
	"webtyp.com/components/decktabs"
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

	tabBoard := newBoardTab(caller, tenantID)
	tabFree := newFreeSearchTab(caller, tenantID)
	tabShifts := newShiftsTab(caller, ids, tenantID)

	tabRooms, err := newRoomsTab(caller, ids, tenantID)
	if err != nil {
		return nil, err
	}

	tabAssign := newAssignTab(caller, tenantID)

	tabCatalogs, err := catalogsPanel(caller, ids)
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

	return platformd.NewUIModule(ID, o.label, Icon(ID), tabs), nil
}
