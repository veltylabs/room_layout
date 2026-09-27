package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)

func (m *Module) SaveShift(tenantID string, f RoomShiftForm) (RoomShift, error) {
	if tenantID == "" {
		return RoomShift{}, ErrTenantRequired
	}
	f.OccupantLabel = fmt.TrimSpace(f.OccupantLabel)

	// 1. Validation of form fields
	if err := model.ValidateFields(byte(model.Create), &f); err != nil {
		return RoomShift{}, &ValidationError{Err: err}
	}

	// 2. Room check
	room, err := m.GetRoom(tenantID, f.RoomId)
	if err != nil || !room.IsActive {
		return RoomShift{}, ErrNotFound
	}

	// 3. Time range check
	start, errStart := parseHour(f.Start)
	end, errEnd := parseHour(f.End)
	if errStart != nil || errEnd != nil || start < 0 || start >= end || end > 1440 {
		return RoomShift{}, ErrInvalidRange
	}

	// 4. Weekly vs Dated check
	var dow int
	var specDate int64
	if f.Date == "" {
		if f.DayOfWeek == "" {
			return RoomShift{}, ErrInvalidWeekday
		}
		var parsed int
		if f.DayOfWeek == WeekdaySunday {
			parsed = 0
		} else if f.DayOfWeek == WeekdayMonday {
			parsed = 1
		} else if f.DayOfWeek == WeekdayTuesday {
			parsed = 2
		} else if f.DayOfWeek == WeekdayWednesday {
			parsed = 3
		} else if f.DayOfWeek == WeekdayThursday {
			parsed = 4
		} else if f.DayOfWeek == WeekdayFriday {
			parsed = 5
		} else if f.DayOfWeek == WeekdaySaturday {
			parsed = 6
		} else {
			return RoomShift{}, ErrInvalidWeekday
		}
		dow = parsed
		specDate = 0
	} else {
		parsedDate, errDate := parseDate(f.Date)
		if errDate != nil || parsedDate <= 0 {
			return RoomShift{}, ErrInvalidDate
		}
		specDate = parsedDate
		dow = time.Weekday(specDate)
	}

	// 5. Category check
	catOpts, err := m.deps.Categories.CategoryOptions(tenantID)
	if err != nil {
		return RoomShift{}, err
	}
	var catFound bool
	for _, opt := range catOpts {
		if opt.Key == f.CategoryId {
			catFound = true
			break
		}
	}
	if !catFound {
		return RoomShift{}, ErrUnknownCategory
	}

	roomCats, err := m.ListRoomCategories(tenantID, f.RoomId)
	if err != nil {
		return RoomShift{}, err
	}
	var roomCatAllowed bool
	for _, rc := range roomCats {
		if rc.Key == f.CategoryId {
			roomCatAllowed = true
			break
		}
	}
	if !roomCatAllowed {
		return RoomShift{}, ErrCategoryNotAllowed
	}

	// 6. Occupant check
	occupantLabel := f.OccupantLabel
	if f.OccupantId != "" {
		occOpts, err := m.deps.Occupants.OccupantOptions(tenantID)
		if err != nil {
			return RoomShift{}, err
		}
		var occFound bool
		for _, opt := range occOpts {
			if opt.Key == f.OccupantId {
				occFound = true
				occupantLabel = opt.Value
				break
			}
		}
		if !occFound {
			return RoomShift{}, ErrUnknownOccupant
		}
	} else {
		if occupantLabel == "" {
			return RoomShift{}, ErrOccupantRequired
		}
	}

	// 7. Bounds check (only for dated shifts)
	if specDate > 0 {
		bounds := time.Unbounded()
		if m.deps.Bounds != nil {
			b, err := m.deps.Bounds.GetDayBounds(specDate)
			if err != nil {
				return RoomShift{}, err
			}
			bounds = b
		}
		if !bounds.Open || start < bounds.OpenMin || end > bounds.CloseMin {
			return RoomShift{}, ErrOutsideBounds
		}
	}

	// 8. Overlap check
	candidate := RoomShift{
		Id:            f.Id,
		TenantId:      tenantID,
		RoomId:        f.RoomId,
		CategoryId:    f.CategoryId,
		OccupantId:    f.OccupantId,
		OccupantLabel: occupantLabel,
		DayOfWeek:     int64(dow),
		SpecificDate:  specDate,
		StartMin:      int64(start),
		EndMin:        int64(end),
		IsActive:      true,
	}

	if err := m.checkOverlap(candidate); err != nil {
		return RoomShift{}, err
	}

	candidate.UpdatedAt = m.now() / 1_000_000_000
	if candidate.Id == "" {
		candidate.Id = m.deps.IDs.NewID()
		if err := m.db.Create(&candidate); err != nil {
			return RoomShift{}, err
		}
	} else {
		if err := m.db.Update(&candidate, orm.Eq("id", candidate.Id), orm.Eq("tenant_id", tenantID)); err != nil {
			return RoomShift{}, err
		}
	}

	m.publish(TopicShiftSaved, &candidate)
	return candidate, nil
}

