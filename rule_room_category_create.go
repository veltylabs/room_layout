package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Create es la regla de negocio de crear room_category: la sala existe y la categoría es una de las que entrega la app.
func (p *RoomCategory) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := checkRoomExists(tx, p.RoomId); err != nil {
		return 0, &model.FieldError{Field: RoomCategory_.RoomId, Err: err}
	}
	options, err := m.deps.Categories.CategoryOptions(m.tenantID)
	if err != nil {
		return 0, err
	}
	if !categoryIn(options, p.CategoryId) {
		return 0, &model.FieldError{Field: RoomCategory_.CategoryId, Err: ErrUnknownCategory}
	}
	return crud.ApplyAndPublish, nil
}
