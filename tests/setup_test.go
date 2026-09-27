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
	date   int64
	bounds time.DayBounds
}

type fakeBounds struct {
	bounds []boundEntry
}

func newFakeBounds() *fakeBounds {
	return &fakeBounds{}
}

func (f *fakeBounds) GetDayBounds(date int64) (time.DayBounds, error) {
	for _, b := range f.bounds {
		if b.date == date {
			return b.bounds, nil
		}
	}
	return time.DayBounds{Open: true, OpenMin: 480, CloseMin: 1200}, nil // 08:00 - 20:00
}

func (f *fakeBounds) setBounds(date int64, b time.DayBounds) {
	for i, entry := range f.bounds {
		if entry.date == date {
			f.bounds[i].bounds = b
			return
		}
	}
	f.bounds = append(f.bounds, boundEntry{date: date, bounds: b})
}

type testEnv struct {
	db        *orm.DB
	ids       *simpleIDs
	cats      *fakeCategories
	occs      *fakeOccupants
	bounds    *fakeBounds
	publisher *mockPublisher
	clockSec  int64
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
	bounds := newFakeBounds()
	pub := &mockPublisher{}

	env := &testEnv{
		db:        db,
		ids:       ids,
		cats:      cats,
		occs:      occs,
		bounds:    bounds,
		publisher: pub,
		clockSec:  1700000000,
	}

	deps := roomlayout.Deps{
		IDs:        ids,
		Categories: cats,
		Occupants:  occs,
		TenantID:   tenantID,
		Timezone:   "America/Santiago",
		Bounds:     bounds,
		Publisher:  pub,
		Clock: func() int64 {
			return env.clockSec * 1_000_000_000
		},
	}

	mod, err := roomlayout.New(db, deps)
	if err != nil {
		panic(err)
	}

	env.mod = mod
	return env
}
