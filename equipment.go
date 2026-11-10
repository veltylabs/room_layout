package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/storage"
)

var equipmentColumns = []string{Equipment_.Name, Equipment_.DeviceId, Equipment_.UpdatedAt}

var artifactColumns = []string{RoomArtifact_.RoomId, RoomArtifact_.Kind, RoomArtifact_.Code, RoomArtifact_.Cell, RoomArtifact_.Status,
	RoomArtifact_.Reason, RoomArtifact_.NextMaintenance, RoomArtifact_.Items}

// SaveEquipment guarda o actualiza un equipamiento.
//
// Ejemplo: eq, err := m.SaveEquipment(eq)
func (m *Module) SaveEquipment(e Equipment) (Equipment, error) {
	store, err := m.storeFor(e.TenantId)
	if err != nil {
		return Equipment{}, err
	}
	e.Name = fmt.TrimSpace(e.Name)
	isNew := e.Id == ""
	if isNew {
		e.Id = m.deps.IDs.NewID()
	}
	if err := model.ValidateFields(byte(model.Create), &e); err != nil {
		return Equipment{}, err
	}
	e.UpdatedAt = m.now()
	err = store.Atomic(func(tx crud.Tx) error {
		if isNew {
			return tx.Create(&e)
		}
		return tx.UpdateFields(&e, e.Id, equipmentColumns)
	})
	if err != nil {
		return Equipment{}, err
	}
	m.publish(TopicEquipmentSaved, &e)
	return e, nil
}

// equipmentOn responde todo el equipamiento, por nombre.
func equipmentOn(tx crud.Tx) ([]Equipment, error) {
	rows, err := allRows(tx, func() model.Model { return &Equipment{} })
	if err != nil {
		return nil, err
	}
	var res []Equipment
	for _, r := range rows {
		res = append(res, *r.(*Equipment))
	}
	sortEquipmentByName(res)
	return res, nil
}

// ListEquipment responde la lista de equipamiento de un tenant ordenados por nombre.
//
// Ejemplo: list, err := m.ListEquipment(tenantID)
func (m *Module) ListEquipment(tenantID string) ([]Equipment, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	return equipmentOn(store)
}

// DeleteEquipment elimina un equipamiento y borra sus vínculos con salas asociadas.
//
// Ejemplo: err := m.DeleteEquipment(tenantID, id)
func (m *Module) DeleteEquipment(tenantID, id string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if id == "" {
		return ErrNotFound
	}
	err = store.Atomic(func(tx crud.Tx) error {
		if err := deleteEquipmentLinks(tx, id); err != nil {
			return err
		}
		return tx.Delete(&Equipment{}, id)
	})
	if err != nil {
		return err
	}
	m.publish(TopicEquipmentDeleted, &Equipment{Id: id, TenantId: tenantID})
	return nil
}

// ListRoomEquipment responde los equipamientos asociados a una sala.
//
// Ejemplo: list, err := m.ListRoomEquipment(tenantID, roomID)
func (m *Module) ListRoomEquipment(tenantID, roomID string) ([]Equipment, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	links, err := allRows(store, func() model.Model { return &RoomEquipment{} }, storage.Eq(RoomEquipment_.RoomId, roomID))
	if err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []Equipment{}, nil
	}
	all, err := equipmentOn(store)
	if err != nil {
		return nil, err
	}
	var res []Equipment
	for _, eq := range all {
		for _, l := range links {
			if l.(*RoomEquipment).EquipmentId == eq.Id {
				res = append(res, eq)
				break
			}
		}
	}
	return res, nil
}

// SetRoomEquipment actualiza la lista de equipamientos vinculados a una sala.
//
// Ejemplo: err := m.SetRoomEquipment(tenantID, roomID, eqIDs)
func (m *Module) SetRoomEquipment(tenantID, roomID string, equipmentIDs []string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if _, err := m.GetRoom(tenantID, roomID); err != nil {
		return err
	}
	return store.Atomic(func(tx crud.Tx) error {
		old, err := allRows(tx, func() model.Model { return &RoomEquipment{} }, storage.Eq(RoomEquipment_.RoomId, roomID))
		if err != nil {
			return err
		}
		for _, r := range old {
			if err := tx.Delete(&RoomEquipment{}, model.JoinKey(roomID, r.(*RoomEquipment).EquipmentId)); err != nil {
				return err
			}
		}
		for _, eqID := range equipmentIDs {
			if err := tx.Create(&RoomEquipment{RoomId: roomID, EquipmentId: eqID}); err != nil {
				return err
			}
		}
		return nil
	})
}

// SaveRoomArtifact guarda o actualiza un artefacto ubicado en el mapa de una sala.
//
// Ejemplo: artifact, err := m.SaveRoomArtifact(artifact)
func (m *Module) SaveRoomArtifact(a RoomArtifact) (RoomArtifact, error) {
	store, err := m.storeFor(a.TenantId)
	if err != nil {
		return RoomArtifact{}, err
	}
	isNew := a.Id == ""
	if isNew {
		a.Id = m.deps.IDs.NewID()
	}
	if err := model.ValidateFields(byte(model.Create), &a); err != nil {
		return RoomArtifact{}, err
	}
	err = store.Atomic(func(tx crud.Tx) error {
		if isNew {
			return tx.Create(&a)
		}
		return tx.UpdateFields(&a, a.Id, artifactColumns)
	})
	if err != nil {
		return RoomArtifact{}, err
	}
	return a, nil
}

// ListRoomArtifacts responde los artefactos de un tenant (opcionalmente filtrados por sala).
//
// Ejemplo: list, err := m.ListRoomArtifacts(tenantID, roomID)
func (m *Module) ListRoomArtifacts(tenantID string, roomID string) ([]RoomArtifact, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	var conds []storage.Condition
	if roomID != "" {
		conds = append(conds, storage.Eq(RoomArtifact_.RoomId, roomID))
	}
	rows, err := allRows(store, func() model.Model { return &RoomArtifact{} }, conds...)
	if err != nil {
		return nil, err
	}
	var res []RoomArtifact
	for _, r := range rows {
		res = append(res, *r.(*RoomArtifact))
	}
	return res, nil
}

// DeleteRoomArtifact elimina un artefacto de una sala por su id.
//
// Ejemplo: err := m.DeleteRoomArtifact(tenantID, id)
func (m *Module) DeleteRoomArtifact(tenantID, id string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if id == "" {
		return ErrNotFound
	}
	return store.Delete(&RoomArtifact{}, id)
}
