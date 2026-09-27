package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/fmt"
	"webtyp.com/model"
)

type dummyArrayWriter struct{}

func (d *dummyArrayWriter) String(val string)           {}
func (d *dummyArrayWriter) Int(val int64)               {}
func (d *dummyArrayWriter) Float(val float64)           {}
func (d *dummyArrayWriter) Bool(val bool)               {}
func (d *dummyArrayWriter) Bytes(val []byte)            {}
func (d *dummyArrayWriter) Object(val model.Encodable) {}
func (d *dummyArrayWriter) Close()                      {}

type dummyWriter struct{}

func (d *dummyWriter) String(name, val string)                 {}
func (d *dummyWriter) Int(name string, val int64)               {}
func (d *dummyWriter) Float(name string, val float64)           {}
func (d *dummyWriter) Bool(name string, val bool)              {}
func (d *dummyWriter) Bytes(name string, val []byte)           {}
func (d *dummyWriter) Null(name string)                        {}
func (d *dummyWriter) Raw(name, val string)                   {}
func (d *dummyWriter) Object(name string, val model.Encodable) {}
func (d *dummyWriter) Array(name string, n int) model.ArrayWriter {
	return &dummyArrayWriter{}
}

type populatedReader struct{}

func (p *populatedReader) String(name string) (string, bool) {
	if name == "day_of_week" {
		return "1", true
	}
	return "val", true
}
func (p *populatedReader) Int(name string) (int64, bool)                 { return 1, true }
func (p *populatedReader) Float(name string) (float64, bool)             { return 1.0, true }
func (p *populatedReader) Bool(name string) (bool, bool)                 { return true, true }
func (p *populatedReader) Bytes(name string) ([]byte, bool)              { return []byte("val"), true }
func (p *populatedReader) Object(name string, into model.Decodable) bool { return false }
func (p *populatedReader) Array(name string) (model.ArrayReader, bool)   { return nil, false }
func (p *populatedReader) Raw(name string) (string, bool)                { return "", false }

type validator interface {
	Validate(action byte) error
}

