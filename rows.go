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
