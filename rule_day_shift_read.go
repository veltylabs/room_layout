package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Read es la regla de negocio de leer day_shift: los turnos de una sala (o de todas) en un día; sin date, hoy.
func (p *DayShift) Read(a crud.CurrentUser, filter *DayShift, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	date, err := parseDate(filter.Date)
	if err != nil {
		return nil, &model.FieldError{Field: "date", Err: err}
	}
	shifts, err := m.ListRoomDay(m.tenantID, filter.RoomId, date)
	if err != nil {
		return nil, err
	}
	list := DayShiftList{}
	for i := range shifts {
		list = append(list, &shifts[i])
	}
	return &list, nil
}
