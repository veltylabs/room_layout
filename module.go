package roomlayout

import (
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)

type Deps struct {
	IDs        model.IDGenerator // requerido
	Categories CategoryReader    // requerido
	Occupants  OccupantReader    // requerido
	TenantID   string            // requerido
	Zone       time.Zone         // requerido
	Hours      HoursReader       // opcional
	Publisher  events.Publisher  // opcional
	Clock      func() time.Nano  // opcional
}

type Module struct {
	db       *orm.DB
	store    *orm.Store
	deps     Deps
	tenantID string
}

// Nombres de las dependencias que New exige: son dato (código), no texto a traducir;
// por eso son variables y no constantes (langc traduce las constantes que recibe fmt.Err).
var (
	errDepsIDs        = "room_layout: Deps.IDs"
	errDepsCategories = "room_layout: Deps.Categories"
	errDepsOccupants  = "room_layout: Deps.Occupants"
	errDepsTenantID   = "room_layout: Deps.TenantID"
	errDepsZone       = "room_layout: Deps.Zone"
)

func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err(errDepsIDs, "is required")
	}
	if deps.Categories == nil {
		return nil, fmt.Err(errDepsCategories, "is required")
	}
	if deps.Occupants == nil {
		return nil, fmt.Err(errDepsOccupants, "is required")
	}
	if deps.TenantID == "" {
		return nil, fmt.Err(errDepsTenantID, "is required")
	}
	if deps.Zone == (time.Zone{}) {
		return nil, fmt.Err(errDepsZone, "is required")
	}
	if deps.Clock == nil {
		deps.Clock = time.NowNano
	}
	return &Module{
		db:       db,
		store:    orm.TenantStore(db, deps.TenantID, deps.IDs),
		deps:     deps,
		tenantID: deps.TenantID,
	}, nil
}

func (m *Module) publish(topic string, payload model.Encodable) {
	if m.deps.Publisher != nil {
		m.deps.Publisher.Publish(events.Event{
			Topic:   topic,
			Payload: payload,
		})
	}
}

func (m *Module) now() time.Nano {
	return m.deps.Clock()
}
