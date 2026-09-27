package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
)

func (m *Module) SaveEquipment(e Equipment) (Equipment, error) {
	if e.TenantId == "" {
		return Equipment{}, ErrTenantRequired
	}
	e.Name = fmt.TrimSpace(e.Name)
	isNew := e.Id == ""
	if isNew {
		e.Id = m.deps.IDs.NewID()
	}

	if err := model.ValidateFields(byte(model.Create), &e); err != nil {
		return Equipment{}, &ValidationError{Err: err}
	}

	e.UpdatedAt = m.now() / 1_000_000_000
	if isNew {
		if err := m.db.Create(&e); err != nil {
			return Equipment{}, err
		}
	} else {
		if err := m.db.Update(&e, orm.Eq("id", e.Id), orm.Eq("tenant_id", e.TenantId)); err != nil {
			return Equipment{}, err
		}
	}

	m.publish(TopicEquipmentSaved, &e)
	return e, nil
}

func (m *Module) ListEquipment(tenantID string) ([]Equipment, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	var res []Equipment
	err := m.db.Query(&Equipment{}).
		Where("tenant_id").Eq(tenantID).
		OrderBy("name").Asc().
		ReadAll(func() model.Model { return &Equipment{} }, func(row model.Model) {
			if item, ok := row.(*Equipment); ok {
				res = append(res, *item)
			}
		})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Module) DeleteEquipment(tenantID, id string) error {
	if tenantID == "" {
		return ErrTenantRequired
	}
	if id == "" {
		return ErrNotFound
	}
	err := m.db.Tx(func(tx *orm.DB) error {
		re := RoomEquipment{TenantId: tenantID, EquipmentId: id}
		if err := tx.Delete(&re, orm.Eq("tenant_id", tenantID), orm.Eq("equipment_id", id)); err != nil {
			return err
		}
		eq := Equipment{Id: id, TenantId: tenantID}
		return tx.Delete(&eq, orm.Eq("id", id), orm.Eq("tenant_id", tenantID))
	})
	if err != nil {
		return err
	}

	m.publish(TopicEquipmentDeleted, &Equipment{Id: id, TenantId: tenantID})
	return nil
}

func (m *Module) ListRoomEquipment(tenantID, roomID string) ([]Equipment, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	var eqIDs []string
	err := m.db.Query(&RoomEquipment{}).
		Where("tenant_id").Eq(tenantID).
		Where("room_id").Eq(roomID).
		ReadAll(func() model.Model { return &RoomEquipment{} }, func(row model.Model) {
			if item, ok := row.(*RoomEquipment); ok {
				eqIDs = append(eqIDs, item.EquipmentId)
			}
		})
	if err != nil {
		return nil, err
	}
	if len(eqIDs) == 0 {
		return []Equipment{}, nil
	}

	allEq, err := m.ListEquipment(tenantID)
	if err != nil {
		return nil, err
	}

	var res []Equipment
	for _, eq := range allEq {
		for _, id := range eqIDs {
			if eq.Id == id {
				res = append(res, eq)
				break
			}
		}
	}
	return res, nil
}

func (m *Module) SetRoomEquipment(tenantID, roomID string, equipmentIDs []string) error {
	if tenantID == "" {
		return ErrTenantRequired
	}
	if _, err := m.GetRoom(tenantID, roomID); err != nil {
		return err
	}

	return m.db.Tx(func(tx *orm.DB) error {
		dummy := RoomEquipment{TenantId: tenantID, RoomId: roomID}
		if err := tx.Delete(&dummy, orm.Eq("tenant_id", tenantID), orm.Eq("room_id", roomID)); err != nil {
			return err
		}
		for _, eqID := range equipmentIDs {
			re := RoomEquipment{
				TenantId:    tenantID,
				RoomId:      roomID,
				EquipmentId: eqID,
			}
			if err := tx.Create(&re); err != nil {
				return err
			}
		}
		return nil
	})
}
