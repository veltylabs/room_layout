package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Read es la regla de negocio de leer occupant_option: los ocupantes que entrega el lector de ocupantes de la app.
func (p *OccupantOption) Read(a crud.CurrentUser, filter *OccupantOption, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	opts, err := m.deps.Occupants.OccupantOptions(m.tenantID)
	if err != nil {
		return nil, err
	}
	list := OccupantOptionList{}
	for _, kv := range opts {
		list = append(list, &OccupantOption{Id: kv.Key, Label: kv.Value})
	}
	return &list, nil
}
