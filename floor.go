package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
)

// floorColumns son las columnas que cambia SaveFloor al actualizar (todas menos id y tenant_id).
var floorColumns = []string{Floor_.Cols, Floor_.Rows, Floor_.GridLocked, Floor_.HabitableCells, Floor_.Name, Floor_.Position, Floor_.UpdatedAt}

// SaveFloor guarda o actualiza un piso en la base de datos.
//
// Ejemplo: floor, err := m.SaveFloor(floor)
func (m *Module) SaveFloor(f Floor) (Floor, error) {
	store, err := m.storeFor(f.TenantId)
	if err != nil {
		return Floor{}, err
	}
	f.Name = fmt.TrimSpace(f.Name)
	isNew := f.Id == ""
	if isNew {
		f.Id = m.deps.IDs.NewID()
	}
	if err := model.ValidateFields(byte(model.Create), &f); err != nil {
		return Floor{}, err
	}
	f.UpdatedAt = m.now()
	err = store.Atomic(func(tx crud.Tx) error {
		if isNew {
			return tx.Create(&f)
		}
		return tx.UpdateFields(&f, f.Id, floorColumns)
	})
	if err != nil {
		return Floor{}, err
	}
	m.publish(TopicFloorSaved, &f)
	return f, nil
}

// ListFloors responde la lista de pisos de un tenant ordenados por posición y nombre.
//
// Ejemplo: floors, err := m.ListFloors(tenantID)
func (m *Module) ListFloors(tenantID string) ([]Floor, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	rows, err := allRows(store, func() model.Model { return &Floor{} })
	if err != nil {
		return nil, err
	}
	var res []Floor
	for _, r := range rows {
		res = append(res, *r.(*Floor))
	}
	sortFloors(res)
	return res, nil
}

// DeleteFloor elimina un piso si no contiene salas asociadas.
//
// Ejemplo: err := m.DeleteFloor(tenantID, id)
func (m *Module) DeleteFloor(tenantID, id string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if id == "" {
		return ErrNotFound
	}
	err = store.Atomic(func(tx crud.Tx) error {
		if err := checkFloorFree(tx, id); err != nil {
			return err
		}
		return tx.Delete(&Floor{}, id)
	})
	if err != nil {
		return err
	}
	m.publish(TopicFloorDeleted, &Floor{Id: id, TenantId: tenantID})
	return nil
}
