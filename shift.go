package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/storage"
	"webtyp.com/time"
)

// SaveShift valida y guarda un turno (semanal o puntual) a partir de los datos de su formulario.
//
// Ejemplo: shift, err := m.SaveShift(tenantID, form)
func (m *Module) SaveShift(tenantID string, f RoomShiftForm) (RoomShift, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return RoomShift{}, err
	}
	f.OccupantLabel = fmt.TrimSpace(f.OccupantLabel)
	if err := model.ValidateFields(byte(model.Create), &f); err != nil {
		return RoomShift{}, err
	}
	cand, err := formToShift(f)
	if err != nil {
		return RoomShift{}, err
	}
	err = store.Atomic(func(tx crud.Tx) error {
		if err := m.checkShift(tx, tenantID, &cand); err != nil {
			return err
		}
		cand.UpdatedAt = m.now()
		if cand.Id == "" {
			cand.Id = m.deps.IDs.NewID()
			return tx.Create(&cand)
		}
		return tx.UpdateFields(&cand, cand.Id, shiftColumns)
	})
	if err != nil {
		return RoomShift{}, err
	}
	m.publish(TopicShiftSaved, &cand)
	return cand, nil
}

// ListRoomShifts responde la lista de turnos activos asignados a una sala.
//
// Ejemplo: list, err := m.ListRoomShifts(tenantID, roomID)
func (m *Module) ListRoomShifts(tenantID, roomID string) ([]RoomShift, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return nil, err
	}
	rows, err := allRows(store, func() model.Model { return &RoomShift{} },
		storage.Eq(RoomShift_.RoomId, roomID), storage.Eq(RoomShift_.IsActive, true))
	if err != nil {
		return nil, err
	}
	var weekly, dated []RoomShift
	for _, r := range rows {
		s := *r.(*RoomShift)
		if s.SpecificDate == 0 {
			weekly = append(weekly, s)
		} else {
			dated = append(dated, s)
		}
	}
	// weekly by day_of_week then start_min; dated by specific_date then start_min
	for i := 0; i < len(weekly); i++ {
		for j := i + 1; j < len(weekly); j++ {
			if weekly[i].DayOfWeek > weekly[j].DayOfWeek ||
				(weekly[i].DayOfWeek == weekly[j].DayOfWeek && weekly[i].StartMin > weekly[j].StartMin) {
				weekly[i], weekly[j] = weekly[j], weekly[i]
			}
		}
	}
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

// DeleteShift elimina un turno y sus cancelaciones asociadas.
//
// Ejemplo: err := m.DeleteShift(tenantID, id)
func (m *Module) DeleteShift(tenantID, id string) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if id == "" {
		return ErrNotFound
	}
	err = store.Atomic(func(tx crud.Tx) error {
		if err := deleteShiftCancellations(tx, id); err != nil {
			return err
		}
		return tx.Delete(&RoomShift{}, id)
	})
	if err != nil {
		return err
	}
	m.publish(TopicShiftDeleted, &RoomShift{Id: id, TenantId: tenantID})
	return nil
}

// CancelShiftOccurrence cancela la ocurrencia de un turno semanal en una fecha específica.
//
// Ejemplo: err := m.CancelShiftOccurrence(tenantID, shiftID, date)
func (m *Module) CancelShiftOccurrence(tenantID, shiftID string, date time.Unix) error {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return err
	}
	if shiftID == "" {
		return ErrNotFound
	}
	c := RoomShiftCancellation{TenantId: tenantID, ShiftId: shiftID, SpecificDate: date}
	err = store.Atomic(func(tx crud.Tx) error {
		if _, err := checkOccurrence(tx, shiftID, date); err != nil {
			return err
		}
		already, err := isShiftCancelled(tx, shiftID, date)
		if err != nil || already {
			return err
		}
		return tx.Create(&c)
	})
	if err != nil {
		return err
	}
	m.publish(TopicShiftOccurrenceCancelled, &c)
	return nil
}

// MoveShiftOccurrence mueve una ocurrencia puntual de un turno a otra sala.
//
// Ejemplo: moved, err := m.MoveShiftOccurrence(tenantID, shiftID, date, targetRoomID)
func (m *Module) MoveShiftOccurrence(tenantID, shiftID string, date time.Unix, targetRoomID string) (RoomShift, error) {
	store, err := m.storeFor(tenantID)
	if err != nil {
		return RoomShift{}, err
	}
	if shiftID == "" || targetRoomID == "" {
		return RoomShift{}, ErrNotFound
	}
	var res RoomShift
	err = store.Atomic(func(tx crud.Tx) error {
		var s RoomShift
		if err := tx.ReadOne(&s, shiftID); err != nil {
			return notFound(err, ErrNotFound)
		}
		if s.SpecificDate > 0 { // con fecha: se cambia de sala el turno mismo
			s.RoomId = targetRoomID
			if err := m.checkShift(tx, tenantID, &s); err != nil {
				return err
			}
			s.UpdatedAt = m.now()
			res = s
			return tx.UpdateFields(&s, s.Id, shiftColumns)
		}
		if _, err := checkOccurrence(tx, shiftID, date); err != nil {
			return err
		}
		moved, err := m.createMovedShift(tx, s, date, targetRoomID)
		if err != nil {
			return err
		}
		already, err := isShiftCancelled(tx, shiftID, date)
		if err != nil {
			return err
		}
		if !already {
			if err := tx.Create(&RoomShiftCancellation{ShiftId: shiftID, SpecificDate: date, MovedToRoomId: targetRoomID}); err != nil {
				return err
			}
		}
		res = moved
		return nil
	})
	if err != nil {
		return RoomShift{}, err
	}
	m.publish(TopicShiftOccurrenceMoved, &res)
	return res, nil
}
