package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/orm"
)

// StoreFloor es el repositorio de los pisos.
//
// Ejemplo: store := m.StoreFloor()
func (m *Module) StoreFloor() crud.Store { return m.store }

// StoreRoom es el repositorio de las salas.
//
// Ejemplo: store := m.StoreRoom()
func (m *Module) StoreRoom() crud.Store { return m.store }

// StoreRoomArtifact es el repositorio de los artefactos de una sala.
//
// Ejemplo: store := m.StoreRoomArtifact()
func (m *Module) StoreRoomArtifact() crud.Store { return m.store }

// StoreEquipment es el repositorio del equipamiento.
//
// Ejemplo: store := m.StoreEquipment()
func (m *Module) StoreEquipment() crud.Store { return m.store }

// StoreRoomEquipment es el repositorio de los vínculos sala–equipamiento.
//
// Ejemplo: store := m.StoreRoomEquipment()
func (m *Module) StoreRoomEquipment() crud.Store { return m.store }

// StoreRoomCategory es el repositorio de los vínculos sala–categoría.
//
// Ejemplo: store := m.StoreRoomCategory()
func (m *Module) StoreRoomCategory() crud.Store { return m.store }

// StoreRoomShift es el repositorio de los turnos.
//
// Ejemplo: store := m.StoreRoomShift()
func (m *Module) StoreRoomShift() crud.Store { return m.store }

// StoreRoomShiftCancellation es el repositorio de las cancelaciones de una ocurrencia de turno.
//
// Ejemplo: store := m.StoreRoomShiftCancellation()
func (m *Module) StoreRoomShiftCancellation() crud.Store { return m.store }

// storeFor es el repositorio de un tenant para los métodos que reciben el tenant como argumento: el del módulo es
// m.store; otro tenant tiene su propio repositorio sellado, que no ve las filas de los demás. Un tenant vacío es un
// error (ErrTenantRequired): un repositorio sin tenant vería todas las filas.
//
// Ejemplo: store, err := m.storeFor("tenant-1")
func (m *Module) storeFor(tenantID string) (*orm.Store, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	if tenantID == m.tenantID {
		return m.store, nil
	}
	return orm.TenantStore(m.db, tenantID, m.deps.IDs), nil
}
