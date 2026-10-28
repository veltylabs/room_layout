package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/lang"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	descSpaces    lang.Text = "Physical spaces, room arrangements and clinical equipment."
	descRooms     lang.Text = "Rooms, consultations and procedure spaces."
	descFloors    lang.Text = "Building levels and architectural floors."
	descEquipment lang.Text = "Clinical and operational equipment allocated to spaces."

	labelRooms     lang.Text = "Rooms"
	labelFloors    lang.Text = "Floors"
	labelEquipment lang.Text = "Equipment"
)

// Screens declara las pantallas que este módulo ofrece al menú de la app.
func Screens(caller router.Caller) []view.Screen {
	return []view.Screen{
		{
			ID:    ID,
			Label: lang.Text(DefaultLabel),
			Icon:  ID,
			Help:  descSpaces,
			Children: []view.Screen{
				{
					ID:         ID + ".rooms",
					Label:      labelRooms,
					Icon:       ID + "_room",
					Permission: roomlayout.ResourceRoom,
					Presenter:  roomlayout.NewRoomView(caller, view.WithDescription(descRooms)),
				},
				{
					ID:         ID + ".floors",
					Label:      labelFloors,
					Icon:       ID + "_floor",
					Permission: roomlayout.ResourceFloor,
					Presenter:  roomlayout.NewFloorView(caller, view.WithDescription(descFloors)),
				},
				{
					ID:         ID + ".equipment",
					Label:      labelEquipment,
					Icon:       ID + "_equipment",
					Permission: roomlayout.ResourceEquipment,
					Presenter:  roomlayout.NewEquipmentView(caller, view.WithDescription(descEquipment)),
				},
			},
		},
	}
}
