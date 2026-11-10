package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/storage"
)

// ListRoomCategories responde las categorías asignadas a una sala.
//
// Ejemplo: list, err := m.ListRoomCategories(tenantID, roomID)
func (m *Module) ListRoomCategories(tenantID, roomID string) ([]fmt.KeyValue, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	links, err := allRows(store, func() model.Model { return &RoomCategory{} }, storage.Eq(RoomCategory_.RoomId, roomID))
	if err != nil {
		return nil, err
	}
	allOptions, err := m.deps.Categories.CategoryOptions(tenantID)
	if err != nil {
		return nil, err
	}
	var res []fmt.KeyValue
	for _, l := range links {
		id := l.(*RoomCategory).CategoryId
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

// SetRoomCategories actualiza la lista de categorías asignadas a una sala.
//
// Ejemplo: err := m.SetRoomCategories(tenantID, roomID, categoryIDs)
func (m *Module) SetRoomCategories(tenantID, roomID string, categoryIDs []string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if _, err := m.GetRoom(tenantID, roomID); err != nil {
		return err
	}
	allOptions, err := m.deps.Categories.CategoryOptions(tenantID)
	if err != nil {
		return err
	}
	for _, catID := range categoryIDs {
		if !categoryIn(allOptions, catID) {
			return ErrUnknownCategory
		}
	}
	return store.Atomic(func(tx crud.Tx) error {
		// an active shift of the room cannot lose the category it uses
		shifts, err := allRows(tx, func() model.Model { return &RoomShift{} },
			storage.Eq(RoomShift_.RoomId, roomID), storage.Eq(RoomShift_.IsActive, true))
		if err != nil {
			return err
		}
		for _, r := range shifts {
			if !containsString(categoryIDs, r.(*RoomShift).CategoryId) {
				return ErrCategoryInUse
			}
		}
		old, err := allRows(tx, func() model.Model { return &RoomCategory{} }, storage.Eq(RoomCategory_.RoomId, roomID))
		if err != nil {
			return err
		}
		for _, r := range old {
			if err := tx.Delete(&RoomCategory{}, model.JoinKey(roomID, r.(*RoomCategory).CategoryId)); err != nil {
				return err
			}
		}
		for _, catID := range categoryIDs {
			if err := tx.Create(&RoomCategory{RoomId: roomID, CategoryId: catID}); err != nil {
				return err
			}
		}
		return nil
	})
}
