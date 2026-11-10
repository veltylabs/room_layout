package roomlayout

import "webtyp.com/crud"

// Create es la regla de negocio de crear room_equipment: el nombre se recorta y no puede quedar vacío.
func (p *Equipment) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	name, err := trimmed(Equipment_.Name, p.Name)
	if err != nil {
		return 0, err
	}
	p.Name = name
	p.UpdatedAt = m.now()
	return crud.ApplyAndPublish, nil
}
