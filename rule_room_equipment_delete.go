package roomlayout

import "webtyp.com/crud"

// Delete es la regla de negocio de borrar room_equipment: sus vínculos con salas se borran primero, en la misma transacción.
func (p *Equipment) Delete(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := deleteEquipmentLinks(tx, p.Id); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
