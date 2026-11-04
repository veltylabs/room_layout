package tests

import (
	"sync/atomic"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/time"
)

type simpleIDs struct {
	counter uint64
}

func (s *simpleIDs) NewID() string {
	val := atomic.AddUint64(&s.counter, 1)
	return fmt.Sprintf("id%d", val)
}

type tenantOpts struct {
	tenantID string
	opts     []fmt.KeyValue
}

type fakeCategories struct {
	list []tenantOpts
}

func newFakeCategories() *fakeCategories {
	return &fakeCategories{}
}

func (f *fakeCategories) CategoryOptions(tenantID string) ([]fmt.KeyValue, error) {
	for _, item := range f.list {
		if item.tenantID == tenantID {
			return item.opts, nil
		}
	}
	return []fmt.KeyValue{
		{Key: "cat1", Value: "Medicina General"},
		{Key: "cat2", Value: "Cardiología"},
		{Key: "cat3", Value: "Ginecología"},
	}, nil
}

type fakeOccupants struct {
	list []tenantOpts
}

func newFakeOccupants() *fakeOccupants {
	return &fakeOccupants{}
}

func (f *fakeOccupants) OccupantOptions(tenantID string) ([]fmt.KeyValue, error) {
	for _, item := range f.list {
		if item.tenantID == tenantID {
			return item.opts, nil
		}
	}
	return []fmt.KeyValue{
		{Key: "occ1", Value: "Dr. Juan Pérez"},
		{Key: "occ2", Value: "Dra. María González"},
	}, nil
}

type boundEntry struct {
	date   time.Unix
	hours time.DayLimits
}

type fakeHours struct {
	hours []boundEntry
}

func newFakeHours() *fakeHours {
	return &fakeHours{}
}

func (f *fakeHours) DayLimits(date time.Unix) (time.DayLimits, error) {
	for _, b := range f.hours {
		if b.date == date {
			return b.hours, nil
		}
	}
	return time.DayLimits{Open: true, OpenMin: 480, CloseMin: 1200}, nil // 08:00 - 20:00
}

func (f *fakeHours) setBounds(date time.Unix, b time.DayLimits) {
	for i, entry := range f.hours {
		if entry.date == date {
			f.hours[i].hours = b
			return
		}
	}
	f.hours = append(f.hours, boundEntry{date: date, hours: b})
}

type testEnv struct {
	db        *orm.DB
	ids       *simpleIDs
	cats      *fakeCategories
	occs      *fakeOccupants
	hours    *fakeHours
	publisher *mockPublisher
	clockSec  time.Unix
	mod       *roomlayout.Module
}

type mockPublisher struct {
	events []events.Event
}

func (m *mockPublisher) Publish(e events.Event) {
	m.events = append(m.events, e)
}

func newTestEnv(tenantID string) *testEnv {
	db := orm.New(mem.New())
	ids := &simpleIDs{}
	cats := newFakeCategories()
	occs := newFakeOccupants()
	hours := newFakeHours()
	pub := &mockPublisher{}

	env := &testEnv{
		db:        db,
		ids:       ids,
		cats:      cats,
		occs:      occs,
		hours:    hours,
		publisher: pub,
		clockSec:  1700000000,
	}

	zone, _ := time.LoadZone("America/Santiago")
	deps := roomlayout.Deps{
		IDs:        ids,
		Categories: cats,
		Occupants:  occs,
		TenantID:   tenantID,
		Zone:       zone,
		Hours:      hours,
		Publisher:  pub,
		Clock: func() time.Nano {
			return env.clockSec.Nano()
		},
	}

	mod, err := roomlayout.New(db, deps)
	if err != nil {
		panic(err)
	}

	env.mod = mod
	return env
}
