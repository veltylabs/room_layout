package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Read es la regla de negocio de leer public_slot: los huecos libres que el público puede ver.
func (p *PublicSlot) Read(a crud.CurrentUser, filter *PublicSlot, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	if filter.Date == "" {
		return nil, &model.FieldError{Field: "date", Err: ErrInvalidDate}
	}
	date, err := parseDate(filter.Date)
	if err != nil {
		return nil, &model.FieldError{Field: "date", Err: ErrInvalidDate}
	}
	slots, err := m.ListPublicAvailability(date, filter.CategoryId)
	if err != nil {
		return nil, err
	}
	list := PublicSlotList{}
	for i := range slots {
		list = append(list, &slots[i])
	}
	return &list, nil
}
