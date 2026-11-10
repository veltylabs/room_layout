//go:build wasm

package seed

import (
	roomlayout "github.com/veltylabs/room_layout"
	"github.com/veltylabs/room_layout/migrate"
	"github.com/veltylabs/room_layout/ui"
	"webtyp.com/crud"
	"webtyp.com/ddl"
	"webtyp.com/fmt"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/time"
	"webtyp.com/unixid"
	"webtyp.com/view"
)

const (
	demoTenantID = "demo-tenant"
	demoZone     = "America/Santiago"
	demoCategory = "cat-general"
	demoOccupant = "occ-perez"
)

// demoCategories es el lector de categorías de la demostración: una sola categoría.
type demoCategories struct{}

func (demoCategories) CategoryOptions(tenantID string) ([]fmt.KeyValue, error) {
	return []fmt.KeyValue{{Key: demoCategory, Value: "Medicina general"}}, nil
}

// demoOccupants es el lector de ocupantes de la demostración: un solo profesional.
type demoOccupants struct{}

func (demoOccupants) OccupantOptions(tenantID string) ([]fmt.KeyValue, error) {
	return []fmt.KeyValue{{Key: demoOccupant, Value: "Dra. Pérez"}}, nil
}

// Demo crea una demostración ejecutable del módulo room_layout sobre una base en memoria con datos semilla.
//
// Ejemplo: demo, err := seed.Demo()
func Demo() (view.Demo, error) {
	db := orm.New(mem.New())
	ids, err := unixid.NewUnixID()
	if err != nil {
		return view.Demo{}, err
	}
	zone, err := time.LoadZone(demoZone)
	if err != nil {
		return view.Demo{}, err
	}
	mod, err := roomlayout.New(db, roomlayout.Deps{
		IDs: ids, TenantID: demoTenantID, Zone: zone, Categories: demoCategories{}, Occupants: demoOccupants{},
	})
	if err != nil {
		return view.Demo{}, err
	}
	if _, err := Load(mod, demoTenantID, LoadOptions{CategoryIDs: []string{demoCategory}, OccupantIDs: []string{demoOccupant}}); err != nil {
		return view.Demo{}, err
	}
	return view.Demo{
		Label:      "Salas",
		Timezone:   demoZone,
		ActingAs:   "u1",
		Modules:    []crud.Module{mod},
		Migrations: []ddl.Step{{Name: roomlayout.ModelName, Run: migrate.Migrate}},
		Screens: func(d view.BrowserDeps) []view.Screen {
			return ui.Screens(d.Caller)
		},
	}, nil
}
