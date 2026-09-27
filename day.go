package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/time"
)

func (m *Module) localNow() (int64, int) {
	nowSec := m.now() / 1_000_000_000
	D := time.MidnightUTC(nowSec)
	min := int((nowSec - time.LocalMinutesToUnixUTC(D, 0, m.deps.Timezone)) / 60)
	if min < 0 {
		D -= 86400
		min = int((nowSec - time.LocalMinutesToUnixUTC(D, 0, m.deps.Timezone)) / 60)
	}
	if min >= 1440 {
		D += 86400
		min = int((nowSec - time.LocalMinutesToUnixUTC(D, 0, m.deps.Timezone)) / 60)
	}
	return D, min
}

func isRoomActive(rooms []Room, roomID string) bool {
	for _, r := range rooms {
		if r.Id == roomID {
			return r.IsActive
		}
	}
	return false
}

func findRoom(rooms []Room, roomID string) (Room, bool) {
	for _, r := range rooms {
		if r.Id == roomID {
			return r, true
		}
	}
	return Room{}, false
}

func findFloor(floors []Floor, floorID string) (Floor, bool) {
	for _, f := range floors {
		if f.Id == floorID {
			return f, true
		}
	}
	return Floor{}, false
}

func findCategoryLabel(catOpts []fmt.KeyValue, catID string) string {
	for _, opt := range catOpts {
		if opt.Key == catID {
			return opt.Value
		}
	}
	return catID
}

