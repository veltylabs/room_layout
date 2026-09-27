package seed

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/fmt"
)

type LoadOptions struct {
	CategoryIDs []string
	OccupantIDs []string
}

type Data struct {
	Floors    []roomlayout.Floor
	Rooms     []roomlayout.Room
	Equipment []roomlayout.Equipment
	Shifts    []roomlayout.RoomShift
}

func Load(m *roomlayout.Module, tenantID string, opts ...LoadOptions) (Data, error) {
	if tenantID == "" {
		return Data{}, fmt.Err("seed: tenantID is required")
	}

	var catIDs []string
	var occIDs []string

	if len(opts) > 0 {
		catIDs = opts[0].CategoryIDs
		occIDs = opts[0].OccupantIDs
	}

	// 1. Create 2 Floors
	f1, err := m.SaveFloor(roomlayout.Floor{TenantId: tenantID, Name: "Primer piso", Position: 1})
	if err != nil {
		return Data{}, err
	}
	f2, err := m.SaveFloor(roomlayout.Floor{TenantId: tenantID, Name: "Segundo piso", Position: 2})
	if err != nil {
		return Data{}, err
	}

	// 2. Create 5 Rooms
	r101, err := m.SaveRoom(roomlayout.Room{TenantId: tenantID, FloorId: f1.Id, Code: "B-101", Name: "Consultorio 101", IsActive: true})
	if err != nil {
		return Data{}, err
	}
	r102, err := m.SaveRoom(roomlayout.Room{TenantId: tenantID, FloorId: f1.Id, Code: "B-102", Name: "Consultorio 102", IsActive: true})
	if err != nil {
		return Data{}, err
	}
	r103, err := m.SaveRoom(roomlayout.Room{TenantId: tenantID, FloorId: f1.Id, Code: "B-103", Name: "Consultorio 103", IsActive: true})
	if err != nil {
		return Data{}, err
	}
	r201, err := m.SaveRoom(roomlayout.Room{TenantId: tenantID, FloorId: f2.Id, Code: "B-201", Name: "Consultorio 201", IsActive: true})
	if err != nil {
		return Data{}, err
	}
	r202, err := m.SaveRoom(roomlayout.Room{TenantId: tenantID, FloorId: f2.Id, Code: "B-202", Name: "Consultorio 202", IsActive: true})
	if err != nil {
		return Data{}, err
	}

	// 3. Create 3 Equipment
	eq1, err := m.SaveEquipment(roomlayout.Equipment{TenantId: tenantID, Name: "Impresora"})
	if err != nil {
		return Data{}, err
	}
	eq2, err := m.SaveEquipment(roomlayout.Equipment{TenantId: tenantID, Name: "Electrocardiógrafo"})
	if err != nil {
		return Data{}, err
	}
	eq3, err := m.SaveEquipment(roomlayout.Equipment{TenantId: tenantID, Name: "Camilla ginecológica"})
	if err != nil {
		return Data{}, err
	}

	// Assign Equipment
	_ = m.SetRoomEquipment(tenantID, r101.Id, []string{eq1.Id, eq2.Id})
	_ = m.SetRoomEquipment(tenantID, r102.Id, []string{eq1.Id})
	_ = m.SetRoomEquipment(tenantID, r103.Id, []string{eq3.Id})
	_ = m.SetRoomEquipment(tenantID, r201.Id, []string{eq2.Id})
	_ = m.SetRoomEquipment(tenantID, r202.Id, []string{eq1.Id, eq3.Id})

	// Assign Categories if present
	if len(catIDs) >= 2 {
		_ = m.SetRoomCategories(tenantID, r101.Id, []string{catIDs[0], catIDs[1]})
		_ = m.SetRoomCategories(tenantID, r102.Id, []string{catIDs[0]})
		_ = m.SetRoomCategories(tenantID, r103.Id, []string{catIDs[1]})
		_ = m.SetRoomCategories(tenantID, r201.Id, []string{catIDs[0]})
		_ = m.SetRoomCategories(tenantID, r202.Id, []string{catIDs[1]})
	}

	// 4. Create 6 Weekly Shifts
	var shifts []roomlayout.RoomShift
	cat1 := "cat1"
	cat2 := "cat1"
	if len(catIDs) > 0 {
		cat1 = catIDs[0]
		cat2 = catIDs[0]
	}
	if len(catIDs) > 1 {
		cat2 = catIDs[1]
	}

	occ1 := ""
	occLabel1 := "Dr. Juan Pérez"
	if len(occIDs) > 0 {
		occ1 = occIDs[0]
	}

	shiftForms := []roomlayout.RoomShiftForm{
		{RoomId: r101.Id, CategoryId: cat1, OccupantId: occ1, OccupantLabel: occLabel1, DayOfWeek: roomlayout.WeekdayMonday, Start: "09:00", End: "13:00"},
		{RoomId: r101.Id, CategoryId: cat1, OccupantId: occ1, OccupantLabel: occLabel1, DayOfWeek: roomlayout.WeekdayTuesday, Start: "09:00", End: "13:00"},
		{RoomId: r102.Id, CategoryId: cat1, OccupantId: occ1, OccupantLabel: occLabel1, DayOfWeek: roomlayout.WeekdayWednesday, Start: "09:00", End: "13:00"},
		{RoomId: r102.Id, CategoryId: cat1, OccupantId: occ1, OccupantLabel: occLabel1, DayOfWeek: roomlayout.WeekdayThursday, Start: "14:00", End: "18:00"},
		{RoomId: r103.Id, CategoryId: cat2, OccupantId: occ1, OccupantLabel: occLabel1, DayOfWeek: roomlayout.WeekdayFriday, Start: "09:00", End: "13:00"},
		{RoomId: r201.Id, CategoryId: cat1, OccupantId: occ1, OccupantLabel: occLabel1, DayOfWeek: roomlayout.WeekdayFriday, Start: "14:00", End: "18:00"},
	}

	for _, sf := range shiftForms {
		s, err := m.SaveShift(tenantID, sf)
		if err == nil {
			shifts = append(shifts, s)
		}
	}

	data := Data{
		Floors:    []roomlayout.Floor{f1, f2},
		Rooms:     []roomlayout.Room{r101, r102, r103, r201, r202},
		Equipment: []roomlayout.Equipment{eq1, eq2, eq3},
		Shifts:    shifts,
	}

	return data, nil
}
