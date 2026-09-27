package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
)

func (m *Module) SaveFloor(f Floor) (Floor, error) {
	if f.TenantId == "" {
		return Floor{}, ErrTenantRequired
	}
	f.Name = fmt.TrimSpace(f.Name)
	isNew := f.Id == ""
	if isNew {
		f.Id = m.deps.IDs.NewID()
	}

	if err := model.ValidateFields(byte(model.Create), &f); err != nil {
		return Floor{}, &ValidationError{Err: err}
	}
	f.UpdatedAt = m.now() / 1_000_000_000

	if isNew {
		if err := m.db.Create(&f); err != nil {
			return Floor{}, err
		}
	} else {
		if err := m.db.Update(&f, orm.Eq("id", f.Id), orm.Eq("tenant_id", f.TenantId)); err != nil {
			return Floor{}, err
		}
	}

	m.publish(TopicFloorSaved, &f)
	return f, nil
}

func (m *Module) ListFloors(tenantID string) ([]Floor, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	var res []Floor
	err := m.db.Query(&Floor{}).
		Where("tenant_id").Eq(tenantID).
		OrderBy("position").Asc().
		OrderBy("name").Asc().
		ReadAll(func() model.Model { return &Floor{} }, func(row model.Model) {
			if item, ok := row.(*Floor); ok {
				res = append(res, *item)
			}
		})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Module) DeleteFloor(tenantID, id string) error {
	if tenantID == "" {
		return ErrTenantRequired
	}
	if id == "" {
		return ErrNotFound
	}
	var hasRoom bool
	err := m.db.Query(&Room{}).
		Where("tenant_id").Eq(tenantID).
		Where("floor_id").Eq(id).
		Limit(1).
		ReadAll(func() model.Model { return &Room{} }, func(row model.Model) {
			hasRoom = true
		})
	if err != nil {
		return err
	}
	if hasRoom {
		return ErrFloorInUse
	}

	f := Floor{Id: id, TenantId: tenantID}
	if err := m.db.Delete(&f, orm.Eq("id", id), orm.Eq("tenant_id", tenantID)); err != nil {
		return err
	}

	m.publish(TopicFloorDeleted, &f)
	return nil
}
