//go:build !wasm

// Package server es room_layout tal como lo enchufa una app.
package server

import (
	"webtyp.com/appkit"
	"webtyp.com/model"

	roomlayout "github.com/veltylabs/room_layout"
	"github.com/veltylabs/room_layout/migrate"
)

// Module es room_layout para appkit. Necesita las categorías de un box
// (roomlayout.CategoryReader) y los ocupantes de un turno
// (roomlayout.OccupantReader): en App.Modules va después de los módulos que las dan.
//
// Ejemplo: Modules: []appkit.Module{catalog.Module, staff.Module, rooms.Module}
var Module = appkit.Module{Name: "room_layout", Migrate: migrate.Migrate, Build: build}

func build(d appkit.Deps) (model.ModuleNaming, error) {
	cats, err := appkit.Find[roomlayout.CategoryReader](d)
	if err != nil {
		return nil, err
	}
	occ, err := appkit.Find[roomlayout.OccupantReader](d)
	if err != nil {
		return nil, err
	}
	return roomlayout.New(d.DB, roomlayout.Deps{
		IDs: d.IDs, Publisher: d.Events, TenantID: d.Tenant, Timezone: d.Timezone,
		Categories: cats, Occupants: occ,
	})
}
