package roomlayout

import "webtyp.com/crud"

// Delete es la regla de negocio de borrar room_category: no se quita una categoría que un turno activo de la sala usa.
func (p *RoomCategory) Delete(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := checkCategoryNotInUse(tx, p.RoomId, p.CategoryId); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
