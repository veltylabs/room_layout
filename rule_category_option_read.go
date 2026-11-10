package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Read es la regla de negocio de leer category_option: las categorías que entrega el lector de categorías de la app.
func (p *CategoryOption) Read(a crud.CurrentUser, filter *CategoryOption, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	opts, err := m.deps.Categories.CategoryOptions(m.tenantID)
	if err != nil {
		return nil, err
	}
	list := CategoryOptionList{}
	for _, kv := range opts {
		list = append(list, &CategoryOption{Id: kv.Key, Label: kv.Value})
	}
	return &list, nil
}
