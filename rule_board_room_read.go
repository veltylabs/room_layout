package roomlayout

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Read es la regla de negocio de leer board_room: el tablero de salas, opcionalmente filtrado por categoría y equipamiento.
func (p *BoardRoom) Read(a crud.CurrentUser, filter *BoardRoom, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	board, err := m.ListBoard(m.tenantID, filter.CategoryId, filter.EquipmentId)
	if err != nil {
		return nil, err
	}
	list := BoardRoomList{}
	for i := range board {
		list = append(list, &board[i])
	}
	return &list, nil
}
