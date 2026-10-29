package roomlayout

import (
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)

type CategoryReader interface {
	CategoryOptions(tenantID string) ([]fmt.KeyValue, error)
}

type OccupantReader interface {
	OccupantOptions(tenantID string) ([]fmt.KeyValue, error)
}

type BoundsReader interface {
	GetDayBounds(date time.Unix) (time.DayBounds, error)
}

type Deps struct {
	IDs        model.IDGenerator // requerido
	Categories CategoryReader    // requerido
	Occupants  OccupantReader    // requerido
	TenantID   string            // requerido
	Timezone   string            // requerido
	Bounds     BoundsReader      // opcional
	Publisher  events.Publisher  // opcional
	Clock      func() time.Nano  // opcional
}

type Module struct {
	db       *orm.DB
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
	errDepsTimezone   = "room_layout: Deps.Timezone"
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
	if deps.Timezone == "" {
		return nil, fmt.Err(errDepsTimezone, "is required")
	}
	if deps.Clock == nil {
		deps.Clock = time.NowNano
	}
	return &Module{
		db:       db,
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
