package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/storage"
)

// Search es la regla de negocio de leer room_category: con room_id en el filtro, solo los vínculos de esa sala.
func (p *RoomCategory) Search(a crud.CurrentUser, filter *RoomCategory, text string, m *Module) ([]storage.Condition, error) {
	if filter.RoomId == "" {
		return nil, nil
	}
	return []storage.Condition{storage.Eq(RoomCategory_.RoomId, filter.RoomId)}, nil
}
