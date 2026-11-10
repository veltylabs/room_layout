package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/storage"
)

// Search es la regla de negocio de leer room_space_equipment: con room_id en el filtro, solo los vínculos de esa sala.
func (p *RoomEquipment) Search(a crud.CurrentUser, filter *RoomEquipment, text string, m *Module) ([]storage.Condition, error) {
	if filter.RoomId == "" {
		return nil, nil
	}
	return []storage.Condition{storage.Eq(RoomEquipment_.RoomId, filter.RoomId)}, nil
}
