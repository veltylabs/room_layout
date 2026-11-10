package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/storage"
)

// Search es la regla de negocio de leer room_shift: con room_id en el filtro, solo los turnos de esa sala.
func (p *RoomShift) Search(a crud.CurrentUser, filter *RoomShift, text string, m *Module) ([]storage.Condition, error) {
	if filter.RoomId == "" {
		return nil, nil
	}
	return []storage.Condition{storage.Eq(RoomShift_.RoomId, filter.RoomId)}, nil
}
