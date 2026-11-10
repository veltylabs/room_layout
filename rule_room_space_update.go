package roomlayout

import "webtyp.com/crud"

// Update es la regla de negocio de actualizar room_space: si cambian código o nombre se recortan y el código sigue libre.
func (next *Room) Update(a crud.CurrentUser, tx crud.Tx, old *Room, fields []string, m *Module) (crud.RuleAnswer, error) {
	if containsString(fields, Room_.Code) {
		code, err := trimmed(Room_.Code, next.Code)
		if err != nil {
			return 0, err
		}
		if err := checkCodeFree(tx, code, old.Id); err != nil {
			return 0, err
		}
		next.Code = code
	}
	if containsString(fields, Room_.Name) {
		name, err := trimmed(Room_.Name, next.Name)
		if err != nil {
			return 0, err
		}
		next.Name = name
	}
	if err := tx.UpdateFields(&Room{UpdatedAt: m.now()}, old.Id, []string{Room_.UpdatedAt}); err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