func containsString(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func (m *Module) shiftsOn(tenantID string, date int64) ([]RoomShift, error) {
	if m.deps.Bounds != nil {
		bounds, err := m.deps.Bounds.GetDayBounds(date)
		if err != nil {
			return nil, err
		}
		if !bounds.Open {
			return []RoomShift{}, nil
		}
	}

	dow := time.Weekday(date)

	var activeShifts []RoomShift
	err := m.db.Query(&RoomShift{}).
		Where("tenant_id").Eq(tenantID).
		Where("is_active").Eq(true).
		ReadAll(func() model.Model { return &RoomShift{} }, func(row model.Model) {
			if item, ok := row.(*RoomShift); ok {
				activeShifts = append(activeShifts, *item)
			}
		})
	if err != nil {
		return nil, err
	}

	// Filter by active room
	rooms, err := m.ListRooms(tenantID, true)
	if err != nil {
		return nil, err
	}

	var res []RoomShift
	for _, s := range activeShifts {
		if !isRoomActive(rooms, s.RoomId) {
			continue
		}
		if s.SpecificDate == 0 { // Weekly
			if s.DayOfWeek == int64(dow) {
				cancelled, err := m.isShiftCancelled(tenantID, s.Id, date)
				if err != nil {
					return nil, err
				}
				if !cancelled {
					res = append(res, s)
				}
			}
		} else { // Dated
			if s.SpecificDate == date {
				res = append(res, s)
			}
		}
	}

	// Sort by room_id asc, start_min asc
	for i := 0; i < len(res); i++ {
		for j := i + 1; j < len(res); j++ {
			if res[i].RoomId > res[j].RoomId ||
				(res[i].RoomId == res[j].RoomId && res[i].StartMin > res[j].StartMin) {
				res[i], res[j] = res[j], res[i]
			}
		}
	}

	return res, nil
}

func (m *Module) ListRoomDay(tenantID, roomID string, date int64) ([]DayShift, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	if date <= 0 {
		date, _ = m.localNow()
	}

	shifts, err := m.shiftsOn(tenantID, date)
	if err != nil {
		return nil, err
	}

	rooms, err := m.ListRooms(tenantID, false)
	if err != nil {
		return nil, err
	}

	catOpts, err := m.deps.Categories.CategoryOptions(tenantID)
	if err != nil {
		return nil, err
	}

	var res []DayShift
	for _, s := range shifts {
		if roomID != "" && s.RoomId != roomID {
			continue
		}
		r, _ := findRoom(rooms, s.RoomId)
		cLabel := findCategoryLabel(catOpts, s.CategoryId)
		res = append(res, DayShift{
			ShiftId:       s.Id,
			RoomId:        s.RoomId,
			RoomCode:      r.Code,
			CategoryId:    s.CategoryId,
			CategoryLabel: cLabel,
			OccupantId:    s.OccupantId,
			OccupantLabel: s.OccupantLabel,
			Start:         formatHour(int(s.StartMin)),
			End:           formatHour(int(s.EndMin)),
			IsWeekly:      s.SpecificDate == 0,
		})
	}
	return res, nil
}

type roomAux struct {
	room       Room
	floorName  string
	floorPos   int64
	catLabels  string
	eqLabels   string
	catIDSlice []string
	eqIDSlice  []string
}

func (m *Module) ListBoard(tenantID, categoryID, equipmentID string) ([]BoardRoom, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	todayDate, nowMin := m.localNow()

	floors, err := m.ListFloors(tenantID)
	if err != nil {
		return nil, err
	}

	rooms, err := m.ListRooms(tenantID, true)
	if err != nil {
		return nil, err
	}

	catOpts, err := m.deps.Categories.CategoryOptions(tenantID)
	if err != nil {
		return nil, err
	}

	var auxList []roomAux
	for _, r := range rooms {
		f, _ := findFloor(floors, r.FloorId)

		// Categories
		rcList, err := m.ListRoomCategories(tenantID, r.Id)
		if err != nil {
			return nil, err
		}
		var catIDSlice []string
		var catNames []string
		for _, rc := range rcList {
			catIDSlice = append(catIDSlice, rc.Key)
			catNames = append(catNames, findCategoryLabel(catOpts, rc.Key))
		}

		// Equipment
		eqList, err := m.ListRoomEquipment(tenantID, r.Id)
		if err != nil {
			return nil, err
		}
		var eqIDSlice []string
		var eqNames []string
		for _, eq := range eqList {
			eqIDSlice = append(eqIDSlice, eq.Id)
			eqNames = append(eqNames, eq.Name)
		}

		if categoryID != "" && !containsString(catIDSlice, categoryID) {
			continue
		}
		if equipmentID != "" && !containsString(eqIDSlice, equipmentID) {
			continue
		}

		auxList = append(auxList, roomAux{
			room:       r,
			floorName:  f.Name,
			floorPos:   f.Position,
			catLabels:  fmt.JoinSlice(catNames, ", "),
			eqLabels:   fmt.JoinSlice(eqNames, ", "),
			catIDSlice: catIDSlice,
			eqIDSlice:  eqIDSlice,
		})
	}

	// Sort auxList by floorPos asc, floorName asc, code asc
	for i := 0; i < len(auxList); i++ {
		for j := i + 1; j < len(auxList); j++ {
			if auxList[i].floorPos > auxList[j].floorPos ||
				(auxList[i].floorPos == auxList[j].floorPos && auxList[i].floorName > auxList[j].floorName) ||
				(auxList[i].floorPos == auxList[j].floorPos && auxList[i].floorName == auxList[j].floorName && auxList[i].room.Code > auxList[j].room.Code) {
				auxList[i], auxList[j] = auxList[j], auxList[i]
			}
		}
	}

	boundsOpen := true
	if m.deps.Bounds != nil {
		b, err := m.deps.Bounds.GetDayBounds(todayDate)
		if err != nil {
			return nil, err
		}
		boundsOpen = b.Open
	}

	todayShifts, err := m.shiftsOn(tenantID, todayDate)
	if err != nil {
		return nil, err
	}

	var res []BoardRoom
	todayStr := formatDate(todayDate)

	for _, aux := range auxList {
		br := BoardRoom{
			RoomId:          aux.room.Id,
			Code:            aux.room.Code,
			Name:            aux.room.Name,
			FloorName:       aux.floorName,
			FloorPosition:   aux.floorPos,
			CategoryLabels:  aux.catLabels,
			EquipmentLabels: aux.eqLabels,
			Today:           todayStr,
		}

		if !boundsOpen {
			br.Status = BoardStatusClosed
			res = append(res, br)
			continue
		}

		var current *RoomShift
		var next *RoomShift

		for _, s := range todayShifts {
			if s.RoomId != aux.room.Id {
				continue
			}
			if s.StartMin <= int64(nowMin) && int64(nowMin) < s.EndMin {
				sCopy := s
				current = &sCopy
			} else if s.StartMin > int64(nowMin) {
				if next == nil || s.StartMin < next.StartMin {
					sCopy := s
					next = &sCopy
				}
			}
		}

		if current != nil {
			br.Status = BoardStatusBusy
			br.CurrentOccupantLabel = current.OccupantLabel
			br.CurrentUntil = formatHour(int(current.EndMin))
		} else {
			br.Status = BoardStatusFree
		}

		if next != nil {
			br.NextStart = formatHour(int(next.StartMin))
			br.NextOccupantLabel = next.OccupantLabel
		}

		res = append(res, br)
	}

	return res, nil
}

func (m *Module) FindFreeRooms(tenantID string, date int64, startMin, endMin int, categoryID string) ([]FreeRoom, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	if startMin < 0 || startMin >= endMin || endMin > 1440 {
		return nil, ErrInvalidRange
	}

	bounds := time.Unbounded()
	if m.deps.Bounds != nil {
		b, err := m.deps.Bounds.GetDayBounds(date)
		if err != nil {
			return nil, err
		}
		bounds = b
	}

	if !bounds.Open || startMin < bounds.OpenMin || endMin > bounds.CloseMin {
		return []FreeRoom{}, nil
	}

	floors, err := m.ListFloors(tenantID)
	if err != nil {
		return nil, err
	}

	rooms, err := m.ListRooms(tenantID, true)
	if err != nil {
		return nil, err
	}

	shifts, err := m.shiftsOn(tenantID, date)
	if err != nil {
		return nil, err
	}

	var auxList []roomAux
	for _, r := range rooms {
		f, _ := findFloor(floors, r.FloorId)

		rcList, err := m.ListRoomCategories(tenantID, r.Id)
		if err != nil {
			return nil, err
		}
		var catIDSlice []string
		for _, rc := range rcList {
			catIDSlice = append(catIDSlice, rc.Key)
		}

		if categoryID != "" && !containsString(catIDSlice, categoryID) {
			continue
		}

		eqList, err := m.ListRoomEquipment(tenantID, r.Id)
		if err != nil {
			return nil, err
		}
		var eqNames []string
		for _, eq := range eqList {
			eqNames = append(eqNames, eq.Name)
		}

		auxList = append(auxList, roomAux{
			room:      r,
			floorName: f.Name,
			floorPos:  f.Position,
			eqLabels:  fmt.JoinSlice(eqNames, ", "),
		})
	}

	// Sort auxList by floorPos asc, floorName asc, code asc
	for i := 0; i < len(auxList); i++ {
		for j := i + 1; j < len(auxList); j++ {
			if auxList[i].floorPos > auxList[j].floorPos ||
				(auxList[i].floorPos == auxList[j].floorPos && auxList[i].floorName > auxList[j].floorName) ||
				(auxList[i].floorPos == auxList[j].floorPos && auxList[i].floorName == auxList[j].floorName && auxList[i].room.Code > auxList[j].room.Code) {
				auxList[i], auxList[j] = auxList[j], auxList[i]
			}
		}
	}

	var res []FreeRoom
	for _, aux := range auxList {
		var overlaps bool
		for _, s := range shifts {
			if s.RoomId == aux.room.Id {
				if int64(startMin) < s.EndMin && s.StartMin < int64(endMin) {
					overlaps = true
					break
				}
			}
		}
		if !overlaps {
			res = append(res, FreeRoom{
				RoomId:          aux.room.Id,
				Code:            aux.room.Code,
				Name:            aux.room.Name,
				FloorName:       aux.floorName,
				EquipmentLabels: aux.eqLabels,
			})
		}
	}

	return res, nil
}

func (m *Module) ListPublicAvailability(date int64, categoryID string) ([]PublicSlot, error) {
	tenantID := m.deps.TenantID
	bounds := time.Unbounded()
	if m.deps.Bounds != nil {
		b, err := m.deps.Bounds.GetDayBounds(date)
		if err != nil {
			return nil, err
		}
		bounds = b
	}
	if !bounds.Open {
		return []PublicSlot{}, nil
	}

	floors, err := m.ListFloors(tenantID)
	if err != nil {
		return nil, err
	}

	rooms, err := m.ListRooms(tenantID, true)
	if err != nil {
		return nil, err
	}

	shifts, err := m.shiftsOn(tenantID, date)
	if err != nil {
		return nil, err
	}

	var res []PublicSlot
	for _, r := range rooms {
		if categoryID != "" {
			rcList, err := m.ListRoomCategories(tenantID, r.Id)
			if err != nil {
				return nil, err
			}
			var enabled bool
			for _, rc := range rcList {
				if rc.Key == categoryID {
					enabled = true
					break
				}
			}
			if !enabled {
				continue
			}
		}

		f, _ := findFloor(floors, r.FloorId)

		// Gather shifts for room r
		var rShifts []RoomShift
		for _, s := range shifts {
			if s.RoomId == r.Id {
				rShifts = append(rShifts, s)
			}
		}

		// Compute free gaps in [bounds.OpenMin, bounds.CloseMin)
		curr := bounds.OpenMin
		for _, s := range rShifts {
			if s.StartMin > int64(curr) {
				gapStart := curr
				gapEnd := int(s.StartMin)
				if gapEnd > bounds.CloseMin {
					gapEnd = bounds.CloseMin
				}
				if gapEnd-gapStart >= MinPublicSlotMinutes {
					res = append(res, PublicSlot{
						RoomCode:  r.Code,
						RoomName:  r.Name,
						FloorName: f.Name,
						Start:     formatHour(gapStart),
						End:       formatHour(gapEnd),
					})
				}
			}
			if int(s.EndMin) > curr {
				curr = int(s.EndMin)
			}
		}
		if curr < bounds.CloseMin {
			gapStart := curr
			gapEnd := bounds.CloseMin
			if gapEnd-gapStart >= MinPublicSlotMinutes {
				res = append(res, PublicSlot{
					RoomCode:  r.Code,
					RoomName:  r.Name,
					FloorName: f.Name,
					Start:     formatHour(gapStart),
					End:       formatHour(gapEnd),
				})
			}
		}
	}

	return res, nil
}
