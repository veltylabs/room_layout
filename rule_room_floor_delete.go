package roomlayout

import "webtyp.com/crud"

// Delete es la regla de negocio de borrar room_floor: un piso con salas no se borra. El receptor es la fila guardada.
func (p *Floor) Delete(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := checkFloorFree(tx, p.Id); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
