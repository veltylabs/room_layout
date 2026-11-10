package roomlayout

import "webtyp.com/crud"

// Delete es la regla de negocio de borrar room_shift: sus cancelaciones se borran primero, en la misma transacción. El
// receptor es la fila guardada.
func (p *RoomShift) Delete(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := deleteShiftCancellations(tx, p.Id); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
