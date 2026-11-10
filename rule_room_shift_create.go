package roomlayout

import "webtyp.com/crud"

// Create es la regla de negocio de crear room_shift: aplica todas las reglas de un turno (sala activa, horas, semanal o
// con fecha, categoría, ocupante, horario de la clínica, solapes) y lo completa con lo que decide el servidor.
func (p *RoomShift) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if err := m.checkShift(tx, m.tenantID, p); err != nil {
		return 0, err
	}
	p.UpdatedAt = m.now()
	return crud.ApplyAndPublish, nil
}
