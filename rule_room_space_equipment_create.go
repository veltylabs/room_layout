package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Create es la regla de negocio de crear room_space_equipment: la sala existe.
func (p *RoomEquipment) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := checkRoomExists(tx, p.RoomId); err != nil {
		return 0, &model.FieldError{Field: RoomEquipment_.RoomId, Err: err}
	}
	return crud.ApplyAndPublish, nil
}
