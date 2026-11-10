package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/storage"
	"webtyp.com/time"
)

// shiftColumns son las columnas que se escriben al actualizar un turno (todas menos id y tenant_id).
var shiftColumns = []string{RoomShift_.RoomId, RoomShift_.CategoryId, RoomShift_.OccupantId, RoomShift_.OccupantLabel,
	RoomShift_.DayOfWeek, RoomShift_.SpecificDate, RoomShift_.StartMin, RoomShift_.EndMin, RoomShift_.IsActive, RoomShift_.UpdatedAt}

// editableShiftFields son los campos que quien llama puede enviar al actualizar un turno.
var editableShiftFields = []string{RoomShift_.RoomId, RoomShift_.CategoryId, RoomShift_.OccupantId, RoomShift_.OccupantLabel,
	RoomShift_.DayOfWeek, RoomShift_.SpecificDate, RoomShift_.StartMin, RoomShift_.EndMin}

// copyShiftFields copia de src a dst solo los campos de fields; uno que no se puede cambiar es un error de campo.
// Sirve en los dos sentidos: mezclar lo enviado en lo guardado y devolver lo normalizado a lo enviado.
// Ejemplo: err := copyShiftFields(&merged, next, []string{"room_id"})
func copyShiftFields(dst, src *RoomShift, fields []string) error {
	for _, f := range fields {
		switch f {
		case RoomShift_.RoomId:
			dst.RoomId = src.RoomId
		case RoomShift_.CategoryId:
			dst.CategoryId = src.CategoryId
		case RoomShift_.OccupantId:
			dst.OccupantId = src.OccupantId
		case RoomShift_.OccupantLabel:
			dst.OccupantLabel = src.OccupantLabel
		case RoomShift_.DayOfWeek:
			dst.DayOfWeek = src.DayOfWeek
		case RoomShift_.SpecificDate:
			dst.SpecificDate = src.SpecificDate
		case RoomShift_.StartMin:
			dst.StartMin = src.StartMin
		case RoomShift_.EndMin:
			dst.EndMin = src.EndMin
		default:
			return &model.FieldError{Field: f, Err: ErrFieldNotEditable}
		}
	}
	return nil
}

// checkShift aplica las reglas de un turno (las de SaveShift, en su orden) y lo completa con lo que decide el servidor:
// la sala existe y está activa; 0 ≤ inicio < fin ≤ 1440; semanal (día 0..6) o con fecha (el día de la semana se
// calcula); la categoría es de la app y la sala la tiene; el ocupante existe (y su nombre queda en el turno) o hay un
// nombre; un turno con fecha cabe en el horario de la clínica; no se solapa con otro turno de la sala ni del ocupante.
// Lo comparten la regla Create/Update de room_shift y SaveShift. s queda con IsActive y los campos calculados.
//
// Ejemplo: err := m.checkShift(tx, "tenant-1", &shift)
func (m *Module) checkShift(tx crud.Tx, tenantID string, s *RoomShift) error {
	var room Room
	if err := tx.ReadOne(&room, s.RoomId); err != nil || !room.IsActive {
		return ErrNotFound
	}
	if s.StartMin < 0 || s.StartMin >= s.EndMin || s.EndMin > 1440 {
		return ErrInvalidRange
	}
	if s.SpecificDate == 0 {
		if s.DayOfWeek < 0 || s.DayOfWeek > 6 {
			return ErrInvalidWeekday
		}
	} else {
		s.DayOfWeek = int64(time.WeekdayUTC(s.SpecificDate))
	}

	catOpts, err := m.deps.Categories.CategoryOptions(tenantID)
	if err != nil {
		return err
	}
	if !categoryIn(catOpts, s.CategoryId) {
		return ErrUnknownCategory
	}
	if err := checkRoomHasCategory(tx, s.RoomId, s.CategoryId); err != nil {
		return err
	}

	if s.OccupantId != "" {
		occOpts, err := m.deps.Occupants.OccupantOptions(tenantID)
		if err != nil {
			return err
		}
		found := false
		for _, opt := range occOpts {
			if opt.Key == s.OccupantId {
				found = true
				s.OccupantLabel = opt.Value
				break
			}
		}
		if !found {
			return ErrUnknownOccupant
		}
	} else if s.OccupantLabel == "" {
		return ErrOccupantRequired
	}

	if s.SpecificDate > 0 {
		if err := m.checkOpen(s.SpecificDate, s.StartMin, s.EndMin); err != nil {
			return err
		}
	}
	s.IsActive = true
	return m.checkOverlap(tx, *s)
}

// checkRoomHasCategory rechaza un turno cuya categoría la sala no tiene habilitada.
// Ejemplo: err := checkRoomHasCategory(tx, "r1", "cat1")   // ErrCategoryNotAllowed
func checkRoomHasCategory(tx crud.Tx, roomId, categoryId string) error {
	_, found, err := firstRow(tx, func() model.Model { return &RoomCategory{} },
		storage.Eq(RoomCategory_.RoomId, roomId), storage.Eq(RoomCategory_.CategoryId, categoryId))
	if err != nil {
		return err
	}
	if !found {
		return ErrCategoryNotAllowed
	}
	return nil
}

// checkOpen rechaza un tramo [start, end) fuera del horario de la clínica en esa fecha; sin lector de horario no hay límite.
// Ejemplo: err := m.checkOpen(date, 540, 600)   // ErrOutsideBounds
func (m *Module) checkOpen(date time.Unix, start, end int64) error {
	bounds := time.Unlimited()
	if m.deps.Hours != nil {
		b, err := m.deps.Hours.DayLimits(date)
		if err != nil {
			return err
		}
		bounds = b
	}
	if !bounds.Open || start < int64(bounds.OpenMin) || end > int64(bounds.CloseMin) {
		return ErrOutsideBounds
	}
	return nil
}

