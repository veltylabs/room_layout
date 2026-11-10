package roomlayout

import "webtyp.com/crud"

// Update es la regla de negocio de actualizar room_equipment: si cambia el nombre se recorta; updated_at siempre.
func (next *Equipment) Update(a crud.CurrentUser, tx crud.Tx, old *Equipment, fields []string, m *Module) (crud.RuleAnswer, error) {
	if containsString(fields, Equipment_.Name) {
		name, err := trimmed(Equipment_.Name, next.Name)
		if err != nil {
			return 0, err
		}
		next.Name = name
	}
	if err := tx.UpdateFields(&Equipment{UpdatedAt: m.now()}, old.Id, []string{Equipment_.UpdatedAt}); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
