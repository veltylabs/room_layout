package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
)

func (m *Module) ListRoomCategories(tenantID, roomID string) ([]fmt.KeyValue, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	var catIDs []string
	err := m.db.Query(&RoomCategory{}).
		Where("tenant_id").Eq(tenantID).
		Where("room_id").Eq(roomID).
		ReadAll(func() model.Model { return &RoomCategory{} }, func(row model.Model) {
			if item, ok := row.(*RoomCategory); ok {
				catIDs = append(catIDs, item.CategoryId)
			}
		})
	if err != nil {
		return nil, err
	}

	allOptions, err := m.deps.Categories.CategoryOptions(tenantID)
	if err != nil {
		return nil, err
	}

	var res []fmt.KeyValue
	for _, id := range catIDs {
		label := id
		for _, opt := range allOptions {
			if opt.Key == id {
				label = opt.Value
				break
			}
		}
		res = append(res, fmt.KeyValue{Key: id, Value: label})
	}
	return res, nil
}

func (m *Module) SetRoomCategories(tenantID, roomID string, categoryIDs []string) error {
	if tenantID == "" {
		return ErrTenantRequired
	}
	if _, err := m.GetRoom(tenantID, roomID); err != nil {
		return err
	}

	allOptions, err := m.deps.Categories.CategoryOptions(tenantID)
	if err != nil {
		return err
	}

	for _, catID := range categoryIDs {
		var found bool
		for _, opt := range allOptions {
			if opt.Key == catID {
				found = true
				break
			}
		}
		if !found {
			return ErrUnknownCategory
		}
	}

	// Check if any active shift uses a category being removed.
	var activeShifts []RoomShift
	err = m.db.Query(&RoomShift{}).
		Where("tenant_id").Eq(tenantID).
		Where("room_id").Eq(roomID).
		Where("is_active").Eq(true).
		ReadAll(func() model.Model { return &RoomShift{} }, func(row model.Model) {
			if item, ok := row.(*RoomShift); ok {
				activeShifts = append(activeShifts, *item)
			}
		})
	if err != nil {
		return err
	}

	for _, s := range activeShifts {
		var kept bool
		for _, catID := range categoryIDs {
			if s.CategoryId == catID {
				kept = true
				break
			}
		}
		if !kept {
			return ErrCategoryInUse
		}
	}

	return m.db.Tx(func(tx *orm.DB) error {
		dummy := RoomCategory{TenantId: tenantID, RoomId: roomID}
		if err := tx.Delete(&dummy, orm.Eq("tenant_id", tenantID), orm.Eq("room_id", roomID)); err != nil {
			return err
		}
		for _, catID := range categoryIDs {
			rc := RoomCategory{
				TenantId:   tenantID,
				RoomId:     roomID,
				CategoryId: catID,
			}
			if err := tx.Create(&rc); err != nil {
				return err
			}
		}
		return nil
	})
}
