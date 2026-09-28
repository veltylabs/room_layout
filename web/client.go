//go:build wasm

package main

import (
	roomlayout "github.com/veltylabs/room_layout"
	"github.com/veltylabs/room_layout/seed"
	"github.com/veltylabs/room_layout/ui"
	"webtyp.com/dom"
	"webtyp.com/events/mock"
	"webtyp.com/fmt"
	"webtyp.com/layout/platformd"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
)

type demoCategories struct{}

func (d *demoCategories) CategoryOptions(tenantID string) ([]fmt.KeyValue, error) {
	return []fmt.KeyValue{
		{Key: "cat1", Value: "Medicina General"},
		{Key: "cat2", Value: "Cardiología"},
		{Key: "cat3", Value: "Ginecología"},
	}, nil
}

type demoOccupants struct{}

func (d *demoOccupants) OccupantOptions(tenantID string) ([]fmt.KeyValue, error) {
	return []fmt.KeyValue{
		{Key: "occ1", Value: "Dr. Juan Pérez"},
		{Key: "occ2", Value: "Dra. María González"},
		{Key: "occ3", Value: "Dr. Carlos Rojas"},
	}, nil
}

type demoUser struct{}

func (u demoUser) UserName() string    { return "Administrador Demo" }
func (u demoUser) UserAvatar() string  { return "" }
func (u demoUser) UserRoles() []string { return []string{"Administrador"} }

func main() {
	db := orm.New(mem.New())
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err)
	}
	pub := &mock.Broker{}

	deps := roomlayout.Deps{
		IDs:        ids,
		Categories: &demoCategories{},
		Occupants:  &demoOccupants{},
		TenantID:   "demo-tenant",
		Timezone:   "America/Santiago",
		Publisher:  pub,
	}

	mod, err := roomlayout.New(db, deps)
	if err != nil {
		panic(err)
	}

	if _, err := seed.Load(mod, "demo-tenant", seed.LoadOptions{
		CategoryIDs: []string{"cat1", "cat2", "cat3"},
		OccupantIDs: []string{"occ1", "occ2", "occ3"},
	}); err != nil {
		panic(err)
	}

	loop := loopback.WithTenant("demo-tenant", mod)
	v, err := ui.Browser(loop, ids, "demo-tenant")
	if err != nil {
		panic(err)
	}

	p := &platformd.Platform{
		AppName:   ui.DefaultLabel + " — demo",
		User:      demoUser{},
		Modules:   []platformd.UIModule{v},
		DefaultID: ui.ID,
	}

	if err := dom.Append("body", p); err != nil {
		panic(err)
	}

	select {}
}
