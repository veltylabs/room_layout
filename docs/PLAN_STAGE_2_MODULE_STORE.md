← [Stage 1](PLAN_STAGE_1_TOOLS_DEFINITION.md) | Next → [Stage 3](PLAN_STAGE_3_RULES.md)

# Stage 2 — `Module.store`, `storeFor`, row helpers, translation table

## 2.1 `module.go`

- `type Module struct`: add `store *orm.Store` after `db` (**`db` stays**: `shift.go` and `day.go` still use it until plan 2).
- In `New`, in the returned literal add `store: orm.TenantStore(db, deps.TenantID, deps.IDs),`.

## 2.2 `store.go` (new), written exactly

```go
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
```
If `definitionc` asks for other `Store…` names (it reads the entity names), keep the names it asks for in its error message.

## 2.3 `rows.go` (new), written exactly

`containsString` already exists in `day.go`: do NOT write it again.

```go
package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/storage"
)

// firstRow responde la primera fila que cumple conds y si existe alguna.
//
// Ejemplo: row, found, err := firstRow(tx, func() model.Model { return &Room{} }, storage.Eq(Room_.Code, "B-101"))
func firstRow(tx crud.Tx, newRow func() model.Model, conds ...storage.Condition) (model.Model, bool, error) {
	var first model.Model
	err := tx.ReadAll(newRow, crud.Query{Conditions: conds, Limit: 1}, func(r model.Model) {
		if first == nil {
			first = r
		}
	})
	return first, first != nil, err
}

// allRows responde todas las filas que cumplen conds, paginando por id porque una lectura trae como máximo crud.MaxPage
// filas. Una tabla de clave compuesta (vínculos, cancelaciones) se lee en una sola página: no admite cursor.
//
// Ejemplo: rows, err := allRows(tx, func() model.Model { return &Floor{} })
func allRows(tx crud.Tx, newRow func() model.Model, conds ...storage.Condition) ([]model.Model, error) {
	var rows []model.Model
	after := ""
	for {
		var page []model.Model
		err := tx.ReadAll(newRow, crud.Query{Conditions: conds, After: after, Limit: crud.MaxPage},
			func(r model.Model) { page = append(page, r) })
		if err != nil {
			return nil, err
		}
		rows = append(rows, page...)
		if len(page) < crud.MaxPage {
			return rows, nil
		}
		after = keyOf(page[len(page)-1])
		if after == "" {
			return rows, nil
		}
	}
}

// keyOf responde el id de una fila de clave simple, o "" si la tabla tiene clave compuesta.
func keyOf(r model.Model) string {
	switch row := r.(type) {
	case *Floor:
		return row.Id
	case *Room:
		return row.Id
	case *RoomArtifact:
		return row.Id
	case *Equipment:
		return row.Id
	case *RoomShift:
		return row.Id
	}
	return ""
}

// notFound cambia el error «no existe» de un repositorio por `missing` y deja pasar cualquier otro error.
//
// Ejemplo: err = notFound(err, ErrNotFound)
func notFound(err error, missing error) error {
	if model.KindOf(err) == model.ErrorNotFound {
		return missing
	}
	return err
}
```

## 2.4 Translation table for Stages 3 and 4 (old `m.db` call → new)

| Old | New |
|---|---|
| `m.db.Query(&X{}).Where("tenant_id").Eq(t).Where("id").Eq(id)…ReadAll` (one row) | `err := tx.ReadOne(&x, id)`, then `notFound(err, ErrNotFound)` (the tenant is the store's) |
| `….Where("a").Eq(v)…ReadAll(newRow, onRow)` | `rows, err := allRows(tx, func() model.Model { return &X{} }, storage.Eq(X_.A, v))`, then `*rows[i].(*X)`; the old `OrderBy` becomes a small insertion sort in Go (Stage 4) |
| `….Limit(1).ReadAll` to ask «is there any?» | `_, found, err := firstRow(tx, func() model.Model { return &X{} }, storage.Eq(X_.A, v))` |
| `m.db.Create(&x)` | `tx.Create(&x)` (`Store.Create` stamps `tenant_id` by itself; never set it on a row of a rule) |
| `m.db.Update(&x, orm.Eq("id", …), orm.Eq("tenant_id", …))` | `tx.UpdateFields(&x, id, cols)` with the columns the old code wrote |
| `m.db.Delete(&x, orm.Eq(…))` | `tx.Delete(&x, id)`; a composite key: `tx.Delete(&x, model.JoinKey(a, b))` |
| `m.db.Tx(func(tx *orm.DB) error {…})` | `store.Atomic(func(tx crud.Tx) error {…})` with `store` from `m.storeFor(tenantID)` |
| `x.UpdatedAt = m.now().Unix()` (Floor, Room, Equipment) | `x.UpdatedAt = m.now()` (Stage 1: `updated_at` is `time.Nano`) |

Commit and push.