func TestORMGeneratedMethodsCoverage(t *testing.T) {
	w := &dummyWriter{}
	r := &populatedReader{}

	// Test ValidationError methods
	veNil := &roomlayout.ValidationError{}
	_ = veNil.Error()
	_ = veNil.Unwrap()

	veErr := &roomlayout.ValidationError{Err: fmt.Err("custom error")}
	_ = veErr.Error()
	_ = veErr.Unwrap()

	models := []model.Model{
		&roomlayout.Floor{Id: "1", TenantId: "t", Name: "N", Position: 1, UpdatedAt: 100},
		&roomlayout.Room{Id: "1", TenantId: "t", FloorId: "f", Code: "C", Name: "N", Notes: "notes", IsActive: true, UpdatedAt: 100},
		&roomlayout.Equipment{Id: "1", TenantId: "t", Name: "E", UpdatedAt: 100},
		&roomlayout.RoomEquipment{TenantId: "t", RoomId: "r", EquipmentId: "e"},
		&roomlayout.RoomCategory{TenantId: "t", RoomId: "r", CategoryId: "c"},
		&roomlayout.RoomShift{Id: "1", TenantId: "t", RoomId: "r", CategoryId: "c", OccupantId: "o", OccupantLabel: "L", DayOfWeek: 1, SpecificDate: 100, StartMin: 600, EndMin: 720, IsActive: true, UpdatedAt: 100},
		&roomlayout.RoomShiftCancellation{TenantId: "t", ShiftId: "s", SpecificDate: 100},
		&roomlayout.RoomShiftForm{Id: "1", RoomId: "r", CategoryId: "c", OccupantId: "o", OccupantLabel: "L", DayOfWeek: "1", Date: "2026-03-16", Start: "09:00", End: "13:00"},
		&roomlayout.IdRef{Id: "1"},
		&roomlayout.ListFloorsArgs{TenantId: "t"},
		&roomlayout.DeleteFloorArgs{TenantId: "t", Id: "1"},
		&roomlayout.ListRoomsArgs{TenantId: "t", ActiveOnly: true},
		&roomlayout.GetRoomArgs{TenantId: "t", Id: "1"},
		&roomlayout.DeactivateRoomArgs{TenantId: "t", Id: "1"},
		&roomlayout.ListEquipmentArgs{TenantId: "t"},
		&roomlayout.DeleteEquipmentArgs{TenantId: "t", Id: "1"},
		&roomlayout.ListRoomEquipmentArgs{TenantId: "t", RoomId: "r"},
		&roomlayout.SetRoomEquipmentArgs{TenantId: "t", RoomId: "r", EquipmentIds: []roomlayout.IdRef{{Id: "1"}}},
		&roomlayout.ListRoomCategoriesArgs{TenantId: "t", RoomId: "r"},
		&roomlayout.SetRoomCategoriesArgs{TenantId: "t", RoomId: "r", CategoryIds: []roomlayout.IdRef{{Id: "1"}}},
		&roomlayout.ListOptionsArgs{TenantId: "t"},
		&roomlayout.ListRoomShiftsArgs{TenantId: "t", RoomId: "r"},
		&roomlayout.DeleteRoomShiftArgs{TenantId: "t", Id: "1"},
		&roomlayout.CancelShiftOccurrenceArgs{TenantId: "t", ShiftId: "s", Date: "2026-03-16"},
		&roomlayout.MoveShiftOccurrenceArgs{TenantId: "t", ShiftId: "s", Date: "2026-03-16", TargetRoomId: "r2"},
		&roomlayout.ListRoomDayArgs{TenantId: "t", RoomId: "r", Date: "2026-03-16"},
		&roomlayout.ListBoardArgs{TenantId: "t", CategoryId: "c", EquipmentId: "e"},
		&roomlayout.FindFreeRoomsArgs{TenantId: "t", Date: "2026-03-16", Start: "09:00", End: "13:00", CategoryId: "c"},
		&roomlayout.ListPublicAvailabilityArgs{Date: "2026-03-16", CategoryId: "c"},
		&roomlayout.Option{Id: "1", Label: "L"},
		&roomlayout.DayShift{ShiftId: "s", RoomId: "r", RoomCode: "C", CategoryId: "c", CategoryLabel: "CL", OccupantId: "o", OccupantLabel: "OL", Start: "09:00", End: "13:00", IsWeekly: true},
		&roomlayout.BoardRoom{RoomId: "r", Code: "C", Name: "N", FloorName: "FN", FloorPosition: 1, CategoryLabels: "CL", EquipmentLabels: "EL", Status: "free", CurrentOccupantLabel: "COL", CurrentUntil: "12:00", NextStart: "14:00", NextOccupantLabel: "NOL", Today: "2026-03-16"},
		&roomlayout.FreeRoom{RoomId: "r", Code: "C", Name: "N", FloorName: "FN", EquipmentLabels: "EL"},
		&roomlayout.PublicSlot{RoomCode: "C", RoomName: "RN", FloorName: "FN", Start: "09:00", End: "13:00"},
	}

	for _, m := range models {
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(w)
		m.DecodeFields(r)
		if v, ok := m.(validator); ok {
			_ = v.Validate(byte(model.Create))
			_ = v.Validate(byte(model.Update))
		}
	}

	// Test nil check
	var nilFloor *roomlayout.Floor
	if !nilFloor.IsNil() {
		t.Fatalf("expected IsNil true for nil floor")
	}

	// Test slices
	slices := []model.FielderSlice{
		&roomlayout.FloorList{&roomlayout.Floor{Id: "1"}},
		&roomlayout.RoomList{&roomlayout.Room{Id: "1"}},
		&roomlayout.EquipmentList{&roomlayout.Equipment{Id: "1"}},
		&roomlayout.RoomEquipmentList{&roomlayout.RoomEquipment{RoomId: "1"}},
		&roomlayout.RoomCategoryList{&roomlayout.RoomCategory{RoomId: "1"}},
		&roomlayout.RoomShiftList{&roomlayout.RoomShift{Id: "1"}},
		&roomlayout.RoomShiftCancellationList{&roomlayout.RoomShiftCancellation{ShiftId: "1"}},
		&roomlayout.RoomShiftFormList{&roomlayout.RoomShiftForm{Id: "1"}},
		&roomlayout.IdRefList{&roomlayout.IdRef{Id: "1"}},
		&roomlayout.ListFloorsArgsList{&roomlayout.ListFloorsArgs{TenantId: "t"}},
		&roomlayout.DeleteFloorArgsList{&roomlayout.DeleteFloorArgs{Id: "1"}},
		&roomlayout.ListRoomsArgsList{&roomlayout.ListRoomsArgs{TenantId: "t"}},
		&roomlayout.GetRoomArgsList{&roomlayout.GetRoomArgs{Id: "1"}},
		&roomlayout.DeactivateRoomArgsList{&roomlayout.DeactivateRoomArgs{Id: "1"}},
		&roomlayout.ListEquipmentArgsList{&roomlayout.ListEquipmentArgs{TenantId: "t"}},
		&roomlayout.DeleteEquipmentArgsList{&roomlayout.DeleteEquipmentArgs{Id: "1"}},
		&roomlayout.ListRoomEquipmentArgsList{&roomlayout.ListRoomEquipmentArgs{RoomId: "r"}},
		&roomlayout.SetRoomEquipmentArgsList{&roomlayout.SetRoomEquipmentArgs{RoomId: "r"}},
		&roomlayout.ListRoomCategoriesArgsList{&roomlayout.ListRoomCategoriesArgs{RoomId: "r"}},
		&roomlayout.SetRoomCategoriesArgsList{&roomlayout.SetRoomCategoriesArgs{RoomId: "r"}},
		&roomlayout.ListOptionsArgsList{&roomlayout.ListOptionsArgs{TenantId: "t"}},
		&roomlayout.ListRoomShiftsArgsList{&roomlayout.ListRoomShiftsArgs{RoomId: "r"}},
		&roomlayout.DeleteRoomShiftArgsList{&roomlayout.DeleteRoomShiftArgs{Id: "1"}},
		&roomlayout.CancelShiftOccurrenceArgsList{&roomlayout.CancelShiftOccurrenceArgs{ShiftId: "s"}},
		&roomlayout.MoveShiftOccurrenceArgsList{&roomlayout.MoveShiftOccurrenceArgs{ShiftId: "s"}},
		&roomlayout.ListRoomDayArgsList{&roomlayout.ListRoomDayArgs{RoomId: "r"}},
		&roomlayout.ListBoardArgsList{&roomlayout.ListBoardArgs{TenantId: "t"}},
		&roomlayout.FindFreeRoomsArgsList{&roomlayout.FindFreeRoomsArgs{Date: "2026-03-16"}},
		&roomlayout.ListPublicAvailabilityArgsList{&roomlayout.ListPublicAvailabilityArgs{Date: "2026-03-16"}},
		&roomlayout.OptionList{&roomlayout.Option{Id: "1"}},
		&roomlayout.DayShiftList{&roomlayout.DayShift{ShiftId: "s"}},
		&roomlayout.BoardRoomList{&roomlayout.BoardRoom{RoomId: "r"}},
		&roomlayout.FreeRoomList{&roomlayout.FreeRoom{RoomId: "r"}},
		&roomlayout.PublicSlotList{&roomlayout.PublicSlot{RoomCode: "C"}},
	}

	for _, sl := range slices {
		if sl.Len() != 1 {
			t.Fatalf("expected Len 1")
		}
		_ = sl.At(0)
		_ = sl.Append()
		if enc, ok := sl.(model.Encodable); ok {
			enc.EncodeFields(w)
		}
		if dec, ok := sl.(model.Decodable); ok {
			dec.DecodeFields(r)
		}
		if v, ok := sl.(validator); ok {
			_ = v.Validate(byte(model.Create))
		}
	}
}
