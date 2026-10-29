package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	roomlayout "github.com/veltylabs/room_layout"
)

func TestEquipment_DeviceIdOptional(t *testing.T) {
	env := newTestEnv("tenantA")

	// save an equipment with no DeviceId
	eq := roomlayout.Equipment{
		TenantId: "tenantA",
		Name:     "Gastroscope",
	}

	savedEq, err := env.mod.SaveEquipment(eq)
	if err != nil {
		t.Fatalf("unexpected error saving equipment: %v", err)
	}

	// read back equipment directly or through list
	res, err := env.mod.ListEquipment("tenantA")
	if err != nil {
		t.Fatalf("unexpected error listing equipment: %v", err)
	}

	var found bool
	for _, item := range res {
		if item.Id == savedEq.Id {
			found = true
			if item.DeviceId != "" {
				t.Errorf("expected DeviceId to be empty, got %q", item.DeviceId)
			}
			break
		}
	}
	if !found {
		t.Errorf("saved equipment not found in list")
	}
}

func TestEquipment_DeviceIdKept(t *testing.T) {
	env := newTestEnv("tenantA")

	// save an equipment with DeviceId
	eq := roomlayout.Equipment{
		TenantId: "tenantA",
		Name:     "Printer",
		DeviceId: "dev-1",
	}

	savedEq, err := env.mod.SaveEquipment(eq)
	if err != nil {
		t.Fatalf("unexpected error saving equipment: %v", err)
	}

	// read back
	res, err := env.mod.ListEquipment("tenantA")
	if err != nil {
		t.Fatalf("unexpected error listing equipment: %v", err)
	}

	var found bool
	for _, item := range res {
		if item.Id == savedEq.Id {
			found = true
			if item.DeviceId != "dev-1" {
				t.Errorf("expected DeviceId to be 'dev-1', got %q", item.DeviceId)
			}
			break
		}
	}
	if !found {
		t.Errorf("saved equipment not found in list")
	}
}

func TestEquipment_DeviceOptionsFromDeviceManager(t *testing.T) {
	field, ok := roomlayout.EquipmentModel.Field("device_id")
	if !ok {
		t.Fatalf("expected field device_id in EquipmentModel")
	}
	if field.Options != &devicemanager.DeviceModel {
		t.Errorf("expected device_id field Options to be &devicemanager.DeviceModel, got %v", field.Options)
	}
}
