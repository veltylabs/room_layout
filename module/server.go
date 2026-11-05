//go:build !browser

package module

import (
	"webtyp.com/appkit"
	"webtyp.com/model"

	roomlayout "github.com/veltylabs/room_layout"
)

// server es lo que el módulo aporta al servidor: cómo se construye.
var server = appkit.ModuleServer{Build: build}

// En el servidor el módulo no aporta pantallas.
var browser appkit.ModuleBrowser

func build(d appkit.Deps) (model.ModuleNaming, error) {
	cats, err := appkit.Find[roomlayout.CategoryReader](d)
	if err != nil {
		return nil, err
	}
	occ, err := appkit.Find[roomlayout.OccupantReader](d)
	if err != nil {
		return nil, err
	}
	hours, _ := appkit.Find[roomlayout.HoursReader](d)
	return roomlayout.New(d.DB, roomlayout.Deps{
		IDs: d.IDs, Publisher: d.Events, TenantID: d.TenantID, Zone: d.Zone,
		Categories: cats, Occupants: occ, Hours: hours,
	})
}
