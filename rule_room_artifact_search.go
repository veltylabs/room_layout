package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/storage"
)

// Search es la regla de negocio de leer room_artifact: con room_id en el filtro, solo los artefactos de esa sala.
func (p *RoomArtifact) Search(a crud.CurrentUser, filter *RoomArtifact, text string, m *Module) ([]storage.Condition, error) {
	if filter.RoomId == "" {
		return nil, nil
	}
	return []storage.Condition{storage.Eq(RoomArtifact_.RoomId, filter.RoomId)}, nil
}