// checkOverlap rechaza un turno que se pisa con otro activo de la misma sala (ErrRoomOverlap) o del mismo ocupante
// (ErrOccupantOverlap). Un turno semanal choca con los de su día y con los de fecha futura de ese día de la semana; uno
// con fecha choca con los semanales de ese día (salvo la ocurrencia cancelada) y con los de su fecha.
func (m *Module) checkOverlap(tx crud.Tx, cand RoomShift) error {
	todayDate, _ := m.localNow()
	rows, err := allRows(tx, func() model.Model { return &RoomShift{} }, storage.Eq(RoomShift_.IsActive, true))
	if err != nil {
		return err
	}
	for _, r := range rows {
		e := *r.(*RoomShift)
		if e.Id == cand.Id {
			continue
		}
		if cand.StartMin >= e.EndMin || e.StartMin >= cand.EndMin {
			continue
		}
		var dateOverlap bool
		if cand.SpecificDate == 0 { // el candidato es semanal
			if e.SpecificDate == 0 {
				dateOverlap = e.DayOfWeek == cand.DayOfWeek
			} else {
				dateOverlap = time.WeekdayUTC(e.SpecificDate) == int(cand.DayOfWeek) && e.SpecificDate >= todayDate
			}
		} else { // el candidato tiene fecha
			if e.SpecificDate == 0 {
				if e.DayOfWeek == int64(time.WeekdayUTC(cand.SpecificDate)) {
					cancelled, err := isShiftCancelled(tx, e.Id, cand.SpecificDate)
					if err != nil {
						return err
					}
					dateOverlap = !cancelled
				}
			} else {
				dateOverlap = e.SpecificDate == cand.SpecificDate
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

// isShiftCancelled dice si la ocurrencia de un turno semanal en esa fecha está cancelada.
// Ejemplo: cancelled, err := isShiftCancelled(tx, "s1", date)
func isShiftCancelled(tx crud.Tx, shiftId string, date time.Unix) (bool, error) {
	_, found, err := firstRow(tx, func() model.Model { return &RoomShiftCancellation{} },
		storage.Eq(RoomShiftCancellation_.ShiftId, shiftId), storage.Eq(RoomShiftCancellation_.SpecificDate, int64(date)))
	return found, err
}

// checkOccurrence valida que se pueda cancelar o mover una ocurrencia: solo un turno semanal (ErrNotWeekly) y en un
// día que sea su día de la semana (ErrInvalidDate). shiftId inexistente es ErrNotFound; responde el turno.
// Ejemplo: s, err := checkOccurrence(tx, "s1", date)
func checkOccurrence(tx crud.Tx, shiftId string, date time.Unix) (RoomShift, error) {
	var s RoomShift
	if err := tx.ReadOne(&s, shiftId); err != nil {
		return RoomShift{}, notFound(err, ErrNotFound)
	}
	if s.SpecificDate > 0 {
		return RoomShift{}, ErrNotWeekly
	}
	if time.WeekdayUTC(date) != int(s.DayOfWeek) {
		return RoomShift{}, ErrInvalidDate
	}
	return s, nil
}

// createMovedShift crea, en la sala destino, el turno con fecha que reemplaza a la ocurrencia movida de un turno
// semanal: la sala destino existe y está activa, tiene la categoría del turno, la fecha cabe en el horario y no hay
// solape. Lo comparten MoveShiftOccurrence y la regla Create de room_shift_cancellation con moved_to_room_id.
// Ejemplo: moved, err := m.createMovedShift(tx, s, date, "r2")
func (m *Module) createMovedShift(tx crud.Tx, s RoomShift, date time.Unix, targetRoomID string) (RoomShift, error) {
	var target Room
	if err := tx.ReadOne(&target, targetRoomID); err != nil || !target.IsActive {
		return RoomShift{}, ErrNotFound
	}
	if err := checkRoomHasCategory(tx, targetRoomID, s.CategoryId); err != nil {
		return RoomShift{}, err
	}
	if err := m.checkOpen(date, s.StartMin, s.EndMin); err != nil {
		return RoomShift{}, err
	}
	moved := RoomShift{
		Id: m.deps.IDs.NewID(), RoomId: targetRoomID, CategoryId: s.CategoryId, OccupantId: s.OccupantId,
		OccupantLabel: s.OccupantLabel, DayOfWeek: s.DayOfWeek, SpecificDate: date, StartMin: s.StartMin, EndMin: s.EndMin,
		IsActive: true, UpdatedAt: m.now(),
	}
	if err := m.checkOverlap(tx, moved); err != nil {
		return RoomShift{}, err
	}
	if err := tx.Create(&moved); err != nil {
		return RoomShift{}, err
	}
	return moved, nil
}

// deleteShiftCancellations borra todas las cancelaciones de un turno (se llama al borrar el turno).
// Ejemplo: err := deleteShiftCancellations(tx, "s1")
func deleteShiftCancellations(tx crud.Tx, shiftId string) error {
	for {
		rows, err := allRows(tx, func() model.Model { return &RoomShiftCancellation{} }, storage.Eq(RoomShiftCancellation_.ShiftId, shiftId))
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		for _, r := range rows {
			c := r.(*RoomShiftCancellation)
			if err := tx.Delete(&RoomShiftCancellation{}, model.JoinKey(c.ShiftId, fmt.Sprint(int64(c.SpecificDate)))); err != nil {
				return err
			}
		}
	}
}