func (m *Module) checkOverlap(cand RoomShift) error {
	todayDate, _ := m.localNow()

	var activeShifts []RoomShift
	err := m.db.Query(&RoomShift{}).
		Where("tenant_id").Eq(cand.TenantId).
		Where("is_active").Eq(true).
		ReadAll(func() model.Model { return &RoomShift{} }, func(row model.Model) {
			if item, ok := row.(*RoomShift); ok {
				if item.Id != cand.Id {
					activeShifts = append(activeShifts, *item)
				}
			}
		})
	if err != nil {
		return err
	}

	for _, e := range activeShifts {
		// Minutes must overlap: cand.StartMin < e.EndMin && e.StartMin < cand.EndMin
		if !(cand.StartMin < e.EndMin && e.StartMin < cand.EndMin) {
			continue
		}

		var dateOverlap bool
		if cand.SpecificDate == 0 { // Candidate is weekly (dow = cand.DayOfWeek)
			if e.SpecificDate == 0 {
				dateOverlap = (e.DayOfWeek == cand.DayOfWeek)
			} else {
				dateOverlap = (time.Weekday(e.SpecificDate) == int(cand.DayOfWeek) && e.SpecificDate >= todayDate)
			}
		} else { // Candidate is dated (date D = cand.SpecificDate)
			if e.SpecificDate == 0 {
				if e.DayOfWeek == int64(time.Weekday(cand.SpecificDate)) {
					// Check if E is cancelled on cand.SpecificDate
					cancelled, err := m.isShiftCancelled(e.TenantId, e.Id, cand.SpecificDate)
					if err != nil {
						return err
					}
					if !cancelled {
						dateOverlap = true
					}
				}
			} else {
				dateOverlap = (e.SpecificDate == cand.SpecificDate)
			}
		}

		if dateOverlap {
			if e.RoomId == cand.RoomId {
				return ErrRoomOverlap
			}
			if cand.OccupantId != "" && e.OccupantId == cand.OccupantId {
				return ErrOccupantOverlap
			}
		}
	}
	return nil
}

func (m *Module) isShiftCancelled(tenantID, shiftID string, date int64) (bool, error) {
	var found bool
	err := m.db.Query(&RoomShiftCancellation{}).
		Where("tenant_id").Eq(tenantID).
		Where("shift_id").Eq(shiftID).
		Where("specific_date").Eq(date).
		Limit(1).
		ReadAll(func() model.Model { return &RoomShiftCancellation{} }, func(row model.Model) {
			found = true
		})
	return found, err
}

func (m *Module) ListRoomShifts(tenantID, roomID string) ([]RoomShift, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	var shifts []RoomShift
	err := m.db.Query(&RoomShift{}).
		Where("tenant_id").Eq(tenantID).
		Where("room_id").Eq(roomID).
		Where("is_active").Eq(true).
		ReadAll(func() model.Model { return &RoomShift{} }, func(row model.Model) {
			if item, ok := row.(*RoomShift); ok {
				shifts = append(shifts, *item)
			}
		})
	if err != nil {
		return nil, err
	}

	var weekly []RoomShift
	var dated []RoomShift
	for _, s := range shifts {
		if s.SpecificDate == 0 {
			weekly = append(weekly, s)
		} else {
			dated = append(dated, s)
		}
	}

	// Sort weekly by day_of_week asc, start_min asc
	for i := 0; i < len(weekly); i++ {
		for j := i + 1; j < len(weekly); j++ {
			if weekly[i].DayOfWeek > weekly[j].DayOfWeek ||
				(weekly[i].DayOfWeek == weekly[j].DayOfWeek && weekly[i].StartMin > weekly[j].StartMin) {
				weekly[i], weekly[j] = weekly[j], weekly[i]
			}
		}
	}

	// Sort dated by specific_date asc, start_min asc
	for i := 0; i < len(dated); i++ {
		for j := i + 1; j < len(dated); j++ {
			if dated[i].SpecificDate > dated[j].SpecificDate ||
				(dated[i].SpecificDate == dated[j].SpecificDate && dated[i].StartMin > dated[j].StartMin) {
				dated[i], dated[j] = dated[j], dated[i]
			}
		}
	}

	return append(weekly, dated...), nil
}

func (m *Module) DeleteShift(tenantID, id string) error {
	if tenantID == "" {
		return ErrTenantRequired
	}
	if id == "" {
		return ErrNotFound
	}

	err := m.db.Tx(func(tx *orm.DB) error {
		c := RoomShiftCancellation{TenantId: tenantID, ShiftId: id}
		if err := tx.Delete(&c, orm.Eq("tenant_id", tenantID), orm.Eq("shift_id", id)); err != nil {
			return err
		}
		s := RoomShift{Id: id, TenantId: tenantID}
		return tx.Delete(&s, orm.Eq("id", id), orm.Eq("tenant_id", tenantID))
	})
	if err != nil {
		return err
	}

	m.publish(TopicShiftDeleted, &RoomShift{Id: id, TenantId: tenantID})
	return nil
}

