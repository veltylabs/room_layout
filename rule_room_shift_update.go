package roomlayout

import "webtyp.com/crud"

// Update es la regla de negocio de actualizar room_shift: mezcla los campos enviados (fields) en el turno guardado, aplica
// todas las reglas al resultado y escribe las columnas calculadas. Mover un turno con fecha a otra sala es un Update de
// ["room_id"].
func (next *RoomShift) Update(a crud.CurrentUser, tx crud.Tx, old *RoomShift, fields []string, m *Module) (crud.RuleAnswer, error) {
	merged := *old
	if err := copyShiftFields(&merged, next, fields); err != nil {
		return 0, err
	}
	if err := m.checkShift(tx, m.tenantID, &merged); err != nil {
		return 0, err
	}
	merged.UpdatedAt = m.now()
	if err := tx.UpdateFields(&merged, old.Id, shiftColumns); err != nil {
		return 0, err
	}
	// crud escribe next con las columnas de fields después de la regla: debe quedar igual a lo ya guardado
	if err := copyShiftFields(next, &merged, fields); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
