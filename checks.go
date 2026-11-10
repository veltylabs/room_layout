package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/storage"
)

// trimmed quita los espacios de un texto y rechaza que quede vacío: es el «Trim» de los antiguos Save* como regla.
// Ejemplo: name, err := trimmed("name", "  Box 1 ")   // "Box 1"
func trimmed(field, value string) (string, error) {
	v := fmt.TrimSpace(value)
	if v == "" {
		return "", &model.FieldError{Field: field, Err: ErrBlankName}
	}
	return v, nil
}

// checkCodeFree rechaza un código de sala que ya usa otra sala (exceptId es la propia sala al actualizar, "" al crear).
// Ejemplo: err := checkCodeFree(tx, "B-101", "")   // ErrCodeAlreadyExists si ya existe
func checkCodeFree(tx crud.Tx, code, exceptId string) error {
	rows, err := allRows(tx, func() model.Model { return &Room{} }, storage.Eq(Room_.Code, code))
	if err != nil {
		return err
	}
	for _, r := range rows {
		if r.(*Room).Id != exceptId {
			return ErrCodeAlreadyExists
		}
	}
	return nil
}

// checkFloorFree rechaza borrar un piso que todavía tiene salas.
// Ejemplo: err := checkFloorFree(tx, "f1")   // ErrFloorInUse si tiene salas
func checkFloorFree(tx crud.Tx, floorId string) error {
	_, found, err := firstRow(tx, func() model.Model { return &Room{} }, storage.Eq(Room_.FloorId, floorId))
	if err != nil {
		return err
	}
	if found {
		return ErrFloorInUse
	}
	return nil
}

// checkRoomExists rechaza un vínculo con una sala que no existe.
// Ejemplo: err := checkRoomExists(tx, "r1")   // ErrNotFound si no existe
func checkRoomExists(tx crud.Tx, roomId string) error {
	var r Room
	if err := tx.ReadOne(&r, roomId); err != nil {
		return notFound(err, ErrNotFound)
	}
	return nil
}

// categoryIn dice si la categoría está entre las opciones que entrega el lector de categorías de la app.
// Ejemplo: ok := categoryIn(options, "cat1")
func categoryIn(options []fmt.KeyValue, categoryId string) bool {
	for _, opt := range options {
		if opt.Key == categoryId {
			return true
		}
	}
	return false
}

// checkCategoryNotInUse rechaza quitar a una sala una categoría que un turno activo suyo todavía usa.
// Ejemplo: err := checkCategoryNotInUse(tx, "r1", "cat1")   // ErrCategoryInUse si hay un turno activo
func checkCategoryNotInUse(tx crud.Tx, roomId, categoryId string) error {
	_, found, err := firstRow(tx, func() model.Model { return &RoomShift{} },
		storage.Eq(RoomShift_.RoomId, roomId), storage.Eq(RoomShift_.CategoryId, categoryId), storage.Eq(RoomShift_.IsActive, true))
	if err != nil {
		return err
	}
	if found {
		return ErrCategoryInUse
	}
	return nil
}

// deleteEquipmentLinks borra todos los vínculos sala–equipamiento de un equipo (se llama al borrar el equipo).
// Ejemplo: err := deleteEquipmentLinks(tx, "e1")
func deleteEquipmentLinks(tx crud.Tx, equipmentId string) error {
	for {
		rows, err := allRows(tx, func() model.Model { return &RoomEquipment{} }, storage.Eq(RoomEquipment_.EquipmentId, equipmentId))
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		for _, r := range rows {
			link := r.(*RoomEquipment)
			if err := tx.Delete(&RoomEquipment{}, model.JoinKey(link.RoomId, link.EquipmentId)); err != nil {
				return err
			}
		}
	}
}
