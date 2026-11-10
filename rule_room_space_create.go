package roomlayout

import "webtyp.com/crud"

// Create es la regla de negocio de crear room_space: código y nombre recortados, y un código que ninguna otra sala usa.
func (p *Room) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	code, err := trimmed(Room_.Code, p.Code)
	if err != nil {
		return 0, err
	}
	name, err := trimmed(Room_.Name, p.Name)
	if err != nil {
		return 0, err
	}
	if err := checkCodeFree(tx, code, ""); err != nil {
		return 0, err
	}
	p.Code, p.Name = code, name
	p.UpdatedAt = m.now()
	return crud.ApplyAndPublish, nil
}
