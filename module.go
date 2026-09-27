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
	GetDayBounds(date int64) (time.DayBounds, error)
}

type Deps struct {
	IDs        model.IDGenerator // requerido
	Categories CategoryReader    // requerido
	Occupants  OccupantReader    // requerido
	TenantID   string            // requerido
	Timezone   string            // requerido
	Bounds     BoundsReader      // opcional
	Publisher  events.Publisher  // opcional
	Clock      func() int64      // opcional
}

type Module struct {
	db       *orm.DB
	deps     Deps
	tenantID string
}

func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("room_layout: Deps.IDs is required")
	}
	if deps.Categories == nil {
		return nil, fmt.Err("room_layout: Deps.Categories is required")
	}
	if deps.Occupants == nil {
		return nil, fmt.Err("room_layout: Deps.Occupants is required")
	}
	if deps.TenantID == "" {
		return nil, fmt.Err("room_layout: Deps.TenantID is required")
	}
	if deps.Timezone == "" {
		return nil, fmt.Err("room_layout: Deps.Timezone is required")
	}
	if deps.Clock == nil {
		deps.Clock = time.Now
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

func (m *Module) now() int64 {
	return m.deps.Clock()
}