func (m *Module) CancelShiftOccurrence(tenantID, shiftID string, date int64) error {
	if tenantID == "" {
		return ErrTenantRequired
	}
	if shiftID == "" {
		return ErrNotFound
	}

	var s RoomShift
	var found bool
	err := m.db.Query(&RoomShift{}).
		Where("tenant_id").Eq(tenantID).
		Where("id").Eq(shiftID).
		Limit(1).
		ReadAll(func() model.Model { return &RoomShift{} }, func(row model.Model) {
			if item, ok := row.(*RoomShift); ok {
				s = *item
				found = true
			}
		})
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}

	if s.SpecificDate > 0 {
		return ErrNotWeekly
	}
	if time.Weekday(date) != int(s.DayOfWeek) {
		return ErrInvalidDate
	}

	c := RoomShiftCancellation{
		TenantId:     tenantID,
		ShiftId:      shiftID,
		SpecificDate: date,
	}
	already, err := m.isShiftCancelled(tenantID, shiftID, date)
	if err != nil {
		return err
	}
	if !already {
		if err := m.db.Create(&c); err != nil {
			return err
		}
	}

	m.publish(TopicShiftOccurrenceCancelled, &c)
	return nil
}

func (m *Module) MoveShiftOccurrence(tenantID, shiftID string, date int64, targetRoomID string) (RoomShift, error) {
	if tenantID == "" {
		return RoomShift{}, ErrTenantRequired
	}
	if shiftID == "" || targetRoomID == "" {
		return RoomShift{}, ErrNotFound
	}

	var resShift RoomShift
	err := m.db.Tx(func(tx *orm.DB) error {
		var s RoomShift
		var found bool
		err := tx.Query(&RoomShift{}).
			Where("tenant_id").Eq(tenantID).
			Where("id").Eq(shiftID).
			Limit(1).
			ReadAll(func() model.Model { return &RoomShift{} }, func(row model.Model) {
				if item, ok := row.(*RoomShift); ok {
					s = *item
					found = true
				}
			})
		if err != nil {
			return err
		}
		if !found {
			return ErrNotFound
		}

		// Check target room
		targetRoom, err := m.GetRoom(tenantID, targetRoomID)
		if err != nil || !targetRoom.IsActive {
			return ErrNotFound
		}

		// Room category check
		roomCats, err := m.ListRoomCategories(tenantID, targetRoomID)
		if err != nil {
			return err
		}
		var roomCatAllowed bool
		for _, rc := range roomCats {
			if rc.Key == s.CategoryId {
				roomCatAllowed = true
				break
			}
		}
		if !roomCatAllowed {
			return ErrCategoryNotAllowed
		}

		// Bounds check
		bounds := time.Unbounded()
		if m.deps.Bounds != nil {
			b, err := m.deps.Bounds.GetDayBounds(date)
			if err != nil {
				return err
			}
			bounds = b
		}
		if !bounds.Open || s.StartMin < int64(bounds.OpenMin) || s.EndMin > int64(bounds.CloseMin) {
			return ErrOutsideBounds
		}

		if s.SpecificDate == 0 { // Weekly
			if time.Weekday(date) != int(s.DayOfWeek) {
				return ErrInvalidDate
			}
			// 1. Cancel occurrence of original shift
			c := RoomShiftCancellation{
				TenantId:     tenantID,
				ShiftId:      shiftID,
				SpecificDate: date,
			}
			already, err := m.isShiftCancelled(tenantID, shiftID, date)
			if err != nil {
				return err
			}
			if !already {
				if err := tx.Create(&c); err != nil {
					return err
				}
			}

			// 2. Create new dated shift in targetRoomID
			newShift := RoomShift{
				Id:            m.deps.IDs.NewID(),
				TenantId:      tenantID,
				RoomId:        targetRoomID,
				CategoryId:    s.CategoryId,
				OccupantId:    s.OccupantId,
				OccupantLabel: s.OccupantLabel,
				DayOfWeek:     s.DayOfWeek,
				SpecificDate:  date,
				StartMin:      s.StartMin,
				EndMin:        s.EndMin,
				IsActive:      true,
				UpdatedAt:     m.now() / 1_000_000_000,
			}
			if err := m.checkOverlap(newShift); err != nil {
				return err
			}
			if err := tx.Create(&newShift); err != nil {
				return err
			}
			resShift = newShift
		} else { // Dated
			updatedShift := s
			updatedShift.RoomId = targetRoomID
			updatedShift.UpdatedAt = m.now() / 1_000_000_000
			if err := m.checkOverlap(updatedShift); err != nil {
				return err
			}
			if err := tx.Update(&updatedShift, orm.Eq("id", s.Id), orm.Eq("tenant_id", tenantID)); err != nil {
				return err
			}
			resShift = updatedShift
		}

		return nil
	})

	if err != nil {
		return RoomShift{}, err
	}

	m.publish(TopicShiftOccurrenceMoved, &resShift)
	return resShift, nil
}
