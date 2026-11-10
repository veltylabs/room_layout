package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Read es la regla de negocio de leer free_room: las salas libres en un día y un tramo de horas (date, start y end son requeridos).
func (p *FreeRoom) Read(a crud.CurrentUser, filter *FreeRoom, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	if filter.Date == "" {
		return nil, &model.FieldError{Field: "date", Err: ErrInvalidDate}
	}
	date, errDate := parseDate(filter.Date)
	startMin, errStart := parseHour(filter.Start)
	endMin, errEnd := parseHour(filter.End)
	if errDate != nil || errStart != nil || errEnd != nil {
		return nil, ErrInvalidRange
	}
	free, err := m.FindFreeRooms(m.tenantID, date, startMin, endMin, filter.CategoryId)
	if err != nil {
		return nil, err
	}
	list := FreeRoomList{}
	for i := range free {
		list = append(list, &free[i])
	}
	return &list, nil
}
