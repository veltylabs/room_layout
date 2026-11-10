package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/storage"
)

// roomColumns son las columnas que cambia SaveRoom al actualizar (todas menos id y tenant_id).
var roomColumns = []string{Room_.RoomType, Room_.Cells, Room_.FloorId, Room_.Code, Room_.Name, Room_.Notes, Room_.IsActive, Room_.UpdatedAt}

// SaveRoom guarda o actualiza una sala.
//
// Ejemplo: room, err := m.SaveRoom(room)
func (m *Module) SaveRoom(r Room) (Room, error) {
	store, err := m.storeFor(r.TenantId)
	if err != nil {
		return Room{}, err
	}
	r.Code = fmt.TrimSpace(r.Code)
	r.Name = fmt.TrimSpace(r.Name)
	isNew := r.Id == ""
	if isNew {
		r.Id = m.deps.IDs.NewID()
	}
	if err := model.ValidateFields(byte(model.Create), &r); err != nil {
		return Room{}, err
	}
	r.UpdatedAt = m.now()
	err = store.Atomic(func(tx crud.Tx) error {
		if err := checkCodeFree(tx, r.Code, r.Id); err != nil {
			return err
		}
		if isNew {
			return tx.Create(&r)
		}
		return tx.UpdateFields(&r, r.Id, roomColumns)
	})
	if err != nil {
		return Room{}, err
	}
	m.publish(TopicRoomSaved, &r)
	return r, nil
}

// GetRoom responde una sala por su id.
//
// Ejemplo: room, err := m.GetRoom(tenantID, id)
func (m *Module) GetRoom(tenantID, id string) (Room, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return Room{}, err
	}
	if id == "" {
		return Room{}, ErrNotFound
	}
	var r Room
	if err := store.ReadOne(&r, id); err != nil {
		return Room{}, notFound(err, ErrNotFound)
	}
	return r, nil
}

// ListRooms responde todas las salas de un tenant (opcionalmente solo las activas), ordenadas por código.
//
// Ejemplo: rooms, err := m.ListRooms(tenantID, true)
func (m *Module) ListRooms(tenantID string, activeOnly bool) ([]Room, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	var conds []storage.Condition
	if activeOnly {
		conds = append(conds, storage.Eq(Room_.IsActive, true))
	}
	rows, err := allRows(store, func() model.Model { return &Room{} }, conds...)
	if err != nil {
		return nil, err
	}
	var res []Room
	for _, r := range rows {
		res = append(res, *r.(*Room))
	}
	sortRoomsByCode(res)
	return res, nil
}

// DeactivateRoom desactiva una sala (is_active = false).
//
// Ejemplo: err := m.DeactivateRoom(tenantID, id)
func (m *Module) DeactivateRoom(tenantID, id string) error {
	r, err := m.GetRoom(tenantID, id)
	if err != nil {
		return err
	}
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	r.IsActive = false
	r.UpdatedAt = m.now()
	if err := store.UpdateFields(&r, id, []string{Room_.IsActive, Room_.UpdatedAt}); err != nil {
		return err
	}
	m.publish(TopicRoomDeactivated, &r)
	return nil
}
