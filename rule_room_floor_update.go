package roomlayout

import "webtyp.com/crud"

// Update es la regla de negocio de actualizar room_floor: si cambia el nombre se recorta; updated_at siempre.
func (next *Floor) Update(a crud.CurrentUser, tx crud.Tx, old *Floor, fields []string, m *Module) (crud.RuleAnswer, error) {
	if containsString(fields, Floor_.Name) {
		name, err := trimmed(Floor_.Name, next.Name)
		if err != nil {
			return 0, err
		}
		next.Name = name
	}
	if err := tx.UpdateFields(&Floor{UpdatedAt: m.now()}, old.Id, []string{Floor_.UpdatedAt}); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
