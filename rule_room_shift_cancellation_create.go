package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Create es la regla de negocio de crear room_shift_cancellation: cancela una ocurrencia de un turno semanal. Con
// moved_to_room_id la ocurrencia se mueve: se crea, en la misma transacción, el turno con fecha en la sala destino. Repetir
// la cancelación de la misma ocurrencia no hace nada (crud no vuelve a correr la regla de una fila que ya existe).
func (p *RoomShiftCancellation) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	s, err := checkOccurrence(tx, p.ShiftId, p.SpecificDate)
	if err != nil {
		return 0, err
	}
	if p.MovedToRoomId != "" {
		if _, err := m.createMovedShift(tx, s, p.SpecificDate, p.MovedToRoomId); err != nil {
			return 0, &model.FieldError{Field: RoomShiftCancellation_.MovedToRoomId, Err: err}
		}
	}
	return crud.ApplyAndPublish, nil
}
